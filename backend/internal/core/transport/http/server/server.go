package server

import (
	"context"
	"net/http"
	"time"

	"pdd-service/internal/core/domain/users"
	"pdd-service/internal/core/transport/http/middlewares"
	authHTTP "pdd-service/internal/features/auth/presentation/http"
	payoutsHTTP "pdd-service/internal/features/payouts/presentation/http"
	reportsHTTP "pdd-service/internal/features/reports/presentation/http"
	usersHTTP "pdd-service/internal/features/users/presentation/http"
	violationTypesHTTP "pdd-service/internal/features/violation_types/presentation/http"

	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	server *http.Server
	config Config
}

func NewHTTPServer(
	cfg Config,
	tokenParser middlewares.AccessTokenParser,
	authH *authHTTP.AuthHTTPHandler,
	usersH *usersHTTP.UsersHTTPHandler,
	violationTypesH *violationTypesHTTP.ViolationTypesHTTPHandler,
	reportsH *reportsHTTP.ReportsHTTPHandler,
	payoutsH *payoutsHTTP.PayoutsHTTPHandler,
) *HTTPServer {
	r := chi.NewRouter()

	r.Use(middlewares.RequestIDMiddleware)
	r.Use(middlewares.RecoverMiddleware)
	r.Use(middlewares.CORSMiddleware)

	r.Get("/_info", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
			r.Post("/logout", authH.Logout)

			r.Group(func(r chi.Router) {
				r.Use(middlewares.AuthMiddleware(tokenParser))
				r.Get("/me", authH.Me)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenParser))

			r.With(middlewares.RequireRoles(users.RoleModerator, users.RoleAdmin)).
				Get("/", usersH.List)
			r.Get("/{id}", usersH.GetByID)
			r.Get("/{id}/payouts", payoutsH.ListByUserID)
			r.With(middlewares.RequireRoles(users.RoleAdmin)).
				Patch("/{id}/role", usersH.UpdateRole)
		})

		r.Route("/violation-types", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenParser))

			r.Get("/", violationTypesH.List)
			r.Get("/{id}", violationTypesH.GetByID)

			r.Group(func(r chi.Router) {
				r.Use(middlewares.RequireRoles(users.RoleAdmin))
				r.Post("/", violationTypesH.Create)
				r.Patch("/{id}", violationTypesH.Update)
				r.Delete("/{id}", violationTypesH.Delete)
				r.Post("/{id}/activate", violationTypesH.Activate)
				r.Post("/{id}/deactivate", violationTypesH.Deactivate)
			})
		})

		r.Route("/reports", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenParser))

			r.Get("/", reportsH.List)
			r.Post("/", reportsH.Create)
			r.Get("/{id}", reportsH.GetByID)
			r.Patch("/{id}", reportsH.Update)
			r.Delete("/{id}", reportsH.Delete)
			r.Post("/{id}/submit", reportsH.Submit)

			r.Group(func(r chi.Router) {
				r.Use(middlewares.RequireRoles(users.RoleModerator, users.RoleAdmin))
				r.Post("/{id}/start-review", reportsH.StartReview)
				r.Post("/{id}/approve", reportsH.Approve)
				r.Post("/{id}/reject", reportsH.Reject)
				r.Post("/{id}/moderate", reportsH.Moderate)
			})
		})

		r.Route("/payouts", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenParser))

			r.With(middlewares.RequireRoles(users.RoleModerator, users.RoleAdmin)).
				Get("/", payoutsH.List)
			r.With(middlewares.RequireRoles(users.RoleAdmin)).
				Post("/from-report", payoutsH.CreateFromReport)
			r.Get("/{id}", payoutsH.GetByID)
			r.With(middlewares.RequireRoles(users.RoleAdmin)).
				Post("/{id}/mark-paid", payoutsH.MarkPaid)
			r.With(middlewares.RequireRoles(users.RoleAdmin)).
				Post("/{id}/mark-failed", payoutsH.MarkFailed)
		})

		r.Route("/payout-rules", func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(tokenParser))

			r.With(middlewares.RequireRoles(users.RoleModerator, users.RoleAdmin)).
				Get("/", payoutsH.ListRules)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.RequireRoles(users.RoleAdmin))
				r.Post("/", payoutsH.CreateRule)
				r.Patch("/{id}", payoutsH.UpdateRule)
				r.Post("/{id}/activate", payoutsH.ActivateRule)
				r.Post("/{id}/deactivate", payoutsH.DeactivateRule)
			})
		})
	})

	return &HTTPServer{
		config: cfg,
		server: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      r,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) GetHandler() http.Handler {
	return s.server.Handler
}
