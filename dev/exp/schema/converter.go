// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"os"
	"strconv"
)

// Converter is an interface used to get and convert values from multi-valued
// sources to populate other types.
type Converter interface  {
	Bool(key string) (bool, os.Error)
	Float(key string) (float64, os.Error)
	Int(key string) (int64, os.Error)
	String(key string) (string, os.Error)
	Uint(key string) (uint64, os.Error)
	BoolMulti(key string) ([]bool, os.Error)
	FloatMulti(key string) ([]float64, os.Error)
	IntMulti(key string) ([]int64, os.Error)
	StringMulti(key string) ([]string, os.Error)
	UintMulti(key string) ([]uint64, os.Error)
}

// StringMapConverter implements Converter for a map[string][]string, such as
// url.Values.
type StringMapConverter struct {
	m map[string][]string
}

// NewStringMapConverter returns a new converter for a map such as url.Values.
func NewStringMapConverter(m map[string][]string) *StringMapConverter {
	return &StringMapConverter{m}
}

// get returns all values for a given key.
func (p *StringMapConverter) get(key string) ([]string, os.Error) {
	if p.m != nil {
		if v, ok := p.m[key]; ok {
			if v == nil || len(v) == 0 {
				return nil, os.NewError("Array is empty.")
			}
			return v, nil
		}
		return nil, os.NewError("Key doesn't exist.")
	}
	return nil, os.NewError("Map is empty.")
}

// Bool returns a single value converted to bool for the given key.
func (p *StringMapConverter) Bool(key string) (bool, os.Error) {
	values, err := p.get(key)
	if err == nil {
		return strconv.Atob(values[0])
	}
	return false, err
}

// Float returns a single value converted to float64 for the given key.
func (p *StringMapConverter) Float(key string) (float64, os.Error) {
	values, err := p.get(key)
	if err == nil {
		return strconv.Atof64(values[0])
	}
	return 0, err
}

// Int returns a single value converted to int64 for the given key.
func (p *StringMapConverter) Int(key string) (int64, os.Error) {
	values, err := p.get(key)
	if err == nil {
		return strconv.Atoi64(values[0])
	}
	return 0, err
}

// String returns a single value converted to string for the given key.
func (p *StringMapConverter) String(key string) (string, os.Error) {
	values, err := p.get(key)
	if err == nil {
		return values[0], nil
	}
	return "", err
}

// Uint returns a single value converted to uint64 for the given key.
func (p *StringMapConverter) Uint(key string) (uint64, os.Error) {
	values, err := p.get(key)
	if err == nil {
		return strconv.Atoui64(values[0])
	}
	return 0, err
}

// BoolMulti returns all values converted to bool for the given key.
func (p *StringMapConverter) BoolMulti(key string) ([]bool, os.Error) {
	values, err := p.get(key)
	if err != nil {
		return nil, err
	}
	res := make([]bool, len(values))
	errors := make(ErrMulti, len(values))
	var hasError bool
	for i, v := range values {
		res[i], errors[i] = strconv.Atob(v)
		if errors[i] != nil {
			hasError = true
		}
	}
	if hasError {
		return res, errors
	}
	return res, nil
}

// FloatMulti returns all values converted to float64 for the given key.
func (p *StringMapConverter) FloatMulti(key string) ([]float64, os.Error) {
	values, err := p.get(key)
	if err != nil {
		return nil, err
	}
	res := make([]float64, len(values))
	errors := make(ErrMulti, len(values))
	var hasError bool
	for i, v := range values {
		res[i], errors[i] = strconv.Atof64(v)
		if errors[i] != nil {
			hasError = true
		}
	}
	if hasError {
		return res, errors
	}
	return res, nil
}

// IntMulti returns all values converted to int64 for the given key.
func (p *StringMapConverter) IntMulti(key string) ([]int64, os.Error) {
	values, err := p.get(key)
	if err != nil {
		return nil, err
	}
	res := make([]int64, len(values))
	errors := make(ErrMulti, len(values))
	var hasError bool
	for i, v := range values {
		res[i], errors[i] = strconv.Atoi64(v)
		if errors[i] != nil {
			hasError = true
		}
	}
	if hasError {
		return res, errors
	}
	return res, nil
}

// StringMulti returns all values converted to string for the given key.
func (p *StringMapConverter) StringMulti(key string) ([]string, os.Error) {
	values, err := p.get(key)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// UintMulti returns all values converted to uint64 for the given key.
func (p *StringMapConverter) UintMulti(key string) ([]uint64, os.Error) {
	values, err := p.get(key)
	if err != nil {
		return nil, err
	}
	res := make([]uint64, len(values))
	errors := make(ErrMulti, len(values))
	var hasError bool
	for i, v := range values {
		res[i], errors[i] = strconv.Atoui64(v)
		if errors[i] != nil {
			hasError = true
		}
	}
	if hasError {
		return res, errors
	}
	return res, nil
}
