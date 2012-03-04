// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"code.google.com/p/gorilla/context"
	"code.google.com/p/gorilla/securecookie"
)

// SessionConfig --------------------------------------------------------------

// SessionConfig stores configuration for a session.
//
// Fields are a subset of http.Cookie fields.
type SessionConfig struct {
	Path     string
	Domain   string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

// Session --------------------------------------------------------------------

// Session stores the values and optional configuration for a session.
type Session struct {
	Value  map[interface{}]interface{}
	Config *SessionConfig
	IsNew  bool
	name   string
	store  Store
}

// Flashes returns a slice of flash messages from the session.
//
// A single variadic argument is accepted, and it is optional: it defines
// the flash key. If not defined "_flash" is used by default.
func (s *Session) Flashes(vars ...string) []interface{} {
	key := "_flash"
	if len(vars) > 0 {
		key = vars[0]
	}
	if flashes, ok := s.Value[key]; ok {
		// Drop the flashes and return it.
		delete(s.Value, key)
		return flashes.([]interface{})
	}
	// Return a new one.
	return make([]interface{}, 0)
}

// AddFlash adds a flash message to the session.
//
// A single variadic argument is accepted, and it is optional: it defines
// the flash key. If not defined "_flash" is used by default.
func (s *Session) AddFlash(value interface{}, vars ...string) {
	key := "_flash"
	if len(vars) > 0 {
		key = vars[0]
	}
	var flashes []interface{}
	if v, ok := s.Value[key]; ok {
		flashes = v.([]interface{})
	}
	s.Value[key] = append(flashes, value)
}

// Name returns the name used to register the session.
func (s *Session) Name() string {
	return s.name
}

// Request Sessions -----------------------------------------------------------

// contextKey is the type used to store Sessions in the context.
type contextKey int

// sessionsKey is the key used to store Sessions in the context.
const sessionsKey contextKey = 0

// GetSessions returns the Sessions instance for the current request.
func GetSessions(r *http.Request) *Sessions {
	s := context.DefaultContext.Get(r, sessionsKey)
	if s != nil {
		return s.(*Sessions)
	}
	sessions := &Sessions{
		m: make(map[string]*Session),
	}
	context.DefaultContext.Set(r, sessionsKey, sessions)
	return sessions
}

// Sessions stores all sessions used during the current request.
type Sessions struct {
	m map[string]*Session
}

// Get returns a registered session for the current request.
//
// It returns (nil, nil) if there are no sessions with the given name.
// It returns (nil, error) if there's a session with the given name but
// it was registered using a different store.
func (s *Sessions) Get(store Store, name string) (*Session, error) {
	if session, ok := s.m[name]; ok {
		if store == nil || session.store != store {
			return nil, errors.New("sessions: session doesn't match the store")
		}
		return session, nil
	}
	return nil, nil
}

// Add registers a session for the current request.
func (s *Sessions) Add(store Store, name string, session *Session) {
	session.store = store
	session.name = name
	s.m[name] = session
}

// Save saves all sessions registered for the current request.
func (s *Sessions) Save(r *http.Request, w http.ResponseWriter) error {
	var errMulti MultiError
	for name, session := range s.m {
		if session.store == nil {
			errMulti = append(errMulti, fmt.Errorf(
				"sessions: missing store for session %q", name))
		} else if err := session.store.Save(r, w, session); err != nil {
			errMulti = append(errMulti, fmt.Errorf(
				"sessions: error saving session %q -- %v", name, err))
		}
	}
	if errMulti != nil {
		return errMulti
	}
	return nil
}

// Helpers --------------------------------------------------------------------

func init() {
	gob.Register([]interface{}{})
}

// Save saves all sessions used during the current request.
func Save(r *http.Request, w http.ResponseWriter) error {
	return GetSessions(r).Save(r, w)
}

// EncodeCookie encodes a cookie value using a group of securecookie codecs.
//
// The codecs are tried in order. Multiple codecs are accepted to allow
// key rotation.
func EncodeCookie(name string, value interface{}, codecs ...securecookie.Codec) (string, error) {
	for _, codec := range codecs {
		if encoded, err := codec.Encode(name, value); err == nil {
			return encoded, nil
		}
	}
	return "", errors.New("sessions: cookie could not be encoded")
}

// DecodeCookie decodes a cookie value using a group of securecookie codecs.
//
// The codecs are tried in order. Multiple codecs are accepted to allow
// key rotation.
func DecodeCookie(name string, value string, codecs ...securecookie.Codec) (map[interface{}]interface{}, error) {
	m := make(map[interface{}]interface{})
	for _, codec := range codecs {
		if err := codec.Decode(name, value, &m); err == nil {
			return m, nil
		}
	}
	return nil, errors.New("sessions: cookie could not be decoded")
}

// CreateCookie returns an *http.Cookie with the given configuration.
func CreateCookie(name, value string, config *SessionConfig) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     config.Path,
		Domain:   config.Domain,
		MaxAge:   config.MaxAge,
		Secure:   config.Secure,
		HttpOnly: config.HttpOnly,
	}
}

// Error ----------------------------------------------------------------------

// MultiError stores multiple errors.
//
// Borrowed from the App Engine SDK.
type MultiError []error

func (m MultiError) Error() string {
	s, n := "", 0
	for _, e := range m {
		if e != nil {
			if n == 0 {
				s = e.Error()
			}
			n++
		}
	}
	switch n {
	case 0:
		return "(0 errors)"
	case 1:
		return s
	case 2:
		return s + " (and 1 other error)"
	}
	return fmt.Sprintf("%s (and %d other errors)", s, n-1)
}
