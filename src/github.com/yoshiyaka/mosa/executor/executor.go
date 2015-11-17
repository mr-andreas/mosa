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
	cmd := exec.Command("/bin/bash")
	cmd.Args = []string{"/bin/bash", "-c", step.Item}
	fmt.Printf("Exec %s\n", step.Item)

	//	cmd.Stdout = os.Stdout
	//	cmd.Stderr = os.Stderr

	//	var stdout, stderr bytes.Buffer

	//	cmd.Stderr = &stderr
	//	cmd.Stdout = &stdout

	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(string(out))

		return err
	}

	return nil
}
