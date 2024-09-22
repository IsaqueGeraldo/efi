package efi

import (
	"fmt"
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
)

var Client *Credentials
var EFI_BASE_URL string

// Credentials holds the authentication information for the client
type Credentials struct {
	ClientID     string
	ClientSecret string
	Timeout      int
	Sandbox      bool
	CA           string
	Key          string
}

// fileExists checks if the specified file exists
func fileExists(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("file %s not found", fileName)
	}
	return nil
}

// NewClient initializes a new client with the provided credentials
func (c Credentials) NewClient() error {
	err := validation.ValidateStruct(&c,
		validation.Field(&c.ClientID, validation.Required),
		validation.Field(&c.ClientSecret, validation.Required),
		validation.Field(&c.Timeout, validation.Required),
		validation.Field(&c.CA, validation.Required),
		validation.Field(&c.Key, validation.Required),
	)
	if err != nil {
		return err
	}

	if err := fileExists(c.CA); err != nil {
		return err
	}

	if err := fileExists(c.Key); err != nil {
		return err
	}

	// Set the base URL based on the environment (production or sandbox)
	EFI_BASE_URL = EFI_PRODUCTION_URL
	if c.Sandbox {
		EFI_BASE_URL = EFI_STAGING_URL
	}

	Client = &c

	return nil
}
