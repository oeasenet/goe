package mail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/domodwyer/mailyak/v3"
	"go.oease.dev/goe/internal/mail/html2text"
	"go.oease.dev/goe/internal/utils"
	"io"
	"net/mail"
	"net/smtp"
	"strings"
)

// Mailer defines a base mail client interface.
type Mailer interface {
	// Send sends an email with the provided Message.
	Send(message *Message) error
}

var _ Mailer = (*SmtpClient)(nil)

const (
	SmtpAuthPlain = "PLAIN"
	SmtpAuthLogin = "LOGIN"
)

// Message defines a generic email message struct.
type Message struct {
	From        *mail.Address        `json:"from"`
	To          []*mail.Address      `json:"to"`
	Bcc         []*mail.Address      `json:"bcc"`
	Cc          []*mail.Address      `json:"cc"`
	Subject     string               `json:"subject"`
	HTML        string               `json:"html"`
	Text        string               `json:"text"`
	Headers     map[string]string    `json:"headers"`
	Attachments map[string]io.Reader `json:"attachments"`
}

// SmtpClient defines a SMTP mail client structure that implements
// `mailer.Mailer` interface.
type SmtpClient struct {
	host     string
	port     int
	username string
	password string
	tls      bool

	// SMTP auth method to use
	// (if not explicitly set, defaults to "PLAIN")
	authMethod string

	// localName is optional domain name used for the EHLO/HELO exchange
	// (if not explicitly set, defaults to "localhost").
	//
	// This is required only by some SMTP servers, such as Gmail SMTP-relay.
	localName string

	devMode bool
}

// Send implements `mailer.Mailer` interface.
func (c *SmtpClient) Send(m *Message) error {
	var smtpAuth smtp.Auth
	if c.username != "" || c.password != "" {
		switch c.authMethod {
		case SmtpAuthLogin:
			smtpAuth = &smtpLoginAuth{c.username, c.password}
		default:
			smtpAuth = smtp.PlainAuth("", c.username, c.password, c.host)
		}
	}

	// create mail instance
	var yak *mailyak.MailYak
	if c.tls {
		// if dev skip cert check
		var tlscfg *tls.Config
		if c.devMode {
			tlscfg = &tls.Config{
				ServerName:         c.host,
				InsecureSkipVerify: true,
			}
		} else {
			tlscfg = nil
		}
		var tlsErr error
		yak, tlsErr = mailyak.NewWithTLS(fmt.Sprintf("%s:%d", c.host, c.port), smtpAuth, tlscfg)
		if tlsErr != nil {
			return tlsErr
		}
	} else {
		yak = mailyak.New(fmt.Sprintf("%s:%d", c.host, c.port), smtpAuth)
	}

	if c.localName != "" {
		yak.LocalName(c.localName)
	}

	if m.From.Name != "" {
		yak.FromName(m.From.Name)
	}
	yak.From(m.From.Address)
	yak.Subject(m.Subject)
	yak.HTML().Set(m.HTML)

	if m.Text == "" {
		// try to generate a plain text version of the HTML
		plain, err := html2text.FromString(m.HTML)
		if err != nil {
			return err
		}
		yak.Plain().Set(plain)
	} else {
		yak.Plain().Set(m.Text)
	}

	if len(m.To) > 0 {
		yak.To(addressesToStrings(m.To, true)...)
	}

	if len(m.Bcc) > 0 {
		yak.Bcc(addressesToStrings(m.Bcc, true)...)
	}

	if len(m.Cc) > 0 {
		yak.Cc(addressesToStrings(m.Cc, true)...)
	}

	// add attachements (if any)
	for name, data := range m.Attachments {
		yak.Attach(name, data)
	}

	// add custom headers (if any)
	var hasMessageId bool
	for k, v := range m.Headers {
		if strings.EqualFold(k, "Message-ID") {
			hasMessageId = true
		}
		yak.AddHeader(k, v)
	}
	if !hasMessageId {
		// add a default message id if missing
		fromParts := strings.Split(m.From.Address, "@")
		if len(fromParts) == 2 {
			yak.AddHeader("Message-ID", fmt.Sprintf("<%s@%s>",
				utils.GenXid(),
				fromParts[1],
			))
		}
	}

	return yak.Send()
}

// -------------------------------------------------------------------
// AUTH LOGIN
// -------------------------------------------------------------------

var _ smtp.Auth = (*smtpLoginAuth)(nil)

// smtpLoginAuth defines an AUTH that implements the LOGIN authentication mechanism.
//
// AUTH LOGIN is obsolete[1] but some mail services like outlook requires it [2].
//
// NB!
// It will only send the credentials if the connection is using TLS or is connected to localhost.
// Otherwise authentication will fail with an error, without sending the credentials.
//
// [1]: https://github.com/golang/go/issues/40817
// [2]: https://support.microsoft.com/en-us/office/outlook-com-no-longer-supports-auth-plain-authentication-07f7d5e9-1697-465f-84d2-4513d4ff0145?ui=en-us&rs=en-us&ad=us
type smtpLoginAuth struct {
	username, password string
}

// Start initializes an authentication with the server.
//
// It is part of the [smtp.Auth] interface.
func (a *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server.
	// Note: If TLS is not true, then we can't trust ANYTHING in ServerInfo.
	// In particular, it doesn't matter if the server advertises LOGIN auth.
	// That might just be the attacker saying
	// "it's ok, you can trust me with your password."
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}

	return "LOGIN", nil, nil
}

// Next "continues" the auth process by feeding the server with the requested data.
//
// It is part of the [smtp.Auth] interface.
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

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}

// addressesToStrings converts the provided address to a list of serialized RFC 5322 strings.
//
// To export only the email part of mail.Address, you can set withName to false.
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
