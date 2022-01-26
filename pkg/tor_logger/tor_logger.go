package tor_logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

type TorLogger struct {
	Info *log.Logger
	Err  *log.Logger
}

// built mainly for a simpler mute on parts of the project
func NewTorLogger(output string) *TorLogger {
	var infoLogOut io.Writer
	var errLogOut io.Writer

	if output == "0" {
		infoLogOut, errLogOut = ioutil.Discard, ioutil.Discard // if user wanted no logs
	} else {
		infoLogOut, errLogOut = os.Stdout, os.Stderr
	}

	// useful.
	colorRed := string("\033[31m")
	colorGreen := string("\033[32m")
	colorWhite := string("\033[37m")

	infoLog := log.New(infoLogOut, colorGreen+"[info] "+colorWhite, log.Ldate|log.Ltime)
	errLog := log.New(errLogOut, colorRed+"[error] "+colorWhite, log.Ldate|log.Ltime|log.Llongfile)

	return &TorLogger{infoLog, errLog}
}
