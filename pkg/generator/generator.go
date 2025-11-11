// Package generator generates encoder/decoder code for various languages.
package generator

import (
	"fmt"

	"github.com/shaban/ffire/pkg/schema"
)

// GenerateSwift generates Swift encoder/decoder code.
// Currently unimplemented - use package system instead.
func GenerateSwift(s *schema.Schema) ([]byte, error) {
	return nil, fmt.Errorf("Swift generation not implemented - use ffire generate --lang swift")
}
