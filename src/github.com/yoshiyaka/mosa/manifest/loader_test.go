package manifest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yoshiyaka/mosa/common"
)

var loadTests = []struct {
	data          string
	expectedSteps []*common.Step
}{
	{
		``,
		[]*common.Step{},
	},

	{
		`
deb[pkg1]:
`,

		[]*common.Step{
			&common.Step{
				Type:    "deb",
				Item:    "pkg1",
				Args:    nil,
				Depends: nil,
			},
		},
	},

	{
		`
file[/path/with some/spaces]:
`,

		[]*common.Step{
			&common.Step{
				Type:    "file",
				Item:    "/path/with some/spaces",
				Args:    nil,
				Depends: nil,
			},
		},
	},

	{
		`
deb[pkg1]:
    depends:
      - deb[pkg2]
      - file[file1]
`,

		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"deb":  {"pkg2"},
					"file": {"file1"},
				},
			},
		},
	},

	{
		`
deb[pkg1]:
    depends: deb[pkg2]
`,

		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"deb": {"pkg2"},
				},
			},
		},
	},

	{
		`
deb[pkg1]:
    depends:
        - deb[pkg2]
        - file[file1]

deb[pkg2]:
    depends:
        - file[file1]
        - file[file2]

file[file1]:
    depends:
        - shell[cmd1]

file[file2]:
    depends:
        - shell[cmd2]

shell[cmd1]:
    depends:
        - shell[cmd2]

shell[cmd2]:
`,

		[]*common.Step{
			&common.Step{
				Type: "deb",
				Item: "pkg1",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"deb":  {"pkg2"},
					"file": {"file1"},
				},
			},

			&common.Step{
				Type: "deb",
				Item: "pkg2",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"file": {"file1", "file2"},
				},
			},

			&common.Step{
				Type: "file",
				Item: "file1",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"shell": {"cmd1"},
				},
			},

			&common.Step{
				Type: "file",
				Item: "file2",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"shell": {"cmd2"},
				},
			},

			&common.Step{
				Type: "shell",
				Item: "cmd1",
				Args: map[string]interface{}{},
				Depends: map[string][]string{
					"shell": {"cmd2"},
				},
			},

			&common.Step{
				Type:    "shell",
				Item:    "cmd2",
				Args:    nil,
				Depends: nil,
			},
		},
	},
}

func TestLoad(t *testing.T) {
	for _, loadTest := range loadTests {
		r := strings.NewReader(loadTest.data)
		if steps, err := Load(r); err != nil {
			t.Errorf("Failed loading %s: %s", loadTest.data, err)
		} else {
			if !common.EqualsAsJSON(loadTest.expectedSteps, steps) {
				js, _ := json.MarshalIndent(steps, "", "  ")
				t.Error("Got bad steps", string(js))
			}
		}
	}
}
