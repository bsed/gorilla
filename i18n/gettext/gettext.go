// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"fmt"
	"io"

	"code.google.com/p/gorilla/i18n"
	"code.google.com/p/gorilla/i18n/gettext/pluralforms"
)

// ContextFunc is used to select the context stored for message disambiguation.
type ContextFunc func(string) bool

// NewCatalog returns a new Catalog, initializing internal fields.
func NewCatalog() *Catalog {
	return &Catalog{
		PluralFunc: pluralforms.DefaultPluralFunc,
		Info:       make(map[string]string),
		msg:        make(map[string][]string),
		trn:        make(map[string][]string),
		ord:        make(map[string][][]int),
	}
}

// Catalog stores gettext translations.
type Catalog struct {
	Fallback    i18n.Translator        // used when a translation is not found
	ContextFunc ContextFunc            // used to select context to load
	PluralFunc  pluralforms.PluralFunc // used to select the plural form index
	Info        map[string]string      // metadata from file header
	msg         map[string][]string    // messages
	trn         map[string][]string    // translations
	ord         map[string][][]int     // translation expansion orderings
	msgOrig     [][]byte               // original messages, unprocessed
	trnOrig     [][]byte               // original translations, unprocessed
}

// Get returns a translation for the given key.
//
// Extra arguments can be passed to format the translation using fmt.Sprintf().
func (c *Catalog) Get(key string, a ...interface{}) string {
	if trn, ok := c.trn[key]; ok {
		if a == nil {
			return trn[0]
		}
		return sprintf(trn[0], c.ord[key][0], a...)
	}
	if c.Fallback != nil {
		return c.Fallback.Get(key, a...)
	}
	if a == nil {
		return key
	}
	return fmt.Sprintf(key, a...)
}

// GetPlural returns a plural translation for the given key and count.
//
// Extra arguments can be passed to format the translation using fmt.Sprintf().
func (c *Catalog) GetPlural(key string, count int, a ...interface{}) string {
	if trn, ok := c.trn[key]; ok && c.PluralFunc != nil {
		if idx := c.PluralFunc(count); idx >= 0 && idx < len(trn) {
			if a == nil {
				return trn[idx]
			}
			return sprintf(trn[idx], c.ord[key][idx], a...)
		}
	}
	if c.Fallback != nil {
		return c.Fallback.GetPlural(key, count, a...)
	}
	if a == nil {
		return key
	}
	return fmt.Sprintf(key, a...)
}

// ----------------------------------------------------------------------------

// Reader wraps the interfaces used to read MO and PO files.
type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// Writer wraps the interfaces used to write MO and PO files.
type Writer interface {
	io.Writer
	io.WriterAt
	io.Seeker
}
