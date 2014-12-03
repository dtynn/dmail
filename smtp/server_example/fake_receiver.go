package main

import (
	"github.com/dtynn/dmail/smtp/server"
)

type faker struct {
	id                string
	local, from, data string
	rcpt              []string
}

func (this *faker) New(id string) (server.Receiver, error) {
	l.Info("Faker New", id)
	return &faker{
		id:   id,
		rcpt: []string{},
	}, nil
}

func (this *faker) Reset() error {
	l.Info("Faker Reset")
	return nil
}

func (this *faker) SetEhlo(local string) error {
	l.Info("Faker Set Ehlo", local)
	this.local = local
	this.rcpt = []string{}
	return nil
}

func (this *faker) SetFrom(from string) error {
	l.Info("Faker Set From", from)
	this.from = from
	return nil
}

func (this *faker) AddRcpt(rcpt string) error {
	l.Info("Faker Add Rcpt", rcpt)
	this.rcpt = append(this.rcpt, rcpt)
	return nil
}

func (this *faker) SetData(data string) error {
	l.Info("Faker Set Data Length", len(data))
	this.data = data
	return nil
}

func (this *faker) Close() error {
	l.Info(this.id, this.local, this.from, this.rcpt, len(this.data))
	l.Info("Faker Closed")
	return nil
}
