package server

type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}

type Receiver interface {
	New(id string) (Receiver, error)
	Reset() error
	SetEhlo(local string) error
	SetFrom(from string) error
	AddRcpt(rcpt string) error
	SetData(data string) error
	Close() error
}
