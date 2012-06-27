// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

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
		Info:       make(map[string]string),
		src:        make(map[string][]string),
		dst:        make(map[string][]string),
		ord:        make(map[string][][]int),
	}
}

// Catalog stores gettext translations.
type Catalog struct {
	Fallback    *Catalog               // used when a translation is not found
	ContextFunc ContextFunc            // used to select context to load
	PluralFunc  pluralforms.PluralFunc // used to select the plural form index
	Info        map[string]string      // metadata from file header
	src         map[string][]string    // original messages
	dst         map[string][]string    // translation messages
	ord         map[string][][]int     // translation expansion orders
}

// Gettext returns a translation for the given message.
func (c *Catalog) Gettext(msg string) string {
	if dst, ok := c.dst[msg]; ok {
		return dst[0]
	}
	if c.Fallback != nil {
		return c.Fallback.Gettext(msg)
	}
	return msg
}

// Gettextf returns a translation for the given message,
// formatted using fmt.Sprintf().
func (c *Catalog) Gettextf(msg string, a ...interface{}) string {
	if dst, ok := c.dst[msg]; ok {
		return sprintf(dst[0], c.ord[msg][0], a...)
	} else if c.Fallback != nil {
		return c.Fallback.Gettextf(msg, a...)
	}
	return fmt.Sprintf(msg, a...)
}

// Ngettext returns a plural translation for a message according to the
// amount n.
//
// msg1 is used to lookup for a translation, and msg2 is used as the plural
// form fallback if a translation is not found.
func (c *Catalog) Ngettext(msg1, msg2 string, n int) string {
	if dst, ok := c.dst[msg1]; ok && c.PluralFunc != nil {
		if idx := c.PluralFunc(n); idx >= 0 && idx < len(dst) {
			return dst[idx]
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

// Ngettextf returns a plural translation for the given message,
// formatted using fmt.Sprintf().
func (c *Catalog) Ngettextf(msg1, msg2 string, n int, a ...interface{}) string {
	if dst, ok := c.dst[msg1]; ok && c.PluralFunc != nil {
		if idx := c.PluralFunc(n); idx >= 0 && idx < len(dst) {
			return sprintf(dst[idx], c.ord[msg1][idx], a...)
		}
	}
	if c.Fallback != nil {
		return c.Fallback.Ngettextf(msg1, msg2, n, a...)
	}
	if n == 1 {
		return fmt.Sprintf(msg1, a...)
	}
	return fmt.Sprintf(msg2, a...)
}

// ReadMO reads a GNU MO file and writes its messages and translations
// to the catalog.
//
// GNU MO file format specification:
//
//     http://www.gnu.org/software/gettext/manual/gettext.html#MO-Files
//
// Inspired by Python's gettext.GNUTranslations.
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
	// Next two words:
	// - major format version number
	// - minor format version number
	v := make([]uint16, 2)
	for i, _ := range v {
		if err := binary.Read(r, order, &v[i]); err != nil {
			return err
		}
	}
	if v[0] > 1 || v[1] > 1 {
		return fmt.Errorf("Major and minor MO revision numbers must be " +
			"0 or 1, got %d and %d", v[0], v[1])
	}
	// Next five words:
	// - number of messages
	// - index of messages table
	// - index of translations table
	// - size of hashing table (ignored)
	// - offset of hashing table (ignored)
	w := make([]uint32, 5)
	for i, _ := range w {
		if err := binary.Read(r, order, &w[i]); err != nil {
			return err
		}
	}
	count, mTableIdx, tTableIdx := w[0], w[1], w[2]
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
			// This is the file header. Parse it.
			c.readMOHeader(tStr)
			continue
		}
		// Check for context.
		if cIdx := strings.Index(mStr, "\x04"); cIdx != -1 {
			ctx := mStr[:cIdx]
			mStr = mStr[cIdx+1:]
			if c.ContextFunc != nil && !c.ContextFunc(ctx) {
				// Context is not valid.
				continue
			}
		}
		// Store messages, plurals and orderings.
		src := strings.Split(mStr, "\x00")
		dst := strings.Split(tStr, "\x00")
		ord := make([][]int, len(dst))
		key := src[0]
		for k, v := range dst {
			dst[k], ord[k] = parseFmt(v, key)
		}
		c.src[key] = src
		c.dst[key] = dst
		c.ord[key] = ord
	}
	return nil
}

// readMOHeader parses the catalog metadata following GNU .mo conventions.
//
// Ported from Python's gettext.GNUTranslations.
func (c *Catalog) readMOHeader(str string) {
	var lastk string
	for _, item := range strings.Split(str, "\n") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if i := strings.Index(item, ":"); i != -1 {
			k := strings.ToLower(strings.TrimSpace(item[:i]))
			v := strings.TrimSpace(item[i+1:])
			c.Info[k] = v
			lastk = k
			switch k {
			// TODO: extract charset from content-type?
			case "plural-forms":
			L1:
				for _, part := range strings.Split(v, ";") {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 && strings.TrimSpace(kv[0]) == "plural" {
						if fn, err := pluralforms.Parse(kv[1]); err == nil {
							c.PluralFunc = fn
						}
						break L1
					}
				}
			}
		} else if lastk != "" {
			c.Info[lastk] += "\n" + item
		}
	}
}

// ----------------------------------------------------------------------------

// parseFmt converts a string that relies on reordering ability to a standard
// format, e.g., the string "%2$d bytes on %1$s." becomes "%d bytes on %s.".
// The returned indices are used to format the resulting string using
// sprintf().
func parseFmt(dst, src string) (string, []int) {
	var idx []int
	end := len(dst)
	buf := new(bytes.Buffer)
	for i := 0; i < end; {
		lasti := i
		for i < end && dst[i] != '%' {
			i++
		}
		if i > lasti {
			buf.WriteString(dst[lasti:i])
		}
		if i >= end {
			break
		}
		i++
		if i < end && dst[i] == '%' {
			// escaped percent
			buf.WriteString("%%")
			i++
		} else {
			buf.WriteByte('%')
			lasti = i
			for i < end && unicode.IsDigit(rune(dst[i])) {
				i++
			}
			if i > lasti {
				if i < end && dst[i] == '$' {
					// extract number, skip dollar sign
					pos, _ := strconv.ParseInt(dst[lasti:i], 10, 0)
					idx = append(idx, int(pos))
					i++
				} else {
					buf.WriteString(dst[lasti:i])
				}
			}
		}
	}
	return buf.String(), idx
}

// sprintf applies fmt.Sprintf() on a string that relies on reordering
// ability, e.g., for the string "%2$d bytes free on %1$s.", the order of
// arguments must be inverted.
func sprintf(format string, order []int, a ...interface{}) string {
	if len(order) == 0 {
		return fmt.Sprintf(format, a...)
	}
	b := make([]interface{}, len(order))
	l := len(a)
	for k, v := range order {
		if v < l {
			b[k] = a[v]
		}
	}
	return fmt.Sprintf(format, b...)
}
