package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsHandler struct {
	pool *pgxpool.Pool
}

func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{pool: pool}
}

// GET /api/stats/public
func (h *StatsHandler) Public(c *gin.Context) {
	ctx := c.Request.Context()
	var totalRaised, totalPeople, activePeople, urgentPeople, closedPeople int64
	var totalDonors, monthDonors int64
	var monthRaisedTiyin int64
	var prevMonthRaisedTiyin int64

	_ = h.pool.QueryRow(ctx, `SELECT COALESCE(SUM(raised),0) FROM people`).Scan(&totalRaised)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM people`).Scan(&totalPeople)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE status='active'`).Scan(&activePeople)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE urgent=true AND status='active'`).Scan(&urgentPeople)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE status='closed' OR raised >= target`).Scan(&closedPeople)
	_ = h.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT COALESCE(NULLIF(donor_phone,''), NULLIF(donor_name,''), id::text)) FROM donations WHERE status='paid'`,
	).Scan(&totalDonors)
	_ = h.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT COALESCE(NULLIF(donor_phone,''), NULLIF(donor_name,''), id::text))
		 FROM donations WHERE status='paid' AND paid_at >= date_trunc('month', NOW())`,
	).Scan(&monthDonors)
	_ = h.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_tiyin),0) FROM donations
		 WHERE status='paid' AND paid_at >= date_trunc('month', NOW())`,
	).Scan(&monthRaisedTiyin)
	_ = h.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_tiyin),0) FROM donations
		WHERE status='paid'
		  AND paid_at >= date_trunc('month', NOW()) - INTERVAL '1 month'
		  AND paid_at <  date_trunc('month', NOW())`,
	).Scan(&prevMonthRaisedTiyin)

	delta := 0
	if prevMonthRaisedTiyin > 0 {
		delta = int(((monthRaisedTiyin - prevMonthRaisedTiyin) * 100) / prevMonthRaisedTiyin)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_raised_som":    totalRaised,
		"total_people":        totalPeople,
		"active_people":       activePeople,
		"urgent_people":       urgentPeople,
		"closed_people":       closedPeople,
		"total_donors":        totalDonors,
		"month_donors":        monthDonors,
		"month_raised_som":    monthRaisedTiyin / 100,
		"month_delta_percent": delta,
	})
}

// GET /api/admin/stats
func (h *StatsHandler) Admin(c *gin.Context) {
	ctx := c.Request.Context()
	var totalRaised, totalPeople, activePeople int64
	var paidCount, pendingCount, cancelledCount int64
	var clickCount, paymeCount int64
	var paidSumTiyin int64

	_ = h.pool.QueryRow(ctx, `SELECT COALESCE(SUM(raised),0) FROM people`).Scan(&totalRaised)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM people`).Scan(&totalPeople)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM people WHERE status='active'`).Scan(&activePeople)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM donations WHERE status='paid'`).Scan(&paidCount)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM donations WHERE status='pending'`).Scan(&pendingCount)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM donations WHERE status='cancelled'`).Scan(&cancelledCount)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM donations WHERE provider='click'`).Scan(&clickCount)
	_ = h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM donations WHERE provider='payme'`).Scan(&paymeCount)
	_ = h.pool.QueryRow(ctx, `SELECT COALESCE(SUM(amount_tiyin),0) FROM donations WHERE status='paid'`).Scan(&paidSumTiyin)

	// Oxirgi 14 kunlik dnevnik
	rows, _ := h.pool.Query(ctx, `
        SELECT DATE(created_at) AS day, COUNT(*), COALESCE(SUM(amount_tiyin),0)
        FROM donations WHERE status='paid' AND created_at >= NOW() - INTERVAL '14 days'
        GROUP BY day ORDER BY day ASC`)
	daily := []gin.H{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var day any
			var count, sum int64
			if err := rows.Scan(&day, &count, &sum); err == nil {
				daily = append(daily, gin.H{"day": day, "count": count, "amount_som": sum / 100})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_raised_som": totalRaised,
		"total_people":     totalPeople,
		"active_people":    activePeople,
		"paid_count":       paidCount,
		"pending_count":    pendingCount,
		"cancelled_count":  cancelledCount,
		"click_count":      clickCount,
		"payme_count":      paymeCount,
		"paid_sum_som":     paidSumTiyin / 100,
		"daily":            daily,
	})
}
