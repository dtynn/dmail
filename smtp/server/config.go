package server

import (
	"crypto/tls"
)

type Config struct {
	Hostname string
	Addr     string
	Verbose  bool
	SConf    *SessionConfig
	Tls      *tls.Config
}

type SessionConfig struct {
	Timeout       int64
	CmdSizeLimit  int
	DataSizeLimit int
	CmdLimit      int
}
