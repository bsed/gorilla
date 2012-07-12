// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"fmt"

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
	if name := msg.Context(); name == nil {
		c.Messages[msg.Key()] = msg
	} else {
		if _, ok := c.Contexts[*name]; !ok {
			c.Contexts[*name] = make(map[string]Message)
		}
		c.Contexts[*name][msg.Key()] = msg
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
	// Clone returns a copy of the message.
	Clone() Message
	// Context returns the message context. Empty strings are valid contexts,
	// so no context can only be nil.
	Context() *string
}

// ----------------------------------------------------------------------------

type MessageInfo struct {
	Ctx            *string
	UserComments   []string
	SourceComments []string
	References     []string
	Flags          []string
	PrevSingular   string
	PrevPlural     string
	PrevCtx        *string
}

func (m *MessageInfo) Clone() *MessageInfo {
	clone := &MessageInfo{
		PrevSingular: m.PrevSingular,
		PrevPlural:   m.PrevPlural,
		PrevCtx:      m.PrevCtx,
	}
	if m.Ctx != nil {
		clone.Ctx = &(*m.Ctx)
	}
	if m.UserComments != nil {
		clone.UserComments = make([]string, len(m.UserComments))
		copy(clone.UserComments, m.UserComments)
	}
	if m.SourceComments != nil {
		clone.SourceComments = make([]string, len(m.SourceComments))
		copy(clone.SourceComments, m.SourceComments)
	}
	if m.References != nil {
		clone.References = make([]string, len(m.References))
		copy(clone.References, m.References)
	}
	if m.Flags != nil {
		clone.Flags = make([]string, len(m.Flags))
		copy(clone.Flags, m.Flags)
	}
	return clone
}

// ----------------------------------------------------------------------------

type BaseMessage struct {
	info *MessageInfo
}

func (m *BaseMessage) Get() string {
	return ""
}

func (m *BaseMessage) GetPlural(idx int) string {
	return ""
}

func (m *BaseMessage) Context() *string {
	if m.info != nil {
		return m.info.Ctx
	}
	return nil
}

func (m *BaseMessage) SetContext(name string) {
	m.Info().Ctx = &name
}

func (m *BaseMessage) Info() *MessageInfo {
	if m.info == nil {
		m.info = &MessageInfo{}
	}
	return m.info
}

// ----------------------------------------------------------------------------

// SimpleMessage is a message without plural forms.
type SimpleMessage struct {
	BaseMessage
	Src string
	Dst string
}

func (m *SimpleMessage) Key() string {
	return m.Src
}

func (m *SimpleMessage) Get() string {
	return m.Dst
}

func (m *SimpleMessage) Format(s string, a ...interface{}) string {
	// TODO: use message formatter
	return fmt.Sprintf(s, a...)
}

func (m *SimpleMessage) Clone() Message {
	clone := &SimpleMessage{
		Src:  m.Src,
		Dst:  m.Dst,
	}
	if m.info != nil {
		clone.info = m.info.Clone()
	}
	return clone
}

// ----------------------------------------------------------------------------

// PluralMessage is a message with plural forms.
type PluralMessage struct {
	BaseMessage
	Src []string
	Dst []string
}

func (m *PluralMessage) Key() string {
	if len(m.Src) > 0 {
		return m.Src[0]
	}
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

func (m *PluralMessage) Clone() Message {
	src := make([]string, len(m.Src))
	copy(src, m.Src)
	dst := make([]string, len(m.Dst))
	copy(dst, m.Dst)
	clone := &PluralMessage{
		Src: src,
		Dst: dst,
	}
	if m.info != nil {
		clone.info = m.info.Clone()
	}
	return clone
}
