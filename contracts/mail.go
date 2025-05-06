package contracts

import (
	"io"
	"net/mail"
)

// MailProvider represents the type of email provider
type MailProvider string

const (
	ProviderSMTP   MailProvider = "smtp"
	ProviderResend MailProvider = "resend"
	ProviderSES    MailProvider = "ses"
	// Add more providers as needed
)

// Mailer is the main interface for sending emails
type Mailer interface {
	// GetSender returns an EmailSender for the specified provider
	GetSender(provider MailProvider) EmailSender
	// DefaultSender returns the default EmailSender
	DefaultSender() EmailSender
	// RegisterProvider registers a new email provider
	RegisterProvider(provider MailProvider, factory EmailProviderFactory) error
}

// EmailSender is the interface for sending emails with a fluent API
type EmailSender interface {
	To(t *[]*mail.Address) EmailSender
	Bcc(b *[]*mail.Address) EmailSender
	Cc(c *[]*mail.Address) EmailSender
	Subject(sub string) EmailSender
	HTML(html string) EmailSender
	Text(text string) EmailSender
	Headers(h map[string]string) EmailSender
	Attachments(a map[string]string) EmailSender
	Send(useQueue ...bool) error
}

// EmailMessage represents an email message
type EmailMessage struct {
	From        *mail.Address
	To          []*mail.Address
	Bcc         []*mail.Address
	Cc          []*mail.Address
	Subject     string
	HTML        string
	Text        string
	Headers     map[string]string
	Attachments map[string]io.Reader
}

// EmailProvider is the interface that all email providers must implement
type EmailProvider interface {
	// Send sends an email message
	Send(message *EmailMessage) error
	// Name returns the name of the provider
	Name() MailProvider
}

// EmailProviderFactory is a function that creates a new EmailProvider
type EmailProviderFactory func() (EmailProvider, error)
