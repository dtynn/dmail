package dmail

type Address struct {
	Address, Name string
}

type Mail struct {
	From    *Address
	To      []*Address
	Cc      []*Address
	Bcc     []*Address
	Subject string
	Body    []byte
}
