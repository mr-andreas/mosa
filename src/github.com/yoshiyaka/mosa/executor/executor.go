package executor

import (
	"fmt"
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
	e := &executor{}

	return e, nil
}

type executor struct {
}

func (e *executor) Execute(stage *common.Stage) error {
	for _, step := range stage.Steps["exec"] {
		if err := e.executeStep(&step); err != nil {
			return err
		}
	}

	return nil
}

func (e *executor) executeStep(step *common.Step) error {
	cmd := exec.Command(step.Item)
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
