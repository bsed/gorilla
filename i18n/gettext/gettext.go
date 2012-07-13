// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"errors"
	"fmt"

	"code.google.com/p/gorilla/i18n/gettext/pluralforms"
)

var ErrMissingContext = errors.New("The message doesn't have a context.")

// Key represents a key for a Catalog translation.
type Key struct {
	Src    string // message
	Ctx    string // message context
	HasCtx bool   // differentiates empty context from no context
}

// NewCatalog returns a new Catalog, initializing internal fields.
func NewCatalog() *Catalog {
	return &Catalog{
		Header:     make(map[string]string),
		Messages:   make(map[Key]Message),
		PluralFunc: pluralforms.DefaultPluralFunc,
	}
}

// Catalog stores gettext translations.
//
// Catalog messages can't be modified in-place; they must be removed and
// re-added using Add() after the modifications, because they message key
// depends on the content of the message.
type Catalog struct {
	Header     map[string]string      // meta-data
	Messages   map[Key]Message        // translations
	PluralFunc pluralforms.PluralFunc // used to select the plural form index
	ctx        string                 // active context
	hasCtx     bool                   // whether to use a context
}

// Add adds a message to the catalog.
func (c *Catalog) Add(msg Message) {
	c.Messages[msg.Key()] = msg
}

// Clone returns a copy of the catalog.
func (c *Catalog) Clone() *Catalog {
	clone := NewCatalog()
	clone.PluralFunc = c.PluralFunc
	clone.ctx = c.ctx
	clone.hasCtx = c.hasCtx
	for k, v := range c.Messages {
		clone.Messages[k] = v.Clone()
	}
	return clone
}

// SetContext activates a given context for messages.
func (c *Catalog) SetContext(ctx string) {
	c.ctx = ctx
	c.hasCtx = true
}

// RemoveContext deactivates any context for messages.
func (c *Catalog) RemoveContext() {
	c.ctx = ""
	c.hasCtx = false
}

// Get returns a translation for the given key, or an empty string if the
// key is not found.
//
// Extra arguments or optional, used to format the translation.
func (c *Catalog) Get(key string, a ...interface{}) string {
	if msg, ok := c.Messages[Key{Src: key, Ctx: c.ctx, HasCtx: c.hasCtx}]; ok {
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
	if msg, ok := c.Messages[Key{Src: key, Ctx: c.ctx, HasCtx: c.hasCtx}]; ok {
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
	Key() Key
	// Context returns the message context, which can be an empty string.
	// If there's no context it returns an error.
	Context() (string, error)
	// Get returns a translation for the message.
	Get() string
	// GetPlural returns a plural translation for the message.
	GetPlural(index int) string
	// Format formats the message. Each message can use a specific formatter.
	Format(s string, a ...interface{}) string
	// Clone returns a copy of the message.
	Clone() Message
	// Info returns the message's meta-data, which can be changed in-place.
	Info() *MessageInfo
}

// ----------------------------------------------------------------------------

type MessageInfo struct {
	PrevSingular   string
	PrevPlural     string
	PrevCtx        string
	HasPrevCtx     bool
	UserComments   []string
	SourceComments []string
	References     []string
	Flags          []string
}

func (m *MessageInfo) Clone() *MessageInfo {
	clone := &MessageInfo{
		PrevSingular: m.PrevSingular,
		PrevPlural:   m.PrevPlural,
		PrevCtx:      m.PrevCtx,
		HasPrevCtx:   m.HasPrevCtx,
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

// SimpleMessage is a message without plural forms.
type SimpleMessage struct {
	Src    string
	Dst    string
	Ctx    string
	HasCtx bool
	info   *MessageInfo
}

func (m *SimpleMessage) Key() Key {
	if m.HasCtx {
		return Key{Src: m.Src, Ctx: m.Ctx, HasCtx: true}
	}
	return Key{Src: m.Src}
}

func (m *SimpleMessage) Context() (string, error) {
	if m.HasCtx {
		return m.Ctx, nil
	}
	return "", ErrMissingContext
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

func (m *SimpleMessage) Clone() Message {
	clone := &SimpleMessage{
		Src:    m.Src,
		Dst:    m.Dst,
		Ctx:    m.Ctx,
		HasCtx: m.HasCtx,
	}
	if m.info != nil {
		clone.info = m.info.Clone()
	}
	return clone
}

func (m *SimpleMessage) Info() *MessageInfo {
	if m.info == nil {
		m.info = &MessageInfo{}
	}
	return m.info
}

// ----------------------------------------------------------------------------

// PluralMessage is a message with plural forms.
type PluralMessage struct {
	Src    []string
	Dst    []string
	Ctx    string
	HasCtx bool
	info   *MessageInfo
}

func (m *PluralMessage) Key() Key {
	src := ""
	if len(m.Src) > 0 {
		src = m.Src[0]
	}
	if m.HasCtx {
		return Key{Src: src, Ctx: m.Ctx, HasCtx: true}
	}
	return Key{Src: src}
}

func (m *PluralMessage) Context() (string, error) {
	if m.HasCtx {
		return m.Ctx, nil
	}
	return "", ErrMissingContext
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

func (m *PluralMessage) Clone() Message {
	src := make([]string, len(m.Src))
	copy(src, m.Src)
	dst := make([]string, len(m.Dst))
	copy(dst, m.Dst)
	clone := &PluralMessage{
		Src:    src,
		Dst:    dst,
		Ctx:    m.Ctx,
		HasCtx: m.HasCtx,
	}
	if m.info != nil {
		clone.info = m.info.Clone()
	}
	return clone
}

func (m *PluralMessage) Info() *MessageInfo {
	if m.info == nil {
		m.info = &MessageInfo{}
	}
	return m.info
}
