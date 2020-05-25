package bootstrap

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	//Log is a global instance of logrus
	Log = log.New()
)

//GetLogger gives access to the centralised loging instance
func init() {
	Log.SetFormatter(&log.JSONFormatter{})
	Log.SetLevel(log.DebugLevel)
	file, err := os.OpenFile(App.IFConfig.String("logging_file"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Log.SetOutput(file)
	} else {
		Log.Info("Failed to log to file, using default stderr")
	}
}
