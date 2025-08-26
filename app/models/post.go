// internal/app/models/post.go
package models

import (
	"time"
)

type Post struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (p *Post) TableName() string {
	return "posts"
}
