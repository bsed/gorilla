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
	magicBigEndian    uint32 = 0xde120495
	magicLittleEndian uint32 = 0x950412de
)

// Reader wraps the interfaces used to read MO and PO files.
//
// Typically catalogs are provided as os.File.
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

// ContextFunc is used to select the context stored for message disambiguation.
type ContextFunc func(string) bool

// NewCatalog returns a new Catalog, initializing its internal fields.
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
	Fallback    *Catalog               // used when a translation is not found
	ContextFunc ContextFunc            // used to select context to load
	PluralFunc  pluralforms.PluralFunc // used to select the plural form index
	Info        map[string]string      // metadata from file header
	msg         map[string][]string    // messages
	trn         map[string][]string    // translations
	ord         map[string][][]int     // translation expansion ordering
	msgOrig     [][]byte               // original messages, unprocessed
	trnOrig     [][]byte               // original translations, unprocessed
}

// Gettext returns a translation for the given message.
func (c *Catalog) Gettext(msg string) string {
	if trn, ok := c.trn[msg]; ok {
		return trn[0]
	}
	if c.Fallback != nil {
		return c.Fallback.Gettext(msg)
	}
	return msg
}

// Gettextf returns a translation for the given message,
// formatted using fmt.Sprintf().
func (c *Catalog) Gettextf(msg string, a ...interface{}) string {
	if trn, ok := c.trn[msg]; ok {
		return sprintf(trn[0], c.ord[msg][0], a...)
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
	if trn, ok := c.trn[msg1]; ok && c.PluralFunc != nil {
		if idx := c.PluralFunc(n); idx >= 0 && idx < len(trn) {
			return trn[idx]
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
	if trn, ok := c.trn[msg1]; ok && c.PluralFunc != nil {
		if idx := c.PluralFunc(n); idx >= 0 && idx < len(trn) {
			return sprintf(trn[idx], c.ord[msg1][idx], a...)
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

// MO -------------------------------------------------------------------------

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
	// - major format revision number
	// - minor format revision number
	v := make([]uint16, 2)
	for i, _ := range v {
		if err := binary.Read(r, order, &v[i]); err != nil {
			return err
		}
	}
	if v[0] > 1 || v[1] > 1 {
		return fmt.Errorf("Major and minor MO revision numbers must be "+
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
	c.msgOrig = make([][]byte, int(count))
	c.trnOrig = make([][]byte, int(count))
	for i := 0; i < int(count); i++ {
		// Get message length and position.
		r.Seek(int64(mTableIdx), 0)
		if err := binary.Read(r, order, &mLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &mIdx); err != nil {
			return err
		}
		// Get message.
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
		// Move cursor to next message.
		mTableIdx += 8
		tTableIdx += 8
		c.msgOrig[i], c.trnOrig[i] = m, t
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
		msg := strings.Split(mStr, "\x00")
		trn := strings.Split(tStr, "\x00")
		ord := make([][]int, len(trn))
		key := msg[0]
		for k, v := range trn {
			trn[k], ord[k] = parseFmt(v, key)
		}
		c.msg[key] = msg
		c.trn[key] = trn
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

// WriteMO writes a compiled catalog to the given writer.
func (c *Catalog) WriteMO(w Writer) error {
	order := binary.LittleEndian
	// Calculate and store initial values.
	count := len(c.msgOrig)
	mTableIdx := 28
	tTableIdx := mTableIdx + ((count - 1) * 8) + 8
	hIdx := tTableIdx + ((count - 1) * 8) + 8
	idx := []interface{}{
		magicLittleEndian, // byte 0:  magic number
		uint16(1),         // byte 4:  major revision number
		uint16(1),         // byte 6:  minor revision number
		uint32(count),     // byte 8:  number of messages
		uint32(mTableIdx), // byte 12: index of messages table (M0)
		uint32(tTableIdx), // byte 16: index of translations table (T0)
		uint32(0),         // byte 20: size of hashing table (HS)
		uint32(0),         // byte 24: offset of hashing table (H0)
	}
	for _, v := range idx {
		if err := binary.Write(w, order, v); err != nil {
			return err
		}
	}
	// Write messages.
	mIdx := uint32(hIdx)
	for _, msg := range c.msgOrig {
		mLen := uint32(len(msg))
		// Write message length and position.
		w.Seek(int64(mTableIdx), 0)
		if err := binary.Write(w, order, mLen); err != nil {
			return err
		}
		if err := binary.Write(w, order, mIdx); err != nil {
			return err
		}
		// Write message, terminating with a NUL byte.
		if _, err := w.WriteAt(append(msg, '\x00'), int64(mIdx)); err != nil {
			return err
		}
		// Move cursor to next message.
		mTableIdx += 8
		mIdx += mLen + 1
	}
	// Write translations.
	tIdx := uint32(mIdx)
	for _, trn := range c.trnOrig {
		tLen := uint32(len(trn))
		// Write translation length and position.
		w.Seek(int64(tTableIdx), 0)
		if err := binary.Write(w, order, tLen); err != nil {
			return err
		}
		if err := binary.Write(w, order, tIdx); err != nil {
			return err
		}
		// Write translation, terminating with a NUL byte.
		if _, err := w.WriteAt(append(trn, '\x00'), int64(tIdx)); err != nil {
			return err
		}
		// Move cursor to next translation.
		tTableIdx += 8
		tIdx += tLen + 1
	}
	return nil
}

// ----------------------------------------------------------------------------

// parseFmt converts a string that relies on reordering ability to a standard
// format, e.g., the string "%2$d bytes on %1$s." becomes "%d bytes on %s.".
// The returned indices are used to format the resulting string using
// sprintf().
func parseFmt(trn, msg string) (string, []int) {
	var idx []int
	end := len(trn)
	buf := new(bytes.Buffer)
	for i := 0; i < end; {
		lasti := i
		for i < end && trn[i] != '%' {
			i++
		}
		if i > lasti {
			buf.WriteString(trn[lasti:i])
		}
		if i >= end {
			break
		}
		i++
		if i < end && trn[i] == '%' {
			// escaped percent
			buf.WriteString("%%")
			i++
		} else {
			buf.WriteByte('%')
			lasti = i
			for i < end && unicode.IsDigit(rune(trn[i])) {
				i++
			}
			if i > lasti {
				if i < end && trn[i] == '$' {
					// extract number, skip dollar sign
					pos, _ := strconv.ParseInt(trn[lasti:i], 10, 0)
					idx = append(idx, int(pos))
					i++
				} else {
					buf.WriteString(trn[lasti:i])
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
