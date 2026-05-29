package main

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Payme error codes (PAYME_DOCS.md)
const (
	paymeOK               = 0
	paymeInvalidAmount    = -31001
	paymeTxNotFound       = -31003
	paymeCannotCancel     = -31007
	paymeCannotPerform    = -31008
	paymeOrderNotFound    = -31050
	paymeOrderAlreadyPaid = -31051
	paymeOrderNotPayable  = -31052
	paymeOrderClosed      = -31053
	paymeOpExpired        = -31054
	paymeInternal         = -31055
	paymeUserErr          = -31099
	paymeInvalidAuth      = -32504
	paymeMethodNotFound   = -32601
	paymeInvalidJSON      = -32700
)

// Payme transaction state
const (
	paymeStateCreated   int16 = 1
	paymeStatePaid      int16 = 2
	paymeStateCancelled int16 = -1
	paymeStateRefunded  int16 = -2
)

type PaymeHandler struct {
	pool *pgxpool.Pool
	cfg  *Config
}

func NewPaymeHandler(pool *pgxpool.Pool, cfg *Config) *PaymeHandler {
	return &PaymeHandler{pool: pool, cfg: cfg}
}

// ----- JSON-RPC types -----

type rpcReq struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcErr struct {
	Code    int            `json:"code"`
	Message map[string]any `json:"message"`
	Data    string         `json:"data,omitempty"`
}

type rpcResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}

func makeErr(code int, uz string) *rpcErr {
	return &rpcErr{Code: code, Message: map[string]any{
		"uz": uz, "ru": uz, "en": uz,
	}}
}

func send(c *gin.Context, id json.RawMessage, result any, e *rpcErr) {
	r := rpcResp{JSONRPC: "2.0", ID: id, Result: result}
	if e != nil {
		r.Error = e
		r.Result = nil
	}
	c.JSON(http.StatusOK, r)
}

// POST /api/payme/create  { person_id, amount, anonim, donor_name, donor_phone }
func (h *PaymeHandler) Create(c *gin.Context) {
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

	var exists bool
	if err := h.pool.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM people WHERE id=$1 AND status='active')`, req.PersonID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}

	var donationID int64
	err := h.pool.QueryRow(c.Request.Context(), `
        INSERT INTO donations(person_id, provider, amount_tiyin, status, anonim, donor_name, donor_phone, tg_user_id, tg_username)
        VALUES($1, 'payme', $2, 'pending', $3, $4, $5, $6, $7)
        RETURNING id`,
		req.PersonID, req.Amount*100, req.Anonim, req.DonorName, req.DonorPhone,
		nullIfZero(req.TgUserID), req.TgUsername).Scan(&donationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "donation insert failed"})
		return
	}

	raw := fmt.Sprintf("m=%s;ac.order_id=%d;a=%d;c=%s",
		h.cfg.PaycomMerchantID, donationID, req.Amount*100, h.cfg.PaymeReturnURL)
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	url := "https://checkout.paycom.uz/" + encoded

	c.JSON(http.StatusOK, gin.H{
		"donation_id":  donationID,
		"redirect_url": url,
	})
}

// POST /api/payme  — JSON-RPC dispatcher
func (h *PaymeHandler) Webhook(c *gin.Context) {
	// Body o'qish
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		send(c, nil, nil, makeErr(paymeInvalidJSON, "parse error"))
		return
	}
	var req rpcReq
	if err := json.Unmarshal(body, &req); err != nil {
		send(c, nil, nil, makeErr(paymeInvalidJSON, "parse error"))
		return
	}

	// Auth
	if !h.checkAuth(c) {
		send(c, req.ID, nil, makeErr(paymeInvalidAuth, "Insufficient privileges"))
		return
	}

	ctx := c.Request.Context()
	switch req.Method {
	case "CheckPerformTransaction":
		h.checkPerform(c, ctx, req)
	case "CreateTransaction":
		h.createTx(c, ctx, req)
	case "PerformTransaction":
		h.performTx(c, ctx, req)
	case "CancelTransaction":
		h.cancelTx(c, ctx, req)
	case "CheckTransaction":
		h.checkTx(c, ctx, req)
	case "GetStatement":
		h.getStatement(c, ctx, req)
	default:
		send(c, req.ID, nil, makeErr(paymeMethodNotFound, "method not found"))
	}
}

func (h *PaymeHandler) checkAuth(c *gin.Context) bool {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return false
	}
	// Constant-time solishtirish (timing attack oldini olish)
	userOK := subtle.ConstantTimeCompare([]byte(parts[0]), []byte("Paycom")) == 1
	passOK := subtle.ConstantTimeCompare([]byte(parts[1]), []byte(h.cfg.PaycomSecretKey)) == 1
	return userOK && passOK
}

// ---- Helper: order_id va donation yuklash ----

type checkParams struct {
	Amount  int64 `json:"amount"`
	Account struct {
		OrderID string `json:"order_id"`
	} `json:"account"`
}

func (h *PaymeHandler) loadDonationByOrder(ctx context.Context, orderID string) (*Donation, *rpcErr) {
	id, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, makeErr(paymeOrderNotFound, "Order not found")
	}
	var d Donation
	row := h.pool.QueryRow(ctx, `
        SELECT id, person_id, provider, amount_tiyin, status, anonim, donor_name, donor_phone,
               click_trans_id, click_paydoc_id, merchant_prepare_id,
               payme_id, payme_state, payme_create_time, payme_perform_time, payme_cancel_time, payme_reason,
               created_at, updated_at, paid_at
        FROM donations WHERE id=$1 AND provider='payme'`, id)
	err = row.Scan(&d.ID, &d.PersonID, &d.Provider, &d.AmountTiyin, &d.Status, &d.Anonim, &d.DonorName, &d.DonorPhone,
		&d.ClickTransID, &d.ClickPaydocID, &d.MerchantPrepareID,
		&d.PaymeID, &d.PaymeState, &d.PaymeCreateTime, &d.PaymePerformTime, &d.PaymeCancelTime, &d.PaymeReason,
		&d.CreatedAt, &d.UpdatedAt, &d.PaidAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, makeErr(paymeOrderNotFound, "Order not found")
		}
		return nil, makeErr(paymeInternal, "db error")
	}
	return &d, nil
}

func (h *PaymeHandler) loadDonationByPaymeID(ctx context.Context, paymeID string) (*Donation, *rpcErr) {
	var d Donation
	row := h.pool.QueryRow(ctx, `
        SELECT id, person_id, provider, amount_tiyin, status, anonim, donor_name, donor_phone,
               click_trans_id, click_paydoc_id, merchant_prepare_id,
               payme_id, payme_state, payme_create_time, payme_perform_time, payme_cancel_time, payme_reason,
               created_at, updated_at, paid_at
        FROM donations WHERE payme_id=$1`, paymeID)
	err := row.Scan(&d.ID, &d.PersonID, &d.Provider, &d.AmountTiyin, &d.Status, &d.Anonim, &d.DonorName, &d.DonorPhone,
		&d.ClickTransID, &d.ClickPaydocID, &d.MerchantPrepareID,
		&d.PaymeID, &d.PaymeState, &d.PaymeCreateTime, &d.PaymePerformTime, &d.PaymeCancelTime, &d.PaymeReason,
		&d.CreatedAt, &d.UpdatedAt, &d.PaidAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, makeErr(paymeTxNotFound, "Transaction not found")
		}
		return nil, makeErr(paymeInternal, "db error")
	}
	return &d, nil
}

// ---- 1. CheckPerformTransaction ----
func (h *PaymeHandler) checkPerform(c *gin.Context, ctx context.Context, req rpcReq) {
	var p checkParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		send(c, req.ID, nil, makeErr(paymeInvalidJSON, "invalid params"))
		return
	}
	d, eErr := h.loadDonationByOrder(ctx, p.Account.OrderID)
	if eErr != nil {
		send(c, req.ID, nil, eErr)
		return
	}
	if d.AmountTiyin != p.Amount {
		send(c, req.ID, nil, makeErr(paymeInvalidAmount, "Incorrect amount"))
		return
	}
	if d.Status == "paid" {
		send(c, req.ID, nil, makeErr(paymeOrderAlreadyPaid, "Already paid"))
		return
	}
	if d.Status == "cancelled" {
		send(c, req.ID, nil, makeErr(paymeOrderNotPayable, "Order is cancelled"))
		return
	}
	send(c, req.ID, gin.H{"allow": true}, nil)
}

// ---- 2. CreateTransaction ----
func (h *PaymeHandler) createTx(c *gin.Context, ctx context.Context, req rpcReq) {
	var p struct {
		ID      string `json:"id"`
		Time    int64  `json:"time"`
		Amount  int64  `json:"amount"`
		Account struct {
			OrderID string `json:"order_id"`
		} `json:"account"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		send(c, req.ID, nil, makeErr(paymeInvalidJSON, "invalid params"))
		return
	}

	// 12 soatdan eski bo'lsa rad etish (PAYME_DOCS — Xavfsizlik § 6)
	nowMs := time.Now().UnixMilli()
	if nowMs-p.Time > 12*60*60*1000 {
		send(c, req.ID, nil, makeErr(paymeCannotPerform, "Transaction expired"))
		return
	}

	// Allaqachon shu payme_id bilan tranzaksiya bormi?
	if existing, _ := h.loadDonationByPaymeID(ctx, p.ID); existing != nil {
		if existing.PaymeState != nil && *existing.PaymeState != paymeStateCreated {
			send(c, req.ID, nil, makeErr(paymeCannotPerform, "Cannot perform"))
			return
		}
		ct := int64(0)
		if existing.PaymeCreateTime != nil {
			ct = *existing.PaymeCreateTime
		}
		send(c, req.ID, gin.H{
			"create_time": ct,
			"transaction": strconv.FormatInt(existing.ID, 10),
			"state":       paymeStateCreated,
		}, nil)
		return
	}

	// Yangi tranzaksiya — order tekshirish
	d, eErr := h.loadDonationByOrder(ctx, p.Account.OrderID)
	if eErr != nil {
		send(c, req.ID, nil, eErr)
		return
	}
	if d.AmountTiyin != p.Amount {
		send(c, req.ID, nil, makeErr(paymeInvalidAmount, "Incorrect amount"))
		return
	}
	if d.Status == "paid" {
		send(c, req.ID, nil, makeErr(paymeOrderAlreadyPaid, "Already paid"))
		return
	}
	if d.Status == "cancelled" {
		send(c, req.ID, nil, makeErr(paymeOrderNotPayable, "Order cancelled"))
		return
	}
	// Boshqa payme_id bilan band bo'lsa
	if d.PaymeID != nil && *d.PaymeID != "" && *d.PaymeID != p.ID {
		send(c, req.ID, nil, makeErr(paymeCannotPerform, "Order locked by another transaction"))
		return
	}

	_, err := h.pool.Exec(ctx, `
        UPDATE donations SET payme_id=$1, payme_state=$2, payme_create_time=$3, updated_at=NOW()
        WHERE id=$4`, p.ID, paymeStateCreated, p.Time, d.ID)
	if err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	send(c, req.ID, gin.H{
		"create_time": p.Time,
		"transaction": strconv.FormatInt(d.ID, 10),
		"state":       paymeStateCreated,
	}, nil)
}

// ---- 3. PerformTransaction ----
func (h *PaymeHandler) performTx(c *gin.Context, ctx context.Context, req rpcReq) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		send(c, req.ID, nil, makeErr(paymeInvalidJSON, "invalid params"))
		return
	}
	d, eErr := h.loadDonationByPaymeID(ctx, p.ID)
	if eErr != nil {
		send(c, req.ID, nil, eErr)
		return
	}
	// Idempotent: allaqachon paid
	if d.PaymeState != nil && *d.PaymeState == paymeStatePaid {
		pt := int64(0)
		if d.PaymePerformTime != nil {
			pt = *d.PaymePerformTime
		}
		send(c, req.ID, gin.H{
			"perform_time": pt,
			"transaction":  strconv.FormatInt(d.ID, 10),
			"state":        paymeStatePaid,
		}, nil)
		return
	}
	if d.PaymeState == nil || *d.PaymeState != paymeStateCreated {
		send(c, req.ID, nil, makeErr(paymeCannotPerform, "Cannot perform"))
		return
	}

	nowMs := time.Now().UnixMilli()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	defer tx.Rollback(ctx)

	// Atomik: faqat 'created' holatdan 'paid' ga o'tkazamiz. Bir vaqtda kelgan
	// ikkita Perform'da pul ikki marta hisoblanmaydi (race-safe).
	tag, err := tx.Exec(ctx, `
        UPDATE donations SET status='paid', payme_state=$1, payme_perform_time=$2, paid_at=NOW(), updated_at=NOW()
        WHERE id=$3 AND payme_state=$4`, paymeStatePaid, nowMs, d.ID, paymeStateCreated)
	if err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}

	if tag.RowsAffected() == 0 {
		// Boshqa so'rov allaqachon bajargan — joriy holatni qaytaramiz (idempotent)
		_ = tx.Rollback(ctx)
		fresh, eErr := h.loadDonationByPaymeID(ctx, p.ID)
		if eErr != nil {
			send(c, req.ID, nil, eErr)
			return
		}
		pt := nowMs
		if fresh.PaymePerformTime != nil {
			pt = *fresh.PaymePerformTime
		}
		send(c, req.ID, gin.H{
			"perform_time": pt,
			"transaction":  strconv.FormatInt(fresh.ID, 10),
			"state":        paymeStatePaid,
		}, nil)
		return
	}

	amountSom := d.AmountTiyin / 100
	if _, err := tx.Exec(ctx,
		`UPDATE people SET raised = raised + $1, donors = donors + 1, updated_at=NOW() WHERE id=$2`,
		amountSom, d.PersonID); err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	if err := tx.Commit(ctx); err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	audit(h.pool, "payme", "donation_paid", strconv.FormatInt(d.ID, 10),
		gin.H{"person_id": d.PersonID, "amount_som": amountSom, "payme_id": p.ID})
	send(c, req.ID, gin.H{
		"perform_time": nowMs,
		"transaction":  strconv.FormatInt(d.ID, 10),
		"state":        paymeStatePaid,
	}, nil)
}

// ---- 4. CancelTransaction ----
func (h *PaymeHandler) cancelTx(c *gin.Context, ctx context.Context, req rpcReq) {
	var p struct {
		ID     string `json:"id"`
		Reason int16  `json:"reason"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		send(c, req.ID, nil, makeErr(paymeInvalidJSON, "invalid params"))
		return
	}
	d, eErr := h.loadDonationByPaymeID(ctx, p.ID)
	if eErr != nil {
		send(c, req.ID, nil, eErr)
		return
	}

	nowMs := time.Now().UnixMilli()
	curState := paymeStateCreated
	if d.PaymeState != nil {
		curState = *d.PaymeState
	}

	// Idempotent — allaqachon bekor qilingan / qaytarilgan
	if curState == paymeStateCancelled || curState == paymeStateRefunded {
		ct := int64(0)
		if d.PaymeCancelTime != nil {
			ct = *d.PaymeCancelTime
		}
		send(c, req.ID, gin.H{
			"cancel_time": ct,
			"transaction": strconv.FormatInt(d.ID, 10),
			"state":       curState,
		}, nil)
		return
	}

	newState := paymeStateCancelled
	cancelPaid := false
	if curState == paymeStatePaid {
		// Refund — to'langan tranzaksiyani qaytarish
		newState = paymeStateRefunded
		cancelPaid = true
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	defer tx.Rollback(ctx)

	// Atomik: faqat o'qigan holatimizdan o'tkazamiz. Bir vaqtda kelgan ikkita
	// Cancel'da refund (raised kamaytirish) ikki marta bajarilmaydi (race-safe).
	tag, err := tx.Exec(ctx, `
        UPDATE donations SET status='cancelled', payme_state=$1, payme_cancel_time=$2, payme_reason=$3, updated_at=NOW()
        WHERE id=$4 AND payme_state=$5`, newState, nowMs, p.Reason, d.ID, curState)
	if err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	if tag.RowsAffected() == 0 {
		// Boshqa so'rov allaqachon bekor qilgan — joriy holatni qaytaramiz
		_ = tx.Rollback(ctx)
		fresh, eErr := h.loadDonationByPaymeID(ctx, p.ID)
		if eErr != nil {
			send(c, req.ID, nil, eErr)
			return
		}
		ct, st := nowMs, newState
		if fresh.PaymeCancelTime != nil {
			ct = *fresh.PaymeCancelTime
		}
		if fresh.PaymeState != nil {
			st = *fresh.PaymeState
		}
		send(c, req.ID, gin.H{
			"cancel_time": ct,
			"transaction": strconv.FormatInt(fresh.ID, 10),
			"state":       st,
		}, nil)
		return
	}

	// Refund bo'lsa, person.raised kamaytirish (faqat shu tx muvaffaqiyatli o'tganda)
	if cancelPaid {
		amountSom := d.AmountTiyin / 100
		if _, err := tx.Exec(ctx, `
            UPDATE people SET raised = GREATEST(0, raised - $1), donors = GREATEST(0, donors - 1), updated_at=NOW()
            WHERE id=$2`, amountSom, d.PersonID); err != nil {
			send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	if cancelPaid {
		audit(h.pool, "payme", "donation_refunded", strconv.FormatInt(d.ID, 10),
			gin.H{"person_id": d.PersonID, "amount_som": d.AmountTiyin / 100, "reason": p.Reason})
	}
	send(c, req.ID, gin.H{
		"cancel_time": nowMs,
		"transaction": strconv.FormatInt(d.ID, 10),
		"state":       newState,
	}, nil)
}

// ---- 5. CheckTransaction ----
func (h *PaymeHandler) checkTx(c *gin.Context, ctx context.Context, req rpcReq) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		send(c, req.ID, nil, makeErr(paymeInvalidJSON, "invalid params"))
		return
	}
	d, eErr := h.loadDonationByPaymeID(ctx, p.ID)
	if eErr != nil {
		send(c, req.ID, nil, eErr)
		return
	}
	ct, pt, ct2 := int64(0), int64(0), int64(0)
	if d.PaymeCreateTime != nil {
		ct = *d.PaymeCreateTime
	}
	if d.PaymePerformTime != nil {
		pt = *d.PaymePerformTime
	}
	if d.PaymeCancelTime != nil {
		ct2 = *d.PaymeCancelTime
	}
	state := int16(0)
	if d.PaymeState != nil {
		state = *d.PaymeState
	}
	var reason any
	if d.PaymeReason != nil {
		reason = *d.PaymeReason
	}
	send(c, req.ID, gin.H{
		"create_time":  ct,
		"perform_time": pt,
		"cancel_time":  ct2,
		"transaction":  strconv.FormatInt(d.ID, 10),
		"state":        state,
		"reason":       reason,
	}, nil)
}

// ---- 6. GetStatement ----
func (h *PaymeHandler) getStatement(c *gin.Context, ctx context.Context, req rpcReq) {
	var p struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	}
	if err := json.Unmarshal(req.Params, &p); err != nil {
		send(c, req.ID, nil, makeErr(paymeInvalidJSON, "invalid params"))
		return
	}
	rows, err := h.pool.Query(ctx, `
        SELECT id, person_id, amount_tiyin, payme_id, payme_state, payme_create_time, payme_perform_time,
               payme_cancel_time, payme_reason
        FROM donations
        WHERE provider='payme' AND payme_create_time BETWEEN $1 AND $2
        ORDER BY payme_create_time ASC`, p.From, p.To)
	if err != nil {
		send(c, req.ID, nil, makeErr(paymeInternal, "db error"))
		return
	}
	defer rows.Close()

	items := []gin.H{}
	for rows.Next() {
		var id, personID, amount int64
		var paymeID *string
		var state *int16
		var ct, pt, cancelT *int64
		var reason *int16
		if err := rows.Scan(&id, &personID, &amount, &paymeID, &state, &ct, &pt, &cancelT, &reason); err != nil {
			continue
		}
		pid := ""
		if paymeID != nil {
			pid = *paymeID
		}
		stateV := int16(0)
		if state != nil {
			stateV = *state
		}
		createT, performT, cancT := int64(0), int64(0), int64(0)
		if ct != nil {
			createT = *ct
		}
		if pt != nil {
			performT = *pt
		}
		if cancelT != nil {
			cancT = *cancelT
		}
		var reasonV any
		if reason != nil {
			reasonV = *reason
		}
		items = append(items, gin.H{
			"id":           pid,
			"time":         createT,
			"amount":       amount,
			"account":      gin.H{"order_id": strconv.FormatInt(id, 10)},
			"create_time":  createT,
			"perform_time": performT,
			"cancel_time":  cancT,
			"transaction":  strconv.FormatInt(id, 10),
			"state":        stateV,
			"reason":       reasonV,
		})
	}
	send(c, req.ID, gin.H{"transactions": items}, nil)
}
