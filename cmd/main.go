package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if _, err := os.Stat(".env"); err == nil {
		godotenv.Load()
	}

	cfg := config{
		Addr: ":" + os.Getenv("PORT"),
		db: dbConfig{
			DSN: os.Getenv("DATABASE_URL"),
		},
	}


	conn, err := pgxpool.New(context.Background(), cfg.db.DSN)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	app := &application {
		config: cfg,
		db: conn,
		validator: validator.New(),
	}

	if err := app.run(app.mount()); err != nil {
		logger.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
	
}
