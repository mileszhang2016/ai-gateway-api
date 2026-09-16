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

package report

import "github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"

// Endpoints holds the report query endpoints. Registered by
// endpoints/openapi_v1 only when the report module is assembled
// ([Report].Backend configured).
var Endpoints = []*xreq.Endpoint{
	ReportOverviewRoute,
	ReportTimeSeriesRoute,
	ReportRankingsRoute,
	ReportDistributionRoute,
	ReportLogsRoute,
}
