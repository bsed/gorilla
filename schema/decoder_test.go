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

func TestBasicValue(t *testing.T) {
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
	_ = NewDecoder().Decode(s, v)
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

func TestSlice(t *testing.T) {
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
	_ = NewDecoder().Decode(s, v)
	if s.F01 == nil || len(s.F01) != 3 {
		t.Errorf("F01: nil or wrong len")
	} else if s.F01[0] != e.F01[0] || s.F01[1] != e.F01[1] || s.F01[2] != e.F01[2] {
		t.Errorf("F01: expected %v, got %v", e.F01, s.F01)
	}
	if s.F02 == nil || len(s.F02) != 3 {
		t.Errorf("F02: nil or wrong len")
	} else if s.F02[0] != e.F02[0] || s.F02[1] != e.F02[1] || s.F02[2] != e.F02[2] {
		t.Errorf("F02: expected %v, got %v", e.F02, s.F02)
	}
	if s.F03 == nil || len(s.F03) != 3 {
		t.Errorf("F03: nil or wrong len")
	} else if s.F03[0] != e.F03[0] || s.F03[1] != e.F03[1] || s.F03[2] != e.F03[2] {
		t.Errorf("F03: expected %v, got %v", e.F03, s.F03)
	}
	if s.F04 == nil || len(s.F04) != 3 {
		t.Errorf("F04: nil or wrong len")
	} else if s.F04[0] != e.F04[0] || s.F04[1] != e.F04[1] || s.F04[2] != e.F04[2] {
		t.Errorf("F04: expected %v, got %v", e.F04, s.F04)
	}
	if s.F05 == nil || len(s.F05) != 3 {
		t.Errorf("F05: nil or wrong len")
	} else if s.F05[0] != e.F05[0] || s.F05[1] != e.F05[1] || s.F05[2] != e.F05[2] {
		t.Errorf("F05: expected %v, got %v", e.F05, s.F05)
	}
	if s.F06 == nil || len(s.F06) != 3 {
		t.Errorf("F06: nil or wrong len")
	} else if s.F06[0] != e.F06[0] || s.F06[1] != e.F06[1] || s.F06[2] != e.F06[2] {
		t.Errorf("F06: expected %v, got %v", e.F06, s.F06)
	}
	if s.F07 == nil || len(s.F07) != 3 {
		t.Errorf("F07: nil or wrong len")
	} else if s.F07[0] != e.F07[0] || s.F07[1] != e.F07[1] || s.F07[2] != e.F07[2] {
		t.Errorf("F07: expected %v, got %v", e.F07, s.F07)
	}
	if s.F08 == nil || len(s.F08) != 3 {
		t.Errorf("F08: nil or wrong len")
	} else if s.F08[0] != e.F08[0] || s.F08[1] != e.F08[1] || s.F08[2] != e.F08[2] {
		t.Errorf("F08: expected %v, got %v", e.F08, s.F08)
	}
	if s.F09 == nil || len(s.F09) != 3 {
		t.Errorf("F09: nil or wrong len")
	} else if s.F09[0] != e.F09[0] || s.F09[1] != e.F09[1] || s.F09[2] != e.F09[2] {
		t.Errorf("F09: expected %v, got %v", e.F09, s.F09)
	}
	if s.F10 == nil || len(s.F10) != 3 {
		t.Errorf("F10: nil or wrong len")
	} else if s.F10[0] != e.F10[0] || s.F10[1] != e.F10[1] || s.F10[2] != e.F10[2] {
		t.Errorf("F10: expected %v, got %v", e.F10, s.F10)
	}
	if s.F11 == nil || len(s.F11) != 3 {
		t.Errorf("F11: nil or wrong len")
	} else if s.F11[0] != e.F11[0] || s.F11[1] != e.F11[1] || s.F11[2] != e.F11[2] {
		t.Errorf("F11: expected %v, got %v", e.F11, s.F11)
	}
	if s.F12 == nil || len(s.F12) != 3 {
		t.Errorf("F12: nil or wrong len")
	} else if s.F12[0] != e.F12[0] || s.F12[1] != e.F12[1] || s.F12[2] != e.F12[2] {
		t.Errorf("F12: expected %v, got %v", e.F12, s.F12)
	}
	if s.F13 == nil || len(s.F13) != 3 {
		t.Errorf("F13: nil or wrong len")
	} else if s.F13[0] != e.F13[0] || s.F13[1] != e.F13[1] || s.F13[2] != e.F13[2] {
		t.Errorf("F13: expected %v, got %v", e.F13, s.F13)
	}
	if s.F14 == nil || len(s.F14) != 3 {
		t.Errorf("F14: nil or wrong len")
	} else if s.F14[0] != e.F14[0] || s.F14[1] != e.F14[1] || s.F14[2] != e.F14[2] {
		t.Errorf("F14: expected %v, got %v", e.F14, s.F14)
	}
}

// ----------------------------------------------------------------------------

type S3 struct {
	F01 S1   `schema:"name1"`
	F02 S2   `schema:"name2"`
	F03 []S1 `schema:"name3"`
	F04 []S2 `schema:"name4"`
}

func TestNestedStruct(t *testing.T) {
	v := map[string][]string{
		"name1.F01": {"true"},
		"name1.F14": {"42"},

		"name2.F01": {"false", "true", "false"},
		"name2.F14": {"43", "44", "45"},

		"name3.0.F01": {"true"},
		"name3.0.F14": {"42"},
		"name3.1.F01": {"false"},
		"name3.1.F14": {"43"},

		"name4.0.F01": {"true", "false", "true"},
		"name4.0.F14": {"42", "43", "44"},
		"name4.1.F01": {"false", "true", "false"},
		"name4.1.F14": {"45", "46", "47"},
	}
	e := S3{
		F01: S1{
			F01: true,
			F14: 42,
		},
		F02: S2{
			F01: []bool{false, true, false},
			F14: []uint64{43, 44, 45},
		},
		F03: []S1{
			S1{
				F01: true,
				F14: 42,
			},
			S1{
				F01: false,
				F14: 43,
			},
		},
		F04: []S2{
			S2{
				F01: []bool{true, false, true},
				F14: []uint64{42, 43, 44},
			},
			S2{
				F01: []bool{false, true, false},
				F14: []uint64{45, 46, 47},
			},
		},
	}
	s := &S3{}
	_ = NewDecoder().Decode(s, v)

	if s.F01.F01 != e.F01.F01 {
		t.Errorf("name1.F01: expected %v, got %v", e.F01.F01, s.F01.F01)
	}
	if s.F01.F14 != e.F01.F14 {
		t.Errorf("name1.F14: expected %v, got %v", e.F01.F14, s.F01.F14)
	}

	if s.F02.F01 == nil || len(s.F02.F01) != 3 {
		t.Errorf("name2.F01: nil or wrong len")
	} else if s.F02.F01[0] != e.F02.F01[0] || s.F02.F01[1] != e.F02.F01[1] || s.F02.F01[2] != e.F02.F01[2] {
		t.Errorf("name2.F01: expected %v, got %v", e.F02.F01, s.F02.F01)
	}
	if s.F02.F14 == nil || len(s.F02.F14) != 3 {
		t.Errorf("name2.F14: nil or wrong len")
	} else if s.F02.F14[0] != e.F02.F14[0] || s.F02.F14[1] != e.F02.F14[1] || s.F02.F14[2] != e.F02.F14[2] {
		t.Errorf("name2.F14: expected %v, got %v", e.F02.F01, s.F02.F14)
	}

	if s.F03 == nil || len(s.F03) != 2 {
		t.Errorf("name3: nil or wrong len")
	} else {
		if s.F03[0].F01 != e.F03[0].F01 {
			t.Errorf("name3.0.F01: expected %v, got %v", e.F03[0].F01, s.F03[0].F01)
		}
		if s.F03[0].F14 != e.F03[0].F14 {
			t.Errorf("name3.0.F14: expected %v, got %v", e.F03[0].F14, s.F03[0].F14)
		}
		if s.F03[1].F01 != e.F03[1].F01 {
			t.Errorf("name3.1.F01: expected %v, got %v", e.F03[1].F01, s.F03[1].F01)
		}
		if s.F03[1].F14 != e.F03[1].F14 {
			t.Errorf("name3.1.F14: expected %v, got %v", e.F03[1].F14, s.F03[1].F14)
		}
	}

	if s.F04 == nil || len(s.F04) != 2 {
		t.Errorf("name4: nil or wrong len")
	} else {
		if s.F04[0].F01 == nil || len(s.F04[0].F01) != 3 {
			t.Errorf("name4.0.F01: nil or wrong len")
		} else if s.F04[0].F01[0] != e.F04[0].F01[0] || s.F04[0].F01[1] != e.F04[0].F01[1] || s.F04[0].F01[2] != e.F04[0].F01[2] {
			t.Errorf("name4.0.F01: expected %v, got %v", e.F04[0].F01, s.F04[0].F01)
		}

		if s.F04[0].F14 == nil || len(s.F04[0].F14) != 3 {
			t.Errorf("name4.0.F14: nil or wrong len")
		} else if s.F04[0].F14[0] != e.F04[0].F14[0] || s.F04[0].F14[1] != e.F04[0].F14[1] || s.F04[0].F14[2] != e.F04[0].F14[2] {
			t.Errorf("name4.0.F14: expected %v, got %v", e.F04[0].F14, s.F04[0].F14)
		}

		if s.F04[1].F01 == nil || len(s.F04[1].F01) != 3 {
			t.Errorf("name4.1.F01: nil or wrong len")
		} else if s.F04[1].F01[0] != e.F04[1].F01[0] || s.F04[1].F01[1] != e.F04[1].F01[1] || s.F04[1].F01[2] != e.F04[1].F01[2] {
			t.Errorf("name4.1.F01: expected %v, got %v", e.F04[1].F01, s.F04[1].F01)
		}

		if s.F04[1].F14 == nil || len(s.F04[1].F14) != 3 {
			t.Errorf("name4.1.F14: nil or wrong len")
		} else if s.F04[1].F14[0] != e.F04[1].F14[0] || s.F04[1].F14[1] != e.F04[1].F14[1] || s.F04[1].F14[2] != e.F04[1].F14[2] {
			t.Errorf("name4.1.F14: expected %v, got %v", e.F04[1].F14, s.F04[1].F14)
		}
	}
}

// ----------------------------------------------------------------------------

type S4 struct {
	F01 *S1
	F02 *S2
	F03 *S4
}

func TestPointer(t *testing.T) {
	v := map[string][]string{
		"F01.F01":         {"true"},
		"F02.F01":         {"true", "false", "true"},
		"F03.F03.F01.F01": {"true"},
	}
	e := S4{
		F01: &S1{
			F01: true,
		},
		F02: &S2{
			F01: []bool{true, false, true},
		},
		F03: &S4{
			F03: &S4{
				F01: &S1{
					F01: true,
				},
			},
		},
	}
	s := &S4{}
	_ = NewDecoder().Decode(s, v)

	if s.F01 == nil || s.F01.F01 != e.F01.F01 {
		t.Errorf("F01.F01: expected %v, got %v", e.F01.F01, s.F01.F01)
	}
	if s.F02 == nil || len(s.F02.F01) != 3 {
		t.Errorf("F02.F01: nil or wrong len")
	} else if s.F02.F01[0] != e.F02.F01[0] || s.F02.F01[1] != e.F02.F01[1] || s.F02.F01[2] != e.F02.F01[2] {
		t.Errorf("F02: expected %v, got %v", e.F01.F01, s.F01.F01)
	}
	if s.F03 == nil || s.F03.F03.F01.F01 != e.F03.F03.F01.F01 {
		t.Errorf("F01.F01: ")
	}
}
