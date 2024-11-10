package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apolunin/slotgame/config"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/time/rate"
	"log/slog"
	"net/http"
	"runtime"
	"time"
)

const (
	headerContentType          = "Content-Type"
	headerAuthorization        = "Authorization"
	contentTypeApplicationJSON = "application/json"
)

type (
	usrCtxKey struct{}

	handlerFn    = func(w http.ResponseWriter, r *http.Request)
	middlewareFn = func(next http.Handler) http.Handler

	authService interface {
		GetLoginFromToken(tokenString string) (string, error)
	}

	userService interface {
		CreateUser(
			ctx context.Context,
			login string,
			password string,
			firstName string,
			lastName string,
			balance int64,
		) (*model.User, error)

		Login(
			ctx context.Context,
			login string,
			password string,
		) (string, error)

		GetUserByLogin(
			ctx context.Context,
			login string,
		) (*model.User, error)

		DepositFunds(
			ctx context.Context,
			login string,
			amount int64,
		) (int64, error)

		WithdrawFunds(
			ctx context.Context,
			login string,
			amount int64,
		) (int64, error)
	}

	slotService interface {
		Spin(
			ctx context.Context,
			user *model.User,
			betAmount int64,
		) (*model.Spin, int64, error)

		GetSpinResults(
			ctx context.Context,
			user *model.User,
			limit int64,
			offset int64,
		) ([]*model.Spin, error)
	}

	Server struct {
		authService  authService
		userService  userService
		slotService  slotService
		rateLimitCfg config.RateLimitConfig
	}

	ErrorResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
)

func NewServer(
	authService authService,
	userService userService,
	slotService slotService,
	rateLimitCfg config.RateLimitConfig,
) *Server {
	return &Server{
		authService:  authService,
		userService:  userService,
		slotService:  slotService,
		rateLimitCfg: rateLimitCfg,
	}
}

func (s *Server) Start(port int) error {
	r := chi.NewRouter()

	setupDefaultMiddlewareLogger()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.With(s.rateLimitMiddleware()).Route("/api", func(r chi.Router) {
		r.Post("/register", s.registerHandler())
		r.Post("/login", s.loginHandler())
		r.With(s.authMiddleware()).Get("/profile", s.profileHandler())

		r.With(s.authMiddleware()).Post("/wallet/deposit", s.depositHandler())
		r.With(s.authMiddleware()).Post("/wallet/withdraw", s.withdrawHandler())

		r.With(s.authMiddleware()).Post("/slot/spin", s.slotSpinHandler())
		r.With(s.authMiddleware()).Get("/slot/history", s.slotHistoryHandler())
	})

	slog.Info(fmt.Sprintf("server is started on port %d...", port))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

func (s *Server) authMiddleware() middlewareFn {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(headerAuthorization)

			if token == "" {
				sendError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			login, err := s.authService.GetLoginFromToken(token)
			if err != nil {
				sendError(w, http.StatusUnauthorized, err.Error())
				return
			}

			usr, err := s.userService.GetUserByLogin(r.Context(), login)
			if err != nil {
				sendError(w, http.StatusUnauthorized, err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), usrCtxKey{}, usr)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *Server) rateLimitMiddleware() middlewareFn {
	var (
		limit   = s.rateLimitCfg.Rate
		burst   = s.rateLimitCfg.Burst
		limiter = rate.NewLimiter(rate.Every(time.Duration(limit)*time.Millisecond), burst)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !limiter.Allow() {
				sendError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func sendResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set(headerContentType, contentTypeApplicationJSON)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set(headerContentType, contentTypeApplicationJSON)
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Message: message,
		Code:    statusCode,
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setupDefaultMiddlewareLogger() {
	color := true

	if runtime.GOOS == "windows" {
		color = false
	}

	middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo),
		NoColor: !color,
	})
}
