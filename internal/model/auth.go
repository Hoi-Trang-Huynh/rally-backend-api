package model

type FirebaseAuthRequest struct {
	IDToken string `json:"id_token"`
} //@name FirebaseAuthRequest

type RegisterResponse struct {
    Message       string         `json:"message" example:"Registration successful"`
    User          *UserResponse  `json:"user,omitempty"`       		// only if user already existed or has profile
    Onboarding    bool           `json:"onboarding" example:"true"` // true if user needs onboarding
} //@name RegisterResponse

type UserResponse struct {
	UserID          string `json:"id" example:"65b8a12c3f5e"`
	Email       string `json:"email" example:"john@example.com"`
	DisplayName string `json:"display_name,omitempty" example:"John Doe"`
} //@name UserResponse

type ErrorResponse struct {
    Message string `json:"message" example:"Invalid email or password"`
} //@name ErrorResponse