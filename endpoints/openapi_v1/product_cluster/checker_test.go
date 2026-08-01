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

package product_cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
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
			name:    "empty models",
			config:  &icluster_conf.LLMConfig{},
			wantErr: true,
		},
		{
			name: "valid with models",
			config: &icluster_conf.LLMConfig{
				Models: []string{"gpt-4"},
			},
			wantErr: false,
		},
		{
			name: "valid model endpoint http",
			config: &icluster_conf.LLMConfig{
				Models:        []string{"gpt-4"},
				ModelEndpoint: &icluster_conf.Endpoint{Schema: "http"},
			},
			wantErr: false,
		},
		{
			name: "invalid model endpoint schema",
			config: &icluster_conf.LLMConfig{
				Models:        []string{"gpt-4"},
				ModelEndpoint: &icluster_conf.Endpoint{Schema: "ftp"},
			},
			wantErr: true,
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

func TestCheckInstancePool(t *testing.T) {
	cases := []struct {
		name      string
		instances []*Instance
		wantErr   bool
	}{
		{
			name:      "empty pool",
			instances: []*Instance{},
			wantErr:   true,
		},
		{
			name: "missing port",
			instances: []*Instance{
				{Name: "host1", Addr: "192.0.2.1", Weight: 100},
			},
			wantErr: true,
		},
		{
			name: "all weights zero",
			instances: []*Instance{
				{Name: "host1", Addr: "192.0.2.1", Weight: 0, Port: 8080},
			},
			wantErr: true,
		},
		{
			name: "valid single instance",
			instances: []*Instance{
				{Name: "host1", Addr: "192.0.2.1", Weight: 100, Port: 8080},
			},
			wantErr: false,
		},
		{
			name: "valid multiple instances",
			instances: []*Instance{
				{Name: "host1", Addr: "192.0.2.1", Weight: 50, Port: 8080},
				{Name: "host2", Addr: "192.0.2.2", Weight: 50, Port: 8080},
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkInstancePool(tc.instances)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
