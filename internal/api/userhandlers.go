package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apolunin/slotgame/internal/model"
	"github.com/apolunin/slotgame/internal/service"
	"github.com/apolunin/slotgame/internal/storage"
	"github.com/apolunin/slotgame/logger"
	"log/slog"
	"net/http"
)

type (
	registerUserRequest struct {
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Login     string `json:"login"`
		Password  string `json:"password"`
		Balance   int64  `json:"balance,omitempty"`
	}

	loginUserRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	loginUserResponse struct {
		Token string `json:"token"`
	}
)

// registerHandler godoc
// @Summary Register a new user
// @Description Register a new user by providing login and password
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body registerUserRequest true "User Data"
// @Success 201 {object} model.User
// @Failure 400 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (s *Server) registerHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		var usr registerUserRequest

		switch err := json.NewDecoder(r.Body).Decode(&usr); {
		case err != nil:
			sendError(w, http.StatusBadRequest, "failed to parse request payload")
			return
		case usr.Login == "" || usr.Password == "":
			sendError(w, http.StatusBadRequest, "'login' and 'password' fields are required")
			return
		}

		res, err := s.userService.CreateUser(
			r.Context(),
			usr.Login,
			usr.Password,
			usr.FirstName,
			usr.LastName,
			usr.Balance,
		)

		log := slog.With(logger.FieldUser, usr.Login)

		switch {
		case errors.Is(err, storage.ErrUserExists):
			sendError(w, http.StatusBadRequest, fmt.Sprintf("user %q already exists", usr.Login))
			return
		case err != nil:
			log.With(logger.FieldError, err).Error("failed to create user")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		res.Password = ""

		log.Info("user registered successfully")
		sendResponse(w, http.StatusCreated, res)
	}
}

// loginHandler godoc
// @Summary Login user into the system
// @Description Login a user by providing login and password
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body loginUserRequest true "User Credentials"
// @Success 200 {object} loginUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /login [post]
func (s *Server) loginHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		var usr loginUserRequest

		switch err := json.NewDecoder(r.Body).Decode(&usr); {
		case err != nil:
			sendError(w, http.StatusBadRequest, "failed to parse request payload")
			return
		case usr.Login == "" || usr.Password == "":
			sendError(w, http.StatusBadRequest, "'login' and 'password' fields are required")
			return
		}

		log := slog.With(logger.FieldUser, usr.Login)

		switch token, err := s.userService.Login(
			r.Context(),
			usr.Login,
			usr.Password,
		); {
		case errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidCredentials):
			sendError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, service.ErrInvalidCredentials):
			sendError(w, http.StatusForbidden, "invalid credentials")
			return
		case err != nil || token == "":
			log.With(logger.FieldError, err).Error("failed to login user")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		default:
			log.Info("user logged in successfully")
			sendResponse(w, http.StatusOK, loginUserResponse{Token: token})
			return
		}
	}
}

// profileHandler godoc
// @Summary Get user profile
// @Description Retrieve the profile info and balance of the authenticated user
// @Tags users
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} model.User
// @Failure 401 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /profile [get]
func (s *Server) profileHandler() handlerFn {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, ok := r.Context().Value(usrCtxKey{}).(*model.User)
		if !ok {
			sendError(w, http.StatusUnauthorized, "unable to extract user from jwt token")
			return
		}

		usr.Password = ""

		slog.With(logger.FieldUser, usr.Login).Info("user profile retrieved successfully")
		sendResponse(w, http.StatusOK, usr)
	}
}
