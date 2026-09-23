// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package ioperlog

import (
	"encoding/json"
)

// ParamToMap marshals a param struct into a map for change_summary
// before/after snapshots. Keys whose value is nil are dropped, so that
// fields omitted from a partial update do not materialize as null entries:
// partial updates omit unchanged fields, and including them produces
// false-positive diffs (see BuildChangeSummary). Explicit zero values
// (e.g. enabled=false) are non-nil and are therefore retained.
func ParamToMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}

	for k, val := range m {
		if val == nil {
			delete(m, k)
		}
	}
	return m
}
