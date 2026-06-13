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
	Logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			Logger.Warn("Failed to load env", "error", err)
		}
	}

	cfg := config{
		Addr: ":" + os.Getenv("PORT"),
		db: dbConfig{
			DSN: os.Getenv("DATABASE_URL"),
		},
	}


	conn, err := pgxpool.New(context.Background(), cfg.db.DSN)
	if err != nil {
		Logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	app := &application {
		config: cfg,
		db: conn,
		validator: validator.New(),
		logger: Logger,
	}

	if err := app.run(app.mount()); err != nil {
		Logger.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
	
}
