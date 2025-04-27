package providers

import (
	"encoding/json"
	"fmt"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/mail/html2text"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

// SESConfig holds the configuration for the AWS SES provider
type SESConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	FromName        string
	FromEmail       string
	Endpoint        string // Optional custom endpoint
}

// SESProvider implements the contracts.EmailProvider interface for AWS SES
type SESProvider struct {
	config SESConfig
	logger contracts.Logger
	client *http.Client
}

// NewSESProvider creates a new AWS SES provider
func NewSESProvider(config SESConfig, logger contracts.Logger) *SESProvider {
	return &SESProvider{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// sesEmailRequest represents the request body for the SES API
type sesEmailRequest struct {
	Source      string            `json:"Source"`
	Destination sesDestination    `json:"Destination"`
	Message     sesMessage        `json:"Message"`
	Tags        map[string]string `json:"Tags,omitempty"`
}

type sesDestination struct {
	ToAddresses  []string `json:"ToAddresses"`
	CcAddresses  []string `json:"CcAddresses,omitempty"`
	BccAddresses []string `json:"BccAddresses,omitempty"`
}

type sesMessage struct {
	Subject sesContent `json:"Subject"`
	Body    sesBody    `json:"Body"`
}

type sesContent struct {
	Data    string `json:"Data"`
	Charset string `json:"Charset,omitempty"`
}

type sesBody struct {
	Text sesContent `json:"Text,omitempty"`
	HTML sesContent `json:"Html,omitempty"`
}

// Send implements the contracts.EmailProvider interface
func (p *SESProvider) Send(message *contracts.EmailMessage) error {
	// This is a simplified implementation
	// In a real-world scenario, you would use the AWS SDK for Go
	// to properly interact with the SES API

	p.logger.Info("Sending email via SES provider")

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

	// Format the from address
	fromAddress := formatAddress(message.From)

	// Create the request body
	reqBody := sesEmailRequest{
		Source: fromAddress,
		Destination: sesDestination{
			ToAddresses:  addressesToStrings(message.To, false),
			CcAddresses:  addressesToStrings(message.Cc, false),
			BccAddresses: addressesToStrings(message.Bcc, false),
		},
		Message: sesMessage{
			Subject: sesContent{
				Data:    message.Subject,
				Charset: "UTF-8",
			},
			Body: sesBody{},
		},
		Tags: message.Headers,
	}

	// Add HTML content if provided
	if message.HTML != "" {
		reqBody.Message.Body.HTML = sesContent{
			Data:    message.HTML,
			Charset: "UTF-8",
		}
	}

	// Add text content if provided
	if text != "" {
		reqBody.Message.Body.Text = sesContent{
			Data:    text,
			Charset: "UTF-8",
		}
	}

	// If there are attachments, log a warning that they're not supported in this simplified implementation
	if len(message.Attachments) > 0 {
		p.logger.Warn("Attachments are not supported in the simplified SES provider. Use the AWS SDK for full functionality.")
	}

	// In a real implementation, you would use the AWS SDK to send the email
	// For now, we'll just log the request and return success
	jsonData, err := json.MarshalIndent(reqBody, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	p.logger.Info("SES request payload: ", string(jsonData))
	p.logger.Info("SES email would be sent to: ", strings.Join(reqBody.Destination.ToAddresses, ", "))

	// In a real implementation, this would make an API call to AWS SES
	// For now, we'll just return success
	return nil
}

// Name implements the contracts.EmailProvider interface
func (p *SESProvider) Name() contracts.MailProvider {
	return contracts.ProviderSES
}

// formatAddress formats a mail.Address for SES
func formatAddress(addr *mail.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s>", addr.Name, addr.Address)
	}
	return addr.Address
}
