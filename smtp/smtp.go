package smtp

import (
	"crypto/tls"
	"net/smtp"
)

const DefaultPort = 25

func SendEmail(addr, local, from string, to []string, msg []byte, tlsConfig *tls.Config) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}

	defer c.Close()

	if err := c.Hello(local); err != nil {
		return err
	}

	if tlsConfig != nil {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsConfig); err != nil {
				return err
			}
		}
	}

	if err := c.Mail(from); err != nil {
		return err
	}

	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
