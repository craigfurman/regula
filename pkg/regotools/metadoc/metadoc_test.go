// Copyright 2021 Fugue, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadoc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_metadoc(t *testing.T) {
	type Test struct {
		input    string
		run      func(*RegoMeta)
		expected string
	}

	tests := []Test{
		{
			// Set package name
			input: `
package foo.bar

default allow = false
`,
			run: func(rego *RegoMeta) {
				assert.Equal(t, "foo.bar", rego.PackageName)
				assert.False(t, rego.HasMetadoc())
				rego.PackageName = "bar.qux"
			},
			expected: `
package bar.qux

default allow = false
`,
		},
		{
			// Insert missing package name
			input: `default allow = false
`,
			run: func(rego *RegoMeta) {
				assert.Equal(t, "", rego.PackageName)
				rego.PackageName = "bar.dux"
			},
			expected: `package bar.dux

default allow = false
`,
		},
		{
			// Metadoc update
			input: `
# Copyright 2020 Fugue, Inc.
package rules.tf_aws_ebs_volume_encrypted

__rego__metadoc__ := {
  "custom": {
    "controls": {
      "CIS-AWS_v1.3.0": [
        "CIS-AWS_v1.3.0_2.2.1"
      ],
      "NIST-800-53_vRev4": [
        "NIST-800-53_vRev4_SC-13"
      ],
    },
    "severity": "High",
    "families": [
      "MyCustomFamily",
      "1172ca4f-6d31-4c46-a085-54ff73c6ed27"
    ],
    "provider": "AWS"
  },
  "description": "EBS volume encryption should be enabled. Enabling encryption on EBS volumes protects data at rest inside the volume, data in transit between the volume and the instance, snapshots created from the volume, and volumes created from those snapshots. EBS volumes are encrypted using KMS keys.",
  "id": "FG_R00016",
  "title": "EBS volume encryption should be enabled"
}

resource_type = "aws_ebs_volume"

default allow = false

allow {
  input.encrypted == true
}
`,
			run: func(rego *RegoMeta) {
				assert.True(t, rego.HasMetadoc())
				assert.Equal(t, "rules.tf_aws_ebs_volume_encrypted", rego.PackageName)
				assert.Equal(t, "FG_R00016", rego.Id)
				assert.Equal(t, "EBS volume encryption should be enabled", rego.Title)
				assert.Equal(t, "EBS volume encryption should be enabled. Enabling encryption on EBS volumes protects data at rest inside the volume, data in transit between the volume and the instance, snapshots created from the volume, and volumes created from those snapshots. EBS volumes are encrypted using KMS keys.", rego.Description)
				assert.Equal(t, "High", rego.Severity)
				assert.Equal(t, []string{
					"MyCustomFamily",
					"1172ca4f-6d31-4c46-a085-54ff73c6ed27",
				}, rego.Families)
				assert.Equal(t,
					map[string][]string{
						"CIS-AWS_v1.3.0":    {"CIS-AWS_v1.3.0_2.2.1"},
						"NIST-800-53_vRev4": {"NIST-800-53_vRev4_SC-13"},
					},
					rego.Controls)
				assert.Equal(t, []string{"AWS"}, rego.Providers)

				rego.Description = "Updated description"
				rego.Severity = "Low"
				rego.Families[1] = "ee1dfd49-a46f-433d-99b9-25805d7e8766"
				delete(rego.Controls, "CIS-AWS_v1.3.0")
				rego.Providers = []string{"AZURE", "REPOSITORY"}
			},
			expected: `
# Copyright 2020 Fugue, Inc.
package rules.tf_aws_ebs_volume_encrypted

__rego__metadoc__ := {
  "custom": {
    "controls": {
      "NIST-800-53_vRev4": [
        "NIST-800-53_vRev4_SC-13"
      ]
    },
    "families": [
      "MyCustomFamily",
      "ee1dfd49-a46f-433d-99b9-25805d7e8766"
    ],
    "providers": [
      "AZURE",
      "REPOSITORY"
    ],
    "severity": "Low"
  },
  "description": "Updated description",
  "id": "FG_R00016",
  "title": "EBS volume encryption should be enabled"
}

resource_type := "aws_ebs_volume"

default allow = false

allow {
  input.encrypted == true
}
`,
		},
		{
			// Metadoc insert
			input: `
# Copyright 2020 Fugue, Inc.
package rules.tf_aws_ebs_volume_encrypted

resource_type = "aws_ebs_volume"
default allow = false
allow {
  input.encrypted == true
}
`,
			run: func(rego *RegoMeta) {
				assert.Equal(t, "rules.tf_aws_ebs_volume_encrypted", rego.PackageName)
				assert.Equal(t, "", rego.Id)
				assert.Equal(t, "", rego.Title)
				assert.Equal(t, "", rego.Description)
				assert.Equal(t, "", rego.Severity)
				assert.Empty(t, rego.Controls)

				rego.Description = "Updated description"
				rego.Severity = "Low"
			},
			expected: `
# Copyright 2020 Fugue, Inc.
package rules.tf_aws_ebs_volume_encrypted

__rego__metadoc__ := {
  "custom": {
    "severity": "Low"
  },
  "description": "Updated description"
}

resource_type := "aws_ebs_volume"
default allow = false
allow {
  input.encrypted == true
}
`,
		},
		{
			// Metadoc + package name insert
			input: `allow { input.encrypted == true }`,
			run: func(rego *RegoMeta) {
				assert.Equal(t, "", rego.PackageName)

				rego.PackageName = "foo.bar"
				rego.Controls = map[string][]string{
					"CIS-AWS_v1.3.0": {"CIS-AWS_v1.3.0_2.2.1"},
				}
			},
			expected: `package foo.bar

__rego__metadoc__ := {
  "custom": {
    "controls": {
      "CIS-AWS_v1.3.0": [
        "CIS-AWS_v1.3.0_2.2.1"
      ]
    }
  }
}

allow { input.encrypted == true }`,
		},
		{
			// Preserve unknown attributes
			input: `
__rego__metadoc__ := {
  "badness": 1000,
  "custom": {
    "rating": "F"
  }
}

deny { true }`,
			run: func(rego *RegoMeta) {
				assert.Equal(t, "", rego.PackageName)

				rego.Severity = "High"
			},
			expected: `
__rego__metadoc__ := {
  "badness": 1000,
  "custom": {
    "rating": "F",
    "severity": "High"
  }
}

deny { true }`,
		},
		{
			// Read resource type, modify it and set input type
			input: `package foo.bar

deny { input.age <= 21 }

resource_type = "MULTIPLE"`,
			run: func(rego *RegoMeta) {
				assert.Equal(t, "MULTIPLE", rego.ResourceType)
				rego.ResourceType = "aws_s3_bucket"
				rego.InputType = "tf"
			},
			expected: `package foo.bar

input_type := "tf"

deny { input.age <= 21 }

resource_type := "aws_s3_bucket"`,
		},
		{
			// Import removal
			input: `import data.foo
import data.foo as bar
deny { input.age <= 21 }`,
			run: func(rego *RegoMeta) {
				assert.Contains(t, rego.Imports, Import{Path: "data.foo"})
				delete(rego.Imports, Import{Path: "data.foo"})
			},
			expected: `import data.foo as bar
deny { input.age <= 21 }`,
		},
		{
			// Import addition
			input: `import data.foo
import data.foo as bar
deny { input.age <= 21 }`,
			run: func(rego *RegoMeta) {
				rego.Imports[Import{Path: "data.fugue"}] = struct{}{}
			},
			expected: `
import data.fugue
import data.foo
import data.foo as bar
deny { input.age <= 21 }`,
		},
		{
			// Check migrating from provider to providers
			input: `
__rego__metadoc__ := {
  "custom": {
    "provider": "AWS"
  }
}

allow = true`,
			run: func(rego *RegoMeta) {
			},
			expected: `
__rego__metadoc__ := {
  "custom": {
    "providers": [
      "AWS"
    ]
  }
}

allow = true`,
		},
	}

	for _, test := range tests {
		rego, err := RegoMetaFromString(test.input)
		if err != nil {
			t.Fatal(err)
		}
		test.run(rego)
		assert.Equal(t, test.expected, rego.String())
	}
}
