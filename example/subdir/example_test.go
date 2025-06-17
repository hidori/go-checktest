package subdir

import (
	"reflect"
	"testing"
)

func TestNewExample(t *testing.T) {
	tests := []struct {
		name string
		want *Example
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewExample(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewExample() = %v, want %v", got, tt.want)
			}
		})
	}
}
