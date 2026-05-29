package main

import "time"

type Admin struct {
	ID           int64     `db:"id"           json:"id"`
	Email        string    `db:"email"        json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	FullName     string    `db:"full_name"    json:"full_name"`
	Role         string    `db:"role"         json:"role"`
	CreatedAt    time.Time `db:"created_at"   json:"created_at"`
}

type Person struct {
	ID               int64     `db:"id"                json:"id"`
	Name             string    `db:"name"              json:"name"`
	Age              int       `db:"age"               json:"age"`
	Region           string    `db:"region"            json:"region"`
	Diagnosis        string    `db:"diagnosis"         json:"diagnosis"`
	Help             string    `db:"help"              json:"help"`
	Facility         string    `db:"facility"          json:"facility"`
	FacilityVerified bool      `db:"facility_verified" json:"facility_verified"`
	Org              string    `db:"org"               json:"org"`
	Story            string    `db:"story"             json:"story"`
	PhotoURL         string    `db:"photo_url"         json:"photo_url"`
	Target           int64     `db:"target"            json:"target"`
	Raised           int64     `db:"raised"            json:"raised"`
	Donors           int       `db:"donors"            json:"donors"`
	DaysLeft         int       `db:"days_left"         json:"days_left"`
	Urgent           bool      `db:"urgent"            json:"urgent"`
	Category         string    `db:"category"          json:"category"`
	AuthorName       string    `db:"author_name"       json:"author_name"`
	AuthorRole       string    `db:"author_role"       json:"author_role"`
	Status           string    `db:"status"            json:"status"`
	CreatedAt        time.Time `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"        json:"updated_at"`
}

type Donation struct {
	ID          int64  `db:"id"           json:"id"`
	PersonID    int64  `db:"person_id"    json:"person_id"`
	Provider    string `db:"provider"     json:"provider"`
	AmountTiyin int64  `db:"amount_tiyin" json:"amount_tiyin"`
	Status      string `db:"status"       json:"status"`
	Anonim      bool   `db:"anonim"       json:"anonim"`
	DonorName   string `db:"donor_name"   json:"donor_name"`
	DonorPhone  string `db:"donor_phone"  json:"donor_phone"`

	ClickTransID      *string `db:"click_trans_id"      json:"click_trans_id,omitempty"`
	ClickPaydocID     *string `db:"click_paydoc_id"     json:"click_paydoc_id,omitempty"`
	MerchantPrepareID *int64  `db:"merchant_prepare_id" json:"merchant_prepare_id,omitempty"`

	PaymeID          *string `db:"payme_id"           json:"payme_id,omitempty"`
	PaymeState       *int16  `db:"payme_state"        json:"payme_state,omitempty"`
	PaymeCreateTime  *int64  `db:"payme_create_time"  json:"payme_create_time,omitempty"`
	PaymePerformTime *int64  `db:"payme_perform_time" json:"payme_perform_time,omitempty"`
	PaymeCancelTime  *int64  `db:"payme_cancel_time"  json:"payme_cancel_time,omitempty"`
	PaymeReason      *int16  `db:"payme_reason"       json:"payme_reason,omitempty"`

	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	PaidAt    *time.Time `db:"paid_at"    json:"paid_at,omitempty"`
}
