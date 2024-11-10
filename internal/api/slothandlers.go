package api

import (
	"encoding/json"
	"errors"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/internal/service"
	"github.com/apolunin/slotgame/logger"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type (
	spinRequest struct {
		BetAmount int64 `json:"bet_amount"`
	}

	spinResponse struct {
		Balance     int64  `json:"balance"`
		Result      string `json:"result"`
		Combination string `json:"combination"`
	}

	spin struct {
		ID          string    `json:"id"`
		UserID      string    `json:"user_id"`
		Combination string    `json:"combination"`
		Result      string    `json:"spin_result"`
		BetAmount   int64     `json:"bet_amount"`
		WinAmount   int64     `json:"win_amount"`
		CreatedAt   time.Time `json:"created_at"`
	}

	spinHistoryResponse struct {
		Results []*spin `json:"results"`
	}
)

const (
	paramLimit  = "limit"
	paramOffset = "offset"
)

// slotSpinHandler godoc
// @Summary Spin a slot machine by authenticated user
// @Description Spin a slot machine by authenticated user with specified bet amount in cents
// @Tags spins
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param user body spinRequest true "Spin Request"
// @Success 201 {object} spinResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /slot/spin [post]
func (s *Server) slotSpinHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, ok := r.Context().Value(usrCtxKey{}).(*model.User)
		if !ok {
			sendError(w, http.StatusUnauthorized, "unable to extract user from jwt token")
			return
		}

		log := slog.With(logger.FieldUser, usr.Login)

		var payload spinRequest

		switch err := json.NewDecoder(r.Body).Decode(&payload); {
		case err != nil:
			sendError(w, http.StatusBadRequest, "failed to parse request payload")
			return
		case payload.BetAmount <= 0:
			sendError(w, http.StatusBadRequest, "'bet_amount' should be positive")
			return
		}

		switch spin, newBalance, err := s.slotService.Spin(r.Context(), usr, payload.BetAmount); {
		case errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInsufficientFunds):
			sendError(w, http.StatusBadRequest, err.Error())
			return
		case err != nil:
			log.With(logger.FieldError, err).Error("failed to make a spin")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		default:
			log.Info("slot spin was created successfully")
			sendResponse(w, http.StatusCreated, &spinResponse{
				Balance:     newBalance,
				Result:      spin.Result.String(),
				Combination: spin.Combination,
			})

			return
		}
	}
}

// slotHistoryHandler godoc
// @Summary Get authenticated user's spin history
// @Description Get authenticated user's spin history
// @Tags spins
// @Security ApiKeyAuth
// @Produce json
// @Param limit query int false "Limit the number of results"
// @Param offset query int false "Offset the results by this number"
// @Success 200 {object} spinHistoryResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /slot/history [get]
func (s *Server) slotHistoryHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, ok := r.Context().Value(usrCtxKey{}).(*model.User)
		if !ok {
			sendError(w, http.StatusUnauthorized, "unable to extract user from jwt token")
			return
		}

		var (
			log           = slog.With(logger.FieldUser, usr.Login)
			limit, offset = parseLimitAndOffset(r, log)
		)

		spins, err := s.slotService.GetSpinResults(r.Context(), usr, limit, offset)
		switch {
		case err != nil:
			log.With(logger.FieldError, err).Error("failed to get spin history")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		case spins == nil:
			spins = []*model.Spin{}
		}

		res := make([]*spin, 0, len(spins))

		for _, sp := range spins {
			res = append(res, &spin{
				ID:          sp.ID,
				UserID:      sp.UserID,
				Combination: sp.Combination,
				Result:      sp.Result.String(),
				BetAmount:   sp.BetAmount,
				WinAmount:   sp.WinAmount,
				CreatedAt:   sp.CreatedAt,
			})
		}

		log.Info("spin history retrieved successfully")
		sendResponse(w, http.StatusOK, spinHistoryResponse{Results: res})
	}
}

func parseLimitAndOffset(r *http.Request, log *slog.Logger) (limit, offset int64) {
	const (
		defaultLimit  = 100
		defaultOffset = 0
	)

	var (
		query     = r.URL.Query()
		limitStr  = query.Get(paramLimit)
		offsetStr = query.Get(paramOffset)
	)

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	switch {
	case err != nil:
		log.With(logger.FieldLimit, limitStr).Debug("failed to parse limit, falling back to default")
		limit = defaultLimit
	case limit < 1:
		log.With(logger.FieldLimit, limit).Debug("invalid limit value, falling back to default")
		limit = defaultLimit
	}

	offset, err = strconv.ParseInt(offsetStr, 10, 64)
	switch {
	case err != nil:
		log.With(logger.FieldOffset, offsetStr).Debug("failed to parse offset, falling back to default")
		offset = defaultOffset
	case offset < 1:
		log.With(logger.FieldOffset, offset).Debug("invalid offset value, falling back to default")
		offset = defaultOffset
	}

	return limit, offset
}
