package dmail

type Mail struct {
	ContentType string   `json:"contentType"`
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
}

func NewMail(contentType string, to []string, subject, body string) *Mail {
	return &Mail{contentType, to, subject, body}
}
