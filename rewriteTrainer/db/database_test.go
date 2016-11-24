package db


import (
	"testing"
)

type TestStruct struct {
	A string
	B float32
}

func TestDb_AuditDb(t *testing.T) {
	Init("_test")

	res := Audit.Get(BUCKET_AUDIT_SENT, "unkown")

	if res != "" {
		t.Error("wrong res")
	}

	Audit.Add(BUCKET_AUDIT_SENT, "someKey", TestStruct{"somefield", 100.0})

	res1 := Audit.Get(BUCKET_AUDIT_SENT, "someKey")

	if res1 != "{\"A\":\"somefield\",\"B\":100}" {
		t.Error(res1)
	}
}
