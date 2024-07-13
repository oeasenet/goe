package core

import (
	"github.com/goccy/go-json"
	"go.oease.dev/goe/contracts"
	moduleMail "go.oease.dev/goe/modules/mail"
	"go.oease.dev/goe/utils"
	"io"
	"net/mail"
)

var emailDeliveryQueueName contracts.QueueName = "goe.mailer.send"

type GoeMailer struct {
	appConfig *GoeConfig
	logger    contracts.Logger
	client    *moduleMail.EmailClient
	queue     contracts.Queue
}

func NewGoeMailer(appConfig *GoeConfig, queueInstance contracts.Queue, logger contracts.Logger) *GoeMailer {
	if appConfig.Features.SMTPMailerEnabled {
		if appConfig.Mailer.Host != "" && appConfig.Mailer.Port != 0 && appConfig.Mailer.Username != "" && appConfig.Mailer.Password != "" && appConfig.Mailer.FromName != "" && appConfig.Mailer.FromEmail != "" {
			m := &GoeMailer{
				appConfig: appConfig,
				logger:    logger,
				client:    moduleMail.NewMailer(appConfig.Mailer.Host, appConfig.Mailer.Port, appConfig.Mailer.Username, appConfig.Mailer.Password, appConfig.Mailer.Tls, appConfig.Mailer.FromName, appConfig.Mailer.FromEmail, appConfig.Mailer.LocalName),
			}
			queueInstance.NewQueue(emailDeliveryQueueName, m.sendMailQueueConsumer)
			m.queue = queueInstance
			return m
		} else {
			logger.Error("failed to initialize SMTP mailer: missing required SMTP configuration")
		}
	}
	return nil
}

func (g *GoeMailer) SMTPSender() contracts.EmailSender {
	return &sender{
		ml: g,
	}
}

func (g *GoeMailer) sendMailQueueConsumer(payload string) bool {
	// process email payload
	s := &sender{}
	err := json.Unmarshal([]byte(payload), s)
	if err != nil {
		g.logger.Error("failed to unmarshal email payload: ", err)
		return false
	}
	// make msg and process file attachments
	// send email directly
	msg := &moduleMail.Message{
		From:        s.From,
		To:          s.ToAddresses,
		Bcc:         s.BccAddresses,
		Cc:          s.CcAddresses,
		Subject:     s.EmailSubject,
		HTML:        s.BodyHTML,
		Text:        s.BodyText,
		Headers:     s.EmailHeaders,
		Attachments: map[string]io.Reader{},
	}
	// process attachments
	for name, path := range s.EmailAttachments {
		read, err := utils.FilePathToIOReader(path)
		if err != nil {
			g.logger.Error("failed to read attachment file: ", err)
			return false
		}
		msg.Attachments[name] = read
	}

	//send
	err = g.client.Send(msg)
	if err != nil {
		g.logger.Error("failed to send email: ", err)
		return false
	}

	return true
}

type sender struct {
	ml               *GoeMailer        `json:"-"`
	From             *mail.Address     `json:"From"`
	ToAddresses      []*mail.Address   `json:"to"`
	BccAddresses     []*mail.Address   `json:"bcc"`
	CcAddresses      []*mail.Address   `json:"cc"`
	EmailSubject     string            `json:"subject"`
	BodyHTML         string            `json:"html"`
	BodyText         string            `json:"text"`
	EmailHeaders     map[string]string `json:"headers"`
	EmailAttachments map[string]string `json:"attachments"`
}

func (s *sender) To(t *[]*mail.Address) contracts.EmailSender {
	s.ToAddresses = *t
	return s
}

func (s *sender) Bcc(b *[]*mail.Address) contracts.EmailSender {
	s.BccAddresses = *b
	return s
}

func (s *sender) Cc(c *[]*mail.Address) contracts.EmailSender {
	s.CcAddresses = *c
	return s
}

func (s *sender) Subject(sub string) contracts.EmailSender {
	s.EmailSubject = sub
	return s
}

func (s *sender) HTML(html string) contracts.EmailSender {
	s.BodyHTML = html
	return s
}

func (s *sender) Text(text string) contracts.EmailSender {
	s.BodyText = text
	return s
}

func (s *sender) Headers(h map[string]string) contracts.EmailSender {
	s.EmailHeaders = h
	return s
}

func (s *sender) Attachments(a map[string]string) contracts.EmailSender {
	s.EmailAttachments = a
	return s
}

func (s *sender) Send(useQueue ...bool) error {
	s.From = &mail.Address{Name: s.ml.appConfig.Mailer.FromName, Address: s.ml.appConfig.Mailer.FromEmail}
	//default to use queue unless specified not to use
	isQueue := true
	if len(useQueue) > 0 {
		isQueue = useQueue[0]
	}
	if isQueue {
		// publish message to queue
		return s.ml.queue.Push(emailDeliveryQueueName, s)
	} else {
		// send email directly
		msg := &moduleMail.Message{
			From:        s.From,
			To:          s.ToAddresses,
			Bcc:         s.BccAddresses,
			Cc:          s.CcAddresses,
			Subject:     s.EmailSubject,
			HTML:        s.BodyHTML,
			Text:        s.BodyText,
			Headers:     s.EmailHeaders,
			Attachments: map[string]io.Reader{},
		}
		// process attachments
		for name, path := range s.EmailAttachments {
			read, err := utils.FilePathToIOReader(path)
			if err != nil {
				return err
			}
			msg.Attachments[name] = read
		}
		return s.ml.client.Send(msg)
	}
	return nil
}
