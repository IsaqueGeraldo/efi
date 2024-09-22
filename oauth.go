package efi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var Authorization Token

// Token represents the authentication credentials
type Token struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Error       error  `json:"-"`
	BadRequest
}

// OAuth performs authentication and returns a Token
func OAuth() Token {
	if Client == nil {
		return Token{Error: errors.New("client not defined")}
	}

	if token := checkToken(); token != nil {
		return *token
	}

	payload := strings.NewReader(`{"grant_type": "client_credentials"}`)

	cert, err := tls.LoadX509KeyPair(Client.CA, Client.Key)
	if err != nil {
		return Token{Error: fmt.Errorf("failed to load certificates: %v", err)}
	}

	client := &http.Client{
		Timeout: time.Second * time.Duration(Client.Timeout),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	path, err := url.JoinPath(EFI_BASE_URL, "oauth", "token")
	if err != nil {
		return Token{Error: fmt.Errorf("failed to construct URL: %v", err)}
	}

	req, err := http.NewRequest(http.MethodPost, path, payload)
	if err != nil {
		return Token{Error: err}
	}

	req.SetBasicAuth(Client.ClientID, Client.ClientSecret)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return Token{Error: err}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Token{Error: err}
	}

	token := Token{}
	if err := json.Unmarshal(body, &token); err != nil {
		return Token{Error: err}
	}

	if res.StatusCode != http.StatusOK {
		return Token{Error: fmt.Errorf("bad request: %s", string(body))}
	}

	Authorization = token
	return token
}

// checkToken verifies if the current token is valid
func checkToken() *Token {
	if Authorization.AccessToken != "" {
		token := strings.Split(authorization(), " ")[1]

		claims, err := decodeJWT(token)
		if err != nil {
			return &Token{Error: err}
		}

		if exp, ok := claims["exp"].(float64); ok && time.Unix(int64(exp), 0).After(time.Now().Add(30*time.Second)) {
			return &Authorization
		}
	}

	return nil
}

// authorization returns the authorization token in the correct format
func authorization() string {
	return fmt.Sprintf("%s %s", Authorization.TokenType, Authorization.AccessToken)
}
