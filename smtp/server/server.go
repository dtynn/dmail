package server

import (
	"net"

	"github.com/dtynn/dmail/safeMap"
	"github.com/dtynn/dmail/utils"
)

const (
	srvName         = "dmail/server"
	sessionIdLength = 16
)

type Server struct {
	cfg      *Config
	l        Logger
	sessions *safeMap.SafeMap
	receiver Receiver
}

func NewServer(cfg *Config, l Logger) *Server {
	sessions := safeMap.NewSafeMap()
	return &Server{
		cfg:      cfg,
		l:        l,
		sessions: sessions,
	}
}

func (this *Server) Run() error {
	listener, err := net.Listen("tcp", this.cfg.Addr)
	if err != nil {
		this.l.Warn("Listen err: ", err)
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			this.l.Warn("Accept err: ", err)
			continue
		}
		sessId := utils.RandString(sessionIdLength)
		sess := newSession(sessId, this.l, conn, this.cfg)
		for {
			err := this.sessions.Setnx(sessId, sess)
			if err == nil {
				break
			}
			sessId = utils.RandString(sessionIdLength)
			sess.id = sessId
		}

		if this.receiver != nil {
			if r, err := this.receiver.New(sessId); err != nil {
				this.l.Warn("receiver.New", err)
			} else {
				sess.registerRecevier(r)
			}
		}
		go func(sess *session, sessions *safeMap.SafeMap) {
			this.l.Info("session err:", sess.handle())
			sessions.Del(sess.id)
		}(sess, this.sessions)
	}
}

func (this *Server) RegisterReceiver(r Receiver) {
	this.receiver = r
}
