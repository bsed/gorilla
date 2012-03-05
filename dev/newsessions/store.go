// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"code.google.com/p/gorilla/securecookie"
	"net/http"
)

// Store ----------------------------------------------------------------------

// Store is an interface for custom session stores.
type Store interface {
	Get(r *http.Request, name string) (*Session, error)
	New(r *http.Request, name string) (*Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *Session) error
}

// CookieStore ----------------------------------------------------------------

// NewCookieStore returns a new CookieStore.
//
// Keys are defined in pairs: one for authentication and the other for
// encryption. The encryption key can be set to nil or omitted in the last
// pair, but the authentication key is required in all pairs.
//
// Multiple pairs are accepted to allow key rotation, but the common case is
// to set a single authentication key and optionally an encryption key.
//
// It is recommended to use an authentication key with 32 or 64 bytes.
// The encryption key, if set, must be either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256 modes.
//
// Use the convenience function securecookie.GenerateRandomKey() to create
// strong keys.
func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	// Initialize it with a default configuration.
	s := &CookieStore{Options: &Options{
		Path:   "/",
		MaxAge: 86400 * 30,
	}}
	for i := 0; i < len(keyPairs); i += 2 {
		var blockKey []byte
		if i+1 < len(keyPairs) {
			blockKey = keyPairs[i+1]
		}
		s.Codecs = append(s.Codecs, securecookie.New(keyPairs[i], blockKey))
	}
	return s
}

// CookieStore stores sessions using secure cookies.
type CookieStore struct {
	Options *Options             // default configuration
	Codecs  []securecookie.Codec
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (s *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	return GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (s *CookieStore) New(r *http.Request, name string) (*Session, error) {
	session := NewSession(s, name)
	var errDecoding error
	if c, err := r.Cookie(name); err == nil {
		errDecoding = DecodeCookie(name, c.Value, &session.Values, s.Codecs...)
	} else {
		session.IsNew = true
	}
	return session, errDecoding
}

// Save saves a single session to the response.
func (s *CookieStore) Save(r *http.Request, w http.ResponseWriter,
	session *Session) error {
	encoded, err := EncodeCookie(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}
	options := s.Options
	if session.Options != nil {
		options = session.Options
	}
	cookie := &http.Cookie{
		Name:     session.Name(),
		Value:    encoded,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	http.SetCookie(w, cookie)
	return nil
}
