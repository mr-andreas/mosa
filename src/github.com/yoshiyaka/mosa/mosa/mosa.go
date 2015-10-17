package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yoshiyaka/mosa/manifest"
	"github.com/yoshiyaka/mosa/reducer"
)

func main() {
	fName := "../testdata/manifest.ms"
	flag.Parse()
	if args := flag.CommandLine.Args(); len(args) == 1 {
		fName = args[0]
	}

	file, fileErr := os.Open(fName)
	if fileErr != nil {
		panic(fileErr)
	}
	defer file.Close()

	mfst, mfstErr := manifest.Lex(fName, file)
	if mfstErr != nil {
		panic(mfstErr)
	}

	reduced, reducedErr := reducer.Reduce(mfst)
	if reducedErr != nil {
		panic(reducedErr)
	}

	fmt.Println(reduced)

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
