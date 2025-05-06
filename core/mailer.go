package core

import (
	"go.oease.dev/goe/contracts"
	mailModule "go.oease.dev/goe/modules/mail"
	"go.oease.dev/goe/modules/mail/providers"
)

// GoeMailer implements the contracts.Mailer interface
type GoeMailer struct {
	manager *mailModule.MailerManager
}

// NewGoeMailer creates a new GoeMailer
func NewGoeMailer(appConfig *GoeConfig, queueInstance contracts.Queue, logger contracts.Logger) *GoeMailer {
	if appConfig.Features.MailerEnabled {
		// Check if we have the required configuration
		if appConfig.Mailer.FromName == "" || appConfig.Mailer.FromEmail == "" {
			logger.Error("failed to initialize mailer: missing required from name or email")
			return nil
		}

		// Create a new mailer manager
		manager := mailModule.NewMailerManager(logger, queueInstance, appConfig.Mailer.FromName, appConfig.Mailer.FromEmail)

		// Register providers based on configuration
		provider := appConfig.Mailer.Provider
		if provider == "" {
			provider = string(contracts.ProviderSMTP) // Default to SMTP if not specified
		}

		// Register the configured provider
		switch contracts.MailProvider(provider) {
		case contracts.ProviderSMTP:
			if appConfig.Mailer.SMTP == nil ||
				appConfig.Mailer.SMTP.Host == "" ||
				appConfig.Mailer.SMTP.Port == 0 ||
				appConfig.Mailer.SMTP.Username == "" ||
				appConfig.Mailer.SMTP.Password == "" {
				logger.Error("failed to initialize SMTP mailer: missing required SMTP configuration")
				return nil
			}

			// Register the SMTP provider
			smtpConfig := providers.SMTPConfig{
				Host:       appConfig.Mailer.SMTP.Host,
				Port:       appConfig.Mailer.SMTP.Port,
				Username:   appConfig.Mailer.SMTP.Username,
				Password:   appConfig.Mailer.SMTP.Password,
				TLS:        appConfig.Mailer.SMTP.Tls,
				LocalName:  appConfig.Mailer.SMTP.LocalName,
				FromName:   appConfig.Mailer.FromName,
				FromEmail:  appConfig.Mailer.FromEmail,
				AuthMethod: appConfig.Mailer.SMTP.AuthMethod,
			}

			// Create the SMTP provider
			smtpProvider := providers.NewSMTPProvider(smtpConfig, logger)

			// Register the provider with the manager
			err := manager.RegisterProvider(contracts.ProviderSMTP, func() (contracts.EmailProvider, error) {
				return smtpProvider, nil
			})

			if err != nil {
				logger.Error("Failed to register SMTP provider: ", err)
				return nil
			}

			// Set the default provider
			manager.SetDefaultProvider(contracts.ProviderSMTP)

		case contracts.ProviderResend:
			if appConfig.Mailer.Resend == nil || appConfig.Mailer.Resend.APIKey == "" {
				logger.Error("failed to initialize Resend mailer: missing required Resend API key")
				return nil
			}

			// Register the Resend provider
			resendConfig := providers.ResendConfig{
				APIKey:    appConfig.Mailer.Resend.APIKey,
				FromName:  appConfig.Mailer.FromName,
				FromEmail: appConfig.Mailer.FromEmail,
			}

			// Create the Resend provider
			resendProvider := providers.NewResendProvider(resendConfig, logger)

			// Register the provider with the manager
			err := manager.RegisterProvider(contracts.ProviderResend, func() (contracts.EmailProvider, error) {
				return resendProvider, nil
			})

			if err != nil {
				logger.Error("Failed to register Resend provider: ", err)
				return nil
			}

			// Set the default provider
			manager.SetDefaultProvider(contracts.ProviderResend)

		case contracts.ProviderSES:
			if appConfig.Mailer.SES == nil ||
				appConfig.Mailer.SES.Region == "" ||
				appConfig.Mailer.SES.AccessKeyID == "" ||
				appConfig.Mailer.SES.SecretAccessKey == "" {
				logger.Error("failed to initialize SES mailer: missing required SES configuration")
				return nil
			}

			// Register the SES provider
			sesConfig := providers.SESConfig{
				Region:          appConfig.Mailer.SES.Region,
				AccessKeyID:     appConfig.Mailer.SES.AccessKeyID,
				SecretAccessKey: appConfig.Mailer.SES.SecretAccessKey,
				Endpoint:        appConfig.Mailer.SES.Endpoint,
				FromName:        appConfig.Mailer.FromName,
				FromEmail:       appConfig.Mailer.FromEmail,
			}

			// Create the SES provider
			sesProvider := providers.NewSESProvider(sesConfig, logger)

			// Register the provider with the manager
			err := manager.RegisterProvider(contracts.ProviderSES, func() (contracts.EmailProvider, error) {
				return sesProvider, nil
			})

			if err != nil {
				logger.Error("Failed to register SES provider: ", err)
				return nil
			}

			// Set the default provider
			manager.SetDefaultProvider(contracts.ProviderSES)

		default:
			logger.Errorf("Unsupported mail provider: %s", provider)
			return nil
		}

		// Set up the queue consumer
		queueInstance.NewQueue(mailModule.EmailDeliveryQueueName, manager.ProcessQueuedMessage)

		return &GoeMailer{
			manager: manager,
		}
	}
	return nil
}

// GetSender returns an EmailSender for the specified provider
func (g *GoeMailer) GetSender(provider contracts.MailProvider) contracts.EmailSender {
	return g.manager.GetSender(provider)
}

// DefaultSender returns the default EmailSender
func (g *GoeMailer) DefaultSender() contracts.EmailSender {
	return g.manager.DefaultSender()
}

// RegisterProvider registers a new email provider
func (g *GoeMailer) RegisterProvider(provider contracts.MailProvider, factory contracts.EmailProviderFactory) error {
	return g.manager.RegisterProvider(provider, factory)
}

// Legacy support for the old interface

// SMTPSender returns an EmailSender for SMTP
func (g *GoeMailer) SMTPSender() contracts.EmailSender {
	return g.GetSender(contracts.ProviderSMTP)
}
