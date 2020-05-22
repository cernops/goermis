package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var (
	l        *log.Logger
	filePath = bootstrap.App.IFConfig.String("logging_file")
)

//GetLogger gives access to the centralised loging instance
func GetLogger() *log.Logger {

	l = log.New()
	l.SetFormatter(&log.JSONFormatter{})
	l.SetLevel(log.DebugLevel)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		l.SetOutput(file)
	} else {
		l.Info("Failed to log to file, using default stderr")
	}

	return l
}
