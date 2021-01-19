package serializer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecoder(t *testing.T) {
	testInputs := []Gensis{{
		Method:   "method",
		Strings:  []string{},
		Numbers:  []int64{},
		Binaries: [][]byte{},
	}, {
		Method:   "method",
		Strings:  []string{"string1", "string2"},
		Numbers:  []int64{},
		Binaries: [][]byte{},
	}, {
		Method:   "method",
		Strings:  []string{},
		Numbers:  []int64{123, 48878432, -213},
		Binaries: [][]byte{},
	}, {
		Method:   "method",
		Strings:  []string{"string1"},
		Numbers:  []int64{},
		Binaries: [][]byte{[]byte("binary1")},
	}, {
		Method:   "method",
		Strings:  []string{"string1"},
		Numbers:  []int64{3847, -42983},
		Binaries: [][]byte{[]byte("binary1")},
	}}

	for _, g := range testInputs {
		gPtr := &g
		bb := gPtr.Encode()
		decoded, err := Decoder(bb)
		if assert.NoError(t, err) {
			assert.Equal(t, g, *decoded)
		}
	}
}
