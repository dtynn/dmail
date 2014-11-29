package dns

import (
	"fmt"
	"net"
)

var domainKeyPrefix = "dkim._domainkey."

type checkResult struct {
	Checked bool
	Detail  string
}

type checks struct {
	AllChecked                                     bool
	KeyChecked, SpfChecked, DkimChecked, MxChecked *checkResult
}

func newChecks() *checks {
	return &checks{false, &checkResult{}, &checkResult{}, &checkResult{}, &checkResult{}}
}

type Checker struct {
	name, key, spf, dkim, mxHost string
	checks                       *checks
}

func NewChecker(name, key, spf, dkim, mxHost string) *Checker {
	return &Checker{name, key, spf, dkim, mxHost, newChecks()}
}

func (this *Checker) CheckAll() *checks {
	this.checkTxt()
	this.checkDkim()
	this.checkMx()
	this.checks.AllChecked = this.checks.KeyChecked.Checked &&
		this.checks.SpfChecked.Checked &&
		this.checks.DkimChecked.Checked &&
		this.checks.MxChecked.Checked
	return this.checks
}

func (this *Checker) checkTxt() {
	txts, err := net.LookupTXT(this.name)
	if err != nil {
		this.checks.KeyChecked.Detail = err.Error()
		this.checks.SpfChecked.Detail = err.Error()
		return
	}

	for _, txt := range txts {
		switch txt {
		case this.key:
			this.checks.KeyChecked.Checked = true
		case this.spf:
			this.checks.SpfChecked.Checked = true
		}
	}

	return
}

func (this *Checker) checkDkim() {
	name := domainKeyPrefix + this.name
	txts, err := net.LookupTXT(name)
	if err != nil {
		this.checks.DkimChecked.Detail = err.Error()
		return
	}
	for _, txt := range txts {
		if txt == this.dkim {
			this.checks.DkimChecked.Checked = true
			break
		}
	}
	return
}

func (this *Checker) checkMx() {
	mxs, err := net.LookupMX(this.name)
	if err != nil {
		this.checks.MxChecked.Detail = err.Error()
		return
	}
	for _, mx := range mxs {
		if mx.Host != this.mxHost {
			this.checks.MxChecked.Detail = fmt.Sprintf("%d %s", mx.Pref, mx.Host)
			return
		}
	}
	this.checks.DkimChecked.Checked = true
	return
}
