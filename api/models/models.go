package models

import "time"

type Transaction struct {
	ID       int       `db:"id" json:"id"`
	Date     time.Time `db:"date" json:"date"`
	Amount   float64   `db:"amount" json:"amount"`
	Category string    `db:"category" json:"category"`
	Type     string    `db:"transaction_type" json:"type"`
	Note     string    `db:"note" json:"note"`
	ImageURL string    `db:"image_url" json:"image_url"`
}

