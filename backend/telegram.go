package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Telegram Mini App foydalanuvchisi (initData ichidagi "user" maydoni)
type TgUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

func (u TgUser) FullName() string {
	n := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if n == "" {
		n = u.Username
	}
	return n
}

// verifyTelegramInitData — Telegram WebApp initData'ni bot token bilan tekshiradi.
// Algoritm: secret = HMAC_SHA256("WebAppData", bot_token),
//
//	hash   = HMAC_SHA256(secret, data_check_string).
//
// Hujjat: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func verifyTelegramInitData(initData, botToken string) (*TgUser, bool) {
	if initData == "" || botToken == "" {
		return nil, false
	}
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, false
	}
	providedHash := values.Get("hash")
	if providedHash == "" {
		return nil, false
	}

	// data_check_string: "hash"dan tashqari barcha juftliklar, kalit bo'yicha saralanган, '\n' bilan
	keys := make([]string, 0, len(values))
	for k := range values {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(values.Get(k))
	}

	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	computed := hex.EncodeToString(hmacSHA256(secret, []byte(sb.String())))
	if !hmac.Equal([]byte(computed), []byte(providedHash)) {
		return nil, false
	}

	// auth_date — 24 soatdan eski bo'lmasin (replay himoyasi)
	if ad := values.Get("auth_date"); ad != "" {
		if ts, err := strconv.ParseInt(ad, 10, 64); err == nil {
			if time.Since(time.Unix(ts, 0)) > 24*time.Hour {
				return nil, false
			}
		}
	}

	var u TgUser
	if err := json.Unmarshal([]byte(values.Get("user")), &u); err != nil || u.ID == 0 {
		return nil, false
	}
	return &u, true
}

func hmacSHA256(key, msg []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return m.Sum(nil)
}

// nullIfZero — 0 bo'lsa NULL (DB'da tg_user_id bo'sh qolishi uchun)
func nullIfZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// ===== HTTP handler =====

type TelegramHandler struct {
	pool *pgxpool.Pool
	cfg  *Config
}

func NewTelegramHandler(pool *pgxpool.Pool, cfg *Config) *TelegramHandler {
	return &TelegramHandler{pool: pool, cfg: cfg}
}

// resolveTgUserID — so'rovdan ishonchli Telegram user ID ni aniqlaydi.
// 1) X-Telegram-Init-Data header (yoki ?initData=) bot token bilan tekshiriladi (xavfsiz).
// 2) Faqat DEBUG (non-release) rejimda, token yo'q bo'lsa — ?tg_id= ga ruxsat (test uchun).
func (h *TelegramHandler) resolveTgUserID(c *gin.Context) (int64, bool) {
	initData := c.GetHeader("X-Telegram-Init-Data")
	if initData == "" {
		initData = c.Query("initData")
	}
	if u, ok := verifyTelegramInitData(initData, h.cfg.TelegramBotToken); ok {
		return u.ID, true
	}
	// Dev fallback — faqat TELEGRAM_DEV_AUTH=true bo'lganda (lokal test uchun).
	// Productionда bu bayroqни o'chiring va TELEGRAM_BOT_TOKEN ni sozlang.
	if h.cfg.TelegramDevAuth {
		if idStr := c.Query("tg_id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id != 0 {
				return id, true
			}
		}
	}
	return 0, false
}

// GET /api/my/donations — joriy Telegram foydalanuvchisining yordamlari tarixi
func (h *TelegramHandler) MyDonations(c *gin.Context) {
	tgID, ok := h.resolveTgUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Telegram identifikatsiya talab qilinadi"})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT d.id, d.person_id, COALESCE(p.name,''), d.provider, d.amount_tiyin,
		       d.status, d.created_at, d.paid_at
		FROM donations d
		LEFT JOIN people p ON p.id = d.person_id
		WHERE d.tg_user_id = $1
		ORDER BY d.created_at DESC
		LIMIT 200`, tgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	items := []gin.H{}
	var totalPaid int64
	people := map[int64]bool{}
	for rows.Next() {
		var id, personID, amount int64
		var name, provider, status string
		var createdAt time.Time
		var paidAt *time.Time
		if err := rows.Scan(&id, &personID, &name, &provider, &amount, &status, &createdAt, &paidAt); err != nil {
			continue
		}
		if status == "paid" {
			totalPaid += amount
			people[personID] = true
		}
		items = append(items, gin.H{
			"id":          id,
			"person_id":   personID,
			"person_name": name,
			"provider":    provider,
			"amount_som":  amount / 100,
			"status":      status,
			"created_at":  createdAt,
			"paid_at":     paidAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":          items,
		"total_paid_som": totalPaid / 100,
		"paid_people":    len(people),
		"count":          len(items),
	})
}
