package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DonationsHandler struct {
	pool *pgxpool.Pool
}

func NewDonationsHandler(pool *pgxpool.Pool) *DonationsHandler {
	return &DonationsHandler{pool: pool}
}

// GET /api/admin/donations?status=...&provider=...&person_id=...&limit=&offset=
func (h *DonationsHandler) List(c *gin.Context) {
	status := c.Query("status")
	provider := c.Query("provider")
	personIDStr := c.Query("person_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	q := `
        SELECT d.id, d.person_id, d.provider, d.amount_tiyin, d.status, d.anonim, d.donor_name, d.donor_phone,
               d.click_trans_id, d.click_paydoc_id, d.merchant_prepare_id,
               d.payme_id, d.payme_state, d.payme_create_time, d.payme_perform_time, d.payme_cancel_time, d.payme_reason,
               d.created_at, d.updated_at, d.paid_at,
               p.name as person_name
        FROM donations d
        LEFT JOIN people p ON p.id = d.person_id
        WHERE 1=1`
	args := []any{}
	idx := 1
	if status != "" {
		q += " AND d.status=$" + strconv.Itoa(idx)
		args = append(args, status)
		idx++
	}
	if provider != "" {
		q += " AND d.provider=$" + strconv.Itoa(idx)
		args = append(args, provider)
		idx++
	}
	if personIDStr != "" {
		if pid, err := strconv.ParseInt(personIDStr, 10, 64); err == nil {
			q += " AND d.person_id=$" + strconv.Itoa(idx)
			args = append(args, pid)
			idx++
		}
	}
	q += " ORDER BY d.created_at DESC LIMIT $" + strconv.Itoa(idx) + " OFFSET $" + strconv.Itoa(idx+1)
	args = append(args, limit, offset)

	rows, err := h.pool.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	items := []gin.H{}
	for rows.Next() {
		var d Donation
		var personName *string
		if err := rows.Scan(&d.ID, &d.PersonID, &d.Provider, &d.AmountTiyin, &d.Status, &d.Anonim, &d.DonorName, &d.DonorPhone,
			&d.ClickTransID, &d.ClickPaydocID, &d.MerchantPrepareID,
			&d.PaymeID, &d.PaymeState, &d.PaymeCreateTime, &d.PaymePerformTime, &d.PaymeCancelTime, &d.PaymeReason,
			&d.CreatedAt, &d.UpdatedAt, &d.PaidAt, &personName); err != nil {
			continue
		}
		name := ""
		if personName != nil {
			name = *personName
		}
		items = append(items, gin.H{
			"id":             d.ID,
			"person_id":      d.PersonID,
			"person_name":    name,
			"provider":       d.Provider,
			"amount_tiyin":   d.AmountTiyin,
			"amount_som":     d.AmountTiyin / 100,
			"status":         d.Status,
			"anonim":         d.Anonim,
			"donor_name":     d.DonorName,
			"donor_phone":    d.DonorPhone,
			"click_trans_id": d.ClickTransID,
			"payme_id":       d.PaymeID,
			"payme_state":    d.PaymeState,
			"created_at":     d.CreatedAt,
			"paid_at":        d.PaidAt,
		})
	}

	// total count
	var total int64
	_ = h.pool.QueryRow(c.Request.Context(), `SELECT COUNT(*) FROM donations`).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
