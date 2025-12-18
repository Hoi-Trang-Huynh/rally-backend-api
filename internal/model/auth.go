package model

type FirebaseAuthRequest struct {
	IDToken string `json:"id_token"`
} //@name FirebaseAuthRequest

type RegisterResponse struct {
    Message       string         `json:"message" example:"Registration successful"`
    User          *UserResponse  `json:"user,omitempty"`       		// only if user already existed or has profile
    Onboarding    bool           `json:"onboarding" example:"true"` // true if user needs onboarding
} //@name RegisterResponse

type LoginResponse struct {
    Message       string         `json:"message" example:"Registration successful"`
    User          *UserResponse  `json:"user,omitempty"`       		// only if user already existed or has profile
} //@name LoginResponse

type UserResponse struct {
	ID          string `json:"id" example:"507f1f77bcf86cd799439011"`
	Email       string `json:"email" example:"john@example.com"`
	Username    string `json:"username,omitempty" example:"John Doe"`
	FirstName   string `json:"first_name,omitempty" example:"John"`
	LastName    string `json:"last_name,omitempty" example:"Doe"`
	AvatarURL   string `json:"avatar_url,omitempty" example:"https://example.com/profile.jpg"`
} //@name UserResponse

type ErrorResponse struct {
    Message string `json:"message" example:"Invalid email or password"`
} //@name ErrorResponse