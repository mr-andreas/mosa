package executor

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/yoshiyaka/mosa/common"
)

type Executor interface {
	Execute(*common.Stage) error
}

func ExecutePlan(p *common.Plan, e Executor) error {
	for _, stage := range p.Stages {
		if err := e.Execute(stage); err != nil {
			return err
		}
	}

	return nil
}

func New(scriptDir string) (Executor, error) {
	e := &executor{
		scriptDir: scriptDir,
		scripts:   map[string]bool{},
	}

	files, err := ioutil.ReadDir(scriptDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Name()[0] == '.' {
			continue
		}

		if file.Mode()&0111 != 0 {
			e.scripts[file.Name()] = true
		}
	}

	return e, nil
}

type executor struct {
	scriptDir string
	scripts   map[string]bool
}

func (e *executor) Execute(stage *common.Stage) error {
	for typ, steps := range stage.Steps {
		if err := e.executeStepsForType(typ, steps); err != nil {
			return err
		}
	}

	return nil
}

func (e *executor) executeStepsForType(typ string, steps []common.Step) error {
	scriptName := typ + "_many.sh"
	if _, ok := e.scripts[scriptName]; !ok {
		return fmt.Errorf("Found no script for type %s, expected %s", typ, scriptName)
	}

	cmd := exec.Command(e.scriptDir + "/" + scriptName)
	for _, item := range steps {
		cmd.Args = append(cmd.Args, item.Item)
	}

	fmt.Println("Exec %s", cmd)

	//	var stdout, stderr bytes.Buffer

	//	cmd.Stderr = &stderr
	//	cmd.Stdout = &stdout

	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(out))

		return err
	}

	return nil
}
