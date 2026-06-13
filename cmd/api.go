package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/jackc/pgx/v5/pgxpool"
)

func (app *application) run(h http.Handler) error {
	app.logger.Info("Server is running and listening at", "addr", app.config.Addr)
	return http.ListenAndServe(app.config.Addr, h)
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello world"))
		if err != nil {
			app.logger.Error("failed to write response", "error", err)
		}
	})

	return r
}

type application struct {
	config    config
	db        *pgxpool.Pool
	validator *validator.Validate
	logger    *slog.Logger
}

type config struct {
	Addr string
	db   dbConfig
}

type dbConfig struct {
	DSN string
}
