package api

import (
	"encoding/json"
	"errors"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/internal/service"
	"github.com/apolunin/slotgame/logger"
	"log/slog"
	"net/http"
)

type (
	walletRequest struct {
		Amount int64 `json:"amount"`
	}

	walletResponse struct {
		Balance int64 `json:"balance"`
	}
)

// depositHandler godoc
// @Summary Deposit funds to authenticated user's balance
// @Description Deposit funds to authenticated user's balance, amount is specified in cents
// @Tags balances
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param user body walletRequest true "Deposit Request"
// @Success 200 {object} walletResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wallet/deposit [post]
func (s *Server) depositHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, ok := r.Context().Value(usrCtxKey{}).(*model.User)
		if !ok {
			sendError(w, http.StatusUnauthorized, "unable to extract user from jwt token")
			return
		}

		log := slog.With(logger.FieldUser, usr.Login)

		var payload walletRequest

		switch err := json.NewDecoder(r.Body).Decode(&payload); {
		case err != nil:
			sendError(w, http.StatusBadRequest, "failed to parse request payload")
			return
		case payload.Amount <= 0:
			sendError(w, http.StatusBadRequest, "'amount' should be positive")
			return
		}

		newBalance, err := s.userService.DepositFunds(
			r.Context(),
			usr.Login,
			payload.Amount,
		)

		switch {
		case errors.Is(err, service.ErrUserNotFound):
			sendError(w, http.StatusBadRequest, err.Error())
			return
		case err != nil:
			log.With(logger.FieldError, err).Error("failed to deposit funds")
			sendResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		log.Info("funds were deposited successfully")
		sendResponse(w, http.StatusOK, walletResponse{Balance: newBalance})
	}
}

// withdrawHandler godoc
// @Summary Withdraw funds from authenticated user's balance
// @Description Withdraw funds from authenticated user's balance, amount is specified in cents
// @Tags balances
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param user body walletRequest true "Withdraw Request"
// @Success 200 {object} walletResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wallet/withdraw [post]
func (s *Server) withdrawHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, ok := r.Context().Value(usrCtxKey{}).(*model.User)
		if !ok {
			sendError(w, http.StatusUnauthorized, "unable to extract user from jwt token")
			return
		}

		log := slog.With(logger.FieldUser, usr.Login)

		var payload walletRequest

		switch err := json.NewDecoder(r.Body).Decode(&payload); {
		case err != nil:
			sendError(w, http.StatusBadRequest, "failed to parse request payload")
			return
		case payload.Amount <= 0:
			sendError(w, http.StatusBadRequest, "'amount' should have positive value")
			return
		}

		newBalance, err := s.userService.WithdrawFunds(
			r.Context(),
			usr.Login,
			payload.Amount,
		)

		switch {
		case errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInsufficientFunds):
			sendError(w, http.StatusBadRequest, err.Error())
			return
		case err != nil:
			log.With(logger.FieldError, err).Error("failed to withdraw funds")
			sendResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}

		log.Info("funds were withdrawn successfully")
		sendResponse(w, http.StatusOK, walletResponse{Balance: newBalance})
	}
}
