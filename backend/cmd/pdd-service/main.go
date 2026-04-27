package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	coreauth "pdd-service/internal/core/auth"
	pgPool "pdd-service/internal/core/repository/postgres/pool"
	httpserver "pdd-service/internal/core/transport/http/server"

	authApp "pdd-service/internal/features/auth/application"
	authRepo "pdd-service/internal/features/auth/infrastructure/postgres"
	authHTTP "pdd-service/internal/features/auth/presentation/http"
	payoutsApp "pdd-service/internal/features/payouts/application"
	payoutsRepo "pdd-service/internal/features/payouts/infrastructure/postgres"
	payoutsHTTP "pdd-service/internal/features/payouts/presentation/http"
	reportsApp "pdd-service/internal/features/reports/application"
	reportsRepo "pdd-service/internal/features/reports/infrastructure/postgres"
	reportsHTTP "pdd-service/internal/features/reports/presentation/http"
	usersApp "pdd-service/internal/features/users/application"
	usersRepo "pdd-service/internal/features/users/infrastructure/postgres"
	usersHTTP "pdd-service/internal/features/users/presentation/http"
	violationTypesApp "pdd-service/internal/features/violation_types/application"
	violationTypesRepo "pdd-service/internal/features/violation_types/infrastructure/postgres"
	violationTypesHTTP "pdd-service/internal/features/violation_types/presentation/http"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer cancel()

	dbCfg := pgPool.LoadConfig()
	srvCfg := httpserver.LoadConfig()
	jwtCfg := coreauth.LoadConfig()

	pool, err := pgPool.NewConnectionPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("failed to init postgres pool: %v", err)
	}
	defer pool.Close()

	tokenManager, err := coreauth.NewTokenManager(jwtCfg)
	if err != nil {
		log.Fatalf("failed to init token manager: %v", err)
	}

	authRepository := authRepo.NewRepository(pool)
	authService := authApp.NewAuthService(authRepository, authRepository, tokenManager)
	authHandler := authHTTP.NewAuthHTTPHandler(authService)

	usersRepository := usersRepo.NewRepository(pool)
	usersService := usersApp.NewService(usersRepository)
	usersHandler := usersHTTP.NewUsersHTTPHandler(usersService)

	violationTypesRepository := violationTypesRepo.NewRepository(pool)
	violationTypesService := violationTypesApp.NewService(violationTypesRepository)
	violationTypesHandler := violationTypesHTTP.NewViolationTypesHTTPHandler(violationTypesService)

	reportsRepository := reportsRepo.NewRepository(pool)
	reportsService := reportsApp.NewService(reportsRepository, reportsRepository)
	reportsHandler := reportsHTTP.NewReportsHTTPHandler(reportsService)

	payoutsRepository := payoutsRepo.NewRepository(pool)
	payoutsService := payoutsApp.NewService(payoutsRepository, payoutsRepository, payoutsRepository, payoutsRepository)
	payoutsHandler := payoutsHTTP.NewPayoutsHTTPHandler(payoutsService)

	srv := httpserver.NewHTTPServer(
		srvCfg,
		tokenManager,
		authHandler,
		usersHandler,
		violationTypesHandler,
		reportsHandler,
		payoutsHandler,
	)

	log.Printf("starting http server on %s", srvCfg.Addr())
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server fatal error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), srvCfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Printf("failed to stop server: %v", err)
	}

	log.Println("server stopped correctly")
}
