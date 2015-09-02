package planner

import (
	"encoding/json"
	"testing"

	"kiliaro.lib/rest/resttest"

	"kiliaro.app/mosa/common"
)

type badPlanTest struct {
	comment       string
	steps         []*common.Step
	expectedError error
}

var badPlans = []badPlanTest{
	{
		"Recursive steps",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Depends: map[string][]string{
					"deb": {"pkg2"},
				},
			},

			&common.Step{
				Type: "deb",
				Item: "pkg2",
				Depends: map[string][]string{
					"deb": {"pkg1"},
				},
			},
		},
		ErrRecursivePlan,
	},

	{
		"Missing dependency",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Depends: map[string][]string{
					"deb": {"pkg2"},
				},
			},
		},
		ErrMissingDependency,
	},

	{
		"Duplicate steps",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},

			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},
		},
		ErrDuplicateDefinition,
	},

	{
		"Deep recursive steps",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Depends: map[string][]string{
					"deb":  {"pkg2"},
					"file": {"file1"},
				},
			},

			&common.Step{
				Type: "deb",
				Item: "pkg2",
				Depends: map[string][]string{
					"file": {"file1", "file2"},
				},
			},

			&common.Step{
				Type: "file",
				Item: "file1",
				Depends: map[string][]string{
					"shell": {"cmd1"},
				},
			},

			&common.Step{
				Type: "file",
				Item: "file2",
				Depends: map[string][]string{
					"shell": {"cmd2"},
				},
			},

			&common.Step{
				Type:    "shell",
				Item:    "cmd1",
				Depends: nil,
			},

			&common.Step{
				Type: "shell",
				Item: "cmd2",
				Depends: map[string][]string{
					// This is the recursve dependency
					"file": {"file2"},
				},
			},
		},
		ErrRecursivePlan,
	},
}

func TestBadPlan(t *testing.T) {
	planner := New()

	for _, badPlan := range badPlans {
		if plan, err := planner.Plan(badPlan.steps); err == nil {
			t.Errorf("Plan for %s worked: %v", badPlan.comment, plan)
		} else {
			if err, errOk := err.(*Error); !errOk {
				t.Errorf("Got non-*Error for %s: %s", badPlan.comment, err.Error())
			} else if err.GeneralError != badPlan.expectedError {
				t.Errorf(
					"Got bad error for %s: %s", badPlan.comment, err.Error(),
				)
			}
		}
	}
}

type goodPlanTest struct {
	comment      string
	steps        []*common.Step
	expectedPlan []map[string][]string
}

var goodPlanTests = []goodPlanTest{
	{
		"Nil plan",
		nil,
		nil,
	},

	{
		"Simple plan",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},
		},
		[]map[string][]string{
			map[string][]string{
				"deb": {"pkg1"},
			},
		},
	},

	{
		"Simple plan with multiple packages",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},
			&common.Step{
				Type: "deb",
				Item: "pkg2",
			},
		},
		[]map[string][]string{
			map[string][]string{
				"deb": {"pkg1", "pkg2"},
			},
		},
	},

	{
		"Simple plan with multiple types",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},
			&common.Step{
				Type: "file",
				Item: "file1",
			},
		},
		[]map[string][]string{
			map[string][]string{
				"deb":  {"pkg1"},
				"file": {"file1"},
			},
		},
	},

	{
		"Simple two-stage",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},
			&common.Step{
				Type: "deb",
				Item: "pkg2",
				Depends: map[string][]string{
					"deb": {"pkg1"},
				},
			},
		},
		[]map[string][]string{
			map[string][]string{
				"deb": {"pkg1"},
			},
			map[string][]string{
				"deb": {"pkg2"},
			},
		},
	},

	{
		"Simple two-stage with multiple types",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
			},
			&common.Step{
				Type: "deb",
				Item: "pkg2",
				Depends: map[string][]string{
					"deb": {"pkg1"},
				},
			},
			&common.Step{
				Type: "file",
				Item: "file1",
			},
		},
		[]map[string][]string{
			map[string][]string{
				"deb":  {"pkg1"},
				"file": {"file1"},
			},
			map[string][]string{
				"deb": {"pkg2"},
			},
		},
	},

	{
		"Complex plan",
		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Depends: map[string][]string{
					"deb":  {"pkg2"},
					"file": {"file1"},
				},
			},

			&common.Step{
				Type: "deb",
				Item: "pkg2",
				Depends: map[string][]string{
					"file": {"file1", "file2"},
				},
			},

			&common.Step{
				Type: "file",
				Item: "file1",
				Depends: map[string][]string{
					"shell": {"cmd1"},
				},
			},

			&common.Step{
				Type: "file",
				Item: "file2",
				Depends: map[string][]string{
					"shell": {"cmd2"},
				},
			},

			&common.Step{
				Type: "shell",
				Item: "cmd1",
				Depends: map[string][]string{
					"shell": {"cmd2"},
				},
			},

			&common.Step{
				Type:    "shell",
				Item:    "cmd2",
				Depends: nil,
			},
		},
		[]map[string][]string{
			map[string][]string{
				"shell": {"cmd2"},
			},
			map[string][]string{
				"shell": {"cmd1"},
				"file":  {"file2"},
			},
			map[string][]string{
				"file": {"file1"},
			},
			map[string][]string{
				"deb": {"pkg2"},
			},
			map[string][]string{
				"deb": {"pkg1"},
			},
		},
	},
}

func TestGoodPlan(t *testing.T) {
	planner := New()
	for _, goodPlan := range goodPlanTests {
		stepsByType, stepsErr := planner.extractSteps(goodPlan.steps)
		if stepsErr != nil {
			t.Fatal(stepsErr)
		}

		expectedPlan := common.Plan{
			Stages: []*common.Stage{},
		}
		for _, goodStage := range goodPlan.expectedPlan {
			steps := map[string][]common.Step{}

			for stepType, stepNames := range goodStage {
				steps[stepType] = make([]common.Step, len(stepNames))
				for i, stepName := range stepNames {
					steps[stepType][i] = *stepsByType[stepType][stepName]
				}
			}

			stage := common.Stage{
				Steps: steps,
			}
			expectedPlan.Stages = append(expectedPlan.Stages, &stage)
		}

		if plan, err := planner.Plan(goodPlan.steps); err != nil {
			t.Errorf("Failed planning %s: %s", goodPlan.comment, err)
		} else {
			if !resttest.EqualsAsJSON(expectedPlan, plan) {
				js, _ := json.MarshalIndent(plan, "", "  ")
				t.Errorf("For %s: got bad plan %s %s", goodPlan.comment, plan, string(js))
			}
		}
	}
}
