// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

var invalidValue = reflect.Value{}

// LoadStruct fills a struct with values from a map.
//
// The first parameter is a map, typically url.Values from an HTTP request.
// Keys are "paths" in dotted notation to the struct fields and nested structs.
//
// The second parameter must be a pointer to a struct.
//
// See the package documentation for a full explanation of the mechanics.
func LoadStruct(src map[string][]string, dst interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("schema: interface must be a pointer to struct")
	}
	rv = rv.Elem()
	for path, values := range src {
		loadStructPath(rv, path, values)
	}
	return nil
}

// loadStructPath
func loadStructPath(rv reflect.Value, path string, values []string) {
	field, rest := getFieldByPath(rv, path)
	switch field.Kind() {
	case reflect.Map:
		loadMap(field, values, rest)
	case reflect.Slice:
		loadSlice(field, values, rest)
	default:
		if len(rest) == 0 {
			// Nothing should be remaining in the path.
			if v := getBasicValue(field.Kind(), values[0]); v.IsValid() {
				field.Set(v)
			}
		}
	}
}

// getFieldByPath returns the last valid struct field corresponding to a path
// in dotted notation.
//
// The returned slice contains the path parts that could not be retrieved as
// struct fields. It is empty for fully valid paths.
func getFieldByPath(v reflect.Value, path string) (reflect.Value, []string) {
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if v.Kind() != reflect.Struct {
			break
		}
		name := cache.getNameByAlias(v.Type(), part)
		if name == "" {
			break
		}
		if v = v.FieldByName(name); v.IsValid() {
			parts = parts[1:]
		}
	}
	return v, parts
}

// loadMap
func loadMap(field reflect.Value, values []string, path []string) {
	fieldType := field.Type()
	elemType := fieldType.Elem()
	elemKind := elemType.Kind()
	if len(path) == 1 {
		// For maps there's at least a rest of 1, the key.
		mapKey := path[0]
		if mapKey == "" {
			return
		}
		if v := getBasicValue(elemKind, values[0]); v.IsValid() {
			if field.IsNil() {
				field.Set(reflect.MakeMap(fieldType))
			}
			field.SetMapIndex(reflect.ValueOf(mapKey), v)
		}
	} else {
		// If there's more path to load, map elem must be struct.
		if elemKind != reflect.Struct {
			return
		}
		if field.IsNil() {
			//...
			//field.Set(...)
		}
	}
}

// loadSlice
func loadSlice(field reflect.Value, values []string, path []string) {
	fieldType := field.Type()
	elemType := fieldType.Elem()
	elemKind := elemType.Kind()
	if len(path) == 0 {
		// Simplest case: a slice of basic values.
		items := make([]reflect.Value, len(values))
		for key, value := range values {
			if item := getBasicValue(elemKind, value); item.IsValid() {
				items[key] = item
			} else {
				// If a single element is invalid should we give up
				// or set a zero value?
				// items[key] = reflect.New(elem)
				break
			}
		}
		if len(values) == len(items) {
			slice := reflect.MakeSlice(fieldType, 0, 0)
			field.Set(reflect.Append(slice, items...))
		}
	} else {
		// If there's more path to load, slice elem must be struct.
		if elemKind != reflect.Struct {
			return
		}
		if field.IsNil() {
			//slice := reflect.MakeSlice(fieldType, 0, 0)
			//field.Set(...)
		}
	}
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

// Load is deprecated. Use LoadStruct instead.
func Load(i interface{}, data map[string][]string) error {
	return LoadStruct(data, i)
}
