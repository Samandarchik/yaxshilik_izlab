package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PeopleHandler struct {
	pool *pgxpool.Pool
}

func NewPeopleHandler(pool *pgxpool.Pool) *PeopleHandler {
	return &PeopleHandler{pool: pool}
}

const peopleCols = `id, name, age, region, diagnosis, help, facility, facility_verified, org, story, photo_url,
                    target, raised, donors, days_left, urgent, category, author_name, author_role, status,
                    created_at, updated_at`

func scanPerson(row pgx.Row, p *Person) error {
	return row.Scan(&p.ID, &p.Name, &p.Age, &p.Region, &p.Diagnosis, &p.Help, &p.Facility, &p.FacilityVerified,
		&p.Org, &p.Story, &p.PhotoURL, &p.Target, &p.Raised, &p.Donors, &p.DaysLeft, &p.Urgent, &p.Category,
		&p.AuthorName, &p.AuthorRole, &p.Status, &p.CreatedAt, &p.UpdatedAt)
}

// GET /api/people  (public) - faqat 'active' bemorlar
func (h *PeopleHandler) ListPublic(c *gin.Context) {
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT `+peopleCols+` FROM people WHERE status='active' ORDER BY urgent DESC, created_at DESC LIMIT 200`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()
	items := []Person{}
	for rows.Next() {
		var p Person
		if err := scanPerson(rows, &p); err == nil {
			items = append(items, p)
		}
	}
	c.JSON(http.StatusOK, items)
}

// GET /api/people/:id (public)
func (h *PeopleHandler) GetPublic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p Person
	err := scanPerson(h.pool.QueryRow(c.Request.Context(),
		`SELECT `+peopleCols+` FROM people WHERE id=$1 AND status='active'`, id), &p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GET /api/admin/people - barcha bemorlar (admin)
func (h *PeopleHandler) ListAdmin(c *gin.Context) {
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT `+peopleCols+` FROM people ORDER BY created_at DESC LIMIT 500`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()
	items := []Person{}
	for rows.Next() {
		var p Person
		if err := scanPerson(rows, &p); err == nil {
			items = append(items, p)
		}
	}
	c.JSON(http.StatusOK, items)
}

type personReq struct {
	Name             string `json:"name"`
	Age              int    `json:"age"`
	Region           string `json:"region"`
	Diagnosis        string `json:"diagnosis"`
	Help             string `json:"help"`
	Facility         string `json:"facility"`
	FacilityVerified bool   `json:"facility_verified"`
	Org              string `json:"org"`
	Story            string `json:"story"`
	PhotoURL         string `json:"photo_url"`
	Target           int64  `json:"target"`
	Raised           int64  `json:"raised"`
	Donors           int    `json:"donors"`
	DaysLeft         int    `json:"days_left"`
	Urgent           bool   `json:"urgent"`
	Category         string `json:"category"`
	AuthorName       string `json:"author_name"`
	AuthorRole       string `json:"author_role"`
	Status           string `json:"status"`
}

// POST /api/admin/people
func (h *PeopleHandler) Create(c *gin.Context) {
	var r personReq
	if err := c.ShouldBindJSON(&r); err != nil || r.Name == "" || r.Target <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name va target majburiy"})
		return
	}
	if r.Status == "" {
		r.Status = "active"
	}
	if r.DaysLeft == 0 {
		r.DaysLeft = 30
	}

	var id int64
	err := h.pool.QueryRow(c.Request.Context(), `
        INSERT INTO people(name, age, region, diagnosis, help, facility, facility_verified, org, story, photo_url,
                           target, raised, donors, days_left, urgent, category, author_name, author_role, status)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
        RETURNING id`,
		r.Name, r.Age, r.Region, r.Diagnosis, r.Help, r.Facility, r.FacilityVerified, r.Org, r.Story, r.PhotoURL,
		r.Target, r.Raised, r.Donors, r.DaysLeft, r.Urgent, r.Category, r.AuthorName, r.AuthorRole, r.Status,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// PUT /api/admin/people/:id
func (h *PeopleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r personReq
	if err := c.ShouldBindJSON(&r); err != nil || r.Name == "" || r.Target <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name va target majburiy"})
		return
	}
	if r.Status == "" {
		r.Status = "active"
	}
	tag, err := h.pool.Exec(c.Request.Context(), `
        UPDATE people SET
            name=$1, age=$2, region=$3, diagnosis=$4, help=$5, facility=$6, facility_verified=$7, org=$8,
            story=$9, photo_url=$10, target=$11, raised=$12, donors=$13, days_left=$14, urgent=$15,
            category=$16, author_name=$17, author_role=$18, status=$19, updated_at=NOW()
        WHERE id=$20`,
		r.Name, r.Age, r.Region, r.Diagnosis, r.Help, r.Facility, r.FacilityVerified, r.Org, r.Story, r.PhotoURL,
		r.Target, r.Raised, r.Donors, r.DaysLeft, r.Urgent, r.Category, r.AuthorName, r.AuthorRole, r.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/admin/people/:id
func (h *PeopleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tag, err := h.pool.Exec(c.Request.Context(), `DELETE FROM people WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
