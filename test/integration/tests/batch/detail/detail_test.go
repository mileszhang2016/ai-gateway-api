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

package detail_test

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	helper "github.com/rainway-ai-gateway/ai-gateway-api/integration/tests/batch/helper"
	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sm *testutil.ServerManager

func TestMain(m *testing.M) {
	var err error
	sm, err = testutil.StartServerAuto()
	if err != nil {
		panic("failed to start server: " + err.Error())
	}
	code := m.Run()
	sm.Shutdown()
	os.Exit(code)
}

// ---------- BT-2-000 可达性冒烟：产品线回退后可进入业务逻辑（ghost → 404） ----------

func TestBatchDetail_Reachable(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/batches/" + helper.Unique("ghost"))
	require.NoError(t, err)
	testutil.AssertErrCode(t, resp, 404)
}

// ---------- BT-2-001 任务不存在 → 404 BATCH_NOT_FOUND ----------

func TestBatchDetail_NotFound(t *testing.T) {
	helper.RequireBatchAPI(t)

	resp, err := testutil.GetClient().Get("/open-api/v1/batches/" + helper.Unique("ghost"))
	require.NoError(t, err)
	testutil.AssertErrCode(t, resp, 404)
}

// ---------- BT-2-002 直写一行 → 200 且字段完整 ----------

func TestBatchDetail_FullFields(t *testing.T) {
	helper.RequireBatchAPI(t)
	db := helper.OpenDB(t, sm)

	tag := helper.Unique("bt-2-002")
	row := helper.BatchTaskRow{
		BatchID:      tag,
		APIKeyID:     tag + "-ak",
		ProductName:  helper.TestProductName,
		EntityID:     tag + "-entity",
		Provider:     "deepseek",
		KeyName:      "key-primary",
		Endpoint:     "/v1/batches",
		InputFileID:  tag + "-in",
		OutputFileID: tag + "-out",
		Status:       "in_progress",
		SettleStatus: "reserved",
		ReserveUnits: 250000000, // 2.5 RMB，1e-8 定点整数
		OverReserved: 1,
		SyncSource:   "log",
		IdemKey:      tag + "-idem",
	}
	helper.InsertBatchTask(t, db, row)
	defer helper.DeleteBatchTask(t, db, tag)

	helper.InsertBatchFile(t, db, helper.BatchFileRow{
		FileID: tag + "-in", Provider: "deepseek", APIKeyID: tag + "-ak",
		ProductName: helper.TestProductName, Direction: "input", Purpose: "batch_input", Lines: 10, Bytes: 1024,
	})
	helper.InsertBatchFile(t, db, helper.BatchFileRow{
		FileID: tag + "-out", Provider: "deepseek", APIKeyID: tag + "-ak",
		ProductName: helper.TestProductName, Direction: "output", Purpose: "batch_output", Lines: 9, Bytes: 2048,
	})
	defer helper.DeleteBatchFile(t, db, tag+"-in", "deepseek")
	defer helper.DeleteBatchFile(t, db, tag+"-out", "deepseek")

	resp, err := testutil.GetClient().Get("/open-api/v1/batches/" + tag)
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	task, ok := data["task"].(map[string]interface{})
	require.True(t, ok, "Data.task must exist")

	assert.Equal(t, tag, task["batch_id"])
	assert.Equal(t, tag+"-ak", task["api_key_id"])
	assert.Equal(t, helper.TestProductName, task["product_name"])
	assert.Equal(t, tag+"-entity", task["entity_id"])
	assert.Equal(t, "deepseek", task["provider"])
	assert.Equal(t, "key-primary", task["key_name"])
	assert.Equal(t, "/v1/batches", task["endpoint"])
	assert.Equal(t, tag+"-in", task["input_file_id"])
	assert.Equal(t, tag+"-out", task["output_file_id"])
	assert.Equal(t, "in_progress", task["status"])
	assert.Equal(t, "reserved", task["settle_status"])
	assert.Equal(t, true, task["over_reserved"])
	assert.Equal(t, "log", task["sync_source"])

	// reserve_units 必须按 1e-8 定点整数输出：值为 250000000 且
	// 原始报文 token 无小数点/科学计数法（不做 ÷1e8 换算）。
	assert.Equal(t, float64(250000000), task["reserve_units"])
	needle := strconv.Quote("reserve_units") + ":"
	idx := strings.LastIndex(string(resp.RawBody), needle)
	require.NotEqual(t, -1, idx)
	rest := string(resp.RawBody)[idx+len(needle):]
	end := strings.IndexAny(rest, ",}")
	require.NotEqual(t, -1, end)
	token := strings.TrimSpace(rest[:end])
	assert.Equal(t, "250000000", token, "reserve_units must serialize as integer token")

	// 关联 batch_files（输入/输出）必须组装在 files 中
	files, ok := data["files"].([]interface{})
	require.True(t, ok, "Data.files must exist")
	require.Equal(t, 2, len(files))
	byID := map[string]map[string]interface{}{}
	for _, item := range files {
		f := item.(map[string]interface{})
		byID[f["file_id"].(string)] = f
	}
	require.Contains(t, byID, tag+"-in")
	require.Contains(t, byID, tag+"-out")
	assert.Equal(t, "input", byID[tag+"-in"]["direction"])
	assert.Equal(t, "output", byID[tag+"-out"]["direction"])
	assert.Equal(t, float64(1024), byID[tag+"-in"]["bytes"])
}

// ---------- BT-2-003 产品线隔离：其它产品线的任务按不存在处理 ----------

func TestBatchDetail_ProductIsolation(t *testing.T) {
	helper.RequireBatchAPI(t)
	db := helper.OpenDB(t, sm)

	tag := helper.Unique("bt-2-003")
	helper.InsertBatchTask(t, db, helper.BatchTaskRow{
		BatchID: tag, APIKeyID: tag + "-ak", ProductName: "OTHER_product",
		Provider: "deepseek", Status: "in_progress", IdemKey: tag + "-idem",
	})
	defer helper.DeleteBatchTask(t, db, tag)

	resp, err := testutil.GetClient().Get("/open-api/v1/batches/" + tag)
	require.NoError(t, err)
	testutil.AssertErrCode(t, resp, 404)
}
