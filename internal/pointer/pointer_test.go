package pointer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikoksr/dbench/internal/pointer"
)

func TestTo(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want interface{}
	}{
		{
			name: "IntValue_HappyPath",
			v:    5,
			want: 5,
		},
		{
			name: "StringValue_HappyPath",
			v:    "test",
			want: "test",
		},
		{
			name: "StructValue_HappyPath",
			v: struct {
				Field string
			}{Field: "value"},
			want: struct {
				Field string
			}{Field: "value"},
		},
		{
			name: "NilValue_EdgeCase",
			v:    nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pointer.To(tt.v)
			assert.NotNil(t, p)
			assert.Equal(t, tt.want, *p)
		})
	}
}
