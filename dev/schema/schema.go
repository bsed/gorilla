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

// getFieldByPath returns the nested field corresponding to a path in dotted
// notation.
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

// Load fills a struct with form values.
//
// The first parameter must be a pointer to a struct. The second is a map,
// typically url.Values, http.Request.Form or http.Request.MultipartForm.
//
// This function is capable of filling nested structs recursivelly using map
// keys as "paths" in dotted notation.
//
// See the package documentation for a full explanation of the mechanics.
func LoadStruct(src map[string][]string, dst interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("schema: interface must be a pointer to struct")
	}
	rv = rv.Elem()
	var mapKey string
	for path, values := range src {
		field, rest := getFieldByPath(rv, path)
		fieldType := field.Type()
		fieldKind := field.Kind()
		mapKey = ""
		if len(rest) == 1 && fieldKind == reflect.Map {
			// For maps there's a rest of 1, the key.
			mapKey = rest[0]
		} else if len(rest) > 0 {
			// Anything else means that the path is invalid.
			continue
		}
		switch fieldKind {
		case reflect.Map:
			mapKind := fieldType.Elem().Kind()
			if v := getBasicValue(mapKind, values[0]); v.IsValid() {
				if field.IsNil() {
					field.Set(reflect.MakeMap(fieldType))
				}
				field.SetMapIndex(reflect.ValueOf(mapKey), v)
			}
		case reflect.Slice:
			if v := getSliceValue(fieldType, values); v.IsValid() {
				field.Set(v)
			}
		default:
			if v := getBasicValue(fieldKind, values[0]); v.IsValid() {
				field.Set(v)
			}
		}
	}
	return nil
}

func getSliceValue(t reflect.Type, values []string) reflect.Value {
	items := make([]reflect.Value, len(values))
	for key, value := range values {
		if item := getBasicValue(t.Elem().Kind(), value); item.IsValid() {
			items[key] = item
		} else {
			// If a single element is invalid we give up.
			break
		}
	}
	if len(values) == len(items) {
		slice := reflect.MakeSlice(t, 0, 0)
		slice = reflect.Append(slice, items...)
		return slice
	}
	return invalidValue
}

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
		if v, err := strconv.ParseInt(value, 16, 8); err == nil {
			return reflect.ValueOf(int16(v))
		}
	case reflect.Int32:
		if v, err := strconv.ParseInt(value, 32, 8); err == nil {
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
