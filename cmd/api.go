package main

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"preuvio/auth"
	"preuvio/form_fields"
	"preuvio/form_sections"
	"preuvio/form_submissions"
	"preuvio/form_templates"
	"preuvio/forms"
	"preuvio/internal/repo"
	appMiddleware "preuvio/middleware"
	"preuvio/organizations"
	"preuvio/partners"
	"preuvio/users"
)

func (app *application) run(h http.Handler) error {
	app.logger.Info("Server is running and listening at", "addr", app.config.Addr)
	app.logger.Info("Documentation at http://localhost:8000/api/docs/index.html")
	return http.ListenAndServe(app.config.Addr, h)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(corsMiddleware)

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

	// Partner routes
	partnerService := partners.NewService(repoQueries)
	partnerHandler := partners.NewPartnerHandler(partnerService, app.validator)

	partnerInviteService := partners.NewInviteService(repoQueries, app.config.jwtSecret)
	partnerInviteHandler := partners.NewInviteHandler(partnerInviteService, app.validator)

	r.Route("/api/partners", func(r chi.Router) {
		r.With(authMiddleware).Post("/", partnerHandler.CreatePartner)
		r.With(authMiddleware).Get("/", partnerHandler.GetAllPartners)
		r.With(authMiddleware).Get("/{id}", partnerHandler.GetPartnerByID)
		r.With(authMiddleware).Put("/{id}", partnerHandler.UpdatePartner)
		r.With(authMiddleware).Delete("/{id}", partnerHandler.DeletePartner)

		// Partner invitation routes
		r.With(authMiddleware).Post("/{partnerId}/invite", partnerInviteHandler.InvitePartnerUser)
		r.With(authMiddleware).Get("/{partnerId}/invite", partnerInviteHandler.GetPartnerInvitations)
	})

	// Public partner invitation acceptance
	r.Post("/api/partners/invite/accept", partnerInviteHandler.AcceptInvitation)

	// Form template routes
	formTemplateService := form_templates.NewServiceWithPool(app.db, repoQueries)
	formTemplateHandler := form_templates.NewHandler(formTemplateService, app.validator)
	templateFieldService := form_templates.NewTemplateFieldService(repoQueries)
	templateFieldHandler := form_templates.NewTemplateFieldHandler(templateFieldService, app.validator)

	sectionService := form_sections.NewService(repoQueries)
	sectionHandler := form_sections.NewHandler(sectionService, app.validator)

	r.Route("/api/templates", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", formTemplateHandler.CreateTemplate)
		r.Get("/", formTemplateHandler.GetAllTemplates)
		r.Get("/{id}", formTemplateHandler.GetTemplateByID)
		r.Get("/{id}/detail", formTemplateHandler.GetTemplateDetail)
		r.Put("/{id}", formTemplateHandler.UpdateTemplate)
		r.Delete("/{id}", formTemplateHandler.DeleteTemplate)
		r.Post("/{id}/clone", formTemplateHandler.CloneTemplateToForm)

		// Template section routes
		r.Post("/{id}/sections", sectionHandler.CreateTemplateSection)
		r.Get("/{id}/sections", sectionHandler.GetTemplateSections)

		// Template field routes
		r.Post("/{id}/fields", templateFieldHandler.CreateField)
		r.Get("/{id}/fields", templateFieldHandler.GetFieldsByTemplateID)
		r.Put("/fields/{fieldId}", templateFieldHandler.UpdateField)
		r.Delete("/fields/{fieldId}", templateFieldHandler.DeleteField)
	})

	// Form routes - unified services (scratch creation via forms, clone via templates)
	formService := forms.NewService(repoQueries)
	formHandler := forms.NewHandler(formService, app.validator)

	// Unified field service - single implementation for both form and template fields (same table)
	formFieldService := form_fields.NewService(repoQueries)
	formFieldHandler := form_fields.NewHandler(formFieldService, app.validator)

	// Unified submission service - single instance reused for form and review routes
	formSubmissionService := form_submissions.NewService(repoQueries)
	formSubmissionHandler := form_submissions.NewHandler(formSubmissionService, app.validator)

	r.Route("/api/forms", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", formHandler.CreateForm) // scratch only; for template use POST /templates/{id}/clone
		r.Get("/", formHandler.GetFormsByOrg)
		r.Get("/{formId}", formHandler.GetFormByID)
		r.Get("/{formId}/detail", formHandler.GetFormDetail)
		r.Put("/{formId}", formHandler.UpdateForm)
		r.Delete("/{formId}", formHandler.DeleteForm)

		// Form section routes
		r.Post("/{formId}/sections", sectionHandler.CreateFormSection)
		r.Get("/{formId}/sections", sectionHandler.GetFormSections)

		// Form field routes
		r.Post("/{formId}/fields", formFieldHandler.CreateField)
		r.Get("/{formId}/fields", formFieldHandler.GetFieldsByFormID)
		r.Put("/{formId}/fields/{fieldId}", formFieldHandler.UpdateField)
		r.Delete("/{formId}/fields/{fieldId}", formFieldHandler.DeleteField)

		// Form submission routes
		r.Post("/{formId}/submissions", formSubmissionHandler.CreateSubmission)
		r.Get("/{formId}/submissions", formSubmissionHandler.GetSubmissionsByFormID)
	})

	// Generic section routes (update/delete both form and template sections)
	r.Route("/api/sections", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Put("/{sectionId}", sectionHandler.UpdateSection)
		r.Delete("/{sectionId}", sectionHandler.DeleteSection)
	})

	// Submission review routes (outside /api/forms for cleaner URLs) - reused service
	r.Route("/api/submissions", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/{id}", formSubmissionHandler.GetSubmissionByID)
		r.Put("/{id}/review", formSubmissionHandler.ReviewSubmission)
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
