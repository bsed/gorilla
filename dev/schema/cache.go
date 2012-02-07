// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"reflect"
	"strings"
	"sync"
)

var cache = &structCache{m: make(map[string]*structInfo)}

type structInfo struct {
	fieldNameByAlias map[string]string
}

type structCache struct {
	l sync.Mutex
	m map[string]*structInfo
}

func (s *structCache) get(t reflect.Type) *structInfo {
	structId := typeName(t)
	s.l.Lock()
	info := s.m[structId]
	if info == nil {
		if info, _ = s.createStructInfo(t); info != nil {
			s.m[structId] = info
		}
	}
	s.l.Unlock()
	return info
}

func (s *structCache) createStructInfo(t reflect.Type) (*structInfo, map[string]reflect.Type){
	if t.Kind() != reflect.Struct {
		return nil, nil
	}
	info := &structInfo{
		fieldNameByAlias: make(map[string]string),
	}
	var nested map[string]reflect.Type
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		alias := field.Name
		// Use the name defined in the tag, if available.
		tag := field.Tag.Get("schema")
		if tag != "" {
			// For now tags only support the name but let's folow the
			// comma convention from encoding/json and others.
			if idx := strings.Index(tag, ","); idx == -1 {
				alias = tag
			} else {
				alias = tag[:idx]
			}
			if alias == "" {
				alias = field.Name
			}
		}
		// Ignore this field?
		if alias != "-" {
			info.fieldNameByAlias[alias] = field.Name
			/*
			// If the field is a struct, store it to be loaded later.
			// Don't need to load recursively, so commented out.
			if field.Type.Kind() == reflect.Ptr {
				field = field.Type.Elem()
			}
			if field.Type.Kind() == reflect.Struct {
				if nested == nil {
					nested = make(map[string]reflect.Type)
				}
				nested[alias] = field.Type
			}
			*/
		}
	}
	return info, nested
}

func (s *structCache) getNameByAlias(t reflect.Type, alias string) string {
	var name string
	if info := s.get(t); info != nil {
		if n, ok := info.fieldNameByAlias[alias]; ok {
			name = n
		}
	}
	return name
}

// typeName returns a string identifier for a type.
func typeName(t reflect.Type) string {
	// Borrowed from gob package.
	// We don't care about pointers.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Default to printed representation for unnamed types.
	name := t.String()
	// But for named types, qualify with import path.
	if t.Name() != "" {
		if t.PkgPath() == "" {
			name = t.Name()
		} else {
			name = t.PkgPath() + "." + t.Name()
		}
	}
	return name
}
