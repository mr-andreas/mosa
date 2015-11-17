package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/yoshiyaka/mosa/ast"
	"github.com/yoshiyaka/mosa/executor"
	"github.com/yoshiyaka/mosa/parser"
	"github.com/yoshiyaka/mosa/planner"
	"github.com/yoshiyaka/mosa/reducer"
	"github.com/yoshiyaka/mosa/stepconverter"
)

func parseDirAsASTRecursively(ast *ast.AST, dirName string) error {
	files, filesErr := ioutil.ReadDir(dirName)
	if filesErr != nil {
		return filesErr
	}

	for _, file := range files {
		if file.Name()[0] == '.' {
			continue
		}

		fullPath := dirName + "/" + file.Name()

		if file.IsDir() {
			if err := parseDirAsASTRecursively(ast, fullPath); err != nil {
				return err
			}
		} else if strings.HasSuffix(file.Name(), ".ms") {
			f, fErr := os.Open(fullPath)
			if fErr != nil {
				return fErr
			}

			if err := parser.Parse(ast, fullPath, f); err != nil {
				f.Close()
				return err
			} else {
				f.Close()
			}
		}
	}

	return nil
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Printf("%s [options] manifest-directory\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	help := false
	flag.BoolVar(&help, "h", false, "Shows this message")

	dirName := "../testdata"
	flag.Parse()

	if help {
		showHelp()
		return
	}

	if args := flag.CommandLine.Args(); len(args) == 1 {
		dirName = args[0]
	}

	mfst := ast.NewAST()
	if err := parseDirAsASTRecursively(mfst, dirName); err != nil {
		panic(err)
	}

	reduced, reducedErr := reducer.Reduce(mfst)
	if reducedErr != nil {
		panic(reducedErr)
	}

	steps, stepsErr := stepconverter.Convert(reduced)
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

	//	realExc, realExcErr := executor.New("../script")
	//	if realExcErr != nil {
	//		panic(realExcErr)
	//	}
	//	if err := executor.ExecutePlan(plan, realExc); err != nil {
	//		panic(err)
	//	}

	//	fmt.Println(plan)
}
