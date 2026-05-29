package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	GinMode    string

	PGUrl string

	JWTSecret          string
	AdminEmail         string
	AdminPassword      string
	AdminResetPassword bool // true => mavjud admin parolini .env'dagiga majburan yangilaydi

	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	MinioPublicHost string

	PaycomMerchantID string
	PaycomSecretKey  string

	TelegramBotToken string // Mini App initData'ni tekshirish uchun
	TelegramDevAuth  bool   // true => ?tg_id= bilan test loginga ruxsat (FAQAT lokal test!)

	ClickMerchantID     string
	ClickServiceID      string
	ClickMerchantUserID string
	ClickSecretKey      string
	ClickReturnURL      string
	PaymeReturnURL      string

	// Xavfsizlik
	CORSOrigins     []string // ruxsat etilgan domenlar (bo'sh = debug'da *, release'da faqat same-origin)
	WebhookAllowIPs []string // Click/Payme webhook IP whitelist (bo'sh = o'chiq)
	TrustProxy      bool     // reverse proxy ortida bo'lsa X-Forwarded-For ga ishonish
}

func (c *Config) IsRelease() bool { return c.GinMode == "release" }

func LoadConfig() *Config {
	// Try .env from cwd, then parent (backend run from project root or from backend/)
	for _, p := range []string{".env", "../.env"} {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err != nil {
				log.Printf("warn: failed to load %s: %v", p, err)
			} else {
				log.Printf(".env loaded from %s", p)
			}
			break
		}
	}

	c := &Config{
		ServerPort: getenv("SERVER_PORT", "8080"),
		GinMode:    getenv("GIN_MODE", "debug"),

		PGUrl: must("PG_URL"),

		JWTSecret:     must("JWT_SECRET"),
		AdminEmail:    getenv("ADMIN_EMAIL", "admin@prava-olamiz.uz"),
		AdminPassword: getenv("ADMIN_PASSWORD", "admin12345"),

		AdminResetPassword: getenv("ADMIN_RESET_PASSWORD", "false") == "true",

		MinioEndpoint:   getenv("MINIO_ENDPOINT", ""),
		MinioAccessKey:  getenv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:  getenv("MINIO_SECRET_KEY", ""),
		MinioBucketName: getenv("MINIO_BUCKET_NAME", ""),
		MinioPublicHost: getenv("MINIO_PUBLIC_HOST", ""),

		PaycomMerchantID: getenv("PAYCOM_MERCHANT_ID", ""),
		PaycomSecretKey:  getenv("PAYCOM_SECRET_KEY", ""),

		TelegramBotToken: getenv("TELEGRAM_BOT_TOKEN", ""),
		TelegramDevAuth:  getenv("TELEGRAM_DEV_AUTH", "false") == "true",

		ClickMerchantID:     getenv("CLICK_MERCHANT_ID", ""),
		ClickServiceID:      getenv("CLICK_SERVICE_ID", ""),
		ClickMerchantUserID: getenv("CLICK_MERCHANT_USER_ID", ""),
		ClickSecretKey:      getenv("CLICK_SECRET_KEY", ""),
		ClickReturnURL:      getenv("CLICK_RETURN_URL", "http://localhost:8080/?paid=1"),
		PaymeReturnURL:      getenv("PAYME_RETURN_URL", ""),

		CORSOrigins:     splitCSV(getenv("CORS_ORIGINS", "")),
		WebhookAllowIPs: splitCSV(getenv("WEBHOOK_ALLOW_IPS", "")),
		TrustProxy:      getenv("TRUST_PROXY", "false") == "true",
	}

	if c.PaymeReturnURL == "" {
		c.PaymeReturnURL = c.ClickReturnURL
	}

	// Production xavfsizlik tekshiruvlari — pul aylanadigan tizim uchun majburiy
	if c.IsRelease() {
		if len(c.JWTSecret) < 32 {
			log.Fatalf("xavfsizlik: JWT_SECRET kamida 32 belgi bo'lishi shart (release). `openssl rand -hex 32` bilan yarating")
		}
		if c.AdminPassword == "" || c.AdminPassword == "admin12345" {
			log.Fatalf("xavfsizlik: ADMIN_PASSWORD default qiymatda. Mustahkam parolga almashtiring")
		}
		if c.PaycomSecretKey == "" && c.ClickSecretKey == "" {
			log.Fatalf("xavfsizlik: na PAYCOM_SECRET_KEY na CLICK_SECRET_KEY sozlangan")
		}
	}

	return c
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("required env %s is not set", k)
	}
	return v
}
