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
	"vendor-guard/vendors"
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
	authMiddleware := appMiddleware.RequireAuth(app.config.jwtSecret)

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

	r.Route("/api/organizations", func(r chi.Router) {
		r.With(authMiddleware).Post("/", orgHandler.CreateOrganization)
		r.With(authMiddleware).Get("/me", orgHandler.GetOrganizationByUserID)

		r.Get("/", orgHandler.GetAllOrganizations)
		r.Get("/{id}", orgHandler.GetOrganizationById)
		r.Put("/{id}", orgHandler.UpdateOrganization)
		r.Delete("/{id}", orgHandler.DeleteOrganization)
	})

	// Vendor routes
	vendorService := vendors.NewService(repoQueries)
	vendorHandler := vendors.NewVendorHandler(vendorService, app.validator)

	vendorInviteService := vendors.NewInviteService(repoQueries, app.config.jwtSecret)
	vendorInviteHandler := vendors.NewInviteHandler(vendorInviteService, app.validator)

	r.Route("/api/vendors", func(r chi.Router) {
		r.With(authMiddleware).Post("/", vendorHandler.CreateVendor)
		r.With(authMiddleware).Get("/", vendorHandler.GetAllVendors)
		r.With(authMiddleware).Get("/{id}", vendorHandler.GetVendorByID)
		r.With(authMiddleware).Put("/{id}", vendorHandler.UpdateVendor)
		r.With(authMiddleware).Delete("/{id}", vendorHandler.DeleteVendor)

		// Vendor invitation routes
		r.With(authMiddleware).Post("/{vendorId}/invite", vendorInviteHandler.InviteVendorUser)
		r.With(authMiddleware).Get("/{vendorId}/invite", vendorInviteHandler.GetVendorInvitations)
	})

	// Public vendor invitation acceptance
	r.Post("/api/vendors/invite/accept", vendorInviteHandler.AcceptInvitation)

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
