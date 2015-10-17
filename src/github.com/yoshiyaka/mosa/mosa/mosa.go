package main

import (
	"os"

	"github.com/yoshiyaka/mosa/manifest"
)

func main() {
	manifest.Lex("-", os.Stdin)
	return

	mfst, mfstErr := os.Open("../testdata/manifest.yaml")
	if mfstErr != nil {
		panic(mfstErr)
	}
	defer mfst.Close()

	//	steps, stepsErr := manifest.Load(mfst)
	//	if stepsErr != nil {
	//		panic(stepsErr)
	//	}

	//	planner := planner.New()
	//	plan, planErr := planner.Plan(steps)
	//	if planErr != nil {
	//		panic(planErr)
	//	}

	//	exc := executor.DryRun()
	//	if err := executor.ExecutePlan(plan, exc); err != nil {
	//		panic(err)
	//	}

	//	realExc, realExcErr := executor.New("../script")
	//	if realExcErr != nil {
	//		panic(realExcErr)
	//	}
	//	if err := executor.ExecutePlan(plan, realExc); err != nil {
	//		panic(err)
	//	}

	//	fmt.Println(plan)
}
