package http

import (
	"encoding/json"
	"errors"
	"net/http"

	appUser "go-ddd/internal/application/user"
	domainUser "go-ddd/internal/domain/user"
)

type UserHandler struct {
	userService *appUser.UserService
}

func NewUserHandler(userService *appUser.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with hashed password and returns an auth token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body user.RegisterRequest true "User Registration Request"
// @Success      201  {object}  user.AuthResponse
// @Failure      400  {object}  map[string]string "Invalid request body or validation error"
// @Failure      409  {object}  map[string]string "User with this email already exists"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /api/v1/auth/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req appUser.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.userService.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainUser.ErrUserAlreadyExists):
			respondError(w, http.StatusConflict, err.Error())
		case errors.Is(err, domainUser.ErrInvalidEmail), errors.Is(err, domainUser.ErrInvalidPassword), errors.Is(err, domainUser.ErrEmptyFields):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusCreated, res)
}

// Login godoc
// @Summary      User login
// @Description  Authenticates user credentials and returns a JWT access token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body user.LoginRequest true "User Login Request"
// @Success      200  {object}  user.AuthResponse
// @Failure      400  {object}  map[string]string "Invalid request body"
// @Failure      401  {object}  map[string]string "Invalid email or password"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /api/v1/auth/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req appUser.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.userService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domainUser.ErrInvalidCredentials):
			respondError(w, http.StatusUnauthorized, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, res)
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Retrieves profile information for the authenticated user.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  user.UserDTO
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "User not found"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domainUser.ErrUserNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
