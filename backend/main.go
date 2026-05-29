package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := LoadConfig()
	gin.SetMode(cfg.GinMode)

	ctx := context.Background()
	pool := InitDB(ctx, cfg.PGUrl)
	defer pool.Close()

	BootstrapAdmin(ctx, pool, cfg.AdminEmail, cfg.AdminPassword, cfg.AdminResetPassword)

	authH := NewAuthHandler(pool, cfg.JWTSecret)
	peopleH := NewPeopleHandler(pool)
	donH := NewDonationsHandler(pool)
	statsH := NewStatsHandler(pool)
	feedH := NewFeedHandler(pool)
	clickH := NewClickHandler(pool, cfg)
	paymeH := NewPaymeHandler(pool, cfg)
	uploadH := NewUploadHandler(cfg)
	tgH := NewTelegramHandler(pool, cfg)

	r := gin.Default()

	// Yuklanadigan body hajmini cheklash (DoS himoyasi)
	r.MaxMultipartMemory = 8 << 20 // 8 MB

	r.Use(SecurityHeaders())

	// CORS — frontend backend bilan bir domendan beriladi (same-origin), shuning uchun
	// CORS faqat tashqi domen kerak bo'lganda yoqiladi:
	//  - CORS_ORIGINS sozlangan bo'lsa => faqat o'sha domenlar
	//  - aks holda, dev (debug) rejimida => barcha domenlar
	//  - release + CORS_ORIGINS bo'sh => CORS umuman qo'shilmaydi (tashqi domenlarga yopiq)
	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           300,
	}
	if len(cfg.CORSOrigins) > 0 {
		corsCfg.AllowOrigins = cfg.CORSOrigins
		r.Use(cors.New(corsCfg))
	} else if !cfg.IsRelease() {
		corsCfg.AllowAllOrigins = true // faqat dev
		r.Use(cors.New(corsCfg))
	}

	// ===== Static files =====
	// Loyiha rootidan ham, backend/ dan ham ishga tushishi mumkin.
	// React build => static/ papka. Admin panel => admin/index.html.
	rootDir := projectRoot()
	staticDir := filepath.Join(rootDir, "static")
	adminFile := filepath.Join(rootDir, "admin", "index.html")
	indexFile := filepath.Join(staticDir, "index.html")

	// React'ning JS/CSS/rasm fayllari
	r.Static("/assets", filepath.Join(staticDir, "assets"))

	// Admin panel
	r.StaticFile("/admin", adminFile)
	r.StaticFile("/admin/", adminFile)

	// SPA: /api va /admin'dan tashqari hamma narsa React index.html'ga tushadi
	// (mavjud statik fayl bo'lsa o'sha beriladi — favicon, robots va h.k.)
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		if p != "/" {
			if f := filepath.Join(staticDir, filepath.Clean(p)); fileExists(f) {
				c.File(f)
				return
			}
		}
		c.File(indexFile) // client-side routing uchun
	})

	// ===== Public API =====
	api := r.Group("/api")
	api.GET("/people", peopleH.ListPublic)
	api.GET("/people/:id", peopleH.GetPublic)
	api.GET("/stats/public", statsH.Public)
	api.GET("/donations/recent", feedH.Recent)
	api.GET("/success-stories", feedH.SuccessStories)

	// Donate (foydalanuvchi to'lov boshlaydi) — provider'ga qarab.
	// Spam/abuse oldini olish uchun yengil rate limit.
	donateRL := RateLimit(cfg, 20, time.Minute)
	api.POST("/click/create", donateRL, clickH.Create)
	api.POST("/payme/create", donateRL, paymeH.Create)

	// Telegram Mini App: joriy foydalanuvchining yordamlari tarixi
	api.GET("/my/donations", tgH.MyDonations)

	// Webhook'lar (Click va Payme serverlari chaqiradi) — IP whitelist + secret/imzo himoyasi
	wh := IPWhitelist(cfg)
	api.POST("/click/webhook", wh, clickH.Webhook)
	api.POST("/payme", wh, paymeH.Webhook)

	// ===== Admin auth ===== (brute-force himoyasi: 10 urinish / 5 daqiqa / IP)
	api.POST("/admin/login", RateLimit(cfg, 10, 5*time.Minute), authH.Login)

	// ===== Admin (JWT himoyalangan) =====
	admin := api.Group("/admin")
	admin.Use(AuthMiddleware(cfg.JWTSecret))
	admin.GET("/me", authH.Me)
	admin.GET("/stats", statsH.Admin)

	admin.GET("/people", peopleH.ListAdmin)
	admin.POST("/people", peopleH.Create)
	admin.PUT("/people/:id", peopleH.Update)
	admin.DELETE("/people/:id", peopleH.Delete)

	admin.GET("/donations", donH.List)

	admin.POST("/upload", uploadH.Upload)

	addr := ":" + cfg.ServerPort
	log.Printf("listening on %s  (admin: http://localhost%s/admin)", addr, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// projectRoot — loyiha ildizini topish (joriy dir yoki bitta yuqori).
// "admin" papkasi barqaror belgi sifatida ishlatiladi.
func projectRoot() string {
	if _, err := os.Stat("admin"); err == nil {
		return "."
	}
	if _, err := os.Stat("../admin"); err == nil {
		return ".."
	}
	return "."
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
