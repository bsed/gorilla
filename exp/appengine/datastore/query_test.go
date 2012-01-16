// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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

func TestTemp(t *testing.T) {
}
