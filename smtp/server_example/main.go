package main

import (
	"github.com/dtynn/dmail/smtp/server"

	"github.com/qiniu/log"
)

var l = log.Std

func main() {
	scfg := server.SessionConfig{}
	scfg.CmdSizeLimit = 512 * 1024
	scfg.DataSizeLimit = 2 * 1024 * 1024
	scfg.CmdLimit = 50
	scfg.Timeout = 20

	cfg := server.Config{}
	cfg.Addr = "0.0.0.0:25"
	cfg.Hostname = "localtest"
	cfg.Verbose = true
	cfg.SConf = &scfg

	srv := server.NewServer(&cfg, log.Std)
	srv.RegisterReceiver(&faker{})
	if err := srv.Run(); err != nil {
		log.Fatalln(err)
	}
}
