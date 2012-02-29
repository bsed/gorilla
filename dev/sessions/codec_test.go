// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"crypto/aes"
	"fmt"
	"testing"
)

var testSessions = []map[interface{}]interface{}{
	{"foo": "bar"},
	{"baz": "ding"},
}

var testStrings = []string{"foo", "bar", "baz"}

func TestAuthentication(t *testing.T) {
}

func TestEncription(t *testing.T) {
	block, err := aes.NewCipher([]byte("1234567890123456"))
	if err != nil {
		t.Fatalf("Block could not be created")
	}
	var encrypted, decrypted []byte
	for _, value := range testStrings {
		if encrypted, err = encrypt(block, []byte(value)); err != nil {
			t.Error(err)
		} else {
			if decrypted, err = decrypt(block, encrypted); err != nil {
				t.Error(err)
			}
			if string(decrypted) != value {
				t.Errorf("Expected %v, got %v.", value, string(decrypted))
			}
		}
	}
}

func TestSerialization(t *testing.T) {
	var (
		serialized []byte
		deserialized map[interface{}]interface{}
		err error
	)
	for _, value := range testSessions {
		if serialized, err = serialize(value); err != nil {
			t.Error(err)
		} else {
			if deserialized, err = deserialize(serialized); err != nil {
				t.Error(err)
			}
			if fmt.Sprintf("%v", deserialized) != fmt.Sprintf("%v", value) {
				t.Errorf("Expected %v, got %v.", value, deserialized)
			}
		}
	}
}

func TestEncoding(t *testing.T) {
	for _, value := range testStrings {
		encoded := encode([]byte(value))
		decoded, err := decode(encoded)
		if err != nil {
			t.Error(err)
		} else if string(decoded) != value {
			t.Errorf("Expected %v, got %s.", value, string(decoded))
		}
	}
}
