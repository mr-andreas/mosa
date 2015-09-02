package common

import (
	"fmt"
	"strings"
)

// A step, for instance a debian package, or a shell command
type Step struct {
	// Step type, such as "deb", "file" or "shell"
	Type string

	// Which item the step handles. Together with Type this is the unique
	// identifier for the step used when setting dependencies. For a
	// deb-package, this will be the package name, for a file it will be the
	// filename, etc.
	Item string

	// Additional arguments. For a file this may for instance be mode or
	// ownership.
	Args map[string]interface{}

	// Dependencies that must be fullfilled before this step can be executed.
	// Each key is a Type, and the values are Item. For instance
	// {"deb": { "apache2", "php" }, "file": { "/etc/php.ini" } }
	Depends map[string][]string
}

func (s *Step) String() string {
	args := ""
	for key, val := range s.Args {
		args += fmt.Sprintf("\t%s: %s\n", key, val)
	}

	depends := ""
	if len(s.Depends) != 0 {
		deps := make([]string, 0, len(s.Depends))
		for key, val := range s.Depends {
			vals := strings.Join(val, ",")
			deps = append(deps, fmt.Sprintf("%s[%s]", key, vals))
		}

		depends = fmt.Sprintf("\tDepends: %s\n", strings.Join(deps, ", "))
	}

	return fmt.Sprintf("%s[%s]:\n%s%s", s.Type, s.Item, args, depends)
}

type Stage struct {
	Steps map[string][]Step
}

func (s *Stage) String() string {
	ret := ""

	for _, steps := range s.Steps {
		for _, step := range steps {
			ret += step.String()
		}
	}

	return ret
}

// A plan consists of a number of stages
type Plan struct {
	Stages []*Stage
}

func (p *Plan) String() string {
	stageStrings := make([]string, len(p.Stages))
	for i, stage := range p.Stages {
		stageStrings[i] = fmt.Sprintf("Stage %d:\n%s", i, stage.String())
	}
	return fmt.Sprintf("Plan:\n%s", strings.Join(stageStrings, "\n"))
}
