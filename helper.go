package efi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// BadRequest represents a detailed error response
type BadRequest struct {
	Name             string   `json:"name,omitempty"`
	Message          string   `json:"message,omitempty"`
	Error            string   `json:"error,omitempty"`
	ErrorDescription string   `json:"error_description,omitempty"`
	Errors           *[]Error `json:"errors,omitempty"`
}

// Error represents a specific error within BadRequest
type Error struct {
	Key     string `json:"key,omitempty"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
}

// decodeJWT decodes a JWT token and returns its claims
func decodeJWT(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token: must have 3 parts")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("failed to decode token payload")
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.New("failed to decode JSON payload")
	}

	return claims, nil
}

const (
	EFI_PRODUCTION_URL = "https://pix.api.efipay.com.br"
	EFI_STAGING_URL    = "https://pix-h.api.efipay.com.br"
)
