package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	redisstore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Vasek18/my_hours/internal/config"
	"github.com/Vasek18/my_hours/internal/db"
	"github.com/Vasek18/my_hours/internal/mailer"
	"github.com/Vasek18/my_hours/internal/server"
)

// Version is set at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	cfg := config.Load()
	log.Printf("my_hours version %s", Version)

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	pool, err := connectDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	store, err := newSessionStore(cfg)
	if err != nil {
		log.Fatalf("session store: %v", err)
	}

	srv := server.New(cfg, pool, mailer.ConsoleMailer{})
	router := srv.Router(store)

	addr := ":" + cfg.AppPort
	log.Printf("listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// connectDB opens a pgx pool, retrying briefly so the API can start alongside a
// Postgres container that is still coming up.
func connectDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for i := 0; i < 15; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		lastErr = pool.Ping(pingCtx)
		cancel()
		if lastErr == nil {
			return pool, nil
		}
		log.Printf("waiting for database... (%v)", lastErr)
		time.Sleep(time.Second)
	}
	pool.Close()
	return nil, lastErr
}

func newSessionStore(cfg config.Config) (sessions.Store, error) {
	store, err := redisstore.NewStore(10, "tcp", cfg.RedisAddr, "", "", []byte(cfg.SessionSecret))
	if err != nil {
		return nil, err
	}
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
	return store, nil
}
