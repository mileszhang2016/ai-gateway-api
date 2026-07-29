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

package icluster_conf

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/iversion_control"
	"github.com/yf-networks/ai-gateway-api/stateful"
)

func init() {
	if stateful.DefaultConfig == nil {
		stateful.DefaultConfig = &stateful.Config{
			RunTime: stateful.RunTimeConfig{
				DefaultAIClusterName: "BFE-AI",
			},
		}
	}
}

func newTestClusterBase() *Cluster {
	return &Cluster{
		ID:   1,
		Name: "c1",
		Basic: &ClusterBasic{
			Protocol: lib.PString("http"),
			Connection: &ClusterBasicConnection{
				MaxIdleConnPerRs:    10,
				CancelOnClientClose: false,
			},
			Retries: &ClusterBasicRetries{
				MaxRetryInSubcluster:    1,
				MaxRetryCrossSubcluster: 1,
			},
			Buffers: &ClusterBasicBuffers{
				ReqWriteBufferSize: 1024,
				ReqFlushInterval:   100,
				ResFlushInterval:   100,
			},
			Timeouts: &ClusterBasicTimeouts{
				TimeoutConnServ:        1000,
				TimeoutResponseHeader:  1000,
				TimeoutReadbodyClient:  1000,
				TimeoutReadClientAgain: 1000,
				TimeoutWriteClient:     1000,
			},
		},
		StickySessions: &ClusterStickySessions{
			SessionSticky: false,
			HashStrategy:  0,
			HashHeader:    "",
		},
		PassiveHealthCheck: &ClusterPassiveHealthCheck{
			Schema:     "http",
			Interval:   10,
			Failnum:    3,
			Statuscode: 500,
			Host:       "",
			Uri:        "/",
		},
		SubClusters: []*SubCluster{
			{
				ID:   1,
				Name: "sc1",
				InstancePool: &Pool{
					ID:   1,
					Name: "p1",
					Instances: []Instance{
						{HostName: "rs1", IP: "127.0.0.1", Port: 80, Weight: 10},
					},
				},
			},
		},
		Scheduler: map[string]map[string]int{
			"bfe1": {
				"sc1": 100,
			},
		},
	}
}

func newTestClusterEPP() *Cluster {
	c := newTestClusterBase()
	c.SubClusters[0].Role = ProductPoolRoleEPP
	c.SubClusters[0].InstancePool.EPPServer = &EPPServer{
		Domain: lib.PString("epp.example.com"),
		Port:   lib.PInt(8080),
	}
	return c
}

func newTestClusterDomain() *Cluster {
	c := newTestClusterBase()
	c.SubClusters[0].InstancePool.Instances[0].IP = "example.com"
	return c
}

func newTestClusterHTTPS() *Cluster {
	c := newTestClusterBase()
	c.Basic.Protocol = lib.PString("https")
	return c
}

func newTestClusterLLM() *Cluster {
	c := newTestClusterBase()
	c.LLMConfig = &LLMConfig{
		ModelEndpoint: &Endpoint{
			Schema: "https",
			URI:    "/models",
			Headers: map[string]string{
				"Authorization": "Bearer token",
			},
		},
		Models: []string{"m1", "m2"},
		ModelMappings: []*Mapping{
			{Key: lib.PString("old"), Value: lib.PString("new")},
		},
		Key:          lib.PString("key"),
		ProviderType: lib.PString("openai"),
	}
	return c
}

func TestCluster_SubClusterNames(t *testing.T) {
	c := newTestClusterBase()
	assert.Equal(t, []string{"sc1"}, c.SubClusterNames())
}

func TestCluster_getBalanceMode(t *testing.T) {
	t.Run("WRR", func(t *testing.T) {
		c := newTestClusterBase()
		assert.Equal(t, "WRR", c.getBalanceMode())
	})

	t.Run("EPP", func(t *testing.T) {
		c := newTestClusterEPP()
		assert.Equal(t, "EPP", c.getBalanceMode())
	})
}

func TestClusterList2MapByName(t *testing.T) {
	c1 := &Cluster{ID: 1, Name: "a"}
	c2 := &Cluster{ID: 2, Name: "b"}
	got := ClusterList2MapByName([]*Cluster{c1, c2})
	assert.Equal(t, map[string]*Cluster{"a": c1, "b": c2}, got)
}

func TestClusterList2MapByID(t *testing.T) {
	c1 := &Cluster{ID: 1, Name: "a"}
	c2 := &Cluster{ID: 2, Name: "b"}
	got := ClusterList2MapByID([]*Cluster{c1, c2})
	assert.Equal(t, map[int64]*Cluster{1: c1, 2: c2}, got)
}

func TestNewClusterManager(t *testing.T) {
	m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
	require.NotNil(t, m)
}

func TestClusterManager_FetchClusterList(t *testing.T) {
	ctx := context.Background()
	expected := []*Cluster{{ID: 1, Name: "c1"}}
	store := &fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
			assert.Equal(t, lib.PString("c1"), param.Name)
			return expected, nil
		},
	}
	m := NewClusterManager(&fakeTxn{}, store, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)

	got, err := m.FetchClusterList(ctx, &ClusterFilter{Name: lib.PString("c1")})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestClusterManager_FetchCluster(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		store := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return []*Cluster{{ID: 1, Name: "c1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, store, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		got, err := m.FetchCluster(ctx, &ClusterFilter{Name: lib.PString("c1")})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(1), got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, store, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		got, err := m.FetchCluster(ctx, &ClusterFilter{Name: lib.PString("c1")})
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("error", func(t *testing.T) {
		store := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewClusterManager(&fakeTxn{}, store, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		_, err := m.FetchCluster(ctx, &ClusterFilter{})
		require.Error(t, err)
	})
}

func TestClusterManager_CreateCluster(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("success with existing sub cluster", func(t *testing.T) {
		bindCalled := false
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, nil
			},
			clusterCreateFn: func(ctx context.Context, product *ibasic.Product, param *ClusterParam, subClusters []*SubCluster) (int64, error) {
				assert.Equal(t, int64(2), product.ID)
				assert.Equal(t, []string{"sc1"}, param.SubClusters)
				return 10, nil
			},
			bindSubClusterFn: func(ctx context.Context, cluster *Cluster, appendSubClusters, unbindSubClusters []*SubCluster) error {
				bindCalled = true
				assert.Equal(t, int64(10), cluster.ID)
				assert.Len(t, appendSubClusters, 1)
				assert.Equal(t, "sc1", appendSubClusters[0].Name)
				return nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				assert.Equal(t, []string{"sc1"}, param.Names)
				return []*SubCluster{{ID: 1, Name: "sc1", Ready: true}}, nil
			},
		}
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.CreateCluster(ctx, product, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1"},
		})
		require.NoError(t, err)
		assert.True(t, bindCalled)
	})

	t.Run("cluster already existed", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return []*Cluster{{Name: "c1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.CreateCluster(ctx, product, &ClusterParam{Name: lib.PString("c1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cluster Record Existed")
	})

	t.Run("success with auto instance pool", func(t *testing.T) {
		bindCalled := false
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, nil
			},
			clusterCreateFn: func(ctx context.Context, product *ibasic.Product, param *ClusterParam, subClusters []*SubCluster) (int64, error) {
				return 20, nil
			},
			bindSubClusterFn: func(ctx context.Context, cluster *Cluster, appendSubClusters, unbindSubClusters []*SubCluster) error {
				bindCalled = true
				return nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{{ID: 2, Name: "c1", Ready: true}}, nil
			},
		}
		poolStore := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return nil, nil
			},
			createPoolFn: func(ctx context.Context, product *ibasic.Product, data *PoolParam) (*Pool, error) {
				assert.Equal(t, "test.c1", *data.Name)
				return &Pool{ID: 1, Name: "test.c1"}, nil
			},
		}
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: stateful.DefaultConfig.RunTime.DefaultAIClusterName}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, bfeClusterStore, poolStore, nil, nil)
		err := m.CreateCluster(ctx, product, &ClusterParam{
			Name: lib.PString("c1"),
			InstancePool: []Instance{
				{HostName: "rs1", IP: "127.0.0.1", Port: 80, Weight: 10},
			},
		})
		require.NoError(t, err)
		assert.True(t, bindCalled)
	})

	t.Run("EPP must single subcluster", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{
					{ID: 1, Name: "sc1", Role: ProductPoolRoleEPP, Ready: true},
					{ID: 2, Name: "sc2", Role: ProductPoolRoleEPP, Ready: true},
				}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.CreateCluster(ctx, product, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1", "sc2"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subcluster is EPP")
	})

	t.Run("sub cluster not exist", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{{ID: 1, Name: "sc1", Ready: true}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.CreateCluster(ctx, product, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1", "sc2"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SubCluster sc2 Not Exist")
	})

	t.Run("manual LB invalid", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
				return nil, nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{{ID: 1, Name: "sc1", Ready: true}}, nil
			},
		}
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.CreateCluster(ctx, product, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1"},
			Scheduler: map[string]map[string]int{
				"bfe1": {"sc1": 50},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Total Rate Is 50")
	})
}

func TestClusterManager_constructDefaultScheduler(t *testing.T) {
	ctx := context.Background()
	bfeClusterStore := &fakeBFEClusterStorager{
		fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
			return []*ibasic.BFECluster{{Name: "bfe1"}, {Name: "bfe2"}}, nil
		},
	}
	m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, bfeClusterStore, &fakePoolStorager{}, nil, nil)
	got, err := m.constructDefaultScheduler(ctx, []*SubCluster{{Name: "sc1"}, {Name: "sc2"}})
	require.NoError(t, err)
	assert.Equal(t, map[string]map[string]int{
		"bfe1": {"sc1": 50, "sc2": 50, BlackHole: 0},
		"bfe2": {"sc1": 50, "sc2": 50, BlackHole: 0},
	}, got)
}

func TestClusterManager_checkManualLB(t *testing.T) {
	ctx := context.Background()

	t.Run("nil scheduler", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.checkManualLB(ctx, nil, &ClusterParam{Name: lib.PString("c1")})
		require.NoError(t, err)
	})

	t.Run("valid scheduler", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.checkManualLB(ctx, nil, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1"},
			Scheduler: map[string]map[string]int{
				"bfe1": {"sc1": 100},
			},
		})
		require.NoError(t, err)
	})

	t.Run("bfe cluster count mismatch", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}, {Name: "bfe2"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.checkManualLB(ctx, nil, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1"},
			Scheduler: map[string]map[string]int{
				"bfe1": {"sc1": 100},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Want All BFE Cluster Exist")
	})

	t.Run("sub cluster not in scheduler", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.checkManualLB(ctx, nil, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1"},
			Scheduler: map[string]map[string]int{
				"bfe1": {"sc2": 100},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SubCluster bfe1 Not In BFE Cluster sc2 Config")
	})

	t.Run("total not 100", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.checkManualLB(ctx, nil, &ClusterParam{
			Name:        lib.PString("c1"),
			SubClusters: []string{"sc1"},
			Scheduler: map[string]map[string]int{
				"bfe1": {"sc1": 90},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Total Rate Is 90")
	})

	t.Run("use old sub clusters", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, bfeClusterStore, &fakePoolStorager{}, nil, nil)
		err := m.checkManualLB(ctx, &Cluster{
			SubClusters: []*SubCluster{{Name: "sc1"}},
		}, &ClusterParam{
			Name: lib.PString("c1"),
			Scheduler: map[string]map[string]int{
				"bfe1": {"sc1": 100},
			},
		})
		require.NoError(t, err)
	})
}

func TestClusterManager_checkBindingSubClusters(t *testing.T) {
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.checkBindingSubClusters(ctx, nil, []string{}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cluster Want At Least On SubCluster")
	})

	t.Run("mounted to other cluster", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.checkBindingSubClusters(ctx, &Cluster{ID: 1}, []string{"sc1"}, []*SubCluster{{ID: 1, Name: "sc1", ClusterID: 2, Ready: true}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "be Mounted With Cluster 2")
	})

	t.Run("not ready", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.checkBindingSubClusters(ctx, nil, []string{"sc1"}, []*SubCluster{{ID: 1, Name: "sc1", Ready: false}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Ready")
	})

	t.Run("success", func(t *testing.T) {
		stateful.IgnoreBNSStatusCheck = true
		defer func() { stateful.IgnoreBNSStatusCheck = false }()
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.checkBindingSubClusters(ctx, &Cluster{ID: 1}, []string{"sc1"}, []*SubCluster{{ID: 1, Name: "sc1", ClusterID: 1, Ready: false}})
		require.NoError(t, err)
	})
}

func TestClusterManager_UpdateCluster(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("success", func(t *testing.T) {
		updated := false
		clusterStore := &fakeClusterStorager{
			clusterUpdateFn: func(ctx context.Context, product *ibasic.Product, old *Cluster, param *ClusterParam) error {
				updated = true
				return nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.UpdateCluster(ctx, product, newTestClusterBase(), &ClusterParam{Name: lib.PString("c1")})
		require.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("update instance pool", func(t *testing.T) {
		poolUpdated := false
		clusterStore := &fakeClusterStorager{
			clusterUpdateFn: func(ctx context.Context, product *ibasic.Product, old *Cluster, param *ClusterParam) error {
				return nil
			},
		}
		poolStore := &fakePoolStorager{
			updatePoolFn: func(ctx context.Context, oldData *Pool, diff *PoolParam) error {
				poolUpdated = true
				assert.Equal(t, "p1", oldData.Name)
				return nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, poolStore, nil, nil)
		err := m.UpdateCluster(ctx, product, newTestClusterBase(), &ClusterParam{
			Name: lib.PString("c1"),
			InstancePool: []Instance{
				{HostName: "rs2", IP: "127.0.0.2", Port: 80, Weight: 10},
			},
		})
		require.NoError(t, err)
		assert.True(t, poolUpdated)
	})

	t.Run("EPP count invalid", func(t *testing.T) {
		c := newTestClusterEPP()
		c.SubClusters = append(c.SubClusters, &SubCluster{ID: 2, Name: "sc2", Role: ProductPoolRoleEPP})
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.UpdateCluster(ctx, product, c, &ClusterParam{Name: lib.PString("c1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subcluster is EPP")
	})
}

func TestClusterManager_RebindSubCluster(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("no change", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.RebindSubCluster(ctx, product, newTestClusterBase(), []string{"sc1"})
		require.NoError(t, err)
	})

	t.Run("unbind with non zero rate", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.RebindSubCluster(ctx, product, newTestClusterBase(), []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rate is 100")
	})

	t.Run("append sub cluster", func(t *testing.T) {
		bindCalled := false
		clusterStore := &fakeClusterStorager{
			clusterUpdateFn: func(ctx context.Context, product *ibasic.Product, old *Cluster, param *ClusterParam) error {
				return nil
			},
			bindSubClusterFn: func(ctx context.Context, cluster *Cluster, appendSubClusters, unbindSubClusters []*SubCluster) error {
				bindCalled = true
				assert.Len(t, appendSubClusters, 1)
				assert.Equal(t, "sc2", appendSubClusters[0].Name)
				return nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{
					{ID: 1, Name: "sc1", ClusterID: 1, Ready: true},
					{ID: 2, Name: "sc2", Ready: true},
				}, nil
			},
		}
		c := newTestClusterBase()
		c.Scheduler["bfe1"]["sc2"] = 0
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		err := m.RebindSubCluster(ctx, product, c, []string{"sc1", "sc2"})
		require.NoError(t, err)
		assert.True(t, bindCalled)
	})
}

func TestClusterManager_DeleteCluster(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("success", func(t *testing.T) {
		deleted := false
		bindCalled := false
		subClusterDeleted := false
		poolDeleted := false
		clusterStore := &fakeClusterStorager{
			bindSubClusterFn: func(ctx context.Context, cluster *Cluster, appendSubClusters, unbindSubClusters []*SubCluster) error {
				bindCalled = true
				return nil
			},
			clusterDeleteFn: func(ctx context.Context, product *ibasic.Product, cluster *Cluster) error {
				deleted = true
				return nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			deleteSubClusterFn: func(ctx context.Context, param *SubCluster) error {
				subClusterDeleted = true
				return nil
			},
		}
		poolStore := &fakePoolStorager{
			deletePoolFn: func(ctx context.Context, pool *Pool) error {
				poolDeleted = true
				return nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, subClusterStore, &fakeBFEClusterStorager{}, poolStore, nil, nil)
		err := m.DeleteCluster(ctx, product, newTestClusterBase())
		require.NoError(t, err)
		assert.True(t, bindCalled)
		assert.True(t, deleted)
		assert.True(t, subClusterDeleted)
		assert.True(t, poolDeleted)
	})

	t.Run("delete checker error", func(t *testing.T) {
		m := NewClusterManager(&fakeTxn{}, &fakeClusterStorager{}, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, map[string]func(context.Context, *ibasic.Product, *Cluster) error{
			"route": func(ctx context.Context, p *ibasic.Product, c *Cluster) error {
				return errors.New("route in use")
			},
		})
		err := m.DeleteCluster(ctx, product, newTestClusterBase())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "route in use")
	})
}

func TestClusterManager_IsBFEClusterUsed(t *testing.T) {
	ctx := context.Background()

	t.Run("used", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchLBMatrixListFn: func(ctx context.Context) (map[int64]map[string]map[string]int, error) {
				return map[int64]map[string]map[string]int{
					1: {"bfe1": {"sc1": 100}},
				}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		used, err := m.IsBFEClusterUsed(ctx, "bfe1")
		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("not used", func(t *testing.T) {
		clusterStore := &fakeClusterStorager{
			fetchLBMatrixListFn: func(ctx context.Context) (map[int64]map[string]map[string]int, error) {
				return map[int64]map[string]map[string]int{
					1: {"bfe2": {"sc1": 100}},
				}, nil
			},
		}
		m := NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, nil, nil)
		used, err := m.IsBFEClusterUsed(ctx, "bfe1")
		require.NoError(t, err)
		assert.False(t, used)
	})
}

func TestAppendAdvancedRuleCluster(t *testing.T) {
	list := []*Cluster{{ID: 1, Name: "c1"}}
	got := AppendAdvancedRuleCluster(list)
	require.Len(t, got, 2)
	assert.Equal(t, RouteAdvancedModeClusterID, got[1].ID)
	assert.Equal(t, RouteAdvancedModeClusterName4DP, got[1].Name)
}

func TestNewBfeClusterConf(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		conf := NewBfeClusterConf("v1", []*Cluster{newTestClusterBase()})
		require.NotNil(t, conf)
		require.NotNil(t, conf.Config)
		assert.Equal(t, "v1", *conf.Version)
		cConf, ok := (*conf.Config)["c1"]
		require.True(t, ok)
		assert.NotNil(t, cConf.BackendConf)
		assert.NotNil(t, cConf.CheckConf)
		assert.NotNil(t, cConf.GslbBasic)
		assert.NotNil(t, cConf.ClusterBasic)
	})

	t.Run("skip system route", func(t *testing.T) {
		conf := NewBfeClusterConf("v1", []*Cluster{
			newTestClusterBase(),
			{ID: RouteAdvancedModeClusterID, Name: RouteAdvancedModeClusterName},
		})
		require.Len(t, *conf.Config, 1)
	})

	t.Run("EPP addrs", func(t *testing.T) {
		conf := NewBfeClusterConf("v1", []*Cluster{newTestClusterEPP()})
		cConf := (*conf.Config)["c1"]
		require.NotNil(t, cConf.GslbBasic.EPPAddr)
		assert.Equal(t, []string{"epp.example.com:8080"}, *cConf.GslbBasic.EPPAddr)
	})

	t.Run("https conf", func(t *testing.T) {
		conf := NewBfeClusterConf("v1", []*Cluster{newTestClusterHTTPS()})
		cConf := (*conf.Config)["c1"]
		require.NotNil(t, cConf.HTTPSConf)
		assert.True(t, *cConf.HTTPSConf.RSInsecureSkipVerify)
	})

	t.Run("domain pool disable checks", func(t *testing.T) {
		conf := NewBfeClusterConf("v1", []*Cluster{newTestClusterDomain()})
		cConf := (*conf.Config)["c1"]
		assert.True(t, *cConf.ClusterBasic.DisableHealthCheck)
		assert.True(t, *cConf.ClusterBasic.DisableHostHeader)
	})

	t.Run("LLM config", func(t *testing.T) {
		conf := NewBfeClusterConf("v1", []*Cluster{newTestClusterLLM()})
		cConf := (*conf.Config)["c1"]
		require.NotNil(t, cConf.AIConf)
		assert.Equal(t, lib.PString("key"), cConf.AIConf.Key)
		require.NotNil(t, cConf.AIConf.ModelMapping)
		assert.Equal(t, "new", (*cConf.AIConf.ModelMapping)["old"])
	})
}

func TestConvertToBFEModelMapping(t *testing.T) {
	t.Run("non empty", func(t *testing.T) {
		got := convertToBFEModelMapping([]*Mapping{
			{Key: lib.PString("k1"), Value: lib.PString("v1")},
		})
		require.NotNil(t, got)
		assert.Equal(t, map[string]string{"k1": "v1"}, *got)
	})

	t.Run("empty", func(t *testing.T) {
		got := convertToBFEModelMapping(nil)
		assert.Nil(t, got)
	})
}

func TestIsDomainPool(t *testing.T) {
	t.Run("IP", func(t *testing.T) {
		assert.False(t, isDomainPool(newTestClusterBase().SubClusters))
	})

	t.Run("domain", func(t *testing.T) {
		assert.True(t, isDomainPool(newTestClusterDomain().SubClusters))
	})
}

func TestBuildEPPAddrsFromSubClusters(t *testing.T) {
	t.Run("domain mode", func(t *testing.T) {
		got := buildEPPAddrsFromSubClusters(newTestClusterEPP().SubClusters)
		require.NotNil(t, got)
		assert.Equal(t, []string{"epp.example.com:8080"}, *got)
	})

	t.Run("endpoints mode", func(t *testing.T) {
		c := newTestClusterEPP()
		c.SubClusters[0].InstancePool.EPPServer = &EPPServer{
			Endpoints: []*EPPEndpoint{
				{IP: lib.PString("127.0.0.1"), Port: lib.PInt(9090)},
			},
		}
		got := buildEPPAddrsFromSubClusters(c.SubClusters)
		require.NotNil(t, got)
		assert.Equal(t, []string{"127.0.0.1:9090"}, *got)
	})

	t.Run("empty", func(t *testing.T) {
		assert.Nil(t, buildEPPAddrsFromSubClusters(nil))
	})
}

func TestClusterPassiveHealthCheck_toBackendCheck(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var c *ClusterPassiveHealthCheck
		assert.Nil(t, c.toBackendCheck())
	})

	t.Run("values", func(t *testing.T) {
		c := &ClusterPassiveHealthCheck{
			Schema:     "http",
			Interval:   10,
			Failnum:    3,
			Statuscode: 500,
			Host:       "h",
			Uri:        "/",
		}
		got := c.toBackendCheck()
		require.NotNil(t, got)
		assert.Equal(t, "http", *got.Schem)
	})
}

func TestClusterTableConf_UpdateVersion(t *testing.T) {
	ctc := &ClusterTableConf{}
	require.NoError(t, ctc.UpdateVersion("v1"))
	assert.Equal(t, "v1", *ctc.Version)
}

func TestGSLBConf_UpdateVersion(t *testing.T) {
	gc := &GSLBConf{}
	require.NoError(t, gc.UpdateVersion("v1"))
	assert.Equal(t, "v1", gc.Version)
	assert.Equal(t, "v1", *gc.Ts)
}

func newClusterManagerForExport(t *testing.T, version string) *ClusterManager {
	t.Helper()
	vcs := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return version, nil
		},
	}
	vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, vcs)
	clusterStore := &fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *ClusterFilter) ([]*Cluster, error) {
			return []*Cluster{newTestClusterBase()}, nil
		},
	}
	return NewClusterManager(&fakeTxn{}, clusterStore, &fakeSubClusterStorager{}, &fakeBFEClusterStorager{}, &fakePoolStorager{}, vcm, nil)
}
