package passwordless

import "time"

// Credentials represents passwordless authentication credentials
type Credentials struct {
	// Delivery method: "email" or "sms"
	Method string `json:"method"`

	// Email or phone number
	Identifier string `json:"identifier"`

	// Code or token sent to user
	Code string `json:"code,omitempty"`

	// Magic link token (alternative to code)
	Token string `json:"token,omitempty"`
}

// User represents a user in passwordless authentication
type User struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Email       string    `json:"email,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
