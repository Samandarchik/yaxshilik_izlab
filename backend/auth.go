package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	pool   *pgxpool.Pool
	secret string
}

func NewAuthHandler(pool *pgxpool.Pool, secret string) *AuthHandler {
	return &AuthHandler{pool: pool, secret: secret}
}

// Login: POST /api/admin/login { email, password } -> { token, admin }
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var a Admin
	row := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, email, password_hash, full_name, role, created_at FROM admins WHERE email=$1`, req.Email)
	if err := row.Scan(&a.ID, &a.Email, &a.PasswordHash, &a.FullName, &a.Role, &a.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			audit(h.pool, req.Email, "admin_login_failed", c.ClientIP(), gin.H{"reason": "no_user"})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(req.Password)); err != nil {
		audit(h.pool, req.Email, "admin_login_failed", c.ClientIP(), gin.H{"reason": "bad_password"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tok, err := h.issueToken(a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token error"})
		return
	}
	audit(h.pool, a.Email, "admin_login_success", c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"token": tok, "admin": a})
}

// Me: GET /api/admin/me -> current admin
func (h *AuthHandler) Me(c *gin.Context) {
	aid := c.GetInt64("admin_id")
	var a Admin
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id, email, password_hash, full_name, role, created_at FROM admins WHERE id=$1`, aid).
		Scan(&a.ID, &a.Email, &a.PasswordHash, &a.FullName, &a.Role, &a.CreatedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AuthHandler) issueToken(a Admin) (string, error) {
	claims := jwtClaims{
		AdminID: a.ID,
		Email:   a.Email,
		Role:    a.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   a.Email,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(h.secret))
}

// BootstrapAdmin: agar admin yo'q bo'lsa, .env'dagi default admin'ni yaratadi.
// resetPassword=true bo'lsa, mavjud admin parolini ham .env'dagiga yangilaydi
// (ADMIN_RESET_PASSWORD=true — bir martalik parolni tiklash uchun).
func BootstrapAdmin(ctx context.Context, pool *pgxpool.Pool, email, password string, resetPassword bool) {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM admins`).Scan(&count); err != nil {
		log.Printf("bootstrap admin: count: %v", err)
		return
	}
	if count > 0 {
		// Admin mavjud — faqat reset so'ralganda parolni yangilaymiz
		if resetPassword {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("bootstrap admin: reset hash: %v", err)
				return
			}
			tag, err := pool.Exec(ctx,
				`UPDATE admins SET password_hash=$1 WHERE email=$2`, string(hash), email)
			if err != nil {
				log.Printf("bootstrap admin: reset failed: %v", err)
				return
			}
			if tag.RowsAffected() > 0 {
				log.Printf("bootstrap admin: %s paroli .env'dagi ADMIN_PASSWORD'ga yangilandi — endi ADMIN_RESET_PASSWORD ni false ga qaytaring!", email)
			} else {
				log.Printf("bootstrap admin: reset uchun %s topilmadi", email)
			}
		}
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bootstrap admin: hash: %v", err)
		return
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO admins(email, password_hash, full_name, role) VALUES($1, $2, $3, 'admin')`,
		email, string(hash), "Bosh administrator")
	if err != nil {
		log.Printf("bootstrap admin: insert: %v", err)
		return
	}
	log.Printf("bootstrap admin: created %s (parolni .env dagi ADMIN_PASSWORD'dan oling)", email)
}
