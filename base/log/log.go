package log

import log "github.com/Sirupsen/logrus"



var Logger = log.WithFields(log.Fields {
	"Calf": "somehostid",
})

var AuditLogger = Logger.WithFields(log.Fields{
	"Stage": "Audit",
})

func LoggerWithField(logger *log.Entry, key string, val string) *log.Entry {
	return logger.WithFields(log.Fields{
		key: val,
	})
}

