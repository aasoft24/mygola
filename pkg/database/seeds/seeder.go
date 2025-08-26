// pkg/database/seeds/seeder.go
package seeds

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Seeder struct {
	db *sql.DB
}

func NewSeeder(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) Run() error {
	log.Println("Seeding database...")

	// Run all seeders
	seeders := []func() error{
		s.UsersTableSeeder,
		s.PostsTableSeeder,
		// Add more seeders here
	}

	for _, seeder := range seeders {
		if err := seeder(); err != nil {
			return err
		}
	}

	log.Println("Database seeded successfully")
	return nil
}

func (s *Seeder) UsersTableSeeder() error {
	log.Println("Seeding users table...")

	// Insert sample users
	users := []struct {
		ID       string
		Name     string
		Email    string
		Password string
	}{
		{"1", "John Doe", "john@example.com", "password123"},
		{"2", "Jane Smith", "jane@example.com", "password123"},
		{"3", "Bob Johnson", "bob@example.com", "password123"},
	}

	for _, user := range users {
		query := `INSERT INTO users (id, name, email, password) VALUES (?, ?, ?, ?)`
		_, err := s.db.Exec(query, user.ID, user.Name, user.Email, user.Password)
		if err != nil {
			return fmt.Errorf("failed to seed users: %v", err)
		}
	}

	return nil
}

func (s *Seeder) PostsTableSeeder() error {
	log.Println("Seeding posts table...")

	// Insert sample posts
	posts := []struct {
		ID      string
		Title   string
		Content string
		UserID  string
	}{
		{"1", "First Post", "This is the first post", "1"},
		{"2", "Second Post", "This is the second post", "1"},
		{"3", "Third Post", "This is the third post", "2"},
	}

	for _, post := range posts {
		query := `INSERT INTO posts (id, title, content, user_id) VALUES (?, ?, ?, ?)`
		_, err := s.db.Exec(query, post.ID, post.Title, post.Content, post.UserID)
		if err != nil {
			return fmt.Errorf("failed to seed posts: %v", err)
		}
	}

	return nil
}

// Add more seeders as needed

// Helper function to generate random data
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
