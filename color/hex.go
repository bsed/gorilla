// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package color

import (
	"fmt"
	"image/color"
	"strconv"
)

var HexModel = color.ModelFunc(hexModel)

// Hex represents an RGB color in hexadecimal format.
//
// The length must be 3 or 6 characters, preceded or not by a '#'.
type Hex string

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the Hex.
func (c Hex) RGBA() (uint32, uint32, uint32, uint32) {
	r8, g8, b8 := HexToRGB(c)
	return uint32(r8), uint32(g8), uint32(b8), 0xffff
}

// hexModel converts a color.Color to Hex.
func hexModel(c color.Color) color.Color {
	if _, ok := c.(Hex); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGBToHex(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// RGBToHex converts an RGB triple to a Hex string.
func RGBToHex(r, g, b uint8) Hex {
	return Hex(fmt.Sprintf("%02X%02X%02X", r, g, b))
}

// HexToRGB converts an Hex string to a RGB triple.
func HexToRGB(h Hex) (uint8, uint8, uint8) {
	size := len(h)
	if size > 0 && h[0] == '#' {
		h = h[1:]
		size -= 1
	}
	if size != 3 && size != 6 {
		return 0, 0, 0
	}
	if size == 3 {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
	}
	r64, err1 := strconv.ParseUint(string(h[:2]), 16, 8)
	g64, err2 := strconv.ParseUint(string(h[2:4]), 16, 8)
	b64, err3 := strconv.ParseUint(string(h[4:]), 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0
	}
	return uint8(r64), uint8(g64), uint8(b64)
}
