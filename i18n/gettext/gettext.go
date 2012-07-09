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

// NewTranslations returns a new Translations, initializing internal fields.
func NewTranslations() *Translations {
	return &Translations{
		PluralFunc: pluralforms.DefaultPluralFunc,
		Info:       make(map[string]string),
		msg:        make(map[string][]string),
		trn:        make(map[string][]string),
		ord:        make(map[string][][]int),
	}
}

// Translations stores gettext translations.
type Translations struct {
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

// Message returns a translation for the given message key.
//
// Extra arguments can be passed to format the translation using fmt.Sprintf().
func (c *Translations) Message(key string, vars ...interface{}) string {
	if trn, ok := c.trn[key]; ok {
		if vars == nil {
			return trn[0]
		}
		return sprintf(trn[0], c.ord[key][0], vars...)
	}
	if c.Fallback != nil {
		return c.Fallback.Message(key, vars...)
	}
	if vars == nil {
		return key
	}
	return fmt.Sprintf(key, vars...)
}

// Plural returns a plural translation for the given message key and count.
//
// Extra arguments can be passed to format the translation using fmt.Sprintf().
func (c *Translations) Plural(key string, count int, vars ...interface{}) string {
	if trn, ok := c.trn[key]; ok && c.PluralFunc != nil {
		if idx := c.PluralFunc(count); idx >= 0 && idx < len(trn) {
			if vars == nil {
				return trn[idx]
			}
			return sprintf(trn[idx], c.ord[key][idx], vars...)
		}
	}
	if c.Fallback != nil {
		return c.Fallback.Plural(key, count, vars...)
	}
	if vars == nil {
		return key
	}
	return fmt.Sprintf(key, vars...)
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
