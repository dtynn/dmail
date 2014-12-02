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
	cmdEhlo = "EHLO"
	cmdHelo = "HELO"
	cmdFrom = "MAIL FROM:"
	cmdRcpt = "RCPT TO:"
	cmdData = "DATA"
	cmdTLS  = "STARTTLS"
	cmdQuit = "QUIT"
)

var (
	errDataSizeLimit = fmt.Errorf("size limit exceeded")

	ehloString = "250-%s\r\n250-SIZE %d\r\n"

	minCmdLimit = 30
)

type permanentResps struct {
	greeting *smtpResponse
}

type command struct {
	cmd, parameter string
}

type session struct {
	id    string
	l     Logger
	conn  net.Conn
	state int
	local string

	conf  *Config
	in    *bufio.Reader
	out   *bufio.Writer
	resps *permanentResps
}

func newSession(id string, l Logger, conn net.Conn, conf *Config) *session {
	if conf.SConf.CmdLimit <= minCmdLimit {
		conf.SConf.CmdLimit = minCmdLimit
	}

	s := session{
		id:   id,
		l:    l,
		conn: conn,

		conf: conf,
		in:   bufio.NewReader(conn),
		out:  bufio.NewWriter(conn),
	}
	return &s
}

func (this *session) handle() error {
	defer this.conn.Close()
	this.l.Info("handling")
	if err := this.greeting(); err != nil {
		return err
	}
	for i := 0; i < this.conf.SConf.CmdLimit; i++ {
		switch this.state {
		case stateWaitForEhlo:
			if err := this.handleWaitForEhlo(); err != nil {
				return err
			}
		case stateWaitForFrom:
			if err := this.handleWaitForFrom(); err != nil {
				return err
			}
		case stateWaitForRcpt:
			if err := this.hanldeWaitForRcpt(); err != nil {
				return err
			}
		case stateWaitForData:
			if err := this.handleWaitForData(); err != nil {
				return err
			}
		case stateWriteData:
			if err := this.handleData(); err != nil {
				return err
			}
		case stateAborted:
			return this.sendResp(respClosing)
		case stateEnded:
			break
		}
	}
	return this.bye()
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
	this.l.Info("Read: ", text)
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
	this.l.Info("cmd", cmd.cmd)
	this.l.Info("param", cmd.parameter)
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

func (this *session) setTimeout(timeout int64) error {
	return this.conn.SetDeadline(time.Now().
		Add(time.Duration(timeout) * time.Second))
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

func (this *session) handleWaitForEhlo() error {
	cmd, err := this.getCmd()
	if err != nil {
		return err
	}

	switch cmd.cmd {
	case cmdEhlo, cmdHelo:
		if len(cmd.parameter) == 0 {
			this.sendResp(respSytaxErr)
		}
		if cmd.cmd == cmdEhlo {
			this.writeString(fmt.Sprintf(ehloString,
				this.conf.Hostname, this.conf.SConf.DataSizeLimit))
		}
		err = this.ok()
		this.state = stateWaitForFrom
	case cmdFrom, cmdRcpt, cmdTLS, cmdData:
		err = this.sendResp(respEhloFirst)
	case cmdQuit:
		this.state = stateEnded
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
		if len(cmd.parameter) == 0 {
			err = this.sendResp(respSytaxErr)
			break
		}
		if cmd.cmd == cmdEhlo {
			this.writeString(fmt.Sprintf(ehloString,
				this.conf.Hostname, this.conf.SConf.DataSizeLimit))
		}
		err = this.ok()
		this.state = stateWaitForFrom
	case cmdFrom:
		if len(cmd.parameter) == 0 || !reMail.MatchString(cmd.parameter) {
			err = this.sendResp(respSytaxErr)
			break
		}
		err = this.ok()
		this.state = stateWaitForRcpt
	case cmdRcpt, cmdData:
		err = this.sendResp(respBadSequense)
	case cmdTLS:
		// todo tls not implemented
		err = this.sendResp(respNotImplemented)
	case cmdQuit:
		this.state = stateEnded
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
		if len(cmd.parameter) == 0 {
			err = this.sendResp(respSytaxErr)
			break
		}
		if cmd.cmd == cmdEhlo {
			this.writeString(fmt.Sprintf(ehloString,
				this.conf.Hostname, this.conf.SConf.DataSizeLimit))
		}
		err = this.ok()
		this.state = stateWaitForFrom
	case cmdFrom, cmdData:
		err = this.sendResp(respBadSequense)
	case cmdRcpt:
		if len(cmd.parameter) == 0 || !reMail.MatchString(cmd.parameter) {
			err = this.sendResp(respSytaxErr)
			break
		}
		err = this.ok()
		this.state = stateWaitForData
	case cmdTLS:
		err = this.sendResp(respEhloFirst)
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
		if len(cmd.parameter) == 0 {
			err = this.sendResp(respSytaxErr)
			break
		}
		if cmd.cmd == cmdEhlo {
			this.writeString(fmt.Sprintf(ehloString,
				this.conf.Hostname, this.conf.SConf.DataSizeLimit))
		}
		err = this.ok()
		this.state = stateWaitForFrom
	case cmdFrom:
		err = this.sendResp(respBadSequense)
	case cmdRcpt:
		if len(cmd.parameter) == 0 || !reMail.MatchString(cmd.parameter) {
			err = this.sendResp(respSytaxErr)
			break
		}
		err = this.ok()
		this.state = stateWaitForData
	case cmdData:
		err = this.sendResp(respReadyForData)
		this.state = stateWriteData
	case cmdTLS:
		err = this.sendResp(respEhloFirst)
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
	this.l.Infof("\n%v", msg)
	err = this.sendResp(NewSmtpResponse(codeOK, "OK queued as "+this.id))
	this.state = stateEnded
	return err
}
