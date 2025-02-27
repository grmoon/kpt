// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package addmergecomment

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMetadataComment(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
		errMsg   string
	}{
		{
			name: "Add kpt merge annotation with name and namespace",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: my-space
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata: # kpt-merge: my-space/nginx-deployment
  name: nginx-deployment
  namespace: my-space
spec:
  replicas: 3
 `,
		},
		{
			name: "Add kpt merge comment with name and no namespace",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata: # kpt-merge: default/nginx-deployment
  name: nginx-deployment
  namespace: default
spec:
  replicas: 3
 `,
		},
		{
			name: "Add kpt merge comment with name and no namespace",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata: # kpt-merge: /nginx-deployment
  name: nginx-deployment
spec:
  replicas: 3
 `,
		},
		{
			name: "Skip adding kpt merge comment if already present",
			input: `
apiVersion: apps/v1
kind: Deployment
metadata: # kpt-merge: my-space/nginx-deployment
  name: nginx-deployment-new
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata: # kpt-merge: my-space/nginx-deployment
  name: nginx-deployment-new
spec:
  replicas: 3
 `,
		},
		{
			name: "Skip adding kpt merge comment if already present",
			input: `
apiVersion: apps/v1
kind: MyKind
spec:
  replicas: 3
 `,
			expected: `
apiVersion: apps/v1
kind: MyKind
spec:
  replicas: 3
 `,
		},
		{
			name: "Skip adding kpt merge comment if non-KRM resource",
			input: `- op: replace
  path: /spec
  value:
    group: kubeflow.org
 `,
			expected: `- op: replace
  path: /spec
  value:
    group: kubeflow.org
 `,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			defer os.RemoveAll(baseDir)

			r, err := ioutil.TempFile(baseDir, "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = Process(baseDir)
			if test.errMsg != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errMsg) {
					t.FailNow()
				}
			}

			if test.errMsg == "" && !assert.NoError(t, err) {
				t.FailNow()
			}

			actualResources, err := ioutil.ReadFile(r.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				strings.TrimSpace(test.expected),
				strings.TrimSpace(string(actualResources))) {
				t.FailNow()
			}
		})
	}
}
