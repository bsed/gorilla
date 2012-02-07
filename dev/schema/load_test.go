// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"testing"
)

type S1 struct {
	F01 bool
	F02 float32
	F03 float64
	F04 int
	F05 int8
	F06 int16
	F07 int32
	F08 int64
	F09 string
	F10 uint
	F11 uint8
	F12 uint16
	F13 uint32
	F14 uint64
}

func TestBasicValues(t *testing.T) {
	v := map[string][]string{
		"F01": {"true"},
		"F02": {"4.2"},
		"F03": {"4.3"},
		"F04": {"-42"},
		"F05": {"-43"},
		"F06": {"-44"},
		"F07": {"-45"},
		"F08": {"-46"},
		"F09": {"foo"},
		"F10": {"42"},
		"F11": {"43"},
		"F12": {"44"},
		"F13": {"45"},
		"F14": {"46"},
	}
	e := S1{
		F01: true,
		F02: 4.2,
		F03: 4.3,
		F04: -42,
		F05: -43,
		F06: -44,
		F07: -45,
		F08: -46,
		F09: "foo",
		F10: 42,
		F11: 43,
		F12: 44,
		F13: 45,
		F14: 46,
	}
	s := &S1{}
	_ = LoadStruct(v, s)
	if s.F01 != e.F01 {	t.Errorf("F01: expected %v, got %v", e.F01, s.F01) }
	if s.F02 != e.F02 {	t.Errorf("F02: expected %v, got %v", e.F02, s.F02) }
	if s.F03 != e.F03 {	t.Errorf("F03: expected %v, got %v", e.F03, s.F03) }
	if s.F04 != e.F04 {	t.Errorf("F04: expected %v, got %v", e.F04, s.F04) }
	if s.F05 != e.F05 {	t.Errorf("F05: expected %v, got %v", e.F05, s.F05) }
	if s.F06 != e.F06 {	t.Errorf("F06: expected %v, got %v", e.F06, s.F06) }
	if s.F07 != e.F07 {	t.Errorf("F07: expected %v, got %v", e.F07, s.F07) }
	if s.F08 != e.F08 {	t.Errorf("F08: expected %v, got %v", e.F08, s.F08) }
	if s.F09 != e.F09 {	t.Errorf("F09: expected %v, got %v", e.F09, s.F09) }
	if s.F10 != e.F10 {	t.Errorf("F10: expected %v, got %v", e.F10, s.F10) }
	if s.F11 != e.F11 {	t.Errorf("F11: expected %v, got %v", e.F11, s.F11) }
	if s.F12 != e.F12 {	t.Errorf("F12: expected %v, got %v", e.F12, s.F12) }
	if s.F13 != e.F13 {	t.Errorf("F13: expected %v, got %v", e.F13, s.F13) }
	if s.F14 != e.F14 {	t.Errorf("F14: expected %v, got %v", e.F14, s.F14) }
}

// ----------------------------------------------------------------------------

type S2 struct {
	F01 []bool
	F02 []float32
	F03 []float64
	F04 []int
	F05 []int8
	F06 []int16
	F07 []int32
	F08 []int64
	F09 []string
	F10 []uint
	F11 []uint8
	F12 []uint16
	F13 []uint32
	F14 []uint64
}

func TestSlices(t *testing.T) {
	v := map[string][]string{
		"F01": {"true", "false", "true"},
		"F02": {"4.2", "4.3", "4.4"},
		"F03": {"4.5", "4.6", "4.7"},
		"F04": {"-42", "-43", "-44"},
		"F05": {"-45", "-46", "-47"},
		"F06": {"-48", "-49", "-50"},
		"F07": {"-51", "-52", "-53"},
		"F08": {"-54", "-55", "-56"},
		"F09": {"foo", "bar", "baz"},
		"F10": {"42", "43", "44"},
		"F11": {"45", "46", "47"},
		"F12": {"48", "49", "50"},
		"F13": {"51", "52", "53"},
		"F14": {"54", "55", "56"},
	}
	e := S2{
		F01: []bool{true, false, true},
		F02: []float32{4.2, 4.3, 4.4},
		F03: []float64{4.5, 4.6, 4.7},
		F04: []int{-42, -43, -44},
		F05: []int8{-45, -46, -47},
		F06: []int16{-48, -49, -50},
		F07: []int32{-51, -52, -53},
		F08: []int64{-54, -55, -56},
		F09: []string{"foo", "bar", "baz"},
		F10: []uint{42, 43, 44},
		F11: []uint8{45, 46, 47},
		F12: []uint16{48, 49, 50},
		F13: []uint32{51, 52, 53},
		F14: []uint64{54, 55, 56},
	}
	s := &S2{}
	_ = LoadStruct(v, s)
	if len(s.F01) != 3 || s.F01[0] != e.F01[0] { t.Errorf("F01: expected %v, got %v", e.F01, s.F01) }
	if len(s.F02) != 3 || s.F02[0] != e.F02[0] { t.Errorf("F02: expected %v, got %v", e.F02, s.F02) }
	if len(s.F03) != 3 || s.F03[0] != e.F03[0] { t.Errorf("F03: expected %v, got %v", e.F03, s.F03) }
	if len(s.F04) != 3 || s.F04[0] != e.F04[0] { t.Errorf("F04: expected %v, got %v", e.F04, s.F04) }
	if len(s.F05) != 3 || s.F05[0] != e.F05[0] { t.Errorf("F05: expected %v, got %v", e.F05, s.F05) }
	if len(s.F06) != 3 || s.F06[0] != e.F06[0] { t.Errorf("F06: expected %v, got %v", e.F06, s.F06) }
	if len(s.F07) != 3 || s.F07[0] != e.F07[0] { t.Errorf("F07: expected %v, got %v", e.F07, s.F07) }
	if len(s.F08) != 3 || s.F08[0] != e.F08[0] { t.Errorf("F08: expected %v, got %v", e.F08, s.F08) }
	if len(s.F09) != 3 || s.F09[0] != e.F09[0] { t.Errorf("F09: expected %v, got %v", e.F09, s.F09) }
	if len(s.F10) != 3 || s.F10[0] != e.F10[0] { t.Errorf("F10: expected %v, got %v", e.F10, s.F10) }
	if len(s.F11) != 3 || s.F11[0] != e.F11[0] { t.Errorf("F11: expected %v, got %v", e.F11, s.F11) }
	if len(s.F12) != 3 || s.F12[0] != e.F12[0] { t.Errorf("F12: expected %v, got %v", e.F12, s.F12) }
	if len(s.F13) != 3 || s.F13[0] != e.F13[0] { t.Errorf("F13: expected %v, got %v", e.F13, s.F13) }
	if len(s.F14) != 3 || s.F14[0] != e.F14[0] { t.Errorf("F14: expected %v, got %v", e.F14, s.F14) }
}

// ----------------------------------------------------------------------------

type S3 struct {
	F01 map[string]bool
	F02 map[string]float32
	F03 map[string]float64
	F04 map[string]int
	F05 map[string]int8
	F06 map[string]int16
	F07 map[string]int32
	F08 map[string]int64
	F09 map[string]string
	F10 map[string]uint
	F11 map[string]uint8
	F12 map[string]uint16
	F13 map[string]uint32
	F14 map[string]uint64
}

func TestMaps(t *testing.T) {
	v := map[string][]string{
		"F01.a": {"true"},
		"F01.b": {"false"},
		"F01.c": {"true"},

		"F02.a": {"4.2"},
		"F02.b": {"4.3"},
		"F02.c": {"4.4"},

		"F03.a": {"4.5"},
		"F03.b": {"4.6"},
		"F03.c": {"4.7"},

		"F04.a": {"-42"},
		"F04.b": {"-43"},
		"F04.c": {"-44"},

		"F05.a": {"-45"},
		"F05.b": {"-46"},
		"F05.c": {"-47"},

		"F06.a": {"-48"},
		"F06.b": {"-49"},
		"F06.c": {"-50"},

		"F07.a": {"-51"},
		"F07.b": {"-52"},
		"F07.c": {"-53"},

		"F08.a": {"-54"},
		"F08.b": {"-55"},
		"F08.c": {"-56"},

		"F09.a": {"foo"},
		"F09.b": {"bar"},
		"F09.c": {"baz"},

		"F10.a": {"42"},
		"F10.b": {"43"},
		"F10.c": {"44"},

		"F11.a": {"45"},
		"F11.b": {"46"},
		"F11.c": {"47"},

		"F12.a": {"48"},
		"F12.b": {"49"},
		"F12.c": {"50"},

		"F13.a": {"51"},
		"F13.b": {"52"},
		"F13.c": {"53"},

		"F14.a": {"54"},
		"F14.b": {"55"},
		"F14.c": {"56"},
	}
	e := S3{
		F01: map[string]bool{"a": true, "b": false, "c": true},
		F02: map[string]float32{"a": 4.2, "b": 4.3, "c": 4.4},
		F03: map[string]float64{"a": 4.5, "b": 4.6, "c": 4.7},
		F04: map[string]int{"a": -42, "b": -43, "c": -44},
		F05: map[string]int8{"a": -45, "b": -46, "c": -47},
		F06: map[string]int16{"a": -48, "b": -49, "c": -50},
		F07: map[string]int32{"a": -51, "b": -52, "c": -53},
		F08: map[string]int64{"a": -54, "b": -55, "c": -56},
		F09: map[string]string{"a": "foo", "b": "bar", "c": "baz"},
		F10: map[string]uint{"a": 42, "b": 43, "c": 44},
		F11: map[string]uint8{"a": 45, "b": 46, "c": 47},
		F12: map[string]uint16{"a": 48, "b": 49, "c": 50},
		F13: map[string]uint32{"a": 51, "b": 52, "c": 53},
		F14: map[string]uint64{"a": 54, "b": 55, "c": 56},
	}
	s := &S3{}
	_ = LoadStruct(v, s)
	if len(s.F01) != 3 || s.F01["a"] != e.F01["a"] { t.Errorf("F01: expected %v, got %v", e.F01, s.F01) }
	if len(s.F02) != 3 || s.F02["a"] != e.F02["a"] { t.Errorf("F02: expected %v, got %v", e.F02, s.F02) }
	if len(s.F03) != 3 || s.F03["a"] != e.F03["a"] { t.Errorf("F03: expected %v, got %v", e.F03, s.F03) }
	if len(s.F04) != 3 || s.F04["a"] != e.F04["a"] { t.Errorf("F04: expected %v, got %v", e.F04, s.F04) }
	if len(s.F05) != 3 || s.F05["a"] != e.F05["a"] { t.Errorf("F05: expected %v, got %v", e.F05, s.F05) }
	if len(s.F06) != 3 || s.F06["a"] != e.F06["a"] { t.Errorf("F06: expected %v, got %v", e.F06, s.F06) }
	if len(s.F07) != 3 || s.F07["a"] != e.F07["a"] { t.Errorf("F07: expected %v, got %v", e.F07, s.F07) }
	if len(s.F08) != 3 || s.F08["a"] != e.F08["a"] { t.Errorf("F08: expected %v, got %v", e.F08, s.F08) }
	if len(s.F09) != 3 || s.F09["a"] != e.F09["a"] { t.Errorf("F09: expected %v, got %v", e.F09, s.F09) }
	if len(s.F10) != 3 || s.F10["a"] != e.F10["a"] { t.Errorf("F10: expected %v, got %v", e.F10, s.F10) }
	if len(s.F11) != 3 || s.F11["a"] != e.F11["a"] { t.Errorf("F11: expected %v, got %v", e.F11, s.F11) }
	if len(s.F12) != 3 || s.F12["a"] != e.F12["a"] { t.Errorf("F12: expected %v, got %v", e.F12, s.F12) }
	if len(s.F13) != 3 || s.F13["a"] != e.F13["a"] { t.Errorf("F13: expected %v, got %v", e.F13, s.F13) }
	if len(s.F14) != 3 || s.F14["a"] != e.F14["a"] { t.Errorf("F14: expected %v, got %v", e.F14, s.F14) }
}

// ----------------------------------------------------------------------------

type S4 struct {
	F01 S1 `schema:"foo"`
	F02 S2 `schema:"bar"`
	F03 S3 `schema:"baz"`
}

func TestNestedWithNames(t *testing.T) {
	v := map[string][]string{
		"foo.F14": {"46"},
		"bar.F14": {"54", "55", "56"},
		"baz.F14.a": {"54"},
		"baz.F14.b": {"55"},
		"baz.F14.c": {"56"},
	}
	e := S4{
		F01: S1{F14: 46},
		F02: S2{F14: []uint64{54, 55, 56}},
		F03: S3{F14: map[string]uint64{"a": 54, "b": 55, "c": 56}},
	}
	s := &S4{}
	_ = LoadStruct(v, s)
	if s.F01.F14 != e.F01.F14 {	t.Errorf("F14: expected %v, got %v", e.F01.F14, s.F01.F14) }
	if len(s.F02.F14) != 3 || s.F02.F14[0] != e.F02.F14[0] { t.Errorf("F14: expected %v, got %v", e.F02.F14, s.F02.F14) }
	if len(s.F03.F14) != 3 || s.F03.F14["a"] != e.F03.F14["a"] { t.Errorf("F14: expected %v, got %v", e.F03.F14, s.F03.F14) }
}

// ----------------------------------------------------------------------------

type S5 struct {
	F02 int
	F03 int
}

type S6 struct {
	F01 []S1
}

// TODO: slices of structs
func TestSlicesOfStructs(t *testing.T) {
	_ = map[string][]string{
		"F01.F02": {"42", "43", "44"},
		"F01.F03": {"45", "46", "47"},
	}
	_ = &S6{}
	//_ = LoadStruct(v, s)
}

// ----------------------------------------------------------------------------

// TODO: maps of structs
func TestMapsOfStructs(t *testing.T) {
}

// ----------------------------------------------------------------------------

// TODO: custom types
func TestCustomTypes(t *testing.T) {
}
