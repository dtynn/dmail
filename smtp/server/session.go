package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/dtynn/dmail/utils"
)

const (
	stateWaitForEhlo int = iota
	stateWaitForFrom
	stateWaitForRcpt
	stateWaitForData
	stateWriteData
	stateEnded
	stateAborted
)

const (
	cmdEhlo  = "EHLO"
	cmdHelo  = "HELO"
	cmdFrom  = "MAIL FROM:"
	cmdRcpt  = "RCPT TO:"
	cmdData  = "DATA"
	cmdTLS   = "STARTTLS"
	cmdQuit  = "QUIT"
	cmdBlank = ""
)

var (
	errDataSizeLimit = fmt.Errorf("size limit exceeded")
	errTimeout       = fmt.Errorf("timeout")

	ehloString = "250-%s\r\n250-SIZE %d\r\n"

	minCmdLimit       = 30
	minTimeout  int64 = 10
)

type permanentResps struct {
	greeting *smtpResponse
}

type command struct {
	cmd, parameter string
}

type session struct {
	id   string
	l    Logger
	conn net.Conn
	conf *Config
	in   *bufio.Reader
	out  *bufio.Writer

	state int
	local string
	from  string
	rcpt  []string
	data  string

	chErr    chan error
	timer    *time.Timer
	timeout  time.Duration
	receiver Receiver
}

func newSession(id string, l Logger, conn net.Conn, conf *Config) *session {
	if conf.SConf.CmdLimit <= minCmdLimit {
		conf.SConf.CmdLimit = minCmdLimit
	}

	if conf.SConf.Timeout < minTimeout {
		conf.SConf.Timeout = minTimeout
	}

	s := session{
		id:   id,
		l:    l,
		conn: conn,
		conf: conf,
		in:   bufio.NewReader(conn),
		out:  bufio.NewWriter(conn),

		rcpt: make([]string, 0),

		chErr:   make(chan error, 1),
		timeout: time.Duration(conf.SConf.Timeout) * time.Second,
	}
	return &s
}

func (this *session) registerRecevier(r Receiver) {
	this.receiver = r
}

func (this *session) handle() error {
	this.timer = time.NewTimer(this.timeout)
	defer this.cleanup()

	go this.do()
	for {
		select {
		case err := <-this.chErr:
			return err
		case <-this.timer.C:
			this.sendResp(respTimeout)
			this.l.Warnf("Id: %s timeout", this.id)
			return errTimeout
		}
	}
}

func (this *session) do() {
	if err := this.greeting(); err != nil {
		this.chErr <- err
	}
	for i := 0; i < this.conf.SConf.CmdLimit; i++ {
		switch this.state {
		case stateWaitForEhlo:
			if err := this.handleWaitForEhlo(); err != nil {
				this.chErr <- err
			}
		case stateWaitForFrom:
			if err := this.handleWaitForFrom(); err != nil {
				this.chErr <- err
			}
		case stateWaitForRcpt:
			if err := this.hanldeWaitForRcpt(); err != nil {
				this.chErr <- err
			}
		case stateWaitForData:
			if err := this.handleWaitForData(); err != nil {
				this.chErr <- err
			}
		case stateWriteData:
			if err := this.handleData(); err != nil {
				this.chErr <- err
			}
		case stateAborted:
			this.l.Info(this.id, "aborted")
			this.chErr <- this.sendResp(respClosing)
		case stateEnded:
			break
		}
	}
	this.l.Infof("Id: %s From %s ; Rcpt: %s; Length: %d",
		this.id, this.from, this.rcpt, len(this.data))
	this.chErr <- this.bye()
}

func (this *session) read() (string, error) {
	suffix := "\r\n"
	limit := this.conf.SConf.CmdSizeLimit

	if this.state == stateWriteData {
		suffix = "\r\n.\r\n"
		limit = this.conf.SConf.DataSizeLimit
	}

	var text, line string
	var err error

	for err == nil {
		this.resetTimeout()
		line, err = this.in.ReadString('\n')
		if err != nil {
			break
		}
		if len(line) != 0 {
			text = text + line
			if len(text) > limit {
				return "", errDataSizeLimit
			}
		}
		if strings.HasSuffix(text, suffix) {
			break
		}
	}
	return text, err
}

func (this *session) parseCmd(s string) *command {
	cmd := command{}
	upper := strings.ToUpper(s)
	if strings.Index(upper, cmdEhlo) == 0 ||
		strings.Index(upper, cmdHelo) == 0 {
		cmd.cmd = upper[0:4]
		if len(s) > 5 && s[4] == ' ' {
			cmd.parameter = utils.Strip(s[5:])
		}
	} else if strings.Index(upper, cmdFrom) == 0 {
		cmd.cmd = cmdFrom
		cmd.parameter = utils.Strip(s[len(cmdFrom):])
	} else if strings.Index(upper, cmdRcpt) == 0 {
		cmd.cmd = cmdRcpt
		cmd.parameter = utils.Strip(s[len(cmdRcpt):])
	} else {
		cmd.cmd = utils.Strip(upper)
	}
	return &cmd
}

func (this *session) getCmd() (*command, error) {
	s, err := this.read()
	if err != nil {
		return nil, err
	}
	cmd := this.parseCmd(s)
	return cmd, nil
}

func (this *session) writeString(s string) error {
	_, err := this.out.WriteString(s)
	this.resetTimeout()
	return err
}

func (this *session) writeLine(line string) error {
	return this.writeLine(line + "\r\n")
}

func (this *session) sendString(s string) error {
	_, err := this.out.WriteString(s)
	if err != nil {
		return err
	}
	return this.out.Flush()
}

func (this *session) sendLine(line string) error {
	return this.sendString(line + "\r\n")
}

func (this *session) sendResp(resp *smtpResponse) error {
	return this.sendLine(resp.String())
}

func (this *session) resetTimeout() bool {
	return this.timer.Reset(this.timeout)
}

func (this *session) greeting() error {
	greeting := NewSmtpResponse(codeGreeting, this.conf.Hostname+" / "+srvName)
	return this.sendResp(greeting)
}

func (this *session) ok() error {
	return this.sendResp(respOK)
}

func (this *session) bye() error {
	return this.sendResp(respBye)
}

func (this *session) resetEhlo(local string) {
	this.local = local
	this.state = stateWaitForFrom
	this.resetRcpt()
}

func (this *session) resetRcpt() {
	this.rcpt = make([]string, 0)
}

func (this *session) handleWaitForEhlo() error {
	cmd, err := this.getCmd()
	if err != nil {
		return err
	}

	switch cmd.cmd {
	case cmdEhlo, cmdHelo:
		err = this.doCmdEhlo(cmd)
	case cmdFrom, cmdRcpt, cmdTLS, cmdData:
		err = this.sendResp(respEhloFirst)
	case cmdQuit:
		this.state = stateEnded
	case cmdBlank:
		break
	default:
		err = this.sendResp(respNotImplemented)
	}

	return err
}

func (this *session) handleWaitForFrom() error {
	cmd, err := this.getCmd()
	if err != nil {
		return err
	}

	switch cmd.cmd {
	case cmdEhlo, cmdHelo:
		err = this.doCmdEhlo(cmd)
	case cmdFrom:
		err = this.doCmdFrom(cmd)
	case cmdRcpt, cmdData:
		err = this.sendResp(respBadSequense)
	case cmdTLS:
		// todo tls not implemented
		err = this.sendResp(respNotImplemented)
	case cmdQuit:
		this.state = stateEnded
	case cmdBlank:
		break
	default:
		err = this.sendResp(respNotImplemented)
	}
	return err
}

func (this *session) hanldeWaitForRcpt() error {
	cmd, err := this.getCmd()
	if err != nil {
		return err
	}

	switch cmd.cmd {
	case cmdEhlo, cmdHelo:
		err = this.doCmdEhlo(cmd)
	case cmdFrom, cmdData:
		err = this.sendResp(respBadSequense)
	case cmdRcpt:
		err = this.doCmdRcpt(cmd)
	case cmdTLS:
		err = this.sendResp(respEhloFirst)
	case cmdBlank:
		break
	default:
		err = this.sendResp(respNotImplemented)
	}

	return err
}

func (this *session) handleWaitForData() error {
	cmd, err := this.getCmd()
	if err != nil {
		return err
	}

	switch cmd.cmd {
	case cmdEhlo, cmdHelo:
		err = this.doCmdEhlo(cmd)
	case cmdFrom:
		err = this.sendResp(respBadSequense)
	case cmdRcpt:
		err = this.doCmdRcpt(cmd)
	case cmdData:
		err = this.sendResp(respReadyForData)
		this.state = stateWriteData
	case cmdTLS:
		err = this.sendResp(respEhloFirst)
	case cmdBlank:
		break
	default:
		err = this.sendResp(respNotImplemented)
	}

	return err
}

func (this *session) handleData() error {
	msg, err := this.read()
	if err == errDataSizeLimit {
		err = this.sendResp(respSizeLimitExceeded)
		this.state = stateAborted
		return err
	}
	if err != nil {
		return err
	}

	this.data = msg

	if this.receiver != nil {
		this.receiver.SetData(this.data)
	}

	err = this.sendResp(NewSmtpResponse(codeOK, "OK queued as "+this.id))
	this.state = stateEnded
	return err
}

func (this *session) doCmdEhlo(cmd *command) error {
	if len(cmd.parameter) == 0 {
		return this.sendResp(respSytaxErr)
	}

	if cmd.cmd == cmdEhlo {
		this.writeString(fmt.Sprintf(ehloString,
			this.conf.Hostname, this.conf.SConf.DataSizeLimit))
	}

	if this.receiver != nil {
		if err := this.receiver.SetEhlo(cmd.parameter); err != nil {
			// todo verbose
		}
	}

	this.resetEhlo(cmd.parameter)
	return this.ok()
}

func (this *session) doCmdFrom(cmd *command) error {
	mail, match := utils.CutMail(cmd.parameter)
	if !match {
		return this.sendResp(respSytaxErr)
	}
	this.from = mail

	if this.receiver != nil {
		if err := this.receiver.SetFrom(this.from); err != nil {
			// todo verbose
		}
	}

	this.state = stateWaitForRcpt
	return this.ok()
}

func (this *session) doCmdRcpt(cmd *command) error {
	mail, match := utils.CutMail(cmd.parameter)
	if !match {
		return this.sendResp(respSytaxErr)
	}

	this.rcpt = append(this.rcpt, mail)

	if this.receiver != nil {
		if err := this.receiver.AddRcpt(mail); err != nil {
			// todo verbose
		}
	}

	this.state = stateWaitForData
	return this.ok()
}

func (this *session) cleanup() {
	this.conn.Close()
	this.timer.Stop()
	if this.receiver != nil {
		this.receiver.Close()
	}
}
