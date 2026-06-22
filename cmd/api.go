package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"vendor-guard/auth"
	"vendor-guard/internal/repo"
	appMiddleware "vendor-guard/middleware"
	"vendor-guard/organizations"
	"vendor-guard/users"
)

func (app *application) run(h http.Handler) error {
	app.logger.Info("Server is running and listening at", "addr", app.config.Addr)
	app.logger.Info("Documentation at http://localhost:8000/api/docs/index.html")
	return http.ListenAndServe(app.config.Addr, h)
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/api/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/api/docs/doc.json"),
	))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello world"))
		if err != nil {
			app.logger.Error("failed to write response", "error", err)
		}
	})

	repoQueries := repo.New(app.db)

	// Auth routes
	authService := auth.NewService(repoQueries, app.config.jwtSecret)
	authHandler := auth.NewHandler(authService, app.validator)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	// User routes
	userService := users.NewService(repoQueries)
	userHandler := users.NewHandler(userService, app.validator)

	r.Route("/api/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/", userHandler.GetAllUsers)
		r.Get("/{id}", userHandler.GetUser)
		r.Put("/{id}", userHandler.UpdateUser)
		r.Delete("/{id}", userHandler.DeleteUser)
	})

	// Organization routes
	orgService := organizations.NewOrganizationService(repoQueries)
	orgHandler := organizations.NewOrganizationHandler(orgService, app.validator)
	authMiddleware := appMiddleware.RequireAuth(app.config.jwtSecret)

	r.Route("/api/organizations", func(r chi.Router) {
		// Protected: must be signed in
		r.With(authMiddleware).Post("/", orgHandler.CreateOrganization)
		r.With(authMiddleware).Get("/me", orgHandler.GetOrganizationByUserID)

		// Public
		r.Get("/", orgHandler.GetAllOrganizations)
		r.Get("/{id}", orgHandler.GetOrganizationById)
		r.Put("/{id}", orgHandler.UpdateOrganization)
		r.Delete("/{id}", orgHandler.DeleteOrganization)
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
	Addr      string
	jwtSecret string
	db        dbConfig
}

type dbConfig struct {
	DSN string
}
