package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	_ "vendor-guard/docs"
)

// @title Vendor Guard API
// @version 1.0
// @description This is a sample server for Vendor Guard.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /
func main() {
	Logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			Logger.Warn("Failed to load env", "error", err)
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecret_default_key"
	}

	cfg := config{
		Addr: ":" + os.Getenv("PORT"),
		db: dbConfig{
			DSN: os.Getenv("DATABASE_URL"),
		},
		jwtSecret: jwtSecret,
	}

	conn, err := pgxpool.New(context.Background(), cfg.db.DSN)
	if err != nil {
		Logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	app := &application{
		config:    cfg,
		db:        conn,
		validator: validator.New(),
		logger:    Logger,
	}

	if err := app.run(app.mount()); err != nil {
		Logger.Error("Failed to start application", "error", err)
		os.Exit(1)
	}

}
