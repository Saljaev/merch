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
	"merch/internal/usecase/shop"
	shoprepo "merch/internal/usecase/shop/repo"
	"merch/internal/usecase/storage"
	"merch/internal/usecase/storage/repo/postgres"
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

	repo := postgres.NewRepo(sharMap)
	defer repo.CloseDB()

	for i := range cfg.ShardNumber {
		err := repo.InitEntity(i)
		if err != nil {
			log.Error("failed to up migrations", slog.Any("error", err))
		}
	}

	// For generate UserID type int
	entity.InitSonyflake()

	merchUseCase := storage.NewStorage(repo)

	log.Info("successful connected to db")

	mapStore := shoprepo.NewMapStore()
	itemShop := shop.NewShop(mapStore)

	c := cache.NewCache[string](cfg.CacheTTL)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.Issuer, cfg.TokenTTL)

	handler := merch.NewMerchHandler(merchUseCase, c, itemShop, jwtManager)

	r := utilapi.NewRouter(log, cfg.SLI)
	r.Handle("/api/sendCoin", http.MethodPost, handler.TokenMiddleware, handler.SendCoin)
	r.Handle("/api/info", http.MethodGet, handler.TokenMiddleware, handler.Info)
	r.Handle("/api/buy/", http.MethodGet, handler.TokenMiddleware, handler.Buy)
	r.Handle("/api/auth", http.MethodPost, handler.Login)

	srv := &http.Server{
		Addr:         cfg.ADDR,
		Handler:      r,
		ReadTimeout:  cfg.SLI,
		WriteTimeout: cfg.SLI,
	}

	go func() {
		srv.ListenAndServe()
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
	var log *slog.Logger

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	return log
}
