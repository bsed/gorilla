// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"fmt"
	"http"
	"os"
	"time"
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"gorilla.googlecode.com/hg/gorilla/sessions"
)

// SetDatastoreSessionStore enables the datastore session store.
func SetDatastoreSessionStore() {
	sessions.SetStore("datastore", new(DatastoreSessionStore))
}

// SetMemcacheSessionStore enables the memcache session store.
func SetMemcacheSessionStore() {
	sessions.SetStore("memcache", new(MemcacheSessionStore))
}

// ----------------------------------------------------------------------------
// DatastoreSessionStore
// ----------------------------------------------------------------------------

type Session struct {
	Date  datastore.Time
	Value []byte
}

// DatastoreSessionStore stores session data in App Engine's datastore.
type DatastoreSessionStore struct {
	// List of encoders registered for this store.
	encoders []sessions.SessionEncoder
}

// Load loads a session for the given key.
func (s *DatastoreSessionStore) Load(r *http.Request, key string,
									 info *sessions.SessionInfo) {
	data := sessions.GetCookie(s, r, key)
	if sidval, ok := data["sid"]; ok {
		// Cleanup session data.
		sid := sidval.(string)
		for k, _ := range data {
			data[k] = nil, false
		}
		// Get session from memcache and deserialize it.
		c := appengine.NewContext(r)
		var session Session
		key := datastore.NewKey("Session", sessionKey(sid), 0, nil)
		if err := datastore.Get(c, key, &session); err == nil {
			data, _ = sessions.DeserializeSessionData(session.Value)
			info.Id = sid
		}
	}
	info.Data = data
}

// Save saves the session in the response.
func (s *DatastoreSessionStore) Save(r *http.Request, w http.ResponseWriter,
									 key string, info *sessions.SessionInfo) (flag bool, err os.Error) {
	// Create a new session id.
	var sid string
	sid, err = sessions.GenerateSessionId(128)
	if err != nil {
		return
	}
	// Serialize session into []byte.
	var serialized []byte
	serialized, err = sessions.SerializeSessionData(&info.Data)
	if err != nil {
		return
	}
	// Save the session.
	c := appengine.NewContext(r)
	entityKey := datastore.NewKey("Session", sessionKey(sid), 0, nil)
	session := Session{
		Date:  datastore.SecondsToTime(time.Seconds()),
		Value: serialized,
	}
	if entityKey, err = datastore.Put(c, entityKey, &session); err != nil {
		return
	}
	// Clone info, setting only sid in data.
	newinfo := &sessions.SessionInfo{
		Data:   sessions.SessionData{"sid": sid},
		Store:  info.Store,
		Config: info.Config,
	}
	return sessions.SetCookie(s, w, key, newinfo)
}

// Encoders returns the encoders for this store.
func (s *DatastoreSessionStore) Encoders() []sessions.SessionEncoder {
	return s.encoders
}

// SetEncoders sets a group of encoders in the store.
func (s *DatastoreSessionStore) SetEncoders(encoders ...sessions.SessionEncoder) {
	s.encoders = encoders
}

// ----------------------------------------------------------------------------
// MemcacheSessionStore
// ----------------------------------------------------------------------------

// MemcacheSessionStore stores session data in App Engine's memcache.
type MemcacheSessionStore struct {
	// List of encoders registered for this store.
	encoders []sessions.SessionEncoder
}

// Load loads a session for the given key.
func (s *MemcacheSessionStore) Load(r *http.Request, key string,
									info *sessions.SessionInfo) {
	data := sessions.GetCookie(s, r, key)
	if sidval, ok := data["sid"]; ok {
		// Cleanup session data.
		sid := sidval.(string)
		for k, _ := range data {
			data[k] = nil, false
		}
		// Get session from memcache and deserialize it.
		c := appengine.NewContext(r)
		if item, err := memcache.Get(c, sessionKey(sid)); err == nil {
			data, _ = sessions.DeserializeSessionData(item.Value)
			info.Id = sid
		}
	}
	info.Data = data
}

// Save saves the session in the response.
func (s *MemcacheSessionStore) Save(r *http.Request, w http.ResponseWriter,
									key string, info *sessions.SessionInfo) (flag bool, err os.Error) {
	// Create a new session id.
	var sid string
	sid, err = sessions.GenerateSessionId(128)
	if err != nil {
		return
	}
	// Serialize session into []byte.
	var serialized []byte
	serialized, err = sessions.SerializeSessionData(&info.Data)
	if err != nil {
		return
	}
	// Add the item to the memcache, if the key does not already exist.
	c := appengine.NewContext(r)
	item := &memcache.Item{
		Key:   sessionKey(sid),
		Value: serialized,
	}
	if err := memcache.Add(c, item); err != nil {
		return false, err
	}
	// Clone info, setting only sid in data.
	newinfo := &sessions.SessionInfo{
		Data:   sessions.SessionData{"sid": sid},
		Store:  info.Store,
		Config: info.Config,
	}
	return sessions.SetCookie(s, w, key, newinfo)
}

// Encoders returns the encoders for this store.
func (s *MemcacheSessionStore) Encoders() []sessions.SessionEncoder {
	return s.encoders
}

// SetEncoders sets a group of encoders in the store.
func (s *MemcacheSessionStore) SetEncoders(encoders ...sessions.SessionEncoder) {
	s.encoders = encoders
}

func sessionKey(sid string) string {
	return fmt.Sprintf("gorilla.appengine.sessions.%s", sid)
}
