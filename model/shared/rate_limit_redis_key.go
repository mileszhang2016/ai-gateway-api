package shared

import (
	"fmt"
	"strings"
)

// BuildBFERateLimitRedisKey 构造 BFE 实际使用的完整 Redis Key。
// 控制面在导出时直接下发完整 Key，BFE 侧优先使用该 Key 而不再拼接前缀。
func BuildBFERateLimitRedisKey(policyID int64, ruleType string, name string) string {
	return fmt.Sprintf("default_bfe_rlp-%d_%s_rlp-%d_%s", policyID, ruleType, policyID, name)
}

// BuildRateLimitRedisKeys 根据限流规则生成 BFE 实际使用的完整 Redis Key 列表。
// 该函数供 API-Key / Entity 删除或更新时清理 Redis Key 使用。
func BuildRateLimitRedisKeys(policyID int64, rules *RateLimitRules) []string {
	if rules == nil {
		return nil
	}
	var keys []string
	for _, tpm := range rules.TpmConfigs {
		keys = append(keys, BuildBFERateLimitRedisKey(policyID, "RL_TPM", tpm.Name))
	}
	for _, rpm := range rules.RpmConfigs {
		keys = append(keys, BuildBFERateLimitRedisKey(policyID, "RL_RPM", rpm.Name))
	}
	return keys
}

// DiffRateLimitRedisKeys 比较新旧限流策略，返回被删除规则的 Redis Key 列表。
// 规则按 name 匹配；旧策略中存在但新策略中不存在的 name 视为被删除。
func DiffRateLimitRedisKeys(policyID int64, oldRules, newRules *RateLimitRules) []string {
	oldKeys := BuildRateLimitRedisKeys(policyID, oldRules)
	newNames := make(map[string]struct{})
	if newRules != nil {
		for _, tpm := range newRules.TpmConfigs {
			newNames[tpm.Name] = struct{}{}
		}
		for _, rpm := range newRules.RpmConfigs {
			newNames[rpm.Name] = struct{}{}
		}
	}

	var deleted []string
	for _, key := range oldKeys {
		name := extractRateLimitNameFromKey(key)
		if _, ok := newNames[name]; !ok {
			deleted = append(deleted, key)
		}
	}
	return deleted
}

// extractRateLimitNameFromKey 从 BuildBFERateLimitRedisKey 生成的 Key 中提取规则 name。
// Key 格式：default_bfe_rlp-<policyID>_RL_<TYPE>_rlp-<policyID>_<name>
func extractRateLimitNameFromKey(key string) string {
	idx := strings.LastIndex(key, "_rlp-")
	if idx == -1 {
		return ""
	}
	// 找到最后一个 "_rlp-" 之后的 "_"，name 在其后
	rest := key[idx+5:]
	underscore := strings.Index(rest, "_")
	if underscore == -1 || underscore+1 >= len(rest) {
		return ""
	}
	return rest[underscore+1:]
}
