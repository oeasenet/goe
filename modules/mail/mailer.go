package mail

import (
	"encoding/json"
	"fmt"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/mail/providers"
	"go.oease.dev/goe/utils"
	"io"
	"net/mail"
	"sync"
)

// EmailDeliveryQueueName is the queue name for email delivery
var EmailDeliveryQueueName contracts.QueueName = "goe.mailer.send"

// MailerManager implements the contracts.Mailer interface
type MailerManager struct {
	providers       map[contracts.MailProvider]contracts.EmailProvider
	defaultProvider contracts.MailProvider
	logger          contracts.Logger
	queue           contracts.Queue
	fromName        string
	fromEmail       string
	mu              sync.RWMutex
}

// NewMailerManager creates a new MailerManager
func NewMailerManager(logger contracts.Logger, queue contracts.Queue, fromName, fromEmail string) *MailerManager {
	return &MailerManager{
		providers:       make(map[contracts.MailProvider]contracts.EmailProvider),
		defaultProvider: contracts.ProviderSMTP, // Default to SMTP
		logger:          logger,
		queue:           queue,
		fromName:        fromName,
		fromEmail:       fromEmail,
		mu:              sync.RWMutex{},
	}
}

// RegisterProvider registers a new email provider
func (m *MailerManager) RegisterProvider(provider contracts.MailProvider, factory contracts.EmailProviderFactory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create the provider
	p, err := factory()
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", provider, err)
	}

	// Register the provider
	m.providers[provider] = p
	return nil
}

// SetDefaultProvider sets the default email provider
func (m *MailerManager) SetDefaultProvider(provider contracts.MailProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.providers[provider]; !ok {
		return fmt.Errorf("provider %s not registered", provider)
	}

	m.defaultProvider = provider
	return nil
}

// GetSender returns an EmailSender for the specified provider
func (m *MailerManager) GetSender(provider contracts.MailProvider) contracts.EmailSender {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &emailSender{
		manager:  m,
		provider: provider,
		from: &mail.Address{
			Name:    m.fromName,
			Address: m.fromEmail,
		},
	}
}

// DefaultSender returns the default EmailSender
func (m *MailerManager) DefaultSender() contracts.EmailSender {
	return m.GetSender(m.defaultProvider)
}

// getProvider returns the specified provider or the default if not found
func (m *MailerManager) getProvider(provider contracts.MailProvider) (contracts.EmailProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, ok := m.providers[provider]; ok {
		return p, nil
	}

	if p, ok := m.providers[m.defaultProvider]; ok {
		m.logger.Warnf("Provider %s not found, using default provider %s", provider, m.defaultProvider)
		return p, nil
	}

	return nil, fmt.Errorf("no providers registered")
}

// emailSender implements the contracts.EmailSender interface
type emailSender struct {
	manager     *MailerManager
	provider    contracts.MailProvider
	from        *mail.Address
	to          []*mail.Address
	cc          []*mail.Address
	bcc         []*mail.Address
	subject     string
	htmlBody    string
	textBody    string
	headers     map[string]string
	attachments map[string]string
}

// To sets the recipients
func (s *emailSender) To(t *[]*mail.Address) contracts.EmailSender {
	s.to = *t
	return s
}

// Bcc sets the BCC recipients
func (s *emailSender) Bcc(b *[]*mail.Address) contracts.EmailSender {
	s.bcc = *b
	return s
}

// Cc sets the CC recipients
func (s *emailSender) Cc(c *[]*mail.Address) contracts.EmailSender {
	s.cc = *c
	return s
}

// Subject sets the subject
func (s *emailSender) Subject(sub string) contracts.EmailSender {
	s.subject = sub
	return s
}

// HTML sets the HTML body
func (s *emailSender) HTML(html string) contracts.EmailSender {
	s.htmlBody = html
	return s
}

// Text sets the text body
func (s *emailSender) Text(text string) contracts.EmailSender {
	s.textBody = text
	return s
}

// Headers sets the headers
func (s *emailSender) Headers(h map[string]string) contracts.EmailSender {
	s.headers = h
	return s
}

// Attachments sets the attachments
func (s *emailSender) Attachments(a map[string]string) contracts.EmailSender {
	s.attachments = a
	return s
}

// Send sends the email
func (s *emailSender) Send(useQueue ...bool) error {
	// Get the provider
	provider, err := s.manager.getProvider(s.provider)
	if err != nil {
		return err
	}

	// Create the message
	message := &contracts.EmailMessage{
		From:        s.from,
		To:          s.to,
		Cc:          s.cc,
		Bcc:         s.bcc,
		Subject:     s.subject,
		HTML:        s.htmlBody,
		Text:        s.textBody,
		Headers:     s.headers,
		Attachments: make(map[string]io.Reader),
	}

	// Process attachments
	for name, path := range s.attachments {
		reader, err := utils.FilePathToIOReader(path)
		if err != nil {
			return fmt.Errorf("failed to read attachment %s: %w", name, err)
		}
		message.Attachments[name] = reader
	}

	// Default to use queue unless specified not to use
	isQueue := true
	if len(useQueue) > 0 {
		isQueue = useQueue[0]
	}

	if isQueue {
		// Queue the message for sending
		// Note: We can't directly queue the message because it contains io.Reader which can't be serialized
		// Instead, we'll queue a simplified version and process the attachments when consuming
		return s.manager.queueMessage(s.provider, message, s.attachments)
	}

	// Send the message directly
	return provider.Send(message)
}

// queueMessage queues a message for sending
func (m *MailerManager) queueMessage(provider contracts.MailProvider, message *contracts.EmailMessage, attachments map[string]string) error {
	// Create a queue-friendly version of the message
	queuedMessage := struct {
		Provider    contracts.MailProvider `json:"provider"`
		From        *mail.Address          `json:"from"`
		To          []*mail.Address        `json:"to"`
		Cc          []*mail.Address        `json:"cc"`
		Bcc         []*mail.Address        `json:"bcc"`
		Subject     string                 `json:"subject"`
		HTML        string                 `json:"html"`
		Text        string                 `json:"text"`
		Headers     map[string]string      `json:"headers"`
		Attachments map[string]string      `json:"attachments"`
	}{
		Provider:    provider,
		From:        message.From,
		To:          message.To,
		Cc:          message.Cc,
		Bcc:         message.Bcc,
		Subject:     message.Subject,
		HTML:        message.HTML,
		Text:        message.Text,
		Headers:     message.Headers,
		Attachments: attachments, // Use the original file paths
	}

	// Push to queue
	return m.queue.Push(EmailDeliveryQueueName, queuedMessage)
}

// ProcessQueuedMessage processes a queued message
func (m *MailerManager) ProcessQueuedMessage(payload string) bool {
	// Parse the queued message
	var queuedMessage struct {
		Provider    contracts.MailProvider `json:"provider"`
		From        *mail.Address          `json:"from"`
		To          []*mail.Address        `json:"to"`
		Cc          []*mail.Address        `json:"cc"`
		Bcc         []*mail.Address        `json:"bcc"`
		Subject     string                 `json:"subject"`
		HTML        string                 `json:"html"`
		Text        string                 `json:"text"`
		Headers     map[string]string      `json:"headers"`
		Attachments map[string]string      `json:"attachments"`
	}

	if err := json.Unmarshal([]byte(payload), &queuedMessage); err != nil {
		m.logger.Error("Failed to unmarshal queued message: ", err)
		return false
	}

	// Get the provider
	provider, err := m.getProvider(queuedMessage.Provider)
	if err != nil {
		m.logger.Error("Failed to get provider: ", err)
		return false
	}

	// Create the message
	message := &contracts.EmailMessage{
		From:        queuedMessage.From,
		To:          queuedMessage.To,
		Cc:          queuedMessage.Cc,
		Bcc:         queuedMessage.Bcc,
		Subject:     queuedMessage.Subject,
		HTML:        queuedMessage.HTML,
		Text:        queuedMessage.Text,
		Headers:     queuedMessage.Headers,
		Attachments: make(map[string]io.Reader),
	}

	// Process attachments
	for name, path := range queuedMessage.Attachments {
		reader, err := utils.FilePathToIOReader(path)
		if err != nil {
			m.logger.Error("Failed to read attachment: ", err)
			return false
		}
		message.Attachments[name] = reader
	}

	// Send the message
	if err := provider.Send(message); err != nil {
		m.logger.Error("Failed to send queued message: ", err)
		return false
	}

	return true
}

// Legacy support for the old EmailClient interface

// EmailClient is the legacy email client
type EmailClient struct {
	manager  *MailerManager
	FromName string
	FromAddr string
}

// NewMailer creates a new legacy email client
func NewMailer(host string, port int, username string, password string, tls bool, fromName string, fromAddress string, localName string) *EmailClient {
	// Create a new mailer manager
	manager := NewMailerManager(nil, nil, fromName, fromAddress)

	// Register the SMTP provider
	smtpConfig := providers.SMTPConfig{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		TLS:        tls,
		LocalName:  localName,
		FromName:   fromName,
		FromEmail:  fromAddress,
		AuthMethod: "PLAIN",
	}

	smtpProvider := providers.NewSMTPProvider(smtpConfig, nil)
	manager.providers[contracts.ProviderSMTP] = smtpProvider

	return &EmailClient{
		manager:  manager,
		FromName: fromName,
		FromAddr: fromAddress,
	}
}

// Send sends an email using the legacy interface
func (c *EmailClient) Send(m *Message) error {
	// Convert the legacy message to the new format
	message := &contracts.EmailMessage{
		From: &mail.Address{
			Name:    c.FromName,
			Address: c.FromAddr,
		},
		To:          m.To,
		Cc:          m.Cc,
		Bcc:         m.Bcc,
		Subject:     m.Subject,
		HTML:        m.HTML,
		Text:        m.Text,
		Headers:     m.Headers,
		Attachments: m.Attachments,
	}

	// Get the SMTP provider
	provider, err := c.manager.getProvider(contracts.ProviderSMTP)
	if err != nil {
		return err
	}

	// Send the message
	return provider.Send(message)
}
