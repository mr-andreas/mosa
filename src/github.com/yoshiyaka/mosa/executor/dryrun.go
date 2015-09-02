package executor

import (
	"fmt"
	"strings"

	"kiliaro.app/mosa/common"
)

type dryRun struct {
}

func DryRun() Executor {
	return &dryRun{}
}

func (dr *dryRun) Execute(stage *common.Stage) error {
	for typ, items := range stage.Steps {
		names := make([]string, len(items))
		for i, item := range items {
			names[i] = item.Item
		}
		fmt.Printf("%s[%s]\n", typ, strings.Join(names, ","))
	}

	return nil
}
