// Package user provides handlers for managing user profile operations
// @title User Profile API
// @version 1.0
// @description API for managing user profile settings
package user

import (
	"encoding/json"
	"net/http"
	"portfolio-manager/pkg/common"
	"portfolio-manager/pkg/logging"
)

// UpdateProfileRequest represents the request body for updating user profile
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

// HandleProfileGet handles retrieving the user profile.
// @Summary Get user profile
// @Description Retrieves the current user profile information
// @Tags user
// @Produce json
// @Success 200 {object} Profile
// @Failure 500 {object} common.ErrorResponse "Failed to get user profile"
// @Router /api/v1/user/profile [get]
func HandleProfileGet(userService *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profile, err := userService.GetProfile()
		if err != nil {
			logging.GetLogger().Errorf("Failed to get user profile: %v", err)
			common.WriteJSONError(w, "Failed to get user profile", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			logging.GetLogger().Errorf("Failed to encode user profile response: %v", err)
			common.WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// HandleProfileUpdate handles updating the user profile.
// @Summary Update user profile
// @Description Updates the user profile information
// @Tags user
// @Accept json
// @Produce json
// @Param profile body UpdateProfileRequest true "User profile data"
// @Success 200 {object} Profile
// @Failure 400 {object} common.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} common.ErrorResponse "Failed to update user profile"
// @Router /api/v1/user/profile [put]
func HandleProfileUpdate(userService *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logging.GetLogger().Errorf("Failed to decode user profile request: %v", err)
			common.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create profile from request
		profile := &Profile{
			Username: request.Username,
			Email:    request.Email,
			Avatar:   request.Avatar,
		}

		// Update the profile
		if err := userService.UpdateProfile(profile); err != nil {
			logging.GetLogger().Errorf("Failed to update user profile: %v", err)
			common.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Return the updated profile
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			logging.GetLogger().Errorf("Failed to encode user profile response: %v", err)
			common.WriteJSONError(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// RegisterHandlers registers the handlers for the user service.
func RegisterHandlers(mux *http.ServeMux, userService *Service) {
	mux.HandleFunc("/api/v1/user/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleProfileGet(userService).ServeHTTP(w, r)
		case http.MethodPut:
			HandleProfileUpdate(userService).ServeHTTP(w, r)
		default:
			common.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}