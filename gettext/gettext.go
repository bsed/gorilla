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
	"regexp"
	"strconv"
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
		Info:       make(map[string]string),
		src:        make(map[string][]string),
		dst:        make(map[string][]string),
		pos:        make(map[string][][]int),
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
	pos         map[string][][]int     // translation expansion orders
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
		return sprintf(dst[0], c.pos[msg][0], a...)
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
			return sprintf(dst[idx], c.pos[msg1][idx], a...)
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
//
// TODO: check if the format version is supported
//
// MO format revisions (to be confirmed):
// Major revision is 0 or 1. Minor revision is also 0 or 1.
//
// - Major revision 1: supports "I" flag for outdigits in string replacements,
//   e.g., translating "%d" to "%Id". The result is that ASCII digits are
//   replaced with the "outdigits" defined in the LC_CTYPE locale category.
//
// - Minor revision 1: supports reordering ability for string replacements,
//   e.g., using "%2$d" to indicate the position of the replacement.
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
		pos := make([][]int, len(dst))
		for k, v := range dst {
			dst[k], pos[k] = parseFmt(v)
		}
		key := src[0]
		c.src[key] = src
		c.dst[key] = dst
		c.pos[key] = pos
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

var fmtRegexp = regexp.MustCompile(`%\d+\$`)

// parseFmt converts a string that relies on reordering ability to a standard
// format, e.g., the string "%2$d bytes on %1$s." becomes "%d bytes on %s.".
// The returned indices are used to format the string using sprintf().
func parseFmt(format string) (string, []int) {
	matches := fmtRegexp.FindAllStringIndex(format, -1)
	if len(matches) == 0 {
		return format, nil
	}
	buf := new(bytes.Buffer)
	idx := make([]int, 0)
	var i int
	for _, v := range matches {
		i1, i2 := v[0], v[1]
		if i1 > 0 && format[i1-1] == '%' {
			// Ignore escaped sequence.
			buf.WriteString(format[i:i2])
		} else {
			buf.WriteString(format[i : i1+1])
			pos, _ := strconv.ParseInt(format[i1+1:i2-1], 10, 0)
			idx = append(idx, int(pos)-1)
		}
		i = i2
	}
	buf.WriteString(format[i:])
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
