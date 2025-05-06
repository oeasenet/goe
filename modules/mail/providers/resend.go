package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/mail/html2text"
	"io"
	"net/http"
	"time"
)

// ResendConfig holds the configuration for the Resend provider
type ResendConfig struct {
	APIKey    string
	FromName  string
	FromEmail string
}

// ResendProvider implements the contracts.EmailProvider interface for Resend
type ResendProvider struct {
	config ResendConfig
	logger contracts.Logger
	client *http.Client
}

// NewResendProvider creates a new Resend provider
func NewResendProvider(config ResendConfig, logger contracts.Logger) *ResendProvider {
	return &ResendProvider{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// resendEmailRequest represents the request body for the Resend API
type resendEmailRequest struct {
	From        string             `json:"from"`
	To          []string           `json:"to"`
	Cc          []string           `json:"cc,omitempty"`
	Bcc         []string           `json:"bcc,omitempty"`
	Subject     string             `json:"subject"`
	HTML        string             `json:"html,omitempty"`
	Text        string             `json:"text,omitempty"`
	Headers     map[string]string  `json:"headers,omitempty"`
	Attachments []resendAttachment `json:"attachments,omitempty"`
}

// resendAttachment represents an attachment for the Resend API
type resendAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"` // Base64 encoded content
}

// resendEmailResponse represents the response from the Resend API
type resendEmailResponse struct {
	ID      string `json:"id"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Send implements the contracts.EmailProvider interface
func (p *ResendProvider) Send(message *contracts.EmailMessage) error {
	// Prepare the request body
	fromAddress := fmt.Sprintf("%s <%s>", message.From.Name, message.From.Address)
	if message.From.Name == "" {
		fromAddress = message.From.Address
	}

	// Convert To, Cc, Bcc to string slices
	to := addressesToStrings(message.To, true)
	cc := addressesToStrings(message.Cc, true)
	bcc := addressesToStrings(message.Bcc, true)

	// Ensure we have plain text if only HTML is provided
	text := message.Text
	if text == "" && message.HTML != "" {
		var err error
		text, err = html2text.FromString(message.HTML)
		if err != nil {
			p.logger.Warn("Failed to convert HTML to text: ", err)
			// Continue anyway, as text is optional
		}
	}

	// Prepare attachments
	attachments := []resendAttachment{}
	for name, reader := range message.Attachments {
		// Read the attachment content
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, reader)
		if err != nil {
			return fmt.Errorf("failed to read attachment %s: %w", name, err)
		}

		// Base64 encode the content
		content := buf.Bytes()
		attachments = append(attachments, resendAttachment{
			Filename: name,
			Content:  string(content), // Resend expects base64 encoded content
		})
	}

	// Create the request
	reqBody := resendEmailRequest{
		From:        fromAddress,
		To:          to,
		Cc:          cc,
		Bcc:         bcc,
		Subject:     message.Subject,
		HTML:        message.HTML,
		Text:        text,
		Headers:     message.Headers,
		Attachments: attachments,
	}

	// Marshal the request body
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resend API error: %s", string(respBody))
	}

	// Parse the response
	var resendResp resendEmailResponse
	err = json.Unmarshal(respBody, &resendResp)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if resendResp.Error != "" {
		return fmt.Errorf("resend API error: %s - %s", resendResp.Error, resendResp.Message)
	}

	return nil
}

// Name implements the contracts.EmailProvider interface
func (p *ResendProvider) Name() contracts.MailProvider {
	return contracts.ProviderResend
}
