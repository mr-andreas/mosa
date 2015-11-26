package reducer

import (
	"os/exec"

	. "github.com/yoshiyaka/mosa/ast"
)

func Reduce(d []Declaration) ([]Declaration, error) {
	ret := make([]Declaration, 0)

	for _, decl := range d {
		if needed, err := declarationNeeded(&decl); err != nil {
			return nil, err
		} else if needed {
			ret = append(ret, decl)
		}
	}

	return ret, nil
}

func declarationNeeded(d *Declaration) (bool, error) {
	if d.Type != "exec" {
		// This is not an exec declaration, so it will not be needed.
		return false, nil
	}

	var unless *Prop
	for _, p := range d.Props {
		if p.Name == "unless" {
			unless = &p
			break
		}
	}

	if unless == nil {
		// This is an exec declaration without the 'unless' parameter. We always
		// need to execute this, since there is no way for us to determine its
		// state.
		return true, nil
	}

	cmd := exec.Command("/bin/bash", "-c", string(unless.Value.(QuotedString)))
	if err := cmd.Run(); err != nil {
		if _, isExitError := err.(*exec.ExitError); isExitError {
			// The command returned a non-0 status. In other words, the "unless"
			// returned false. This declaration is needed.
			return true, nil
		} else {
			// Uh-oh, some more serious error (such as trouble forking)
			return false, err
		}
	} else {
		// 'unless' returned true, the declaration is not needed.
		return false, nil
	}
}
