// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"fmt"
	"io"

	"code.google.com/p/gorilla/i18n/gettext/pluralforms"
)

// NewCatalog returns a new Catalog, initializing internal fields.
func NewCatalog() *Catalog {
	return &Catalog{
		Header:     make(map[string]string),
		Messages:   make(map[string]Message),
		Contexts:   make(map[string]map[string]Message),
		PluralFunc: pluralforms.DefaultPluralFunc,
	}
}

// Catalog stores gettext translations.
type Catalog struct {
	Header     map[string]string             // meta-data
	Messages   map[string]Message            // translations, including meta-data
	Contexts   map[string]map[string]Message // context-specific messages
	PluralFunc pluralforms.PluralFunc        // used to select the plural form index
}

// Add adds a message to the catalog.
func (c *Catalog) Add(msg Message) {
	if ctx := msg.Context(); ctx == "" {
		c.Messages[msg.Key()] = msg
	} else {
		if _, ok := c.Contexts[ctx]; !ok {
			c.Contexts[ctx] = make(map[string]Message)
		}
		c.Contexts[ctx][msg.Key()] = msg
	}
}

// Clone returns a copy of the catalog.
func (c *Catalog) Clone() *Catalog {
	clone := NewCatalog()
	for k, v := range c.Messages {
		clone.Messages[k] = v.Clone()
	}
	for k, v := range c.Contexts {
		clone.Contexts[k] = make(map[string]Message)
		for km, vm := range v {
			clone.Contexts[k][km] = vm
		}
	}
	return clone
}

// Context returns a catalog for the given context key, or nil if the
// context doesn't exist.
func (c *Catalog) Context(key string) *Catalog {
	if ctx, ok := c.Contexts[key]; ok {
		clone := c.Clone()
		for k, v := range ctx {
			clone.Messages[k] = v.Clone()
		}
		return clone
	}
	return nil
}

// Get returns a translation for the given key, or an empty string if the
// key is not found.
//
// Extra arguments or optional, used to format the translation.
func (c *Catalog) Get(key string, a ...interface{}) string {
	if msg, ok := c.Messages[key]; ok {
		if a == nil {
			return msg.Get()
		}
		return msg.Format(msg.Get(), a...)
	}
	return ""
}

// GetPlural returns a plural translation for the given key and num,
// or an empty string if the key is not found.
//
// Extra arguments or optional, used to format the translation.
func (c *Catalog) GetPlural(key string, num int, a ...interface{}) string {
	if msg, ok := c.Messages[key]; ok {
		if a == nil {
			return msg.GetPlural(c.PluralIndex(num))
		}
		return msg.Format(msg.GetPlural(c.PluralIndex(num)), a...)
	}
	return ""
}

// PluralIndex returns the plural index for a given number.
//
// This evaluates a Plural-Forms expression.
func (c *Catalog) PluralIndex(num int) int {
	return c.PluralFunc(num)
}

// ----------------------------------------------------------------------------

// Message represents a translation, including meta-data.
type Message interface {
	// Key returns the message's key.
	Key() string
	// Get returns a translation for the message.
	Get() string
	// GetPlural returns a plural translation for the message.
	GetPlural(index int) string
	// Format formats the message. Each message can use a specific formatter.
	Format(s string, a ...interface{}) string
	// Context returns the context of the message, if any.
	Context() string
	// Clone returns a copy of the message.
	Clone() Message
}

// ----------------------------------------------------------------------------

// SimpleMessage is a message without plural forms.
type SimpleMessage struct {
	Src string
	Dst string
	Ctx string
}

func (m *SimpleMessage) Key() string {
	return m.Src
}

func (m *SimpleMessage) Get() string {
	return m.Dst
}

func (m *SimpleMessage) GetPlural(idx int) string {
	return ""
}

func (m *SimpleMessage) Format(s string, a ...interface{}) string {
	// TODO: use message formatter
	return fmt.Sprintf(s, a...)
}

func (m *SimpleMessage) Context() string {
	return m.Ctx
}

func (m *SimpleMessage) Clone() Message {
	return &SimpleMessage{
		Src: m.Src,
		Dst: m.Dst,
		Ctx: m.Ctx,
	}
}

// ----------------------------------------------------------------------------

// PluralMessage is a message with plural forms.
type PluralMessage struct {
	Src []string
	Dst []string
	Ctx string
}

func (m *PluralMessage) Key() string {
	if len(m.Src) > 0 {
		return m.Src[0]
	}
	return ""
}

func (m *PluralMessage) Get() string {
	return ""
}

func (m *PluralMessage) GetPlural(index int) string {
	if index >= 0 && index < len(m.Dst) {
		return m.Dst[index]
	}
	return ""
}

func (m *PluralMessage) Format(s string, a ...interface{}) string {
	// TODO: use message formatter
	return fmt.Sprintf(s, a...)
}

func (m *PluralMessage) Context() string {
	return m.Ctx
}

func (m *PluralMessage) Clone() Message {
	src := make([]string, len(m.Src))
	copy(src, m.Src)
	dst := make([]string, len(m.Dst))
	copy(dst, m.Dst)
	return &PluralMessage{
		Src: src,
		Dst: dst,
		Ctx: m.Ctx,
	}
}

// ----------------------------------------------------------------------------

// Reader wraps the interfaces used to read MO and PO files.
type Reader interface {
	io.Reader
	io.Seeker
}

// Writer wraps the interfaces used to write MO and PO files.
type Writer interface {
	io.Writer
	io.Seeker
}
