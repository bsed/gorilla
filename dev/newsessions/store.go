// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"net/http"
	"code.google.com/p/gorilla/securecookie"
)

// Store ----------------------------------------------------------------------

// Store is an interface for custom session stores.
type Store interface {
	Get(r *http.Request, name string) (*Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *Session) error
}

// CookieStore ----------------------------------------------------------------

// Default configuration for the cookie store.
var defaultConfig = &SessionConfig{}

// NewCookieStore returns a new CookieStore.
func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	config := *defaultConfig
	s := &CookieStore{Config: &config}
	for i := 0; i < len(keyPairs); i += 2 {
		var blockKey []byte
		if i+1 < len(keyPairs) {
			blockKey = keyPairs[i+1]
		}
		s.codecs = append(s.codecs, securecookie.New(keyPairs[i], blockKey))
	}
	return s
}

// CookieStore stores sessions using secure cookies.
type CookieStore struct {
	codecs []securecookie.Codec
	Config *SessionConfig       // default configuration
}

// Get returns a session with the given name.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns an error if the session exists but could not be decoded.
func (s *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	sessions := GetSessions(r)
	if session, err := sessions.Get(s, name); err != nil {
		return nil, err
	} else if session != nil {
		return session, nil
	}
	session := &Session{}
	if c, err := r.Cookie(name); err == nil {
		if m, err := DecodeCookie(name, c.Value, s.codecs...); err == nil {
			session.Value = m
		} else {
			return nil, err
		}
	} else {
		session.Value = make(map[interface{}]interface{})
		session.IsNew = true
	}
	sessions.Add(s, name, session)
	return session, nil
}

// Save saves a single session to the response.
func (s *CookieStore) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	encoded, err := EncodeCookie(session.Name(), session.Value, s.codecs...)
	if err != nil {
		return err
	}
	config := s.Config
	if session.Config != nil {
		config = session.Config
	}
	http.SetCookie(w, CreateCookie(session.Name(), encoded, config))
	return nil
}
