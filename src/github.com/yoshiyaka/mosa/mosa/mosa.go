package main

import (
	"fmt"
	"os"

	"kiliaro.app/mosa/executor"
	"kiliaro.app/mosa/manifest"
	"kiliaro.app/mosa/planner"
)

func main() {
	mfst, mfstErr := os.Open("../testdata/manifest.yaml")
	if mfstErr != nil {
		panic(mfstErr)
	}
	defer mfst.Close()

	steps, stepsErr := manifest.Load(mfst)
	if stepsErr != nil {
		panic(stepsErr)
	}

	planner := planner.New()
	plan, planErr := planner.Plan(steps)
	if planErr != nil {
		panic(planErr)
	}

	exc := executor.DryRun()
	if err := executor.ExecutePlan(plan, exc); err != nil {
		panic(err)
	}

	realExc, realExcErr := executor.New("../script")
	if realExcErr != nil {
		panic(realExcErr)
	}
	if err := executor.ExecutePlan(plan, realExc); err != nil {
		panic(err)
	}

	fmt.Println(plan)
}
