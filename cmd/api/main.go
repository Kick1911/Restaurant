package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    chimiddleware "github.com/go-chi/chi/v5/middleware"
    "github.com/jmoiron/sqlx"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/redis/go-redis/v9"

    "github.com/kick/sigma-connected/internal/config"
    "github.com/kick/sigma-connected/internal/handler"
    "github.com/kick/sigma-connected/internal/middleware"
    "github.com/kick/sigma-connected/internal/repository"
    "github.com/kick/sigma-connected/internal/service"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
    slog.SetDefault(logger)

    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    db, err := sqlx.Connect("pgx", cfg.Database.URL)
    if err != nil {
        slog.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    rdb := redis.NewClient(&redis.Options{
        Addr: cfg.Redis.URL,
    })

    bfProtector := middleware.NewBruteForceProtector(rdb)

    tenantRepo := repository.NewTenantRepository(db)
    userRepo := repository.NewUserRepository(db)
    dishRepo := repository.NewDishRepository(db)
    ratingRepo := repository.NewRatingRepository(db)

    userService := service.NewUserService(userRepo, tenantRepo, bfProtector, cfg.JWT.Secret, cfg.JWT.Expiration, db)
    dishService := service.NewDishService(dishRepo)
    ratingService := service.NewRatingService(ratingRepo)

    userHandler := handler.NewUserHandler(userService)
    dishHandler := handler.NewDishHandler(dishService)
    ratingHandler := handler.NewRatingHandler(ratingService)

    r := chi.NewRouter()

    r.Use(chimiddleware.RequestID)
    r.Use(chimiddleware.RealIP)
    r.Use(middleware.Logging(logger))
    r.Use(chimiddleware.Recoverer)
    r.Use(chimiddleware.Timeout(30 * time.Second))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    r.Handle("/metrics", promhttp.Handler())

    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/register", userHandler.Register)
        r.Post("/login", userHandler.Login)

        r.Group(func(r chi.Router) {
            r.Use(middleware.Auth(cfg.JWT.Secret))
            r.Use(middleware.RateLimit)

            r.Get("/dishes", dishHandler.Search)
            r.Get("/dishes/{id}", dishHandler.GetByID)

            r.Route("/dishes/{id}/ratings", func(r chi.Router) {
                r.Get("/", ratingHandler.GetByDishID)
                r.Post("/", ratingHandler.Create)
            })

            r.Group(func(r chi.Router) {
                r.Use(middleware.RequireRole("admin"))

                r.Post("/dishes", dishHandler.Create)
                r.Put("/dishes/{id}", dishHandler.Update)
                r.Delete("/dishes/{id}", dishHandler.Delete)
            })
        })
    })

    srv := &http.Server{
        Addr:         cfg.Server.Addr(),
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        slog.Info("server starting", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "error", err)
            os.Exit(1)
        }
    }()

    <-ctx.Done()
    slog.Info("shutting down server")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        slog.Error("server shutdown error", "error", err)
        os.Exit(1)
    }

    slog.Info("server stopped")
}
