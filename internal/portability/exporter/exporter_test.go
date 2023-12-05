package exporter_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikoksr/dbench/internal/pointer"
	"github.com/nikoksr/dbench/internal/portability/exporter"
)

type (
	dummy struct {
		ID      string  `json:"id" csv:"id"`
		Name    *string `json:"name" csv:"name"`
		Age     int     `json:"age" csv:"age"`
		Married bool    `json:"married" csv:"married"`
	}

	testCase struct {
		name     string
		data     []dummy
		expected string
		err      error
	}
)

func TestToCSV(t *testing.T) {
	t.Parallel()

	testCases := []testCase{
		{
			name:     "Nil Writer",
			data:     []dummy{{ID: "1", Name: pointer.To("Test"), Age: 1, Married: true}},
			expected: "id,name,age,married\n1,Test,1,true\n",
		},
		{
			name:     "Empty Data",
			data:     []dummy{},
			expected: "id,name,age,married\n",
		},
		{
			name:     "Nil Data",
			data:     nil,
			expected: "id,name,age,married\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.Buffer{}
			err := exporter.ToCSV(&buf, tc.data)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestToJSON(t *testing.T) {
	t.Parallel()

	testCases := []testCase{
		{
			name:     "Nil Writer",
			data:     []dummy{{ID: "1", Name: pointer.To("Test"), Age: 1, Married: true}},
			expected: "[{\"id\":\"1\",\"name\":\"Test\",\"age\":1,\"married\":true}]\n",
		},
		{
			name:     "Empty Data",
			data:     []dummy{},
			expected: "[]\n",
		},
		{
			name:     "Nil Data",
			data:     nil,
			expected: "null\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.Buffer{}
			err := exporter.ToJSON(&buf, tc.data)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}
