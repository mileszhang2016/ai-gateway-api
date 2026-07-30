// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package xreq

import (
	ut "github.com/go-playground/universal-translator"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
)

// Validator is implemented by request parameter structs that need extra
// validation beyond the struct-tag rules enforced by go-playground/validator.
// When a parameter value implements this interface, xreq.Bind* will invoke
// Validate() after the standard struct validation succeeds.
type Validator interface {
	Validate() error
}

// validateData runs the standard struct validation and then calls the optional
// custom Validate() method registered by the parameter type.
func validateData(data interface{}, lang ut.Translator) error {
	if err := ValidateData(data, lang); err != nil {
		return err
	}

	if v, ok := data.(Validator); ok {
		if err := v.Validate(); err != nil {
			return xerror.WrapParamError(err)
		}
	}

	return nil
}
