package app

import (
	"context"
	_ "embed"
	"log/slog"
	"merch/internal/api/handlers/auth"
	"merch/internal/api/handlers/merch"
	"merch/internal/api/utilapi"
	"merch/internal/cache"
	"merch/internal/config"
	"merch/internal/entity"
	"merch/internal/usecase"
	"merch/internal/usecase/repo/postgres"
	shoprepo "merch/internal/usecase/repo/shop"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	cfg := config.ConfigLoad()
	log := setupLogger()

	sharMap := make(map[int]string, len(cfg.DBPath))

	for sharNumber, db := range cfg.DBPath {
		sharMap[sharNumber] = db
	}

	log.Info("starting server")

	repo := postgres.NewRepo(sharMap, cfg.MaxConn, cfg.MinConn, cfg.ShardNumber, cfg.LifeConn)
	defer repo.CloseDB()

	// For generate UserID type int
	entity.InitSonyflake()

	log.Info("successful connected to db")

	mapStore := shoprepo.NewMapStore()

	c := cache.NewCache[string](cfg.CacheTTL)

	merchUseCase := usecase.NewStorage(repo, c, mapStore)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.Issuer, cfg.TokenTTL)

	handler := merch.NewMerchHandler(merchUseCase, jwtManager)

	r := utilapi.NewRouter(log, cfg.SLI)
	r.Handle("/api/sendCoin", http.MethodPost, handler.TokenMiddleware, handler.SendCoin)
	r.Handle("/api/info", http.MethodGet, handler.TokenMiddleware, handler.Info)
	r.Handle("/api/buy/", http.MethodGet, handler.TokenMiddleware, handler.Buy)
	r.Handle("/api/auth", http.MethodPost, handler.Login)

	srv := &http.Server{
		Addr:         cfg.ADDR,
		Handler:      r,
		ReadTimeout:  cfg.SLI * 10,
		WriteTimeout: cfg.SLI * 10,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Error("failed to listen and server", slog.Any("error", err))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", slog.Any("error", err))

		return
	}

	log.Info("server stopped")
}

func setupLogger() *slog.Logger {
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	return log
}
