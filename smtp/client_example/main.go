package main

import (
	"fmt"

	"github.com/dtynn/dmail/smtp"
)

func main() {
	err := smtp.SendEmail("127.0.0.1:25", "dtynn.me", "a@a.a",
		[]string{"b@b.b"}, []byte("From: <a@a.a>\r\nTo: <b@b.b>\r\nSubject: test\r\n\r\nTestBody\r\n"), nil)
	fmt.Println(err)
}
