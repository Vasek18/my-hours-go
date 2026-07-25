package server

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Vasek18/my_hours/internal/auth"
	"github.com/Vasek18/my_hours/internal/config"
	"github.com/Vasek18/my_hours/internal/db"
	"github.com/Vasek18/my_hours/internal/mailer"
)

// Server holds the dependencies shared by all HTTP handlers. It depends on the
// db.Querier interface so handlers can be tested against a fake data layer.
type Server struct {
	cfg    config.Config
	pool   *pgxpool.Pool
	q      db.Querier
	mailer mailer.Mailer
}

// New constructs a Server from its dependencies.
func New(cfg config.Config, pool *pgxpool.Pool, m mailer.Mailer) *Server {
	return &Server{
		cfg:    cfg,
		pool:   pool,
		q:      db.New(pool),
		mailer: m,
	}
}

// Router builds the Gin engine with all routes and middleware wired up.
func (s *Server) Router(store sessions.Store) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(sessions.Sessions(auth.SessionName, store))

	api := r.Group("/api")
	api.GET("/health", s.health)

	a := api.Group("/auth")
	a.POST("/register", s.register)
	a.POST("/login", s.login)
	a.POST("/logout", s.logout)
	a.POST("/password/forgot", s.forgotPassword)
	a.POST("/password/reset", s.resetPassword)
	a.POST("/email/confirm", s.confirmEmailChange)

	// Authenticated account routes.
	authed := a.Group("")
	authed.Use(s.requireAuth())
	authed.GET("/me", s.me)
	authed.POST("/profile", s.updateProfile)
	authed.POST("/password/change", s.changePassword)
	authed.POST("/email/change", s.changeEmail)

	// Authenticated activity-type CRUD (rows are scoped to the current user).
	at := api.Group("/activity-types")
	at.Use(s.requireAuth())
	at.GET("", s.listActivityTypes)
	at.POST("", s.createActivityType)
	at.GET("/:id", s.getActivityType)
	at.PUT("/:id", s.updateActivityType)
	at.DELETE("/:id", s.deleteActivityType)

	// Authenticated activity CRUD (rows are scoped to the current user).
	act := api.Group("/activities")
	act.Use(s.requireAuth())
	act.GET("", s.listActivities)
	act.POST("", s.createActivity)
	act.GET("/:id", s.getActivity)
	act.PUT("/:id", s.updateActivity)
	act.DELETE("/:id", s.deleteActivity)

	// Authenticated day notes (one per user per day; upsert with empty clears it).
	dn := api.Group("/day-notes")
	dn.Use(s.requireAuth())
	dn.GET("", s.listDayNotes)
	dn.GET("/:date", s.getDayNote)
	dn.PUT("/:date", s.upsertDayNote)

	return r
}
