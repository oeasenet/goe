package mail

type EmailClient struct {
	mailer   *SmtpClient
	fromName string
	fromAddr string
}

func NewMailer(host string, port int, username string, password string, tls bool, fromName string, fromAddress string, localName string) *EmailClient {
	emailClient := &EmailClient{
		mailer: &SmtpClient{
			host:       host,
			port:       port,
			username:   username,
			password:   password,
			tls:        tls,
			authMethod: SmtpAuthPlain,
			localName:  localName,
		},
	}
	return emailClient
}

func (c *EmailClient) Send(m *Message) error {
	return c.mailer.Send(m)
}
