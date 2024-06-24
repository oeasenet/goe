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

//type Sender struct {
//	client           *EmailClient
//	From             *mail.Address        `json:"from"`
//	ToAddresses      []*mail.Address      `json:"to"`
//	BccAddresses     []*mail.Address      `json:"bcc"`
//	CcAddresses      []*mail.Address      `json:"cc"`
//	EmailSubject     string               `json:"subject"`
//	BodyHTML         string               `json:"html"`
//	BodyText         string               `json:"text"`
//	EmailHeaders     map[string]string    `json:"headers"`
//	EmailAttachments map[string]io.Reader `json:"attachments"`
//}
//
//func (c *EmailClient) Sender() *Sender {
//	return &Sender{
//		client: c,
//		From:   &mail.Address{Name: c.fromName, Address: c.fromAddr},
//	}
//}
//
//func (s *Sender) To(t *[]*mail.Address) *Sender {
//	s.ToAddresses = *t
//	return s
//}
//
//func (s *Sender) Bcc(b *[]*mail.Address) *Sender {
//	s.BccAddresses = *b
//	return s
//}
//
//func (s *Sender) Cc(c *[]*mail.Address) *Sender {
//	s.CcAddresses = *c
//	return s
//}
//
//func (s *Sender) Subject(sub string) *Sender {
//	s.EmailSubject = sub
//	return s
//}
//
//func (s *Sender) HTML(html string) *Sender {
//	s.BodyHTML = html
//	return s
//}
//
//func (s *Sender) Text(text string) *Sender {
//	s.BodyText = text
//	return s
//}
//
//func (s *Sender) Headers(h map[string]string) *Sender {
//	s.EmailHeaders = h
//	return s
//}
//
//func (s *Sender) Attachments(a map[string]io.Reader) *Sender {
//	s.EmailAttachments = a
//	return s
//}
//
//func (s *Sender) Send() error {
//	return s.client.mailer.Send(&Message{
//		From:        s.From,
//		To:          s.ToAddresses,
//		Bcc:         s.BccAddresses,
//		Cc:          s.CcAddresses,
//		Subject:     s.EmailSubject,
//		HTML:        s.BodyHTML,
//		Text:        s.BodyText,
//		Headers:     s.EmailHeaders,
//		Attachments: s.EmailAttachments,
//	})
//}
