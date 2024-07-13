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
		From:        s.from,
		To:          s.toAddresses,
		Bcc:         s.bccAddresses,
		Cc:          s.ccAddresses,
		Subject:     s.emailSubject,
		HTML:        s.bodyHTML,
		Text:        s.bodyText,
		Headers:     s.emailHeaders,
		Attachments: map[string]io.Reader{},
	}
	// process attachments
	for name, path := range s.emailAttachments {
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
	from             *mail.Address     `json:"from"`
	toAddresses      []*mail.Address   `json:"to"`
	bccAddresses     []*mail.Address   `json:"bcc"`
	ccAddresses      []*mail.Address   `json:"cc"`
	emailSubject     string            `json:"subject"`
	bodyHTML         string            `json:"html"`
	bodyText         string            `json:"text"`
	emailHeaders     map[string]string `json:"headers"`
	emailAttachments map[string]string `json:"attachments"`
}

func (s *sender) To(t *[]*mail.Address) contracts.EmailSender {
	s.toAddresses = *t
	return s
}

func (s *sender) Bcc(b *[]*mail.Address) contracts.EmailSender {
	s.bccAddresses = *b
	return s
}

func (s *sender) Cc(c *[]*mail.Address) contracts.EmailSender {
	s.ccAddresses = *c
	return s
}

func (s *sender) Subject(sub string) contracts.EmailSender {
	s.emailSubject = sub
	return s
}

func (s *sender) HTML(html string) contracts.EmailSender {
	s.bodyHTML = html
	return s
}

func (s *sender) Text(text string) contracts.EmailSender {
	s.bodyText = text
	return s
}

func (s *sender) Headers(h map[string]string) contracts.EmailSender {
	s.emailHeaders = h
	return s
}

func (s *sender) Attachments(a map[string]string) contracts.EmailSender {
	s.emailAttachments = a
	return s
}

func (s *sender) Send(useQueue ...bool) error {
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
			From:        s.from,
			To:          s.toAddresses,
			Bcc:         s.bccAddresses,
			Cc:          s.ccAddresses,
			Subject:     s.emailSubject,
			HTML:        s.bodyHTML,
			Text:        s.bodyText,
			Headers:     s.emailHeaders,
			Attachments: map[string]io.Reader{},
		}
		// process attachments
		for name, path := range s.emailAttachments {
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
