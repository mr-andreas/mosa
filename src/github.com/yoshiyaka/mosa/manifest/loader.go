package manifest

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/yoshiyaka/mosa/common"
)

var nameRxp = regexp.MustCompile(`([A-Za-z0-9]+)\[([^[\]]+)\]$`)

type loadedStep struct {
}

// Loads steps from a yaml file
func Load(r io.Reader) ([]*common.Step, error) {
	var loadedMap map[string]map[string]interface{}

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(bytes, &loadedMap); err != nil {
		return nil, err
	}

	ret := make([]*common.Step, 0, len(loadedMap))
	for key, stepMap := range loadedMap {
		matches := nameRxp.FindStringSubmatch(key)
		if matches == nil {
			return nil, errors.New("Invalid step identifier: " + key)
		}

		step := &common.Step{
			Type: matches[1],
			Item: matches[2],
			Args: stepMap,
		}

		if err := setDepends(step); err != nil {
			return nil, err
		}

		ret = append(ret, step)
	}

	return ret, nil
}

func setDepends(step *common.Step) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Invalid values for depends in step %v: %s", step, r)
		}
	}()

	if _, ok := step.Args["depends"]; !ok {
		return nil
	}

	var dependsArray []string
	if da, ok := step.Args["depends"].([]interface{}); ok {
		dependsArray = make([]string, len(da))
		for i, val := range da {
			dependsArray[i] = val.(string)
		}
	} else {
		dependsArray = []string{step.Args["depends"].(string)}
	}

	delete(step.Args, "depends")

	step.Depends = map[string][]string{}

	for _, dependency := range dependsArray {
		matches := nameRxp.FindStringSubmatch(dependency)
		if matches == nil {
			return errors.New("Invalid dependency identifier: " + dependency)
		}

		typ := matches[1]
		item := matches[2]

		if _, ok := step.Depends[typ]; !ok {
			step.Depends[typ] = []string{item}
		} else {
			step.Depends[typ] = append(step.Depends[typ], item)
		}
	}

	return nil
}
