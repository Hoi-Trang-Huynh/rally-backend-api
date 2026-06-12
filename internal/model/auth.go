package model

// RegisterRequest is the optional payload for POST /auth/register.
// Authentication comes from the Authorization header (verified Firebase ID
// token); the body only carries initial profile fields so signup completes
// in a single request.
type RegisterRequest struct {
	Username  *string `json:"username,omitempty" example:"johndoe"`
	FirstName *string `json:"firstName,omitempty" example:"John"`
	LastName  *string `json:"lastName,omitempty" example:"Doe"`
} //@name RegisterRequest

type RegisterResponse struct {
	Message string        `json:"message" example:"Registration successful"`
	User    *UserResponse `json:"user,omitempty"`
} //@name RegisterResponse

type UserResponse struct {
	ID              string `json:"id" example:"507f1f77bcf86cd799439011"`
	Email           string `json:"email" example:"john@example.com"`
	Username        string `json:"username,omitempty" example:"johndoe"`
	FirstName       string `json:"firstName,omitempty" example:"John"`
	LastName        string `json:"lastName,omitempty" example:"Doe"`
	AvatarUrl       string `json:"avatarUrl,omitempty" example:"https://example.com/avatar.jpg"`
	IsActive        bool   `json:"isActive" example:"true"`
	IsEmailVerified bool   `json:"isEmailVerified" example:"false"`
	IsOnboarding    bool   `json:"isOnboarding" example:"true"`
} //@name UserResponse

type ErrorResponse struct {
	Message string `json:"message" example:"Invalid email or password"`
} //@name ErrorResponse

type AvailabilityResponse struct {
	Available bool   `json:"available" example:"true"`
	Message   string `json:"message" example:"Email is available"`
} //@name AvailabilityResponse
