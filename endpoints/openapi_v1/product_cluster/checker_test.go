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

package product_cluster

import (
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/stretchr/testify/assert"
)

func TestCheckLLMConfig(t *testing.T) {
	cases := []struct {
		name    string
		config  *icluster_conf.LLMConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "empty provider and models",
			config:  &icluster_conf.LLMConfig{},
			wantErr: true,
		},
		{
			name: "valid with provider and models",
			config: &icluster_conf.LLMConfig{
				Provider: lib.PString("openai"),
				Models:   []string{"gpt-4"},
			},
			wantErr: false,
		},
		{
			name: "strip prefix without match prefix",
			config: &icluster_conf.LLMConfig{
				Provider:    lib.PString("openai"),
				Models:      []string{"gpt-4"},
				StripPrefix: lib.PBool(true),
			},
			wantErr: true,
		},
		{
			name: "match prefix missing trailing slash",
			config: &icluster_conf.LLMConfig{
				Provider:    lib.PString("openai"),
				Models:      []string{"gpt-4"},
				MatchPrefix: lib.PString("openrouter"),
			},
			wantErr: true,
		},
		{
			name: "valid prefix config",
			config: &icluster_conf.LLMConfig{
				Provider:    lib.PString("openai"),
				Models:      []string{"gpt-4"},
				MatchPrefix: lib.PString("openrouter/"),
				StripPrefix: lib.PBool(true),
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkLLMConfig(tc.config)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckEPPServer(t *testing.T) {
	validDomain := "example.com"
	invalidDomain := "example"
	validPort := 8080
	invalidPortHigh := 70000
	invalidPortLow := 0
	validIP := "192.0.2.1"
	invalidIP := "999.999.999.999"

	cases := []struct {
		name    string
		server  *icluster_conf.EPPServer
		wantErr bool
	}{
		{
			name:    "nil server",
			server:  nil,
			wantErr: false,
		},
		{
			name:    "empty server",
			server:  &icluster_conf.EPPServer{},
			wantErr: false,
		},
		{
			name: "domain and port valid",
			server: &icluster_conf.EPPServer{
				Domain: &validDomain,
				Port:   &validPort,
			},
			wantErr: false,
		},
		{
			name: "domain missing port",
			server: &icluster_conf.EPPServer{
				Domain: &validDomain,
			},
			wantErr: true,
		},
		{
			name: "invalid domain",
			server: &icluster_conf.EPPServer{
				Domain: &invalidDomain,
				Port:   &validPort,
			},
			wantErr: true,
		},
		{
			name: "port out of range high",
			server: &icluster_conf.EPPServer{
				Domain: &validDomain,
				Port:   &invalidPortHigh,
			},
			wantErr: true,
		},
		{
			name: "port out of range low",
			server: &icluster_conf.EPPServer{
				Domain: &validDomain,
				Port:   &invalidPortLow,
			},
			wantErr: true,
		},
		{
			name: "valid endpoints",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{
					{IP: &validIP, Port: &validPort},
				},
			},
			wantErr: false,
		},
		{
			name: "endpoint nil",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{nil},
			},
			wantErr: true,
		},
		{
			name: "endpoint missing ip",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{
					{Port: &validPort},
				},
			},
			wantErr: true,
		},
		{
			name: "endpoint empty ip",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{
					{IP: lib.PString(""), Port: &validPort},
				},
			},
			wantErr: true,
		},
		{
			name: "endpoint invalid ip",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{
					{IP: &invalidIP, Port: &validPort},
				},
			},
			wantErr: true,
		},
		{
			name: "endpoint missing port",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{
					{IP: &validIP},
				},
			},
			wantErr: true,
		},
		{
			name: "endpoint port out of range",
			server: &icluster_conf.EPPServer{
				Endpoints: []*icluster_conf.EPPEndpoint{
					{IP: &validIP, Port: &invalidPortHigh},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckEPPServer(tc.server)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
