package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedHandler struct {
	pool *pgxpool.Pool
}

func NewFeedHandler(pool *pgxpool.Pool) *FeedHandler {
	return &FeedHandler{pool: pool}
}

// GET /api/donations/recent?limit=10
// Public: faqat to'langan donation'lar, masking bilan (anonimlar uchun)
func (h *FeedHandler) Recent(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := h.pool.Query(c.Request.Context(), `
        SELECT d.id, d.amount_tiyin, d.anonim, d.donor_name, d.created_at, d.paid_at,
               p.id, p.name
        FROM donations d
        JOIN people p ON p.id = d.person_id
        WHERE d.status='paid'
        ORDER BY COALESCE(d.paid_at, d.created_at) DESC
        LIMIT $1`, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var did, amount, pid int64
		var anonim bool
		var donor, pname string
		var createdAt, paidAt any
		if err := rows.Scan(&did, &amount, &anonim, &donor, &createdAt, &paidAt, &pid, &pname); err != nil {
			continue
		}
		display := donor
		if anonim || display == "" {
			display = "Anonim"
		} else if len(display) > 1 {
			// "Aziz Karimov" -> "Aziz K."
			parts := []rune(display)
			space := -1
			for i, r := range parts {
				if r == ' ' {
					space = i
					break
				}
			}
			if space > 0 && space+1 < len(parts) {
				display = string(parts[:space]) + " " + string(parts[space+1]) + "."
			}
		}
		ts := paidAt
		if ts == nil {
			ts = createdAt
		}
		items = append(items, gin.H{
			"id":          did,
			"donor":       display,
			"anonim":      anonim,
			"amount_som":  amount / 100,
			"person_id":   pid,
			"person_name": pname,
			"at":          ts,
		})
	}
	c.JSON(http.StatusOK, items)
}

// GET /api/success-stories?limit=6
// Public: maqsadga yetgan yoki 'closed' bemorlar
func (h *FeedHandler) SuccessStories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if limit <= 0 || limit > 24 {
		limit = 6
	}
	rows, err := h.pool.Query(c.Request.Context(), `
        SELECT id, name, age, region, diagnosis, photo_url, raised, target
        FROM people
        WHERE status='closed' OR raised >= target
        ORDER BY updated_at DESC
        LIMIT $1`, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, raised, target int64
		var age int
		var name, region, diagnosis, photo string
		if err := rows.Scan(&id, &name, &age, &region, &diagnosis, &photo, &raised, &target); err != nil {
			continue
		}
		items = append(items, gin.H{
			"id":        id,
			"name":      name,
			"age":       age,
			"region":    region,
			"diagnosis": diagnosis,
			"photo_url": photo,
			"raised":    raised,
			"target":    target,
		})
	}
	c.JSON(http.StatusOK, items)
}
