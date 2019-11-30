package main

import (
	"bytes"
	"testing"
)

func getBuf() *bytes.Buffer {
	buf := &bytes.Buffer{}
	buf.WriteString(`{
		"1575097102": {
		  "213.87.144.131": 1
		},
		"1575097110": {
		  "213.87.135.29": 1
		},
		"1575097112": {
		  "213.87.135.29": 1
		},
		"1575097114": {
		  "213.87.135.29": 1
		},
		"1575097115": {
		  "213.87.135.29": 1
		},
		"1575097116": {
		  "213.87.135.29": 3
		}
	  }
	`)
	return buf
}

func Test_OriginCalc(t *testing.T) {

	cnt, err := GetOriginReqCount(getBuf(), "213.87.135.29", 1575097110, 1575097112)
	if err != nil {
		t.Error(err)
	}
	if cnt != 2 {
		t.Fatalf("Shoud be 2")
	}

	cnt, err = GetOriginReqCount(getBuf(), "213.87.135.29", 0, 0xFFFFFFFF)
	if err != nil {
		t.Error(err)
	}
	if cnt != 7 {
		t.Fatalf("Shoud be 7")
	}

	cnt, err = GetOriginReqCount(getBuf(), "127.0.0.1", 0, 0xFFFFFFFF)
	if err != nil {
		t.Error(err)
	}
	if cnt != 0 {
		t.Fatalf("Shoud be 0")
	}

}
