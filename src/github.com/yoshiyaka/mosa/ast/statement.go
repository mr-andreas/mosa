package ast

import "fmt"

type If struct {
	LineNum int

	Expression Value
	Block      Block
	Else       *Block
}

func (i *If) String() string {
	if i.Else == nil {
		return fmt.Sprintf("if %s %s", valToStr(i.Expression), i.Block.String())
	} else {
		return fmt.Sprintf(
			"if %s %s else %s", valToStr(i.Expression), i.Block.String(),
			i.Else.String(),
		)
	}
}

// Returns whether the if statements are equal. Line numbers and filenames are
// not taken into consideration.
func IfEquals(i1, i2 *If) bool {
	return ValueEquals(&i1.Expression, &i2.Expression) &&
		BlockEquals(&i1.Block, &i2.Block) &&
		BlockEquals(i1.Else, i2.Else)
}

// Returns whether the if lists are equal. Order is important.
func IfsEquals(i1, i2 []If) bool {
	if len(i1) != len(i2) {
		return false
	}

	for i, _ := range i1 {
		if !IfEquals(&i1[i], &i2[i]) {
			return false
		}
	}

	return true
}
