// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/gob"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"time"
)

// Codec defines an interface to encode and decode session values.
type Codec interface {
	Encode(key string, value map[interface{}]interface{}) (string, error)
	Decode(key, value string) (map[interface{}]interface{}, error)
}

// SecureCookie is the default codec implementation.
//
// It can be used standalone to create and read authenticated and optionally
// encrypted cookie values.
type SecureCookie struct {
	hashKey []byte
	block cipher.Block
	// For testing purposes, the function that returns the current timestamp.
	// If not set, it will use time.UTC().Unix().
	timeFunc func() int64
}

func (c *SecureCookie) Encode(key string, value map[interface{}]interface{}) (string, error) {
	if c.hashKey == nil {
		return "", errors.New("sessions: missing hash key to encode session")
	}
	var err error
	var b []byte
	// 1. Serialize.
	if b, err = serialize(value); err != nil {
		return "", err
	}
	// 2. Encrypt and encode (optional).
	if c.block != nil {
		if b, err = encrypt(c.block, b); err != nil {
			return "", err
		}
		b = encode(b)
	}
	// 3. Create MAC for "date|key|value" and append the result.
	buf := bytes.NewBufferString(fmt.Sprintf("%d|%s|", c.timestamp(), key))
	buf.Write(b)
	mac := createMac(hmac.New(sha256.New, c.hashKey), buf.Bytes())
	buf.WriteString("|")
	buf.Write(mac)
	// 4. Encode to base64.
	b = encode(buf.Bytes())
	// Done.
	return string(b), nil
}

func (c *SecureCookie) Decode(key, value string) (map[interface{}]interface{}, error) {
	if c.hashKey == nil {
		return nil, errors.New("sessions: missing hash key to encode session")
	}
	// 1. verify length
	// TODO
	// ...
	// 2. Decode from base64.
	b, err := decode([]byte(value))
	if err != nil {
		return nil, err
	}
	// 3. Value is "date|key|value|mac". Split into parts.
	parts := bytes.SplitN(b, []byte("|"), 4)
	if len(parts) != 4 {
		return nil, errors.New("sessions: invalid value")
	}
	// 4. Verify expiration date against parts[0].
	// TODO
	// ...
	// 5. Verify MAC: "date|key|parts[2]" against parts[3].
	h := hmac.New(sha256.New, c.hashKey)
	if err = verifyMac(h, b[:len(b)-len(parts[3])], parts[3]); err != nil {
		return nil, err
	}
	// 6. Decode and decrypt parts[2] (optional).
	if c.block != nil {
		if b, err = decode(parts[2]); err != nil {
			return nil, err
		}
		if b, err = decrypt(c.block, b); err != nil {
			return nil, err
		}
	}
	// 7. Deserialize.
	var data map[interface{}]interface{}
	if data, err = deserialize(b); err != nil {
		return nil, err
	}
	// Done.
	return data, nil
}

// timestamp returns the current timestamp, in seconds.
//
// For testing purposes, the function that generates the timestamp can be
// overridden. If not set, it will return time.Now().UTC().Unix().
func (c *SecureCookie) timestamp() int64 {
	if c.timeFunc == nil {
		return time.Now().UTC().Unix()
	}
	return c.timeFunc()
}

// Authentication -------------------------------------------------------------

// createMac creates a message authentication code (MAC).
func createMac(h hash.Hash, value []byte) []byte {
	h.Write(value)
	return h.Sum(nil)
}

// verifyMac verifies that a message authentication code (MAC) is valid.
func verifyMac(h hash.Hash, value []byte, mac []byte) error {
	h.Write(value)
	mac2 := h.Sum(nil)
	if len(mac) == len(mac2) && subtle.ConstantTimeCompare(mac, mac2) == 1 {
		return nil
	}
	return errors.New("The value is not valid")
}

// Encryption -----------------------------------------------------------------

// encrypt encrypts a value using the given block in counter mode.
//
// A random initialization vector with the length of the block size is
// prepended to the resulting ciphertext.
func encrypt(block cipher.Block, value []byte) ([]byte, error) {
	// Initialization vector on wikipedia: http://goo.gl/zF67k
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	// Encrypt it.
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(value, value)
	// Return iv + ciphertext.
	return append(iv, value...), nil
}

// decrypt decrypts a value using the given block in counter mode.
//
// The value to be decrypted must be prepended by a initialization vector
// with the length of the block size.
func decrypt(block cipher.Block, value []byte) ([]byte, error) {
	size := block.BlockSize()
	if len(value) > size {
		// Extract iv.
		iv := value[:size]
		// Extract ciphertext.
		value = value[size:]
		// Decrypt it.
		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(value, value)
		return value, nil
	}
	return nil, errors.New("The value could not be decrypted.")
}

// Serialization --------------------------------------------------------------

// serialize encodes a value using gob.
func serialize(value map[interface{}]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserialize decodes a value using gob.
func deserialize(value []byte) (map[interface{}]interface{}, error) {
	var m map[interface{}]interface{}
	dec := gob.NewDecoder(bytes.NewBuffer(value))
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

// Encoding -------------------------------------------------------------------

// encode encodes a value to a format suitable for cookie transmission.
func encode(value []byte) []byte {
	encoded := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
	base64.URLEncoding.Encode(encoded, value)
	return encoded
}

// decode decodes a value received as a session cookie.
func decode(value []byte) ([]byte, error) {
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		return nil, err
	}
	return decoded[:b], nil
}
