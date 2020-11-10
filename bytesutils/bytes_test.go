package bytesutils

import (
	"fmt"
	"testing"
	"time"
)

type TestStruct struct {
	Tmp string
}

func Test_Sha1(t *testing.T) {
	date1, _ := time.Parse(time.RFC3339, "2020-01-01T01:01:01Z")
	date2, _ := time.Parse(time.RFC3339, "2020-01-01T01:01:02Z")

	testBo := struct {
		AfterRow  map[string]interface{}
		BeforeRow map[string]interface{}
	}{
		AfterRow: map[string]interface{}{
			"1234": date1,
			"ABC":  []int{1, 2, 3},
			"DEF":  []string{"1", "2", "3"},
		},
		BeforeRow: map[string]interface{}{
			"5678": &date2,
			"DEF": map[string]interface{}{
				"int64":   int64(4),
				"float64": float64(-0.1),
				"struct": struct {
					MapObj      map[string]interface{}
					MapSliceObj []map[int]interface{}
					Str         string
					Pstr        *TestStruct
				}{
					MapObj:      map[string]interface{}{"ob1": &date2},
					MapSliceObj: []map[int]interface{}{{1: "1"}, {2: "2"}},
					Str:         "hello",
					Pstr: &TestStruct{
						Tmp: "OK",
					},
				},
			},
		},
	}

	buf := []interface{}{testBo.AfterRow, testBo.BeforeRow}

	ret, err := Sha1(buf)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(ret)
}
