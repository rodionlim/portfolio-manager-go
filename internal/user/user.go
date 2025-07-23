package user

import (
	"errors"
	"fmt"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
)

const (
	UserProfileKey = "user_profile"
)

// Profile represents a user profile
type Profile struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

// Service manages user profile operations
type Service struct {
	db     dal.Database
	logger *logging.Logger
}

// NewService creates a new user service instance
func NewService(db dal.Database) *Service {
	return &Service{
		db:     db,
		logger: logging.GetLogger(),
	}
}

// GetProfile retrieves the user profile, returning default values if not found
func (s *Service) GetProfile() (*Profile, error) {
	var profile Profile
	err := s.db.Get(UserProfileKey, &profile)
	if err != nil {
		// Return default profile if not found
		return &Profile{
			Username: "User",
			Email:    "user@example.com",
			Avatar:   "", // Empty avatar will use initials
		}, nil
	}
	return &profile, nil
}

// UpdateProfile saves the user profile
func (s *Service) UpdateProfile(profile *Profile) error {
	if profile == nil {
		return errors.New("profile cannot be nil")
	}

	// Validate required fields
	if profile.Username == "" {
		return errors.New("username cannot be empty")
	}
	if profile.Email == "" {
		return errors.New("email cannot be empty")
	}

	err := s.db.Put(UserProfileKey, profile)
	if err != nil {
		s.logger.Errorf("Failed to save user profile: %v", err)
		return fmt.Errorf("failed to save user profile: %w", err)
	}

	s.logger.Infof("User profile updated: %s <%s>", profile.Username, profile.Email)
	return nil
}

// UpdateUsername updates only the username
func (s *Service) UpdateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	profile, err := s.GetProfile()
	if err != nil {
		return err
	}

	profile.Username = username
	return s.UpdateProfile(profile)
}

// UpdateEmail updates only the email
func (s *Service) UpdateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	profile, err := s.GetProfile()
	if err != nil {
		return err
	}

	profile.Email = email
	return s.UpdateProfile(profile)
}

// UpdateAvatar updates only the avatar
func (s *Service) UpdateAvatar(avatar string) error {
	profile, err := s.GetProfile()
	if err != nil {
		return err
	}

	profile.Avatar = avatar
	return s.UpdateProfile(profile)
}