// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"errors"
	"reflect"
	"strconv"
)

var invalidValue = reflect.Value{}

// TODO: make cache part of a loader instance
var cache = structCache{m: make(map[string]*structInfo)}

// LoadStruct fills a struct with values from a map.
//
// The first parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
//
// The second parameter must be a pointer to a struct.
//
// See the package documentation for a full explanation of the mechanics.
func LoadStruct(src map[string][]string, dst interface{}) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("schema: interface must be a pointer to struct")
	}
	v = v.Elem()
	t := v.Type()
	for path, values := range src {
		if parts, err := cache.parsePath(path, t); err == nil {
			loadPath(v, parts, values)
		}
	}
	return nil
}

func loadPath(v reflect.Value, parts []pathPart, values []string) {
	field := v.FieldByIndex(parts[0].path)
	if len(parts) == 1 {
		// Simple case.
		switch field.Kind() {
		case reflect.Slice:
			kind := field.Type().Elem().Kind()
			items := make([]reflect.Value, len(values))
			for key, value := range values {
				if item := getBasicValue(kind, value); item.IsValid() {
					items[key] = item
				} else {
					// If a single element is invalid should we give up
					// or set a zero value?
					// items[key] = reflect.New(elem)
					break
				}
			}
			if len(values) == len(items) {
				slice := reflect.MakeSlice(field.Type(), 0, 0)
				field.Set(reflect.Append(slice, items...))
			}
		default:
			if v := getBasicValue(field.Kind(), values[0]); v.IsValid() {
				field.Set(v)
			}
		}
		return
	}
	// Slice of structs. Let's go recursive.
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
	loadPath(field.Index(idx), parts[1:], values)
}

// getBasicValue returns a reflect.Value for a basic type.
func getBasicValue(kind reflect.Kind, value string) reflect.Value {
	switch kind {
	case reflect.Bool:
		if v, err := strconv.ParseBool(value); err == nil {
			return reflect.ValueOf(v)
		}
	case reflect.Float32:
		if v, err := strconv.ParseFloat(value, 32); err == nil {
			return reflect.ValueOf(float32(v))
		}
	case reflect.Float64:
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return reflect.ValueOf(v)
		}
	case reflect.Int:
		if v, err := strconv.ParseInt(value, 10, 0); err == nil {
			return reflect.ValueOf(int(v))
		}
	case reflect.Int8:
		if v, err := strconv.ParseInt(value, 10, 8); err == nil {
			return reflect.ValueOf(int8(v))
		}
	case reflect.Int16:
		if v, err := strconv.ParseInt(value, 10, 16); err == nil {
			return reflect.ValueOf(int16(v))
		}
	case reflect.Int32:
		if v, err := strconv.ParseInt(value, 10, 32); err == nil {
			return reflect.ValueOf(int32(v))
		}
	case reflect.Int64:
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return reflect.ValueOf(v)
		}
	case reflect.String:
		return reflect.ValueOf(value)
	case reflect.Uint:
		if v, err := strconv.ParseUint(value, 10, 0); err == nil {
			return reflect.ValueOf(uint(v))
		}
	case reflect.Uint8:
		if v, err := strconv.ParseUint(value, 10, 8); err == nil {
			return reflect.ValueOf(uint8(v))
		}
	case reflect.Uint16:
		if v, err := strconv.ParseUint(value, 10, 16); err == nil {
			return reflect.ValueOf(uint16(v))
		}
	case reflect.Uint32:
		if v, err := strconv.ParseUint(value, 10, 32); err == nil {
			return reflect.ValueOf(uint32(v))
		}
	case reflect.Uint64:
		if v, err := strconv.ParseUint(value, 10, 64); err == nil {
			return reflect.ValueOf(v)
		}
	}
	return invalidValue
}
