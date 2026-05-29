package main

import (
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	clickActionPrepare  = 0
	clickActionComplete = 1
)

// Click webhook javob kodlari (CLICK_DOCS.md bo'yicha)
const (
	clickOK              = 0
	clickSignCheckFailed = -1
	clickAmountMismatch  = -2
	clickActionNotFound  = -3
	clickAlreadyPaid     = -4
	clickUserNotFound    = -5
	clickTxNotFound      = -6
	clickUpdateFailed    = -7
	clickBadRequest      = -8
	clickCancelled       = -9
)

type ClickHandler struct {
	pool *pgxpool.Pool
	cfg  *Config
}

func NewClickHandler(pool *pgxpool.Pool, cfg *Config) *ClickHandler {
	return &ClickHandler{pool: pool, cfg: cfg}
}

// POST /api/click/create  { person_id, amount, anonim, donor_name, donor_phone }
// Donation yaratadi va redirect URL qaytaradi
func (h *ClickHandler) Create(c *gin.Context) {
	var req struct {
		PersonID   int64  `json:"person_id"`
		Amount     int64  `json:"amount"` // so'm
		Anonim     bool   `json:"anonim"`
		DonorName  string `json:"donor_name"`
		DonorPhone string `json:"donor_phone"`
		TgUserID   int64  `json:"tg_user_id"`
		TgUsername string `json:"tg_username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PersonID == 0 || req.Amount < 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id va amount (>=1000) majburiy"})
		return
	}
	if req.Amount > 10_000_000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maksimal 10 000 000 so'm"})
		return
	}

	// person mavjudligini tekshirish
	var exists bool
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM people WHERE id=$1 AND status='active')`, req.PersonID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}

	var donationID int64
	err := h.pool.QueryRow(c.Request.Context(), `
        INSERT INTO donations(person_id, provider, amount_tiyin, status, anonim, donor_name, donor_phone, tg_user_id, tg_username)
        VALUES($1, 'click', $2, 'pending', $3, $4, $5, $6, $7)
        RETURNING id`,
		req.PersonID, req.Amount*100, req.Anonim, req.DonorName, req.DonorPhone,
		nullIfZero(req.TgUserID), req.TgUsername).Scan(&donationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "donation insert failed"})
		return
	}

	params := url.Values{}
	params.Set("service_id", h.cfg.ClickServiceID)
	params.Set("merchant_id", h.cfg.ClickMerchantID)
	params.Set("amount", strconv.FormatInt(req.Amount, 10))
	params.Set("transaction_param", strconv.FormatInt(donationID, 10))
	params.Set("return_url", h.cfg.ClickReturnURL)

	redirect := "https://my.click.uz/services/pay?" + params.Encode()
	c.JSON(http.StatusOK, gin.H{
		"donation_id":  donationID,
		"redirect_url": redirect,
	})
}

// POST /api/click/webhook (form-encoded)
// Prepare (action=0) yoki Complete (action=1)
func (h *ClickHandler) Webhook(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusOK, clickResp("", 0, clickBadRequest, "bad request"))
		return
	}
	f := c.Request.PostForm

	clickTransID := f.Get("click_trans_id")
	serviceID := f.Get("service_id")
	merchantTransID := f.Get("merchant_trans_id")
	amount := f.Get("amount")
	actionStr := f.Get("action")
	signTime := f.Get("sign_time")
	signString := f.Get("sign_string")
	merchantPrepareID := f.Get("merchant_prepare_id")
	clickPaydocID := f.Get("click_paydoc_id")
	clientErr := f.Get("error")

	action, _ := strconv.Atoi(actionStr)

	// 1) Imzo tekshirish
	if !verifyClickSignature(h.cfg.ClickSecretKey, clickTransID, serviceID, merchantTransID,
		merchantPrepareID, amount, actionStr, signTime, signString, action) {
		c.JSON(http.StatusOK, clickResp(clickTransID, merchantTransIDInt(merchantTransID), clickSignCheckFailed, "SIGN CHECK FAILED"))
		return
	}

	// 2) Donation topish
	donationID, err := strconv.ParseInt(merchantTransID, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, clickResp(clickTransID, 0, clickTxNotFound, "transaction not found"))
		return
	}
	var d Donation
	row := h.pool.QueryRow(c.Request.Context(), `
        SELECT id, person_id, provider, amount_tiyin, status, anonim, donor_name, donor_phone,
               click_trans_id, click_paydoc_id, merchant_prepare_id,
               payme_id, payme_state, payme_create_time, payme_perform_time, payme_cancel_time, payme_reason,
               created_at, updated_at, paid_at
        FROM donations WHERE id=$1 AND provider='click'`, donationID)
	if err := row.Scan(&d.ID, &d.PersonID, &d.Provider, &d.AmountTiyin, &d.Status, &d.Anonim, &d.DonorName, &d.DonorPhone,
		&d.ClickTransID, &d.ClickPaydocID, &d.MerchantPrepareID,
		&d.PaymeID, &d.PaymeState, &d.PaymeCreateTime, &d.PaymePerformTime, &d.PaymeCancelTime, &d.PaymeReason,
		&d.CreatedAt, &d.UpdatedAt, &d.PaidAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickTxNotFound, "transaction not found"))
			return
		}
		c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickUpdateFailed, "db error"))
		return
	}

	// 3) Summa tekshirish (Click so'm da yuboradi, biz tiyin saqlaymiz)
	amountSomFloat, _ := strconv.ParseFloat(amount, 64)
	if int64(amountSomFloat*100) != d.AmountTiyin {
		c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickAmountMismatch, "incorrect amount"))
		return
	}

	switch action {
	case clickActionPrepare:
		// Allaqachon to'langan
		if d.Status == "paid" {
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickAlreadyPaid, "already paid"))
			return
		}
		// Allaqachon bekor qilingan
		if d.Status == "cancelled" {
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickCancelled, "cancelled"))
			return
		}
		// Idempotent: agar prepared bo'lsa, OK
		_, err := h.pool.Exec(c.Request.Context(),
			`UPDATE donations SET status='prepared', click_trans_id=$1, click_paydoc_id=$2,
                merchant_prepare_id=$3, updated_at=NOW() WHERE id=$4`,
			clickTransID, clickPaydocID, donationID, donationID)
		if err != nil {
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickUpdateFailed, "update failed"))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"click_trans_id":      mustInt(clickTransID),
			"merchant_trans_id":   merchantTransID,
			"merchant_prepare_id": donationID,
			"error":               clickOK,
			"error_note":          "Success",
		})
		return

	case clickActionComplete:
		// Idempotent
		if d.Status == "paid" {
			c.JSON(http.StatusOK, gin.H{
				"click_trans_id":      mustInt(clickTransID),
				"merchant_trans_id":   merchantTransID,
				"merchant_confirm_id": donationID,
				"error":               clickAlreadyPaid,
				"error_note":          "Already paid",
			})
			return
		}
		if d.Status != "prepared" {
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickBadRequest, "not prepared"))
			return
		}
		// Mijoz xatoligi (user bekor qildi)
		if clientErr != "" && clientErr != "0" {
			_, _ = h.pool.Exec(c.Request.Context(),
				`UPDATE donations SET status='cancelled', updated_at=NOW() WHERE id=$1`, donationID)
			ec, _ := strconv.Atoi(clientErr)
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, ec, "cancelled"))
			return
		}

		// Muvaffaqiyatli — transactionda yangilash
		if err := h.markPaid(c.Request.Context(), d); err != nil {
			c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickUpdateFailed, "update failed"))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"click_trans_id":      mustInt(clickTransID),
			"merchant_trans_id":   merchantTransID,
			"merchant_confirm_id": donationID,
			"error":               clickOK,
			"error_note":          "Success",
		})
		return

	default:
		c.JSON(http.StatusOK, clickResp(clickTransID, donationID, clickActionNotFound, "action not found"))
	}
}

// markPaid — donatsiyani atomik ravishda 'paid' qiladi. Faqat 'prepared' holatdan
// o'tkazadi, shuning uchun bir vaqtda kelgan ikkita webhook'da pul ikki marta
// hisoblanmaydi (race-safe). people.raised cheklanmaydi — refund to'g'ri ishlashi uchun.
func (h *ClickHandler) markPaid(ctx context.Context, d Donation) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE donations SET status='paid', paid_at=NOW(), updated_at=NOW()
		 WHERE id=$1 AND status='prepared'`, d.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Boshqa so'rov allaqachon to'lab bo'lgan — qayta hisoblamaymiz (idempotent)
		return tx.Commit(ctx)
	}

	amountSom := d.AmountTiyin / 100
	if _, err := tx.Exec(ctx,
		`UPDATE people SET raised = raised + $1, donors = donors + 1, updated_at=NOW() WHERE id=$2`,
		amountSom, d.PersonID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	audit(h.pool, "click", "donation_paid", strconv.FormatInt(d.ID, 10),
		gin.H{"person_id": d.PersonID, "amount_som": amountSom, "click_trans_id": d.ClickTransID})
	return nil
}

func verifyClickSignature(secret, clickTransID, serviceID, merchantTransID, merchantPrepareID,
	amount, action, signTime, signString string, actionInt int) bool {
	var raw string
	if actionInt == clickActionPrepare {
		raw = clickTransID + serviceID + secret + merchantTransID + amount + action + signTime
	} else {
		raw = clickTransID + serviceID + secret + merchantTransID + merchantPrepareID + amount + action + signTime
	}
	sum := md5.Sum([]byte(raw))
	expected := hex.EncodeToString(sum[:])
	// Constant-time solishtirish (timing attack oldini olish)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signString)) == 1
}

func clickResp(clickTransID string, merchantTransID int64, errCode int, note string) gin.H {
	resp := gin.H{
		"error":      errCode,
		"error_note": note,
	}
	if clickTransID != "" {
		resp["click_trans_id"] = mustInt(clickTransID)
	}
	if merchantTransID != 0 {
		resp["merchant_trans_id"] = strconv.FormatInt(merchantTransID, 10)
	}
	return resp
}

func merchantTransIDInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func mustInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
