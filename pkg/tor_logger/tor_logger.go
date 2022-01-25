package tor_logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

type TorLogger struct {
	info *log.Logger
	err  *log.Logger
}

// built mainly for a simpler mute on parts of the project
func NewTorLogger(mute string) *TorLogger {
	var infoLogOut io.Writer
	var errLogOut io.Writer

	if mute == "1" {
		infoLogOut, errLogOut = ioutil.Discard, ioutil.Discard // if user wanted no logs
	} else {
		infoLogOut, errLogOut = os.Stdout, os.Stderr
	}

	infoLog := log.New(infoLogOut, "[info] ", log.Ldate|log.Ltime)
	errLog := log.New(errLogOut, "[error] ", log.Ldate|log.Ltime|log.Llongfile)

	return &TorLogger{infoLog, errLog}
}
