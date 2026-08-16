// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

package imodel_price

import (
	"io"

	"gopkg.in/yaml.v3"
)

// ModelListFile is the top-level structure of model-list.yaml.
type ModelListFile struct {
	Version         string        `json:"version" yaml:"version"`
	DefaultCurrency string        `json:"default_currency" yaml:"default_currency"`
	Models          []*ModelPrice `json:"models" yaml:"models"`
}

// ParseModelListYAML parses a model-list.yaml stream into ModelPrice entries.
func ParseModelListYAML(r io.Reader) (*ModelListFile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	file := &ModelListFile{}
	if err := yaml.Unmarshal(data, file); err != nil {
		return nil, err
	}

	for _, m := range file.Models {
		if m != nil && m.PriceCurrency == "" {
			m.PriceCurrency = file.DefaultCurrency
		}
	}

	return file, nil
}
