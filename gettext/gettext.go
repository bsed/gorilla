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
type Catalog struct {
	Fallback *Catalog
	Contexts map[string]string
	Messages map[string]string
	Plurals  map[string][]string
}

func (c *Catalog) Gettext(msg string) string {
	if trans, ok := c.Messages[msg]; ok {
		return trans
	}
	if c.Fallback != nil {
		return c.Fallback.Gettext(msg)
	}
	return msg
}

func (c *Catalog) Ngettext(msg1, msg2 string, n int) string {
	if plurals, ok := c.Plurals[msg1]; ok {
		idx := c.pluralIndex(n)
		if idx < len(plurals) {
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

func (c *Catalog) pluralIndex(n int) int {
	// TODO
	return n
}

// New catalog builds a translations catalog from a GNU MO file contents.
//
// GNU MO file format specification:
//
//     http://www.gnu.org/software/gettext/manual/gettext.html#MO-Files
//
// TODO: check if the format version is supported
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

// WriteCatalog parses the GNU MO contents from the reader and writes it to
// the given catalog.
//
// GNU MO file format specification:
//
//     http://www.gnu.org/software/gettext/manual/gettext.html#MO-Files
//
// TODO: check if the format version is supported
// TODO: parse file metadata
func WriteCatalog(c *Catalog, r Reader) error {
	// First word identifies the byte order.
	var magic uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return err
	}
	var order binary.ByteOrder
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
		// The original message length and position.
		r.Seek(int64(mTableIdx), 0)
		if err := binary.Read(r, order, &mLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &mIdx); err != nil {
			return err
		}
		// The translation length and position.
		r.Seek(int64(tTableIdx), 0)
		if err := binary.Read(r, order, &tLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &tIdx); err != nil {
			return err
		}
		// The original message.
		s := make([]byte, mLen)
		if _, err := r.ReadAt(s, int64(mIdx)); err != nil {
			return err
		}
		// The translation.
		t := make([]byte, tLen)
		if _, err := r.ReadAt(t, int64(tIdx)); err != nil {
			return err
		}
		mStr, tStr := string(s), string(t)
		// Check for context.
		ctx := ""
		if cIdx := strings.Index(mStr, "\x04"); cIdx != -1 {
			ctx = mStr[:cIdx]
			mStr = mStr[cIdx+1:]
		}
		// Check for plurals.
		if pIdx := strings.Index(mStr, "\x00"); pIdx != -1 {
			// Store only the singular of the original string and translation
			// in the strings map, and all plural forms in the plurals map.
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
		// Move to next string.
		mTableIdx += 8
		tTableIdx += 8
	}
	return nil
}
