package pix

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Key represents the structure for storing PIX keys.
type Key struct {
	Chaves     []string `json:"chaves,omitempty"` // List of PIX keys
	Chave      string   `json:"chave,omitempty"`  // Selected PIX key
	BadRequest          // Embedding BadRequest for error handling
}

// Fetch retrieves the available PIX keys from the server.
func (k *Key) Fetch() error {
	// Obtain an OAuth token for authentication.
	token := OAuth()
	if token.Error != nil {
		return token.Error
	}

	// Load the client certificate for secure communication.
	cert, err := tls.LoadX509KeyPair(Client.CA, Client.Key)
	if err != nil {
		return err
	}

	// Set up the HTTP client with a timeout and TLS configuration.
	client := &http.Client{
		Timeout: time.Second * time.Duration(Client.Timeout),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	// Construct the request path for fetching PIX keys.
	path, err := url.JoinPath(EFI_BASE_URL, "v2", "gn", "evp")
	if err != nil {
		return err
	}

	// Create a new HTTP GET request.
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	// Set the appropriate headers for the request.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", authorization())

	// Execute the HTTP request.
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() // Ensure the response body is closed after reading.

	// Read the response body.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Unmarshal the response body into the Key object.
	if err := json.Unmarshal(body, &k); err != nil {
		return err
	}

	// Check if the response status is successful.
	if res.StatusCode != http.StatusOK {
		return errors.New("bad request")
	}

	return nil // Return nil if the keys were fetched successfully.
}
