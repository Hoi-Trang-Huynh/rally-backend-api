package model

type RegisterEmailRequest struct {
    Email    string `json:"email" example:"john@example.com"`
    Password string `json:"password" example:"secret123"`
} //@name RegisterEmailRequest

type RegisterOAuthRequest struct {
    Provider    string `json:"provider" example:"google"`
    AccessToken string `json:"access_token" example:"ya29.a0AfH6SM..."`
} //@name RegisterOAuthRequest

type RegisterResponse struct {
    Message       string         `json:"message" example:"Registration successful"`
    Token         string         `json:"token" example:"eyJhbGciOiJI..."`
    User          *UserResponse  `json:"user,omitempty"`       		// only if user already existed or has profile
    Onboarding    bool           `json:"onboarding" example:"true"` // true if user needs onboarding
} //@name RegisterResponse

type UserResponse struct {
	ID          string `json:"id" example:"65b8a12c3f5e"`
	Email       string `json:"email" example:"john@example.com"`
	DisplayName string `json:"display_name,omitempty" example:"John Doe"`
} //@name UserResponse

type ErrorResponse struct {
    Error string `json:"error" example:"Invalid email or password"`
} //@name ErrorResponse
