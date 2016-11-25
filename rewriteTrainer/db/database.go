package db


import (
	"github.com/boltdb/bolt"
	Logger "gatoor/orca/rewriteTrainer/log"
	"encoding/json"
	"time"
)

var DbLogger = Logger.LoggerWithField(Logger.Logger, "module", "db")

type OrcaDb struct {
	DbName string
	Db *bolt.DB
}

const (
	BUCKET_AUDIT_CURRENT_LAYOUT = "CurrentLayout"
	BUCKET_AUDIT_DESIRED_LAYOUT = "DesiredLayout"
	BUCKET_AUDIT_RECEIVED_STATS = "StatsReceived"
	BUCKET_AUDIT_RECEIVED_HOST_INFO = "HostInfoReceived"
	BUCKET_AUDIT_SENT = "PushSent"
)


var Audit OrcaDb

func Init(postfix string) {
	audit, err := bolt.Open("/orca/data/audit.db" + postfix, 0600, nil)

	if err != nil {
		DbLogger.Error("Cannot open database audit.db")
	}
	Audit = OrcaDb{
		"audit.db" + postfix, audit,
	}

	buckets := []string{BUCKET_AUDIT_RECEIVED_STATS, BUCKET_AUDIT_RECEIVED_HOST_INFO, BUCKET_AUDIT_SENT, BUCKET_AUDIT_DESIRED_LAYOUT, BUCKET_AUDIT_CURRENT_LAYOUT}

	for _, bucketName := range buckets {
		Audit.Db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				DbLogger.Errorf("Bucket %s could not be created in audit.db", bucketName)
				return nil
			}
			DbLogger.Info("Created DB Bucket %s", bucketName)
			return nil
		})
	}
}

func Close() {
	Audit.Db.Close()
}

func (a OrcaDb) Add(bucket string, key string, obj interface{}){
	encoded, jerr := json.Marshal(obj)
	if jerr != nil {
		DbLogger.Errorf("Failed to json encode '%+v'", obj)
		return
	}

	err := a.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), []byte(encoded))
		return err
	})
	if err != nil {
		DbLogger.Errorf("Failed to save '%s':'%s' to '%s'", key, encoded, bucket)
	}
}

func (a OrcaDb) Get(bucket string, key string) string {
	var res []byte
	a.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		res = b.Get([]byte(key))
		return nil
	})
	return string(res)
}

func GetNow() (string, time.Time) {
	t := time.Now().UTC()
	return t.Format(time.RFC3339), t
}