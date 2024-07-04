package contracts

import (
	"net/mail"
)

type Mailer interface {
	SMTPSender() EmailSender
}

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
