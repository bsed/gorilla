// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"

	"code.google.com/p/gorilla/i18n/gettext/pluralforms"
)

const (
	magicBigEndian    uint32 = 0xde120495
	magicLittleEndian uint32 = 0x950412de
)

// ReadMo fills a catalog with translations from a GNU MO file.
func ReadMo(c *Catalog, r Reader) error {
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
	count, mTableIdx, tTableIdx := int(idx[0]), int64(idx[1]), int64(idx[2])
	// Build a translations table of strings and translations.
	// Plurals are stored separately with the first message as key.
	var mLen, mIdx, tLen, tIdx uint32
	for i := 0; i < count; i++ {
		// Get message length and position.
		r.Seek(mTableIdx, 0)
		if err := binary.Read(r, order, &mLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &mIdx); err != nil {
			return err
		}
		// Get message.
		mb := make([]byte, mLen)
		r.Seek(int64(mIdx), 0)
		if err := binary.Read(r, order, mb); err != nil {
			return err
		}
		// Get translation length and position.
		r.Seek(tTableIdx, 0)
		if err := binary.Read(r, order, &tLen); err != nil {
			return err
		}
		if err := binary.Read(r, order, &tIdx); err != nil {
			return err
		}
		// Get translation.
		tb := make([]byte, tLen)
		r.Seek(int64(tIdx), 0)
		if err := binary.Read(r, order, tb); err != nil {
			return err
		}
		// Move cursor to next message.
		mTableIdx += 8
		tTableIdx += 8
		// Is this is the file header?
		if len(mb) == 0 {
			readMoHeader(c, string(tb))
			continue
		}
		// Check for context.
		mStr, tStr := string(mb), string(tb)
		var ctx string
		if ctxIdx := strings.Index(mStr, "\x04"); ctxIdx != -1 {
			ctx = mStr[:ctxIdx]
			mStr = mStr[ctxIdx+1:]
		}

		var msg Message
		if keyIdx := strings.Index(mStr, "\x00"); keyIdx == -1 {
			// Singular.
			msg = &SimpleMessage{Src: mStr, Dst: tStr, Ctx: ctx}
		} else {
			// Plural.
			msg = &PluralMessage{
				Src: strings.Split(mStr, "\x00"),
				Dst: strings.Split(tStr, "\x00"),
				Ctx: ctx,
			}
		}
		c.Add(msg)
	}
	return nil
}

// readMoHeader parses the translations metadata following GNU .mo conventions.
//
// Ported from Python's gettext.GNUTranslations.
func readMoHeader(c *Catalog, str string) {
	var lastk string
	for _, item := range strings.Split(str, "\n") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if i := strings.Index(item, ":"); i != -1 {
			k := strings.ToLower(strings.TrimSpace(item[:i]))
			v := strings.TrimSpace(item[i+1:])
			c.Header[k] = v
			lastk = k
			switch k {
			// TODO: extract and apply charset from content-type?
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
			c.Header[lastk] += "\n" + item
		}
	}
}

// ----------------------------------------------------------------------------

type moMessages struct {
	src     *bytes.Buffer
	dst     *bytes.Buffer
	srcIdx  uint32
	dstIdx  uint32
	srcList []uint32
	dstList []uint32
}

// newMoMessages returns pre-computed values for WriteMo.
func newMoMessages(c *Catalog) (count int, idxs []uint32, msgs []byte) {
	// Count messages, sort keys.
	keyMap := make(map[string]bool)
	for k, _ := range c.Messages {
		keyMap[k] = true
		count++
	}
	for _, v := range c.Contexts {
		for k, _ := range v {
			keyMap[k] = true
			count++
		}
	}
	keys := make([]string, len(keyMap))
	i := 0
	for k, _ := range keyMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	// Write all messages.
	m := &moMessages{
		src:    new(bytes.Buffer),
		dst:    new(bytes.Buffer),
		srcIdx: uint32(28 + count*16),
	}
	for _, k := range keys {
		if msg, ok := c.Messages[k]; ok {
			m.append(msg)
		}
		for _, ctx := range c.Contexts {
			if msg, ok := ctx[k]; ok {
				m.append(msg)
			}
		}
	}
	for i := 0; i < len(m.dstList); i += 2 {
		// Increment offset for translations.
		m.dstList[i+1] += m.srcIdx
	}
	m.src.Write(m.dst.Bytes())
	idxs = append(m.srcList, m.dstList...)
	return count, idxs, m.src.Bytes()
}

func (m *moMessages) append(msg Message) {
	src := msg.Context()
	dst := ""
	if src != "" {
		src += "\x04"
	}
	switch t := msg.(type) {
	case *SimpleMessage:
		src += t.Src
		dst = t.Dst
	case *PluralMessage:
		src += strings.Join(t.Src, "\x00")
		dst = strings.Join(t.Dst, "\x00")
	}
	m.src.WriteString(src + "\x00")
	m.dst.WriteString(dst + "\x00")
	sLen, dLen := uint32(len(src)), uint32(len(dst))
	m.srcList = append(m.srcList, sLen, m.srcIdx)
	m.dstList = append(m.dstList, dLen, m.dstIdx)
	m.srcIdx += sLen + 1
	m.dstIdx += dLen + 1
}

// WriteMo writes a compiled catalog to the given writer.
func WriteMo(c *Catalog, w Writer) error {
	order := binary.LittleEndian
	count, idxs, msgs := newMoMessages(c)
	mTableIdx := 28
	tTableIdx := mTableIdx + count*8
	table := []uint32{
		magicLittleEndian, // byte 0:  magic number
		uint32(0),         // byte 4:  major+minor revision number
		uint32(count),     // byte 8:  number of messages
		uint32(mTableIdx), // byte 12: index of messages table
		uint32(tTableIdx), // byte 16: index of translations table
		uint32(0),         // byte 20: size of hashing table
		uint32(0),         // byte 24: offset of hashing table
	}
	if err := binary.Write(w, order, table); err != nil {
		return err
	}
	// At byte 28
	if err := binary.Write(w, order, idxs); err != nil {
		return err
	}
	// At byte 28 + (count * 8) + (count * 8)
	if err := binary.Write(w, order, msgs); err != nil {
		return err
	}
	return nil
}
