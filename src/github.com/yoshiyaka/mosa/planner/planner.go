package planner

import (
	"errors"
	"fmt"

	"github.com/yoshiyaka/mosa/common"
)

var (
	ErrRecursivePlan       = errors.New("Recursive plan")
	ErrMissingDependency   = errors.New("Missing dependency")
	ErrDuplicateDefinition = errors.New("Duplicate definition")
)

type Error struct {
	// One of the Err* variables
	GeneralError error

	// Failing step
	Step *common.Step

	// Details, for instance the name of the missing dependency on
	// ErrMissingDependency
	Details string
}

func newError(ge error, s *common.Step, details string) *Error {
	return &Error{ge, s, details}
}

func (e *Error) Error() string {
	err := fmt.Sprintf("Error processsing %s: %s", e.Step, e.GeneralError.Error())
	if e.Details != "" {
		err += " (" + e.Details + ")"
	}
	return err
}

type Planner struct {
}

func New() *Planner {
	return &Planner{}
}

// If an error is returned, it will always be a *Error
func (p *Planner) Plan(steps []common.Step) (*common.Plan, error) {
	stepsByType, extractErr := p.extractSteps(steps)
	if extractErr != nil {
		return nil, extractErr
	}

	for _, step := range steps {
		if recursionErr := p.stepIsRecursive(stepsByType, nil, &step); recursionErr != nil {
			return nil, recursionErr
		}
	}

	plan := common.Plan{
		Stages: make([]*common.Stage, 0),
	}

	stepsLeft := make([]common.Step, len(steps))
	for i, step := range steps {
		stepsLeft[i] = step
	}

	for len(stepsLeft) > 0 {
		var stage *common.Stage
		stage, stepsLeft = p.extractNextStage(stepsLeft)

		plan.Stages = append(plan.Stages, stage)
	}

	return &plan, nil
}

// Groups all step by type and returns them. On duplicate definitions, an error
// is returned.
func (p *Planner) extractSteps(steps []common.Step) (map[string]map[string]*common.Step, error) {
	stepsByType := map[string]map[string]*common.Step{}

	for _, step := range steps {
		if _, ok := stepsByType[step.Type]; !ok {
			stepsByType[step.Type] = map[string]*common.Step{}
		}

		if _, stepExists := stepsByType[step.Type][step.Item]; stepExists {
			return nil, &Error{
				GeneralError: ErrDuplicateDefinition,
				Step:         &step,
			}
		}

		stepCopy := step
		stepsByType[step.Type][step.Item] = &stepCopy
	}

	//	js, _ := json.MarshalIndent(stepsByType, "", "  ")
	//	fmt.Println("Loaded", string(js))

	return stepsByType, nil
}

// Returns the recursive chain if a step is recursive.
func (p *Planner) stepIsRecursive(stepsByType map[string]map[string]*common.Step, seenSteps []*common.Step, step *common.Step) error {
	seenStepsAfterThis := make([]*common.Step, 0, len(seenSteps)+1)
	seenStepsAfterThis = append(seenStepsAfterThis, seenSteps...)
	seenStepsAfterThis = append(seenStepsAfterThis, step)

	for _, seenStep := range seenSteps {
		if seenStep.Type == step.Type && seenStep.Item == step.Item {
			return newError(
				ErrRecursivePlan, seenSteps[0],
				fmt.Sprintf("%#v", seenStepsAfterThis),
			)
		}
	}

	for typ, items := range step.Depends {
		for _, item := range items {
			nextStep, nextStepOk := stepsByType[typ][item]
			if !nextStepOk {
				return newError(
					ErrMissingDependency, step, fmt.Sprintf("%s[%s]", typ, item),
				)
			}

			if recursion := p.stepIsRecursive(stepsByType, seenStepsAfterThis, nextStep); recursion != nil {
				return recursion
			}
		}
	}

	return nil
}

// Find all steps without and dependencies and creates a new stage from the.
// stepsLeft will then be returned without any of the steps in the new stage,
// and any dependencies on the steps moved to the stage will be removed.
func (p *Planner) extractNextStage(steps []common.Step) (stage *common.Stage, stepsLeft []common.Step) {
	stageSteps := map[string][]common.Step{}
	stageStepsNames := map[string]map[string]bool{}
	stepsLeft = make([]common.Step, 0, len(steps))

	// Filter out all steps without dependencies
	for _, step := range steps {
		if len(step.Depends) == 0 {
			if _, ok := stageSteps[step.Type]; ok {
				stageSteps[step.Type] = append(stageSteps[step.Type], step)
			} else {
				stageSteps[step.Type] = []common.Step{step}
			}

			if _, ok := stageStepsNames[step.Type]; !ok {
				stageStepsNames[step.Type] = map[string]bool{}
			}

			stageStepsNames[step.Type][step.Item] = true
		} else {
			stepsLeft = append(stepsLeft, step)
		}
	}

	// Remove the staged steps as dependencies from all remaining steps
	for _, step := range stepsLeft {
		for depType, depItems := range step.Depends {
			depItemsRemoved, depTypeRemoved := stageStepsNames[depType]
			if !depTypeRemoved {
				continue
			}

			depItemsCopy := make([]string, 0, len(depItems))
			for _, depItem := range depItems {
				if _, depRemoved := depItemsRemoved[depItem]; !depRemoved {
					depItemsCopy = append(depItemsCopy, depItem)
				}
			}

			if len(depItemsCopy) == 0 {
				delete(step.Depends, depType)
			} else {
				step.Depends[depType] = depItemsCopy
			}
		}
	}

	return &common.Stage{Steps: stageSteps}, stepsLeft
}
