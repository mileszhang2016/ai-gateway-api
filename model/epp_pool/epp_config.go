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

package epp_pool

import (
	"bytes"
	"encoding/json"
	"net"
	"regexp"
	"strings"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
)

// Simplified epp_config field values and defaults (see api-changes.md §3.2).
const (
	LoadProfileQueueFirst = "queue-first"
	LoadProfileBalanced   = "balanced"
	LoadProfileKVFirst    = "kv-first"

	AffinityOff    = "off"
	AffinityLow    = "low"
	AffinityMedium = "medium"
	AffinityHigh   = "high"

	DefaultLoadProfile                 = LoadProfileBalanced
	DefaultAffinity                    = AffinityMedium
	DefaultPrefixCacheAffinity         = true
	DefaultKVCacheUtilizationMax       = 0.9
	DefaultWaitingQueueMax             = 0
	DefaultRunningRequestsMax          = 0
	DefaultFallbackOnEmpty             = false
	DefaultMetricsStalenessThresholdMs = 200
	DefaultSessionAffinityEnabled      = false

	// Detector internal fixed values (not exposed this round): stalenessPolicy
	// "ignore" avoids "monitoring jitter halts everything"; queue depth defaults
	// to 5 when waiting_queue_max is not enabled.
	defaultSaturationQueueDepthThreshold = 5
	stalenessPolicyIgnore                = "ignore"

	// Saturation detector modes (issue #1482): "utilization" is the pre-fix
	// behavior; "concurrency" judges saturation by in-flight load.
	SaturationDetectorUtilization = "utilization"
	SaturationDetectorConcurrency = "concurrency"

	// Concurrency counting modes, same names and values as llm-d's
	// concurrency-detector concurrencyMode.
	ConcurrencyModeRequests = "requests"
	ConcurrencyModeTokens   = "tokens"

	// Defaults mirroring llm-d concurrency-detector apiConfig defaults
	// (pkg/epp/framework/plugins/flowcontrol/saturationdetector/concurrency/config.go).
	DefaultSaturationDetectorMode = SaturationDetectorUtilization
	DefaultMaxConcurrency         = 100
	DefaultMaxTokenConcurrency    = 1_000_000
	DefaultConcurrencyHeadroom    = 0.0
	DefaultConcurrencyMode        = ConcurrencyModeRequests
)

// FlowControlUnlimited marks an explicit "no limit" max_requests.
const FlowControlUnlimited = -1

// EppConfigSimplified is the user-facing simplified shape of the per-cluster
// EPP scheduling configuration stored in clusters.epp_config. All fields are
// pointers so that "not set" (follow compiler defaults) is distinguishable
// from an explicit value; unset fields are not persisted.
type EppConfigSimplified struct {
	// ── ① 调度倾向 ──
	LoadProfile            *string `json:"load_profile,omitempty"`
	Affinity               *string `json:"affinity,omitempty"`
	PrefixCacheAffinity    *bool   `json:"prefix_cache_affinity,omitempty"`
	SessionAffinityEnabled *bool   `json:"session_affinity_enabled,omitempty"`
	SessionAffinityHeader  *string `json:"session_affinity_header,omitempty"`

	// ── ② 谁能接（准入 + 坏后端剔除） ──
	KVCacheUtilizationMax       *float64 `json:"kv_cache_utilization_max,omitempty"`
	WaitingQueueMax             *float64 `json:"waiting_queue_max,omitempty"`
	RunningRequestsMax          *float64 `json:"running_requests_max,omitempty"`
	FallbackOnEmpty             *bool    `json:"fallback_on_empty,omitempty"`
	MetricsStalenessThresholdMs *int     `json:"metrics_staleness_threshold_ms,omitempty"`

	// ── ③ 容量与排队 ──
	FlowControl *FlowControlSimplified `json:"flow_control,omitempty"`

	// ── ④ 饱和判定（issue #1482） ──
	SaturationDetector *SaturationDetectorSimplified `json:"saturation_detector,omitempty"`
}

// SaturationDetectorSimplified is the per-cluster saturation detector
// selection. Unset or mode=utilization keeps the pre-fix behavior; mode=
// concurrency compiles the llm-d concurrency-detector, which judges pool
// saturation by in-flight load (pool capacity = endpoints x maxConcurrency).
type SaturationDetectorSimplified struct {
	// Mode selects the detector: "utilization" (default) or "concurrency".
	Mode *string `json:"mode,omitempty"`
	// MaxConcurrency is the ideal per-endpoint concurrency when mode=
	// concurrency. Pool capacity = endpoints * MaxConcurrency.
	MaxConcurrency *int `json:"max_concurrency,omitempty"`
	// MaxTokenConcurrency bounds per-endpoint token concurrency when
	// concurrency_mode is "tokens".
	MaxTokenConcurrency *int `json:"max_token_concurrency,omitempty"`
	// Headroom is the allowed burst above the ideal capacity (e.g. 0.2 = 20%),
	// in [0, 1]; only relaxes the scheduling filter, not pool-level backpressure.
	Headroom *float64 `json:"headroom,omitempty"`
	// ConcurrencyMode counts requests or tokens ("requests" default).
	ConcurrencyMode *string `json:"concurrency_mode,omitempty"`
}

// FlowControlSimplified is the simplified flow-control section of epp_config.
type FlowControlSimplified struct {
	MaxRequests        *int  `json:"max_requests,omitempty"`
	QueueTTL           *int  `json:"queue_ttl,omitempty"`
	NoEndpointQueueTTL *int  `json:"no_endpoint_queue_ttl,omitempty"`
	EnableEviction     *bool `json:"enable_eviction,omitempty"`
}

// ParseEppConfig decodes the raw stored JSON of epp_config with strict field
// checking. Empty input returns nil (no config).
func ParseEppConfig(raw string) (*EppConfigSimplified, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
	decoder.DisallowUnknownFields()

	conf := &EppConfigSimplified{}
	if err := decoder.Decode(conf); err != nil {
		return nil, xerror.WrapParamErrorWithMsg("epp_config is invalid: %s", err.Error())
	}

	return conf, nil
}

// IsEmpty reports whether no field of the simplified config is set (an empty
// JSON object {} decodes to an all-nil struct).
func (c *EppConfigSimplified) IsEmpty() bool {
	return c == nil ||
		(c.LoadProfile == nil &&
			c.Affinity == nil &&
			c.PrefixCacheAffinity == nil &&
			c.SessionAffinityEnabled == nil &&
			c.SessionAffinityHeader == nil &&
			c.KVCacheUtilizationMax == nil &&
			c.WaitingQueueMax == nil &&
			c.RunningRequestsMax == nil &&
			c.FallbackOnEmpty == nil &&
			c.MetricsStalenessThresholdMs == nil &&
			c.FlowControl == nil &&
			c.SaturationDetector == nil)
}

// Validate checks all field-level rules of the simplified epp_config. It is
// enforced whenever epp_config is non-empty, regardless of balance_mode.
func (c *EppConfigSimplified) Validate() error {
	if c == nil {
		return nil
	}

	if c.LoadProfile != nil {
		switch *c.LoadProfile {
		case LoadProfileQueueFirst, LoadProfileBalanced, LoadProfileKVFirst:
		default:
			return xerror.WrapParamErrorWithMsg("epp_config.load_profile must be one of %q/%q/%q, got %q",
				LoadProfileQueueFirst, LoadProfileBalanced, LoadProfileKVFirst, *c.LoadProfile)
		}
	}

	if c.Affinity != nil {
		switch *c.Affinity {
		case AffinityOff, AffinityLow, AffinityMedium, AffinityHigh:
		default:
			return xerror.WrapParamErrorWithMsg("epp_config.affinity must be one of %q/%q/%q/%q, got %q",
				AffinityOff, AffinityLow, AffinityMedium, AffinityHigh, *c.Affinity)
		}
	}

	if c.KVCacheUtilizationMax != nil {
		if *c.KVCacheUtilizationMax <= 0 || *c.KVCacheUtilizationMax > 1 {
			return xerror.WrapParamErrorWithMsg("epp_config.kv_cache_utilization_max must be in (0, 1], got %v", *c.KVCacheUtilizationMax)
		}
	}

	if c.WaitingQueueMax != nil && *c.WaitingQueueMax < 0 {
		return xerror.WrapParamErrorWithMsg("epp_config.waiting_queue_max must be >= 0, got %v", *c.WaitingQueueMax)
	}

	if c.RunningRequestsMax != nil && *c.RunningRequestsMax < 0 {
		return xerror.WrapParamErrorWithMsg("epp_config.running_requests_max must be >= 0, got %v", *c.RunningRequestsMax)
	}

	if c.MetricsStalenessThresholdMs != nil && *c.MetricsStalenessThresholdMs <= 0 {
		return xerror.WrapParamErrorWithMsg("epp_config.metrics_staleness_threshold_ms must be > 0, got %d", *c.MetricsStalenessThresholdMs)
	}

	if c.SessionAffinityEnabled != nil && *c.SessionAffinityEnabled {
		if c.SessionAffinityHeader == nil || strings.TrimSpace(*c.SessionAffinityHeader) == "" {
			return xerror.WrapParamErrorWithMsg("epp_config.session_affinity_header is required when session_affinity_enabled is true")
		}
	}

	if c.SessionAffinityHeader != nil {
		if err := validateHeaderName(*c.SessionAffinityHeader); err != nil {
			return xerror.WrapParamErrorWithMsg("epp_config.session_affinity_header: %s", err.Error())
		}
	}

	if c.FlowControl != nil {
		if err := c.FlowControl.Validate(); err != nil {
			return err
		}
	}

	if c.SaturationDetector != nil {
		if err := c.SaturationDetector.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks the saturation_detector section (issue #1482). Concurrency
// knobs under mode=utilization are rejected explicitly — silently ignoring
// them would repeat the exact confusion reported as "flow control has no
// effect" when a value looks configured but never applies.
func (s *SaturationDetectorSimplified) Validate() error {
	if s == nil {
		return nil
	}

	switch s.EffectiveMode() {
	case SaturationDetectorUtilization:
		if s.MaxConcurrency != nil || s.MaxTokenConcurrency != nil ||
			s.Headroom != nil || s.ConcurrencyMode != nil {
			return xerror.WrapParamErrorWithMsg("epp_config.saturation_detector.max_concurrency/max_token_concurrency/headroom/concurrency_mode require mode=%q",
				SaturationDetectorConcurrency)
		}
	case SaturationDetectorConcurrency:
		if s.MaxConcurrency != nil && *s.MaxConcurrency <= 0 {
			return xerror.WrapParamErrorWithMsg("epp_config.saturation_detector.max_concurrency must be > 0, got %d", *s.MaxConcurrency)
		}
		if s.MaxTokenConcurrency != nil && *s.MaxTokenConcurrency <= 0 {
			return xerror.WrapParamErrorWithMsg("epp_config.saturation_detector.max_token_concurrency must be > 0, got %d", *s.MaxTokenConcurrency)
		}
		if s.Headroom != nil && (*s.Headroom < 0 || *s.Headroom > 1) {
			return xerror.WrapParamErrorWithMsg("epp_config.saturation_detector.headroom must be in [0, 1], got %v", *s.Headroom)
		}
		if s.ConcurrencyMode != nil && *s.ConcurrencyMode != ConcurrencyModeRequests && *s.ConcurrencyMode != ConcurrencyModeTokens {
			return xerror.WrapParamErrorWithMsg("epp_config.saturation_detector.concurrency_mode must be %q or %q, got %q",
				ConcurrencyModeRequests, ConcurrencyModeTokens, *s.ConcurrencyMode)
		}
	default:
		return xerror.WrapParamErrorWithMsg("epp_config.saturation_detector.mode must be %q or %q, got %q",
			SaturationDetectorUtilization, SaturationDetectorConcurrency, *s.Mode)
	}

	return nil
}

// Validate checks the flow_control section.
func (f *FlowControlSimplified) Validate() error {
	if f == nil {
		return nil
	}

	if f.MaxRequests != nil && *f.MaxRequests != FlowControlUnlimited && *f.MaxRequests <= 0 {
		return xerror.WrapParamErrorWithMsg("epp_config.flow_control.max_requests must be > 0 or -1, got %d", *f.MaxRequests)
	}

	if f.QueueTTL != nil && *f.QueueTTL < 0 {
		return xerror.WrapParamErrorWithMsg("epp_config.flow_control.queue_ttl must be >= 0, got %d", *f.QueueTTL)
	}

	if f.NoEndpointQueueTTL != nil && *f.NoEndpointQueueTTL < 0 {
		return xerror.WrapParamErrorWithMsg("epp_config.flow_control.no_endpoint_queue_ttl must be >= 0, got %d", *f.NoEndpointQueueTTL)
	}

	// enable_eviction (in-flight negative-priority eviction) is not wired in
	// EPP yet; keep it false (the compiled output never emits it). See the EPP
	// overload-protection plan (issue #1482) for the planned follow-up card.
	if f.EnableEviction != nil && *f.EnableEviction {
		return xerror.WrapParamErrorWithMsg("epp_config.flow_control.enable_eviction is not supported yet, keep false")
	}

	return nil
}

// EffectiveLoadProfile returns the load profile value with default applied.
func (c *EppConfigSimplified) EffectiveLoadProfile() string {
	if c == nil || c.LoadProfile == nil {
		return DefaultLoadProfile
	}
	return *c.LoadProfile
}

// EffectiveAffinity returns the affinity strength with default applied.
func (c *EppConfigSimplified) EffectiveAffinity() string {
	if c == nil || c.Affinity == nil {
		return DefaultAffinity
	}
	return *c.Affinity
}

// EffectivePrefixCacheAffinity returns the prefix cache affinity switch with default applied.
func (c *EppConfigSimplified) EffectivePrefixCacheAffinity() bool {
	if c == nil || c.PrefixCacheAffinity == nil {
		return DefaultPrefixCacheAffinity
	}
	return *c.PrefixCacheAffinity
}

// EffectiveKVCacheUtilizationMax returns the utilization filter threshold with default applied.
func (c *EppConfigSimplified) EffectiveKVCacheUtilizationMax() float64 {
	if c == nil || c.KVCacheUtilizationMax == nil {
		return DefaultKVCacheUtilizationMax
	}
	return *c.KVCacheUtilizationMax
}

// EffectiveWaitingQueueMax returns the waiting-queue filter threshold with default applied.
func (c *EppConfigSimplified) EffectiveWaitingQueueMax() float64 {
	if c == nil || c.WaitingQueueMax == nil {
		return DefaultWaitingQueueMax
	}
	return *c.WaitingQueueMax
}

// EffectiveRunningRequestsMax returns the running-requests filter threshold with default applied.
func (c *EppConfigSimplified) EffectiveRunningRequestsMax() float64 {
	if c == nil || c.RunningRequestsMax == nil {
		return DefaultRunningRequestsMax
	}
	return *c.RunningRequestsMax
}

// EffectiveFallbackOnEmpty returns the filter fallback switch with default applied.
func (c *EppConfigSimplified) EffectiveFallbackOnEmpty() bool {
	if c == nil || c.FallbackOnEmpty == nil {
		return DefaultFallbackOnEmpty
	}
	return *c.FallbackOnEmpty
}

// EffectiveMetricsStalenessThresholdMs returns the bad-endpoint eviction threshold (ms).
func (c *EppConfigSimplified) EffectiveMetricsStalenessThresholdMs() int {
	if c == nil || c.MetricsStalenessThresholdMs == nil {
		return DefaultMetricsStalenessThresholdMs
	}
	return *c.MetricsStalenessThresholdMs
}

// EffectiveSessionAffinityEnabled returns the session affinity switch with default applied.
func (c *EppConfigSimplified) EffectiveSessionAffinityEnabled() bool {
	if c == nil || c.SessionAffinityEnabled == nil {
		return DefaultSessionAffinityEnabled
	}
	return *c.SessionAffinityEnabled
}

var (
	// hostnameLabel matches a single DNS label (RFC 1123) with length 1-63.
	hostnameLabel = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	// httpHeaderNameToken matches valid HTTP header name characters (RFC 7230 token).
	httpHeaderNameToken = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`)
)

// validateHost checks that host is a hostname or an IP literal (IPv6 without
// brackets). Implemented locally (instead of lib/validate) to avoid an
// import cycle with the cluster model in later wiring.
func validateHost(host string) error {
	if host == "" {
		return xerror.WrapParamErrorWithMsg("host is required")
	}
	if len(host) > 255 {
		return xerror.WrapParamErrorWithMsg("host length must be <= 255")
	}
	if net.ParseIP(host) != nil {
		return nil
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if !hostnameLabel.MatchString(label) {
			return xerror.WrapParamErrorWithMsg("host %q contains invalid label %q", host, label)
		}
	}
	return nil
}

// validatePort checks a TCP port number.
func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return xerror.WrapParamErrorWithMsg("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

// validateHeaderName checks a non-empty HTTP header name.
// EffectiveSaturationDetectorMode returns the detector mode with default applied.
func (s *SaturationDetectorSimplified) EffectiveMode() string {
	if s == nil || s.Mode == nil {
		return DefaultSaturationDetectorMode
	}
	return *s.Mode
}

// EffectiveMaxConcurrency returns the per-endpoint concurrency with default applied.
func (s *SaturationDetectorSimplified) EffectiveMaxConcurrency() int {
	if s == nil || s.MaxConcurrency == nil {
		return DefaultMaxConcurrency
	}
	return *s.MaxConcurrency
}

// EffectiveMaxTokenConcurrency returns the per-endpoint token concurrency with default applied.
func (s *SaturationDetectorSimplified) EffectiveMaxTokenConcurrency() int {
	if s == nil || s.MaxTokenConcurrency == nil {
		return DefaultMaxTokenConcurrency
	}
	return *s.MaxTokenConcurrency
}

// EffectiveHeadroom returns the burst allowance with default applied.
func (s *SaturationDetectorSimplified) EffectiveHeadroom() float64 {
	if s == nil || s.Headroom == nil {
		return DefaultConcurrencyHeadroom
	}
	return *s.Headroom
}

// EffectiveConcurrencyMode returns the counting mode with default applied.
func (s *SaturationDetectorSimplified) EffectiveConcurrencyMode() string {
	if s == nil || s.ConcurrencyMode == nil {
		return DefaultConcurrencyMode
	}
	return *s.ConcurrencyMode
}

func validateHeaderName(name string) error {
	if strings.TrimSpace(name) == "" {
		return xerror.WrapParamErrorWithMsg("header name is required")
	}
	if !httpHeaderNameToken.MatchString(name) {
		return xerror.WrapParamErrorWithMsg("invalid HTTP header name %q", name)
	}
	return nil
}
