package models

// JWTAuthResponse is the response to a successful login and token refresh. The expires field is used on by the client to
// know when to refresh the token.
//
//nolint:lll
type JWTAuthResponse struct {
	Token     string `json:"token" yaml:"token" validate:"required"`         // JWT string sent to and validated by the server.
	Expires   int64  `json:"expires" yaml:"expires" validate:"required"`     // Expiration time as unix time stamp. Strictly used by client to gauge when to refresh the token.
	Threshold int64  `json:"threshold" yaml:"threshold" validate:"required"` // The window in seconds before expiration during which the token can be refreshed.
}

// HTTPError is a generic error message that is returned to the requester.
type HTTPError struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}

// HTTPSuccess is a generic success message that is returned to the requester.
type HTTPSuccess struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}
