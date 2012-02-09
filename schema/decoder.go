// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"errors"
	"reflect"
)

// NewDecoder returns a new Decoder.
func NewDecoder() *Decoder {
	return &Decoder{cache: newCache()}
}

// Decoder decodes values from a map[string][]string to a struct.
type Decoder struct {
	cache *cache
}

// RegisterConverter registers a converter function for a custom type.
func (d *Decoder) RegisterConverter(value interface{}, converterFunc Converter) {
	d.cache.conv[reflect.TypeOf(value)] = converterFunc
}

// Decode decodes a map[string][]string to a struct.
//
// The first parameter must be a pointer to a struct.
//
// The second parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
//
// See the package documentation for a full explanation of the mechanics.
func (d *Decoder) Decode(dst interface{}, src map[string][]string) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("schema: interface must be a pointer to struct")
	}
	v = v.Elem()
	t := v.Type()
	for path, values := range src {
		if parts, err := d.cache.parsePath(path, t); err == nil {
			d.decode(v, parts, values)
		}
	}
	return nil
}

// decode fills a struct field using a parsed path.
func (d *Decoder) decode(v reflect.Value, parts []pathPart, values []string) {
	field := fieldByIndex(v, parts[0].path)
	if len(parts) == 1 {
		// Simple case.
		switch field.Kind() {
		case reflect.Slice:
			items := make([]reflect.Value, len(values))
			elemT := field.Type().Elem()
			for key, value := range values {
				if conv := d.cache.conv[elemT]; conv != nil {
					if item := conv(value); item.IsValid() {
						items[key] = item
					} else {
						// If a single element is invalid should we give up
						// or set a zero value?
						// items[key] = reflect.New(elem)
						break
					}
				} else {
					break
				}
			}
			if len(values) == len(items) {
				slice := reflect.MakeSlice(field.Type(), 0, 0)
				field.Set(reflect.Append(slice, items...))
			}
		default:
			if conv := d.cache.conv[field.Type()]; conv != nil {
				if v := conv(values[0]); v.IsValid() {
					field.Set(v)
				}
			}
		}
		return
	}
	// Let's go recursive.
	idx := parts[0].index
	if field.IsNil() {
		slice := reflect.MakeSlice(field.Type(), idx+1, idx+1)
		field.Set(slice)
	} else if field.Len() < idx+1 {
		// Resize it.
		slice := reflect.MakeSlice(field.Type(), idx+1, idx+1)
		reflect.Copy(slice, field)
		field.Set(slice)
	}
	d.decode(field.Index(idx), parts[1:], values)
}

// fieldByIndex returns the nested field corresponding to index.
// It panics if v's Kind is not struct.
func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, x := range index {
		if v.Type().Kind() == reflect.Ptr {
			newV := reflect.New(v.Type().Elem())
			v.Set(newV)
			v = newV.Elem()
		}
		v = v.Field(x)
	}
	return v
}
