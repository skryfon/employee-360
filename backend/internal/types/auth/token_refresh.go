package auth

import "time"

// TokenRefreshRequest contains parameters for rotating a refresh token.
type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// TokenRefreshResponse contains new tokens generated after refresh.
type TokenRefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}
