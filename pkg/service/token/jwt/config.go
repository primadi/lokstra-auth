package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/primadi/lokstra/common/json"
)

// Config holds JWT configuration
type Config struct {
	// SigningMethod is the signing algorithm (HS256, RS256, ES256, etc.)
	SigningMethod jwt.SigningMethod `json:"-"` // Populated via UnmarshalJSON

	// SigningMethodStr is the string representation of the signing method
	SigningMethodStr string `json:"signing_method"`

	// SigningKey is the key used to sign tokens
	SigningKey []byte `json:"signing_key"`

	// VerifyingKey is the key used to verify tokens (can be same as SigningKey)
	VerifyingKey []byte `json:"verifying_key"`

	// Issuer is the token issuer
	Issuer string `json:"issuer"`

	// Audience is the intended audience
	Audience []string `json:"audience"`

	// AccessTokenDuration is how long access tokens are valid
	AccessTokenDuration time.Duration `json:"access_token_duration"`

	// RefreshTokenDuration is how long refresh tokens are valid
	RefreshTokenDuration time.Duration `json:"refresh_token_duration"`
}

// UnmarshalJSON custom unmarshaler to handle signing method conversion
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert signing method string to jwt.SigningMethod
	if c.SigningMethod == nil {
		if c.SigningMethodStr == "" {
			c.SigningMethodStr = "HS256" // Default to HS256
		}

		c.SigningMethod = StringToSigningMethod(c.SigningMethodStr)
		if c.SigningMethod == nil {
			return fmt.Errorf("unsupported signing method: %s", c.SigningMethodStr)
		}
	}

	if c.Issuer == "" {
		c.Issuer = "lokstra-auth"
	}
	if len(c.Audience) == 0 {
		c.Audience = []string{"lokstra-clients"}
	}

	if c.AccessTokenDuration == 0 {
		c.AccessTokenDuration = 15 * time.Minute
	}

	if c.RefreshTokenDuration == 0 {
		c.RefreshTokenDuration = 7 * 24 * time.Hour
	}

	return nil
}

// StringToSigningMethod converts a string to jwt.SigningMethod
func StringToSigningMethod(method string) jwt.SigningMethod {
	switch method {
	case "HS256":
		return jwt.SigningMethodHS256
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	case "RS256":
		return jwt.SigningMethodRS256
	case "RS384":
		return jwt.SigningMethodRS384
	case "RS512":
		return jwt.SigningMethodRS512
	case "ES256":
		return jwt.SigningMethodES256
	case "ES384":
		return jwt.SigningMethodES384
	case "ES512":
		return jwt.SigningMethodES512
	case "PS256":
		return jwt.SigningMethodPS256
	case "PS384":
		return jwt.SigningMethodPS384
	case "PS512":
		return jwt.SigningMethodPS512
	default:
		return nil
	}
}
