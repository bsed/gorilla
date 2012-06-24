// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"code.google.com/p/gorilla/gettext/pluralforms"
)

const (
	magicBigEndian    = 0xde120495
	magicLittleEndian = 0x950412de
)

// Reader wraps the interfaces used to read compiled catalogs.
//
// Typically catalogs are provided as os.File.
type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// ContextFunc is used to select the context stored for message disambiguation.
type ContextFunc func(string) bool

// NewCatalog returns a new Catalog, initializing its internal fields.
func NewCatalog() *Catalog {
	return &Catalog{
		PluralFunc: pluralforms.DefaultPluralFunc,
		messages:   make(map[string]string),
		mPlurals:   make(map[string][]string),
		tPlurals:   make(map[string][]string),
	}
}

// Catalog stores gettext translations.
//
// Inspired by Python's gettext.GNUTranslations.
//
// TODO: Gettextf(msg, replacements...) to use with fmt.Sprintf?
type Catalog struct {
	Fallback    *Catalog               // used when a translation is not found
	ContextFunc ContextFunc            // used to select context to load
	PluralFunc  pluralforms.PluralFunc // used to select the plural form index
	messages    map[string]string      // original messages
	mPlurals    map[string][]string    // message plurals
	tPlurals    map[string][]string	   // translation plurals
}

// Gettext returns a translation for the given message.
func (c *Catalog) Gettext(msg string) string {
	if trans, ok := c.messages[msg]; ok {
		return trans
	}
	if c.Fallback != nil {
		return c.Fallback.Gettext(msg)
	}
	return msg
}

// Ngettext returns a plural translation for a message according to the
// amount n.
//
// msg1 is used to lookup for a translation, and msg2 is used as the plural
// form fallback if a translation is not found.
func (c *Catalog) Ngettext(msg1, msg2 string, n int) string {
	if plurals, ok := c.tPlurals[msg1]; ok {
		if idx := c.PluralFunc(n); idx < len(plurals) {
			return plurals[idx]
		}
	}
	if c.Fallback != nil {
		return c.Fallback.Ngettext(msg1, msg2, n)
	}
	if n == 1 {
		return msg1
	}
	return msg2
}

// ReadMO reads a GNU MO file and writes its messages and translations
// to the catalog.
//
// GNU MO file format specification:
//
//     http://www.gnu.org/software/gettext/manual/gettext.html#MO-Files
//
// TODO: check if the format version is supported
//
// TODO: parse file header; specially Content-Type and Plural-Forms values.
func (c *Catalog) ReadMO(r Reader) error {
	// First word identifies the byte order.
	var order binary.ByteOrder
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return err
	}
	if magic == magicLittleEndian {
		order = binary.LittleEndian
	} else if magic == magicBigEndian {
		order = binary.BigEndian
	} else {
		return errors.New("Unable to identify the file byte order")
	}
	// Next six words:
	// - major+minor format version numbers (ignored)
	// - number of messages
	// - index of messages table
	// - index of translations table
	// - size of hashing table (ignored)
	// - offset of hashing table (ignored)
	w := make([]uint32, 6)
	for i, _ := range w {
		if err := binary.Read(r, order, &w[i]); err != nil {
			return err
		}
	}
	count, mTableIdx, tTableIdx := w[1], w[2], w[3]
	// Build a translations table of strings and translations.
	// Plurals are stored separately with the first message as key.
	var mLen, mIdx, tLen, tIdx uint32
	for i := 0; i < int(count); i++ {
		// Get original message length and position.
		r.Seek(int64(mTableIdx), 0)
		if err := binary.Read(r, order, &mLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &mIdx); err != nil {
			return err
		}
		// Get original message.
		m := make([]byte, mLen)
		if _, err := r.ReadAt(m, int64(mIdx)); err != nil {
			return err
		}
		// Get translation length and position.
		r.Seek(int64(tTableIdx), 0)
		if err := binary.Read(r, order, &tLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &tIdx); err != nil {
			return err
		}
		// Get translation.
		t := make([]byte, tLen)
		if _, err := r.ReadAt(t, int64(tIdx)); err != nil {
			return err
		}
		// Move cursor to next string.
		mTableIdx += 8
		tTableIdx += 8
		mStr, tStr := string(m), string(t)
		if mStr == "" {
			// TODO: this is the file header. Parse it.
			continue
		}
		// Check for context.
		ctx := ""
		if cIdx := strings.Index(mStr, "\x04"); cIdx != -1 {
			ctx = mStr[:cIdx]
			mStr = mStr[cIdx+1:]
			if c.ContextFunc != nil && !c.ContextFunc(ctx) {
				// Context is not valid.
				continue
			}
		}
		// Check for plurals.
		if pIdx := strings.Index(mStr, "\x00"); pIdx != -1 {
			// Store only the first original string and translation in the
			// messages map, and all versions in the two other maps.
			mPlurals := strings.Split(mStr, "\x00")
			tPlurals := strings.Split(tStr, "\x00")
			mStr = mPlurals[0]
			c.messages[mStr] = tPlurals[0]
			c.mPlurals[mStr] = mPlurals
			c.tPlurals[mStr] = tPlurals
		} else {
			c.messages[mStr] = tStr
		}
	}
	return nil
}
