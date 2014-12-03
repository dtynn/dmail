package main

import (
	"github.com/dtynn/dmail/smtp/server"
	"github.com/ian-kent/Go-MailHog/data"
)

type faker struct {
	id      string
	message *data.SMTPMessage
}

func (this *faker) New(id string) (server.Receiver, error) {
	l.Info("Faker New", id)
	return &faker{
		id: id,
		message: &data.SMTPMessage{
			To: []string{},
		},
	}, nil
}

func (this *faker) Reset() error {
	l.Info("Faker Reset")
	return nil
}

func (this *faker) SetEhlo(local string) error {
	l.Info("Faker Set Ehlo", local)
	this.message.Helo = local
	this.message.To = []string{}
	return nil
}

func (this *faker) SetFrom(from string) error {
	l.Info("Faker Set From", from)
	this.message.From = from
	return nil
}

func (this *faker) AddRcpt(rcpt string) error {
	l.Info("Faker Add Rcpt", rcpt)
	this.message.To = append(this.message.To, rcpt)
	return nil
}

func (this *faker) SetData(data string) error {
	l.Info("Faker Set Data Length", len(data))
	this.message.Data = data
	return nil
}

func (this *faker) Close() error {
	l.Info("Faker Closed")
	parsed := this.message.Parse("dtynn.me")
	l.Info("From: ", parsed.From.Domain, parsed.From.Mailbox, parsed.From.Params, parsed.From.Relays)
	l.Info("Heaader: ", parsed.Content.Headers)
	return nil
}
