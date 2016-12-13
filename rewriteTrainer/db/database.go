/*
Copyright Alex Mack
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

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


	DB_PATH = "/orca/data/audit.db"
)


var Audit OrcaDb

func Init(postfix string) {
	audit, err := bolt.Open(DB_PATH + postfix, 0600, nil)

	if err != nil {
		DbLogger.Panicf("Cannot open database %s", DB_PATH)
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
			DbLogger.Infof("Created DB Bucket %s", bucketName)
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
	return t.Format(time.RFC3339Nano), t
}