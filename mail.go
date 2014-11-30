package dmail

type Mail struct {
	ContentType string   `json:"contentType"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
}

func NewMail(contentType, from string, to []string, subject, body string) *Mail {
	return &Mail{contentType, from, to, subject, body}
}
