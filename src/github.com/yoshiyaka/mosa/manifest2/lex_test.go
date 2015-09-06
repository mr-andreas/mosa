package manifest2

import (
	"strings"
	"testing"
)

func TestLex(t *testing.T) {
	manifest := `
		class Test {
			foo = bar,
			baz = yup
		}
		
		class Class2 {
			good = text
		}
	`

	if err := Lex(strings.NewReader(manifest)); err != nil {
		t.Fatal(err)
	}
}
