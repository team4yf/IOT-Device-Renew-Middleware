package main

import "testing"

func TestRenew(t *testing.T) {
	isOk, err := renew("foo", 10)
	if err != nil {
		t.Errorf("Renew('foo', 10) error: %v", err)
	}
	if !isOk {
		t.Error("Renew('foo', 10) should be ok")
	}
}

func TestCheck(t *testing.T) {
	isOk, err := check("foo")
	if err != nil {
		t.Errorf("Check('foo') error: %v", err)
	}
	if !isOk {
		t.Error("Check('foo') should be ok")
	}
}
