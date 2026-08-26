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

package quota

import "time"

// Clock provides the current time. It is introduced to make time-based
// quota reset logic testable without waiting for real calendar periods.
type Clock interface {
	Now() time.Time
}

// realClock uses the system clock.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// NewRealClock returns a Clock backed by the system clock.
func NewRealClock() Clock { return realClock{} }
