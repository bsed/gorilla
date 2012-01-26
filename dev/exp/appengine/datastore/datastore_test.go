package datastore

import (
	"testing"
	"gae-go-testing.googlecode.com/git/appenginetesting"
)

func getContext(t *testing.T) *appenginetesting.Context {
	c, err := appenginetesting.NewContext(nil)
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	return c
}
