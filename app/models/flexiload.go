package models

import (
	"mygola/pkg/helpers"
	"time"
)

type Flexiload struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (m *Flexiload) TableName() string {
	return "flexiload"
}

func NewFlexiload() *Flexiload {
	return &Flexiload{
		ID:        helpers.GenerateID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Add your model methods here
