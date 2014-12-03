package server

import (
	"fmt"
)

const (
	codeGreeting          = 220
	codeBye               = 221
	codeOK                = 250
	codeRedyForData       = 354
	codeTimeout           = 420
	codeTryAgain          = 421
	codeRequestNotTaken   = 450
	codeSyntaxErr         = 500
	codeSyntaxErrInParams = 501
	codeCmdNotImplemented = 502
	codeBadSequense       = 503
	codeAuthenticationErr = 530
)

var (
	respOK             = NewSmtpResponse(codeOK, "OK")
	respBye            = NewSmtpResponse(codeBye, "Bye")
	respEhloFirst      = NewSmtpResponse(codeBadSequense, "EHLO/HELO first")
	respBadSequense    = NewSmtpResponse(codeBadSequense, "Bad sequense")
	respNotImplemented = NewSmtpResponse(
		codeCmdNotImplemented, "Cmd not implemented")
	respSytaxErr     = NewSmtpResponse(codeSyntaxErr, "Syntax error")
	respReadyForData = NewSmtpResponse(
		codeRedyForData, "End data with <CR><LF>.<CR><LF>")
	respSizeLimitExceeded = NewSmtpResponse(
		codeRequestNotTaken, "Size limit exceeded")
	respClosing = NewSmtpResponse(codeTryAgain, "closing transmission channel")
	respTimeout = NewSmtpResponse(codeTimeout, "action timeout")
)

type smtpResponse struct {
	statusCode int
	detail     string
}

func NewSmtpResponse(code int, detail string) *smtpResponse {
	return &smtpResponse{code, detail}
}

func (this *smtpResponse) String() string {
	return fmt.Sprintf("%d %s", this.statusCode, this.detail)
}
