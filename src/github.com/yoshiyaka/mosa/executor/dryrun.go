package executor

import (
	"fmt"

	"github.com/yoshiyaka/mosa/common"
)

type dryRun struct {
}

func DryRun() Executor {
	return &dryRun{}
}

func (dr *dryRun) Execute(stage *common.Stage) error {
	fmt.Println("Realized types:")
	for typ, items := range stage.Steps {
		if typ == "exec" {
			continue
		}

		fmt.Printf("%s:\n", typ)
		for _, item := range items {
			fmt.Printf("\t%s\n", item.Item)
		}
		fmt.Println("")
	}

	fmt.Println("Execute:")
	for typ, items := range stage.Steps {
		if typ != "exec" {
			continue
		}

		for _, item := range items {
			fmt.Printf("%s\n", item.Item)
		}
	}

	return nil
}
