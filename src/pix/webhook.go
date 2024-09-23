package pix

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Webhook represents the structure for managing PIX webhooks.
type Webhook struct {
	WebhookURL string      `json:"webhookUrl,omitempty"` // The URL where the webhook will send notifications
	Chave      string      `json:"chave,omitempty"`      // The key associated with the webhook
	SkipMTLS   bool        `json:"-"`                    // Option to skip mutual TLS check
	Criacao    string      `json:"criacao,omitempty"`    // Timestamp of webhook creation
	Parametros *Parametros `json:"parametros,omitempty"` // Parameters for filtering webhook events
	Paginacao  *Paginacao  `json:"paginacao,omitempty"`  // Pagination information
	Webhooks   *[]Webhooks `json:"webhooks,omitempty"`   // List of webhooks
	BadRequest             // Embedding for error handling
}

// Create registers a new webhook for a PIX key.
func (w *Webhook) Create() error {
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

	// Construct the request path for creating the webhook.
	path, err := url.JoinPath(EFI_BASE_URL, "v2", "webhook", w.Chave)
	if err != nil {
		return err
	}

	// Clear the Chave field in the Webhook structure to avoid sending it.
	w.Chave = ""

	// Marshal the Webhook structure to JSON.
	data, err := json.Marshal(w)
	if err != nil {
		return err
	}

	// Create a new HTTP PUT request to register the webhook.
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// Set the appropriate headers for the request.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", authorization())
	req.Header.Set("x-skip-mtls-checking", strconv.FormatBool(w.SkipMTLS))

	// Execute the HTTP request.
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Read the response body.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Unmarshal the response body into the Webhook structure.
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}

	// Check if the request was successful.
	if res.StatusCode != http.StatusCreated {
		return errors.New("bad request")
	}

	return nil // Return nil if the webhook was successfully created.
}

// Delete removes an existing webhook for a PIX key.
func (w *Webhook) Delete() error {
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

	// Construct the request path for deleting the webhook.
	path, err := url.JoinPath(EFI_BASE_URL, "v2", "webhook", w.Chave)
	if err != nil {
		return err
	}

	// Create a new HTTP DELETE request to remove the webhook.
	req, err := http.NewRequest(http.MethodDelete, path, nil)
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

	// Read the response body.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Unmarshal the response body and check the response status.
	if err := json.Unmarshal(body, &w); err != nil && res.StatusCode != http.StatusNoContent {
		return err
	}

	// Check if the request was successful.
	if res.StatusCode != http.StatusNoContent {
		return errors.New("bad request")
	}

	return nil // Return nil if the webhook was successfully deleted.
}

// Parametros defines the parameters for filtering webhook events.
type Parametros struct {
	Inicio    string    `json:"inicio"`    // Start date for the webhook events
	Fim       string    `json:"fim"`       // End date for the webhook events
	Paginacao Paginacao `json:"paginacao"` // Pagination details for the results
}

// Paginacao contains pagination information for the webhook responses.
type Paginacao struct {
	PaginaAtual            int `json:"paginaAtual"`            // Current page number
	ItensPorPagina         int `json:"itensPorPagina"`         // Number of items per page
	QuantidadeDePaginas    int `json:"quantidadeDePaginas"`    // Total number of pages
	QuantidadeTotalDeItens int `json:"quantidadeTotalDeItens"` // Total number of items
}

// Webhooks represents a list of registered webhooks.
type Webhooks struct {
	WebhookUrl string `json:"webhookUrl,omitempty"` // URL of the webhook
	Chave      string `json:"chave,omitempty"`      // Key associated with the webhook
	Criacao    string `json:"criacao,omitempty"`    // Creation timestamp of the webhook
}
