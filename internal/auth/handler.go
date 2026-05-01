package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/akhilr007/tasks/internal/utils"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(s *Service, log *slog.Logger) *Handler {
	return &Handler{
		service: s,
		logger:  log,
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID         int       `json:"id"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		"method", r.Method,
		"path", r.URL.Path,
		"handler", "Register",
	)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		log.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"))
		utils.WriteError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	var req authRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		log.Warn("invalid request body", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "email and password required")
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Error("user registration failed", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	log.Info("user registered", "user_id", user.ID)
	utils.WriteSuccess(w, http.StatusCreated, user)

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		"method", r.Method,
		"path", r.URL.Path,
		"handler", "Login",
	)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		log.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"))
		utils.WriteError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	var req authRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		log.Warn("invalid request body", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "email and password required")
		return
	}

	user, accessToken, refreshToken, expiry, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Warn("login failed", "error", err)
		utils.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	resp := map[string]any{
		"user": userResponse{
			ID:         user.ID,
			Email:      user.Email,
			IsVerified: user.IsVerified,
			CreatedAt:  user.CreatedAt,
		},
		"accessToken": accessToken,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // true in production (https)
		SameSite: http.SameSiteLaxMode,
		Expires:  expiry,
	})

	log.Info("login successful", "user", user.ID)

	utils.WriteSuccess(w, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		"method", r.Method,
		"path", r.URL.Path,
		"handler", "Refresh",
	)

	// get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Warn("missing refresh token cookie", "error", err)
		utils.WriteError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	rawToken := cookie.Value

	accessToken, newRefreshToken, err := h.service.Refresh(r.Context(), rawToken)
	if err != nil {
		log.Warn("refresh failed", "error", err)

		// clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: false,
			Secure:   true,
			Expires:  time.Unix(0, 0),
		})

		utils.WriteError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: false, // true for production
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	resp := map[string]any{
		"accessToken": accessToken,
	}

	log.Info("token refresh successfully")

	utils.WriteSuccess(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		"method", r.Method,
		"path", r.URL.Path,
		"handler", "Logout",
	)

	// 1. Read refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Warn("missing refresh token", "error", err)
		utils.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	rawToken := cookie.Value

	// 2. Revoke in DB
	err = h.service.Logout(r.Context(), rawToken)
	if err != nil {
		log.Warn("logout failed", "error", err)
		utils.WriteError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	// 3. Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false, // true for production
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})

	log.Info("logout successful")

	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}
