package ast

import "fmt"

type QuotedString string

func (qs QuotedString) String() string { return fmt.Sprintf("'%s'", string(qs)) }

// A double-quoted interpolated string which may contain variables. For instance
// "php5-$module" or "/home/$user".
//
// It consists of a number of segments which are parsed directly in bison, where
// each segment is either a raw string, or a variable name. For instance, the
// string "/home/$user/.config-{$app}" will be interpreted as
// [ "/home/", $user, "/.config-", $app ].
type InterpolatedString struct {
	LineNum int

	// Each segment will be either a string, or a VariableName.
	Segments []interface{}
}

func (is *InterpolatedString) String() string {
	str := `"`
	for _, seg := range is.Segments {
		switch seg.(type) {
		case string:
			str += seg.(string)
		case VariableName:
			str += seg.(VariableName).Str
		default:
			panic("Bad segment type")
		}
	}
	str += `"`

	return str
}
