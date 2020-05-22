package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

//Log global
var Log *log.Logger

func init() {
	Log = log.New()
	file, err := os.OpenFile("/var/log/logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.SetOutput(file)
	} else {
		Log.Info("Failed to log to file, using default stderr")
	}
	Log.SetFormatter(&log.JSONFormatter{})
	Log.SetLevel(log.DebugLevel)

}
