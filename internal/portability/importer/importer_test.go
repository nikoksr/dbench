package importer_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/nikoksr/dbench/internal/portability/importer"
)

type dummy struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Married bool   `json:"married"`
}

func TestFromJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		want     []dummy
		wantErr  bool
		errMatch error
	}{
		{
			name:  "Valid JSON",
			input: `[{"id":"1","name":"Test","age":30,"married":false}]`,
			want:  []dummy{{ID: "1", Name: "Test", Age: 30, Married: false}},
		},
		{
			name:     "Nil Reader",
			input:    "",
			wantErr:  true,
			errMatch: io.EOF,
		},
		{
			name:    "Invalid JSON",
			input:   `{"id":"1","name":"Test","age":30,"married":false]`,
			wantErr: true,
		},
		{
			name:  "Empty JSON",
			input: `[]`,
			want:  []dummy{},
		},
		{
			name:    "Invalid JSON Format",
			input:   `not a json`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewBufferString(tc.input)
			got, err := importer.FromJSON[[]dummy](reader)

			if (err != nil) != tc.wantErr {
				t.Fatalf("FromJSON() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && tc.errMatch != nil && !errors.Is(err, tc.errMatch) {
				t.Errorf("FromJSON() error = %v, wantErrMatch %v", err, tc.errMatch)
			}

			if !tc.wantErr && !jsonEqual(got, tc.want) {
				t.Errorf("FromJSON() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func jsonEqual(a, b interface{}) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aJSON, bJSON)
}
