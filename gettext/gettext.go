// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	magicBigEndian    = 0xde120495
	magicLittleEndian = 0x950412de
)

type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// Catalog stores translations for a GNU MO file.
//
// Inspired by Python's gettext.GNUTranslations.
//
// TODO: Gettextf(msg, replacements...) to use with fmt.Sprintf?
type Catalog struct {
	Fallback *Catalog
	Contexts map[string]string
	Messages map[string]string
	Plurals  map[string][]string
}

// Gettext returns a translation for msg.
func (c *Catalog) Gettext(msg string) string {
	if trans, ok := c.Messages[msg]; ok {
		return trans
	}
	if c.Fallback != nil {
		return c.Fallback.Gettext(msg)
	}
	return msg
}

// Ngettext returns a plural translation for msg1 according to the amount n.
func (c *Catalog) Ngettext(msg1, msg2 string, n int) string {
	if plurals, ok := c.Plurals[msg1]; ok {
		if idx := c.pluralIndex(n); idx < len(plurals) {
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

// pluralIndex returns the index of the plural form for the amount n.
//
// This depends on parsing the Plural-Forms header and is still not supported.
// E.g., expression for English and many others:
//
//     Plural-Forms: nplurals=2; plural=n != 1;
//
// ...which translates to the exoression in the body of this function.
func (c *Catalog) pluralIndex(n int) int {
	if n != 1 {
		return 1
	}
	return 0
}

// New catalog builds a translations catalog from a GNU MO file contents.
//
// GNU MO file format specification:
//
//     http://www.gnu.org/software/gettext/manual/gettext.html#MO-Files
func NewCatalog(r Reader) (*Catalog, error) {
	c := &Catalog{
		Contexts: make(map[string]string),
		Messages: make(map[string]string),
		Plurals:  make(map[string][]string),
	}
	if err := WriteCatalog(c, r); err != nil {
		return nil, err
	}
	return c, nil
}

// WriteCatalog parses the GNU MO contents from the reader and writes
// the messages and translations to the given catalog.
//
// GNU MO file format specification:
//
//     http://www.gnu.org/software/gettext/manual/gettext.html#MO-Files
//
// TODO: check if the format version is supported
// TODO: parse file header; specially Content-Type and Plural-Forms values.
func WriteCatalog(c *Catalog, r Reader) error {
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
	// - number of strings
	// - offset of strings table
	// - offset of translations table
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
	// Plurals are stored separately because they don't belong to the
	// same lookup table, per spec.
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
		}
		// Check for plurals.
		if pIdx := strings.Index(mStr, "\x00"); pIdx != -1 {
			// Store only the singular of the original string and translation
			// in the messages map, and all plural forms in the plurals map.
			mStr = mStr[:pIdx]
			tPlurals := strings.Split(tStr, "\x00")
			c.Messages[mStr] = tPlurals[0]
			c.Plurals[mStr] = tPlurals
		} else {
			c.Messages[mStr] = tStr
		}
		if ctx != "" {
			c.Contexts[mStr] = ctx
		}
	}
	return nil
}
