// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"code.google.com/p/gorilla/i18n/gettext/pluralforms"
)

const (
	magicBigEndian    uint32 = 0xde120495
	magicLittleEndian uint32 = 0x950412de
)

// ReadMO loads a catalog from a GNU MO file.
func ReadMO(t *Translations, r Reader) error {
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
	// byte 4: major revision number
	// byte 6: minor revision number
	rev := make([]uint16, 2)
	for k, _ := range rev {
		if err := binary.Read(r, order, &rev[k]); err != nil {
			return err
		}
	}
	if rev[0] > 1 || rev[1] > 1 {
		return fmt.Errorf("Major and minor MO revision numbers must be "+
			"0 or 1, got %d and %d", rev[0], rev[1])
	}
	// Next five words:
	// byte 8:  number of messages
	// byte 12: index of messages table
	// byte 16: index of translations table
	// byte 20: size of hashing table
	// byte 24: offset of hashing table
	idx := make([]uint32, 5)
	for k, _ := range idx {
		if err := binary.Read(r, order, &idx[k]); err != nil {
			return err
		}
	}
	count, mTableIdx, tTableIdx := idx[0], idx[1], idx[2]
	// Build a translations table of strings and translations.
	// Plurals are stored separately with the first message as key.
	var mLen, mIdx, tLen, tIdx uint32
	t.msgOrig = make([][]byte, int(count))
	t.trnOrig = make([][]byte, int(count))
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
		mb := make([]byte, mLen)
		if _, err := r.ReadAt(mb, int64(mIdx)); err != nil {
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
		tb := make([]byte, tLen)
		if _, err := r.ReadAt(tb, int64(tIdx)); err != nil {
			return err
		}
		// Move cursor to next message.
		mTableIdx += 8
		tTableIdx += 8
		t.msgOrig[i], t.trnOrig[i] = mb, tb
		mStr, tStr := string(mb), string(tb)
		if mStr == "" {
			// This is the file header. Parse it.
			readMOHeader(t, tStr)
			continue
		}
		// Check for context.
		if cIdx := strings.Index(mStr, "\x04"); cIdx != -1 {
			if t.ContextFunc != nil && !t.ContextFunc(mStr[:cIdx]) {
				// Context is not valid.
				continue
			}
			mStr = mStr[cIdx+1:]
		}
		// Store messages, plurals and orderings.
		msg := strings.Split(mStr, "\x00")
		trn := strings.Split(tStr, "\x00")
		ord := make([][]int, len(trn))
		key := msg[0]
		for k, v := range trn {
			trn[k], ord[k] = parseFmt(v, key)
		}
		t.msg[key] = msg
		t.trn[key] = trn
		t.ord[key] = ord
	}
	return nil
}

// readMOHeader parses the catalog metadata following GNU .mo conventions.
//
// Ported from Python's gettext.GNUTranslations.
func readMOHeader(t *Translations, str string) {
	var lastk string
	for _, item := range strings.Split(str, "\n") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if i := strings.Index(item, ":"); i != -1 {
			k := strings.ToLower(strings.TrimSpace(item[:i]))
			v := strings.TrimSpace(item[i+1:])
			t.Info[k] = v
			lastk = k
			switch k {
			// TODO: extract charset from content-type?
			case "plural-forms":
			L1:
				for _, part := range strings.Split(v, ";") {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 && strings.TrimSpace(kv[0]) == "plural" {
						if fn, err := pluralforms.Parse(kv[1]); err == nil {
							t.PluralFunc = fn
						}
						break L1
					}
				}
			}
		} else if lastk != "" {
			t.Info[lastk] += "\n" + item
		}
	}
}

// WriteMO writes a compiled catalog to the given writer.
func WriteMO(t *Translations, w Writer) error {
	order := binary.LittleEndian
	// Calculate and store initial values.
	count := len(t.msgOrig)
	mTableIdx := 28
	tTableIdx := mTableIdx + ((count - 1) * 8) + 8
	hIdx := tTableIdx + ((count - 1) * 8) + 8
	idx := []interface{}{
		magicLittleEndian, // byte 0:  magic number
		uint16(1),         // byte 4:  major revision number
		uint16(1),         // byte 6:  minor revision number
		uint32(count),     // byte 8:  number of messages
		uint32(mTableIdx), // byte 12: index of messages table
		uint32(tTableIdx), // byte 16: index of translations table
		uint32(0),         // byte 20: size of hashing table
		uint32(0),         // byte 24: offset of hashing table
	}
	for _, v := range idx {
		if err := binary.Write(w, order, v); err != nil {
			return err
		}
	}
	// Write messages.
	mIdx := uint32(hIdx)
	for _, msg := range t.msgOrig {
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
	for _, trn := range t.trnOrig {
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
