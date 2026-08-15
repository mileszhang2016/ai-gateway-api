// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

package icluster_conf

import (
	"context"
	"errors"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterManager_ExportClusterTable(t *testing.T) {
	ctx := context.Background()

	t.Run("unchanged", func(t *testing.T) {
		m := newClusterManagerForExport(t, iversion_control.ZeroVersion)
		got, err := m.ExportClusterTable(ctx, iversion_control.ZeroVersion)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("changed", func(t *testing.T) {
		m := newClusterManagerForExport(t, "20240102000000")
		got, err := m.ExportClusterTable(ctx, "old")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "20240102000000", *got.Version)
	})

	t.Run("export error", func(t *testing.T) {
		vcs := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return "", errors.New("upsert failed")
			},
		}
		vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, vcs)
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return []*Cluster{newTestClusterBase()}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, vcm, nil, nil)
		_, err := m.ExportClusterTable(ctx, "")
		require.Error(t, err)
	})

	t.Run("generator fetch error", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil, nil)
		_, err := m.clusterTableConfGenerator(ctx)
		require.Error(t, err)
	})
}

func TestClusterManager_ExportGSLB(t *testing.T) {
	ctx := context.Background()

	t.Run("unchanged", func(t *testing.T) {
		m := newClusterManagerForExportGSLB(t, iversion_control.ZeroVersion)
		got, err := m.ExportGSLB(ctx, iversion_control.ZeroVersion, "bfe1")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("changed", func(t *testing.T) {
		m := newClusterManagerForExportGSLB(t, "20240102000000")
		got, err := m.ExportGSLB(ctx, "old", "bfe1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "20240102000000", got.Version)
	})

	t.Run("bfe cluster not exist", func(t *testing.T) {
		m := newClusterManagerForExportGSLB(t, iversion_control.ZeroVersion)
		_, err := m.ExportGSLB(ctx, "", "bfe-missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BFECluster bfe-missing Not Exist")
	})

	t.Run("export error", func(t *testing.T) {
		vcs := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return "", errors.New("upsert failed")
			},
		}
		vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, vcs)
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				c := newTestClusterBase()
				return []*Cluster{c}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, vcm, nil, nil)
		_, err := m.ExportGSLB(ctx, "", "bfe1")
		require.Error(t, err)
	})
}

func TestClusterManager_gslbConfGenerator(t *testing.T) {
	ctx := context.Background()
	m := newClusterManagerForExportGSLB(t, iversion_control.ZeroVersion)
	ed, err := m.gslbConfGenerator("bfe1")(ctx)
	require.NoError(t, err)
	require.NotNil(t, ed)
	assert.Equal(t, ConfigTopicGSLB+".bfe1", ed.Topic)
	conf := ed.DataWithoutVersion.(*GSLBConf)
	require.NotNil(t, conf.Clusters)
	require.NotNil(t, (*conf.Clusters)["c1"])
}

func newClusterManagerForExportGSLB(t *testing.T, version string) *ClusterManager {
	t.Helper()
	vcs := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return version, nil
		},
	}
	vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, vcs)
	c := newTestClusterBase()
	c.Scheduler = map[string]map[string]int{
		"bfe1": {"sc1": 100},
	}
	clusterStore := &fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
			return []*Cluster{c}, nil
		},
	}
	return NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, vcm, nil, nil)
}

func TestClusterTableConf_clusterWithIPv6(t *testing.T) {
	ctx := context.Background()
	vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "20240102000000", nil
		},
	})
	c := newTestClusterBase()
	c.SubClusters[0].InstancePool.Instances[0].Addr = "2001:db8::1"
	clusterStore := &fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
			return []*Cluster{c}, nil
		},
	}
	m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, vcm, nil, nil)
	conf, err := m.ExportClusterTable(ctx, "old")
	require.NoError(t, err)
	require.NotNil(t, conf)
	require.NotNil(t, conf.Config)
	backends := (*conf.Config)["c1"]["sc1"]
	require.Len(t, backends, 1)
	assert.Equal(t, "[2001:db8::1]", *backends[0].Addr)
}
