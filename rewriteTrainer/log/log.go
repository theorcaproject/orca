package log

import log "github.com/Sirupsen/logrus"



var Logger = log.WithFields(log.Fields {
	"Orca": "Trainer",
})

func LoggerWithField(logger *log.Entry, key string, val string) *log.Entry {
	return logger.WithFields(log.Fields{
		key: val,
	})
}

var InitLogger = LoggerWithField(Logger, "module", "init")

var AuditLogger = Logger.WithFields(log.Fields{
	"module": "audit",
})

