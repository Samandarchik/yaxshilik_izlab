package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ===== Security headers =====

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-XSS-Protection", "0")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// ===== Client IP (proxy-aware) =====

// clientIP — TrustProxy yoqilgan bo'lsa X-Forwarded-For dagi birinchi IP,
// aks holda to'g'ridan-to'g'ri ulanish manzili.
func clientIP(c *gin.Context, trustProxy bool) string {
	if trustProxy {
		if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			// "client, proxy1, proxy2" — birinchisi haqiqiy mijoz
			if i := indexByte(xff, ','); i >= 0 {
				return trimSpace(xff[:i])
			}
			return trimSpace(xff)
		}
		if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
			return trimSpace(xrip)
		}
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// ===== IP whitelist (webhook'lar uchun) =====

// IPWhitelist — ruxsat etilgan IP/CIDR ro'yxati bo'sh bo'lmasa, faqat shularga ruxsat.
// Bo'sh bo'lsa hamma o'tadi (imzo/secret himoyasi baribir bor).
func IPWhitelist(cfg *Config) gin.HandlerFunc {
	allowed := cfg.WebhookAllowIPs
	// Oldindan parslab olamiz
	var nets []*net.IPNet
	var ips []net.IP
	for _, a := range allowed {
		if _, n, err := net.ParseCIDR(a); err == nil {
			nets = append(nets, n)
			continue
		}
		if ip := net.ParseIP(a); ip != nil {
			ips = append(ips, ip)
		}
	}
	return func(c *gin.Context) {
		if len(nets) == 0 && len(ips) == 0 {
			c.Next()
			return
		}
		raw := clientIP(c, cfg.TrustProxy)
		ip := net.ParseIP(raw)
		if ip != nil {
			for _, n := range nets {
				if n.Contains(ip) {
					c.Next()
					return
				}
			}
			for _, a := range ips {
				if a.Equal(ip) {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "ip not allowed"})
	}
}

// ===== Rate limiter (login brute-force himoyasi) =====

type rateLimiter struct {
	mu       sync.Mutex
	hits     map[string][]time.Time
	max      int
	window   time.Duration
	lastSwap time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		hits:     make(map[string][]time.Time),
		max:      max,
		window:   window,
		lastSwap: time.Now(),
	}
}

// allow — key (IP) uchun limitdan oshmaganini tekshiradi.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Vaqti o'tgan yozuvlarni tozalash
	recent := rl.hits[key][:0]
	for _, t := range rl.hits[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	rl.hits[key] = recent

	// Map o'sib ketmasligi uchun davriy tozalash
	if now.Sub(rl.lastSwap) > 10*rl.window {
		for k, v := range rl.hits {
			if len(v) == 0 {
				delete(rl.hits, k)
			}
		}
		rl.lastSwap = now
	}

	if len(recent) >= rl.max {
		return false
	}
	rl.hits[key] = append(rl.hits[key], now)
	return true
}

// RateLimit — har bir IP uchun belgilangan oynada max so'rovga ruxsat beradi.
func RateLimit(cfg *Config, max int, window time.Duration) gin.HandlerFunc {
	rl := newRateLimiter(max, window)
	return func(c *gin.Context) {
		ip := clientIP(c, cfg.TrustProxy)
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "juda ko'p urinish, biroz kuting",
			})
			return
		}
		c.Next()
	}
}

// ===== Xavfsiz tasodifiy hex =====

func secureRandHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand ishlamasa — vaqtga asoslangan zaxira (deyarli bo'lmaydi)
		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> (i % 8))
		}
	}
	return hex.EncodeToString(b)
}

// ===== Audit log =====

// audit — muhim hodisalarni audit_log jadvaliga yozadi (xato bo'lsa jim o'tadi).
func audit(pool *pgxpool.Pool, actor, action, target string, payload any) {
	var raw []byte
	if payload != nil {
		raw, _ = json.Marshal(payload)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = pool.Exec(ctx,
		`INSERT INTO audit_log(actor, action, target, payload) VALUES($1,$2,$3,$4)`,
		actor, action, target, raw)
}
