package providers

import (
	"crypto/tls"
	"fmt"
	"github.com/domodwyer/mailyak/v3"
	"go.oease.dev/goe/contracts"
	"go.oease.dev/goe/modules/mail/html2text"
	"go.oease.dev/goe/utils"
	"net/mail"
	"net/smtp"
	"strings"
)

// SMTPConfig holds the configuration for the SMTP provider
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	TLS        bool
	LocalName  string
	FromName   string
	FromEmail  string
	AuthMethod string
	DevMode    bool
}

// SMTPProvider implements the contracts.EmailProvider interface for SMTP
type SMTPProvider struct {
	config SMTPConfig
	logger contracts.Logger
}

// NewSMTPProvider creates a new SMTP provider
func NewSMTPProvider(config SMTPConfig, logger contracts.Logger) *SMTPProvider {
	return &SMTPProvider{
		config: config,
		logger: logger,
	}
}

// Send implements the contracts.EmailProvider interface
func (p *SMTPProvider) Send(message *contracts.EmailMessage) error {
	var smtpAuth smtp.Auth
	if p.config.Username != "" || p.config.Password != "" {
		switch p.config.AuthMethod {
		case "LOGIN":
			smtpAuth = &smtpLoginAuth{p.config.Username, p.config.Password}
		default:
			smtpAuth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
		}
	}

	// create mail instance
	var yak *mailyak.MailYak
	if p.config.TLS {
		// if dev skip cert check
		var tlscfg *tls.Config
		if p.config.DevMode {
			tlscfg = &tls.Config{
				ServerName:         p.config.Host,
				InsecureSkipVerify: true,
			}
		} else {
			tlscfg = nil
		}
		var tlsErr error
		yak, tlsErr = mailyak.NewWithTLS(fmt.Sprintf("%s:%d", p.config.Host, p.config.Port), smtpAuth, tlscfg)
		if tlsErr != nil {
			return tlsErr
		}
	} else {
		yak = mailyak.New(fmt.Sprintf("%s:%d", p.config.Host, p.config.Port), smtpAuth)
	}

	if p.config.LocalName != "" {
		yak.LocalName(p.config.LocalName)
	}

	if message.From.Name != "" {
		yak.FromName(message.From.Name)
	}
	yak.From(message.From.Address)
	yak.Subject(message.Subject)
	yak.HTML().Set(message.HTML)

	if message.Text == "" {
		// try to generate a plain text version of the HTML
		plain, err := html2text.FromString(message.HTML)
		if err != nil {
			return err
		}
		yak.Plain().Set(plain)
	} else {
		yak.Plain().Set(message.Text)
	}

	if len(message.To) > 0 {
		yak.To(addressesToStrings(message.To, true)...)
	}

	if len(message.Bcc) > 0 {
		yak.Bcc(addressesToStrings(message.Bcc, true)...)
	}

	if len(message.Cc) > 0 {
		yak.Cc(addressesToStrings(message.Cc, true)...)
	}

	// add attachements (if any)
	for name, data := range message.Attachments {
		yak.Attach(name, data)
	}

	// add custom headers (if any)
	var hasMessageId bool
	for k, v := range message.Headers {
		if strings.EqualFold(k, "Message-ID") {
			hasMessageId = true
		}
		yak.AddHeader(k, v)
	}
	if !hasMessageId {
		// add a default message id if missing
		fromParts := strings.Split(message.From.Address, "@")
		if len(fromParts) == 2 {
			yak.AddHeader("Message-ID", fmt.Sprintf("<%s@%s>",
				utils.GenXid(),
				fromParts[1],
			))
		}
	}

	return yak.Send()
}

// Name implements the contracts.EmailProvider interface
func (p *SMTPProvider) Name() contracts.MailProvider {
	return contracts.ProviderSMTP
}

// smtpLoginAuth implements the smtp.Auth interface for LOGIN authentication
type smtpLoginAuth struct {
	username, password string
}

// Start initializes an authentication with the server.
func (a *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server.
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, fmt.Errorf("unencrypted connection")
	}

	return "LOGIN", nil, nil
}

// Next continues the authentication process
func (a *smtpLoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		}
	}

	return nil, nil
}

// isLocalhost checks if the given name is localhost
func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}

// addressesToStrings converts mail.Address to strings
func addressesToStrings(addresses []*mail.Address, withName bool) []string {
	result := make([]string, len(addresses))

	for i, addr := range addresses {
		if withName && addr.Name != "" {
			result[i] = addr.String()
		} else {
			// keep only the email part to avoid wrapping in angle-brackets
			result[i] = addr.Address
		}
	}

	return result
}
