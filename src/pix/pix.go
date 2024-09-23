package pix

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Pix represents the main structure for a PIX transaction.
type Pix struct {
	TipoCob            string           `json:"tipoCob,omitempty"`            // Type of charge
	Status             string           `json:"status,omitempty"`             // Transaction status
	Calendario         *Calendario      `json:"calendario,omitempty"`         // Calendar information
	Location           string           `json:"location,omitempty"`           // Location string
	TxID               string           `json:"txid,omitempty"`               // Transaction ID
	Revisao            int              `json:"revisao,omitempty"`            // Revision number
	Devedor            *Devedor         `json:"devedor,omitempty"`            // Debtor information
	Pagador            *Pagador         `json:"pagador,omitempty"`            // Payer information
	Valor              interface{}      `json:"valor,omitempty"`              // Transaction value
	Chave              string           `json:"chave,omitempty"`              // Key for the transaction
	SolicitacaoPagador string           `json:"solicitacaoPagador,omitempty"` // Payer's request
	PixCopiaECola      string           `json:"pixCopiaECola,omitempty"`      // Copy and paste PIX
	InfoAdicionais     *[]InfoAdicional `json:"infoAdicionais,omitempty"`     // Additional information
	Loc                *Loc             `json:"loc,omitempty"`                // Location information
	Favorecido         *Favorecido      `json:"favorecido,omitempty"`         // Recipient information
	BadRequest
}

// Create initializes and sends a PIX transaction request.
func (p *Pix) Create() error {
	// Check if the PIX key is provided; if not, fetch available keys.
	if p.Chave == "" {
		keys := Key{}

		// Fetch the keys, return an error if the fetch fails.
		if err := keys.Fetch(); err != nil {
			return err
		}

		// Ensure at least one key is available.
		if len(keys.Chaves) == 0 {
			return errors.New("no pix keys found")
		}

		// Assign the first available key to the transaction.
		p.Chave = keys.Chaves[0]
	}

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

	// Construct the request path for the PIX transaction.
	path, err := url.JoinPath(EFI_BASE_URL, "v2", "cob", p.TxID)
	if err != nil {
		return err
	}

	// Marshal the Pix object to JSON format.
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	// Determine the HTTP method: POST for new transactions, PUT for updates.
	method := http.MethodPost
	if p.TxID != "" {
		method = http.MethodPut
	}

	// Create a new HTTP request with the specified method, path, and body.
	req, err := http.NewRequest(method, path, bytes.NewBuffer(data))
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

	// Unmarshal the response body into the Pix object.
	if err = json.Unmarshal(body, &p); err != nil {
		return err
	}

	// Check if the response status is successful.
	if res.StatusCode != http.StatusCreated {
		return errors.New("bad request")
	}

	return nil // Return nil if the transaction was created successfully.
}

// Fetch retrieves the details of a PIX transaction using its TxID.
func (p *Pix) Fetch() error {
	// Ensure that TxID is provided; it is required to fetch the transaction.
	if p.TxID == "" {
		return errors.New("txid is required")
	}

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

	// Construct the request path for fetching transaction details.
	path, err := url.JoinPath(EFI_BASE_URL, "v2", "cob", p.TxID)
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

	// Unmarshal the response body into the Pix object.
	if err = json.Unmarshal(body, &p); err != nil {
		return err
	}

	// Check if the response status is successful.
	if res.StatusCode != http.StatusOK {
		return errors.New("bad request")
	}

	return nil // Return nil if the transaction details were fetched successfully.
}

// Calendario contains information about the transaction's calendar.
type Calendario struct {
	Criacao                string `json:"criacao,omitempty"`                // Creation date
	Expiracao              int    `json:"expiracao,omitempty"`              // Expiration time in minutes
	DataDeVencimento       string `json:"dataDeVencimento,omitempty"`       // Due date
	ValidadeAposVencimento int    `json:"validadeAposVencimento,omitempty"` // Validity after expiration
}

// Devedor represents the debtor's information.
type Devedor struct {
	CPF        string `json:"cpf,omitempty"`        // CPF (individual taxpayer ID)
	CNPJ       string `json:"cnpj,omitempty"`       // CNPJ (business taxpayer ID)
	Nome       string `json:"nome,omitempty"`       // Name of the debtor
	Logradouro string `json:"logradouro,omitempty"` // Address
	Cidade     string `json:"cidade,omitempty"`     // City
	UF         string `json:"uf,omitempty"`         // State
	CEP        string `json:"cep,omitempty"`        // ZIP code
}

// Valor represents the value details of the transaction.
type Valor struct {
	Original string    `json:"original,omitempty"` // Original amount
	Multa    *Multa    `json:"multa,omitempty"`    // Penalty information
	Juros    *Juros    `json:"juros,omitempty"`    // Interest information
	Desconto *Desconto `json:"desconto,omitempty"` // Discount information
}

// Multa contains information about penalties.
type Multa struct {
	Modalidade int    `json:"modalidade,omitempty"` // Penalty modality
	ValorPerc  string `json:"valorPerc,omitempty"`  // Penalty percentage
}

// Juros contains information about interest.
type Juros struct {
	Modalidade int    `json:"modalidade,omitempty"` // Interest modality
	ValorPerc  string `json:"valorPerc,omitempty"`  // Interest percentage
}

// Desconto contains information about discounts.
type Desconto struct {
	Modalidade       int                `json:"modalidade,omitempty"`       // Discount modality
	DescontoDataFixa []DescontoDataFixa `json:"descontoDataFixa,omitempty"` // Fixed date discounts
}

// DescontoDataFixa represents a fixed date discount.
type DescontoDataFixa struct {
	Data      string `json:"data,omitempty"`      // Discount date
	ValorPerc string `json:"valorPerc,omitempty"` // Discount percentage
}

// InfoAdicional represents additional information related to the transaction.
type InfoAdicional struct {
	Nome  string `json:"nome,omitempty"`  // Name of the additional info
	Valor string `json:"valor,omitempty"` // Value of the additional info
}

// Loc represents location details for the transaction.
type Loc struct {
	ID       int    `json:"id,omitempty"`       // Location ID
	Location string `json:"location,omitempty"` // Location string
	TipoCob  string `json:"tipoCob,omitempty"`  // Type of charge
}

// Pagador represents payer's information.
type Pagador struct {
	Chave       string `json:"chave,omitempty"`       // Payer key
	InfoPagador string `json:"infoPagador,omitempty"` // Payer information
}

// Favorecido represents the recipient's information.
type Favorecido struct {
	Chave string `json:"chave,omitempty"` // Recipient key
}
