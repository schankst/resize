// Copyright (c) 2014, Charlie Vieth <charlie.vieth@gmail.com>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

package resize

import "testing"

// Covers [ARCH-3.8].
func Test_FloatToUint8(t *testing.T) {
	var testData = []struct {
		in       float32
		expected uint8
	}{
		{0, 0},
		{255, 255},
		{128, 128},
		{1, 1},
		{256, 255},
	}
	for _, test := range testData {
		actual := floatToUint8(test.in)
		if actual != test.expected {
			t.Fail()
		}
	}
}

// Covers [ARCH-3.8].
func Test_FloatToUint16(t *testing.T) {
	var testData = []struct {
		in       float32
		expected uint16
	}{
		{0, 0},
		{65535, 65535},
		{128, 128},
		{1, 1},
		{65536, 65535},
	}
	for _, test := range testData {
		actual := floatToUint16(test.in)
		if actual != test.expected {
			t.Fail()
		}
	}
}
