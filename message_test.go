package dmail

import (
	"testing"
)

var pemBytes = []byte(
	`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQC8LDUi99u+LOhD/T4QYh4xjaf//wx6eOgM8YhAHzSYywIscehI
k/7YXoZwIYvXT1eQCG5RSV7N+bp5IcGZ96tyXfMjUWvEJtLmrOBjMeKtlSqzBIih
IXaGucngXljt41RoAmNwm3ycfWtfsKRChbPR1sLQlv8su0y0LtoMiR6rkQIDAQAB
AoGBAIXFb6kCR0c1KZFb8Mk413om2C3XJQnT9jNtaY0cIgoVF+B8wcMG4v7yg+Qn
FQDluLv+Il7LKAiJ5hTC+Jz6Qvh81jjg4oAn8QmqETyyeNtqJDfPW0SvqSujQwcX
fSKIng301ut6miMcjoMaqodz3FRuEnFlgka2NI+oa91YeaMhAkEA3oMMzNM8lQkC
KgfunHaYcoyZutD3uwUlcEO1EGPFuQauJwNh/a2O+l9jkxaU8HkY7VKS9KCVh8O0
asJsnB+LxQJBANh+IlNYKJAfZ2lpMZUYT3NvrfHtcaraa94hl2qNEHv5Zopu0HDJ
eMEjs7XTYjnUv9JxksrB9NFsz9MQBXHaoV0CQF4RVQX6f3AaINoYBF4NHSHAIvWB
hlmAMXWmihNlup8gHdvMaE7QYtOiI/x43XpUF5+s+weEI/MDX3CKxVOzWmkCQBfD
sMzpTnqTl+xwSasOIhqP1c5KvEF+/HxDv7VIitixBdqIU4Ut+H1rB90buRqUCgJ1
ySFMrS0X/rAygAaBc1kCQDEMGiemJeQF+/ee8B7AoW2Ei6IqRkIyyklxtCX4z836
aGuSbLZpX/EcVR3d3yd+uSF52fuamuZK7nQo5Y1AEqo=
-----END RSA PRIVATE KEY-----`)

func TestSign(t *testing.T) {
	res := "v=1; a=rsa-sha256; c=relaxed/simple; d=dtynn.me; i=@dtynn.me; \r\n q=dns/txt; s=abc; t=1416912369; h=From : To : Subject : to : To : \r\n subject : Subject : to : From : Subject : subject : Subject; \r\n bh=P/65m6skPz/Wri2VBU/j1ViQK3o7hPVzpfeEh4cYpKw=; \r\n b=ESVTjf/jN4czA3Vx3B1T6rwuDl9H3M7mfKA9WsfXUNSSC6XV2auGIYXXDzdo1qYmsU8Uj/\r\n HbcMALpMFtHphgaPSTwp9sK9ieptXrs5XamnomZTGXG75JmWoidyaL0UiQpkmFG6BFEfS3BW\r\n TYbfr26LT3ATKGc1M76rx6HQ3N56M="

	msg := NewMessage(Unencoded, CharsetUTF8)
	msg.AddNormalHeader("From", "a")
	msg.AddNormalHeader("To", "b")
	msg.AddNormalHeader("Subject", "c")
	msg.AddNormalHeader("to", "d")
	msg.AddNormalHeader("To", "e")
	msg.AddNormalHeader("subject", "f")
	msg.AddNormalHeader("Subject", "g")
	msg.AddNormalHeader("to", "h")
	msg.AddNormalHeader("Test", "i")
	msg.AddNormalHeader("test", "j")
	body := "\r\nabc\r\nabc\r\n\r\n"
	msg.SetBody(body)

	d := NewDefaultDkim(msg, "dtynn.me", "", "abc", false)
	d.t = int64(1416912369)

	sig, err := d.Sign(pemBytes)
	if err != nil {
		t.Error("sign err: ", err)
		return
	}

	if sig != res {
		t.Error("sign failed")
	}
}
