package dmail

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/dtynn/dmail/dkim"
	"github.com/dtynn/dmail/dns"
	"github.com/dtynn/dmail/message"
	"github.com/dtynn/dmail/smtp"
	"github.com/qiniu/log"
)

var (
	errInvalidFromAddress = fmt.Errorf("invalid from address")
	errInvalidRcptAddress = fmt.Errorf("invalid rcpt address")
)

const (
	cacheExpire        int64 = 60 * 5
	defaultContentType       = "text/html"
)

type senderConfig struct {
	from      string
	local     string
	retry     int
	encoding  message.Encoding
	charset   message.Charset
	enableTls bool
}

func NewSenderConfig(from string, retry int, encoding message.Encoding, charset message.Charset, enableTls bool) (*senderConfig, error) {
	pieces := strings.Split(from, "@")
	if len(pieces) != 2 || pieces[0] == "" || pieces[1] == "" {
		return nil, errInvalidFromAddress
	}
	return &senderConfig{from, pieces[1], retry, encoding, charset, enableTls}, nil
}

func NewDefaultSenderConfig(from string, retry int, enableTls bool) (*senderConfig, error) {
	return NewSenderConfig(from, retry, message.Base64, message.CharsetUTF8, enableTls)
}

type Sender struct {
	conf     *senderConfig
	dkimConf *dkim.DkimConf
	dnsCache *safeMap
}

type fail struct {
	Email, Detail string
}

func NewSender(conf *senderConfig, dkimConf *dkim.DkimConf) *Sender {
	s := Sender{
		conf:     conf,
		dkimConf: dkimConf,
		dnsCache: NewSafeMap(),
	}
	return &s
}

func (this *Sender) Send(to []string, subject string, body string) error {
	mail := NewMail(defaultContentType, to, subject, body)
	_, err := this.SendMail(mail)
	return err
}

func (this *Sender) SendMail(mail *Mail) ([]*fail, error) {
	msg := message.NewMessage(this.conf.encoding, this.conf.charset, mail.ContentType)
	msg.AddContentType()
	msg.AddTransferEncodingHeader()
	msg.AddDate()

	msg.AddAddressHeader("From", this.conf.from, "")
	for _, t := range mail.To {
		msg.AddAddressHeader("To", t, "")
	}
	msg.AddNormalHeader("Subject", mail.Subject)
	msg.SetBody(mail.Body)

	if this.dkimConf != nil {
		d := dkim.NewDefaultDkim(msg, this.dkimConf)
		header, err := d.SignatureHeader()
		if err != nil {
			return nil, err
		}
		msg.AddHeader(header)
	}

	b := msg.Bytes()
	fails := make([]*fail, 0)
	for _, rcpt := range mail.To {
		piece := strings.Split(rcpt, "@")
		if len(piece) != 2 || piece[0] == "" || piece[1] == "" {
			fails = append(fails, &fail{rcpt, errInvalidRcptAddress.Error()})
			continue
		}
		err := this.send(piece[1], []string{rcpt}, b)
		if err != nil {
			fails = append(fails, &fail{rcpt, err.Error()})
		}
	}
	return fails, nil
}

func (this *Sender) getMxHost(name string) (string, error) {
	if host, err := this.dnsCache.Get(name); err == nil {
		return host.(string), nil
	}

	mx, err := dns.ChoseMx(name)
	if err != nil {
		return "", err
	}
	this.dnsCache.Setex(name, mx.Host, cacheExpire)
	return mx.Host, nil
}

func (this *Sender) send(hostname string, to []string, msg []byte) error {
	host, err := this.getMxHost(hostname)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", host, smtp.DefaultPort)

	var tlsConfig *tls.Config
	if this.conf.enableTls {
		tlsConfig = &tls.Config{
			ServerName: hostname,
		}
	} else {
		tlsConfig = nil
	}

	log.Info("send: addr ", addr)
	log.Info("send: local ", this.conf.local)
	log.Info("send: from ", this.conf.from)
	log.Info("send: to ", to)
	log.Info("send: msg ", string(msg))
	log.Info("send: tls", tlsConfig)
	return smtp.SendEmail(addr, this.conf.local, this.conf.from, to, msg, tlsConfig)
}
