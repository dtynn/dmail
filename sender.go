package dmail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/dtynn/dmail/dkim"
	"github.com/dtynn/dmail/dns"
	"github.com/dtynn/dmail/message"
	. "github.com/dtynn/dmail/safeMap"
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
	retry     int
	encoding  message.Encoding
	charset   message.Charset
	enableTls bool
}

func NewSenderConfig(retry int, encoding message.Encoding, charset message.Charset, enableTls bool) *senderConfig {
	return &senderConfig{retry, encoding, charset, enableTls}
}

func NewDefaultSenderConfig(retry int, enableTls bool) *senderConfig {
	return NewSenderConfig(retry, message.Base64, message.CharsetUTF8, enableTls)
}

type Sender struct {
	conf     *senderConfig
	dkimConf *dkim.DkimConf
	dnsCache *SafeMap
}

type fail struct {
	Email, Detail string
}

type Fails []*fail

func (this Fails) Error() string {
	buf := bytes.NewBufferString("\n")
	for _, f := range this {
		buf.WriteString(fmt.Sprintf("%s: %s\n", f.Email, f.Detail))
	}
	return buf.String()
}

func NewSender(conf *senderConfig, dkimConf *dkim.DkimConf) *Sender {
	s := Sender{
		conf:     conf,
		dkimConf: dkimConf,
		dnsCache: NewSafeMap(),
	}
	return &s
}

func (this *Sender) Send(from string, to []string, subject string, body string) error {
	mail := NewMail(defaultContentType, from, to, subject, body)
	return this.SendMail(mail)
}

func (this *Sender) SendMail(mail *Mail) error {
	pieces := strings.Split(mail.From, "@")
	if len(pieces) != 2 || pieces[0] == "" || pieces[1] == "" {
		return errInvalidFromAddress
	}
	msg := message.NewMessage(this.conf.encoding, this.conf.charset, mail.ContentType)
	msg.AddContentType()
	msg.AddTransferEncodingHeader()
	msg.AddDate()

	msg.AddAddressHeader("From", mail.From, "")
	for _, t := range mail.To {
		msg.AddAddressHeader("To", t, "")
	}
	msg.AddNormalHeader("Subject", mail.Subject)
	msg.SetBody(mail.Body)

	if this.dkimConf != nil {
		d := dkim.NewDefaultDkim(msg, this.dkimConf)
		header, err := d.SignatureHeader()
		if err != nil {
			return err
		}
		msg.AddHeader(header)
	}

	b := msg.Bytes()
	fails := Fails{}
	for _, rcpt := range mail.To {
		piece := strings.Split(rcpt, "@")
		if len(piece) != 2 || piece[0] == "" || piece[1] == "" {
			fails = append(fails, &fail{rcpt, errInvalidRcptAddress.Error()})
			continue
		}
		err := this.send(piece[1], mail.From, []string{rcpt}, b)
		if err != nil {
			fails = append(fails, &fail{rcpt, err.Error()})
		}
	}
	if len(fails) == 0 {
		return nil
	}
	return fails
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

func (this *Sender) send(hostname, from string, to []string, msg []byte) error {
	local := strings.Split(from, "@")[1]
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
	log.Info("send: local ", local)
	log.Info("send: from ", from)
	log.Info("send: to ", to)
	log.Info("send: msg ", string(msg))
	log.Info("send: tls", tlsConfig)
	return smtp.SendEmail(addr, local, from, to, msg, tlsConfig)
}
