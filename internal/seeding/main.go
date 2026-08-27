package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Phone     string
	RoleName  string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("warn: failed to load .env: %v", err)
		}
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("connected to database")

	// Ensure roles exist
	roleIDs := make(map[string]string)
	roles := []struct{ name, desc string }{
		{"owner", "Full access to the organization"},
		{"admin", "Can manage members and settings"},
		{"member", "Standard access"},
		{"manager", "Can manage resources across the platform"},
		{"viewer", "Read-only access across assigned domains"},
	}
	for _, r := range roles {
		var id string
		err := pool.QueryRow(ctx,
			`INSERT INTO roles (name, description) VALUES ($1, $2)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			r.name, r.desc,
		).Scan(&id)
		if err != nil {
			log.Fatalf("failed to ensure role %q: %v", r.name, err)
		}
		roleIDs[r.name] = id
		fmt.Printf("  role %-12s -> %s\n", r.name, id)
	}

	// Users from env — format: FirstName:LastName:email:password:phone:role (comma-separated)
	usersRaw := getEnv("SEED_USERS", "")
	defaultPassword := getEnv("SEED_DEFAULT_PASSWORD", "password123")

	var users []seedUser

	if usersRaw != "" {
		entries := strings.Split(usersRaw, ",")
		for _, entry := range entries {
			parts := strings.Split(strings.TrimSpace(entry), ":")
			if len(parts) < 6 {
				log.Fatalf("invalid SEED_USERS entry %q — expected FirstName:LastName:email:password:phone:role", entry)
			}
			pw := parts[3]
			if pw == "" {
				pw = defaultPassword
			}
			users = append(users, seedUser{
				FirstName: parts[0],
				LastName:  parts[1],
				Email:     parts[2],
				Password:  pw,
				Phone:     parts[4],
				RoleName:  parts[5],
			})
		}
	} else {
		// Default seed users if SEED_USERS is not set
		users = []seedUser{
			{FirstName: "John", LastName: "Owner", Email: "john@acme.com", Password: defaultPassword, Phone: "+1234567890", RoleName: "owner"},
			{FirstName: "Jane", LastName: "Admin", Email: "jane@acme.com", Password: defaultPassword, Phone: "+1234567891", RoleName: "admin"},
			{FirstName: "Bob", LastName: "Member", Email: "bob@acme.com", Password: defaultPassword, Phone: "+1234567892", RoleName: "member"},
		}
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("failed to hash password for %s: %v", u.Email, err)
		}

		roleID := roleIDs[u.RoleName]

		var userID string
		err = pool.QueryRow(ctx,
			`INSERT INTO users (first_name, last_name, email, password, phone, role_id)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (email) DO NOTHING
			 RETURNING id`,
			u.FirstName, u.LastName, u.Email, string(hash), u.Phone, roleID,
		).Scan(&userID)
		if err != nil {
			err2 := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, u.Email).Scan(&userID)
			if err2 != nil {
				log.Fatalf("failed to get user %s: %v", u.Email, err2)
			}
			fmt.Printf("  user    %-25s -> %s (already exists, skipped)\n", u.Email, userID)
			continue
		}
		fmt.Printf("  user    %-25s -> %s\n", u.Email, userID)
	}

	fmt.Println("\nseeding complete!")
}
