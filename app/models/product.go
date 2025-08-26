package models

import (
	"mygola/pkg/helpers"
	"time"
)

type Product struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (m *Product) TableName() string {
	return "product"
}

func NewProduct() *Product {
	return &Product{
		ID:        helpers.GenerateID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Add your model methods here
