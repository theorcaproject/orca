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
