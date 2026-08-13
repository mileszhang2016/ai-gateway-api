package lib

// QuotaToRedisValue converts a float64 quota value to a fixed-point integer for Redis storage.
// For RMB quotas, the value is multiplied by 1e8 to avoid Lua floating point errors.
// For total_token quotas, the value is stored as-is (must be an integer).
func QuotaToRedisValue(quota *float64, unit *string) int64 {
	if quota == nil {
		return 0
	}
	if unit != nil && *unit == "RMB" {
		return int64(*quota * 1e8)
	}
	return int64(*quota)
}

// RedisValueToQuota converts a Redis fixed-point integer back to a float64 quota value.
// For RMB quotas, the value is divided by 1e8.
// For total_token quotas, the value is returned as-is.
func RedisValueToQuota(value int64, unit *string) float64 {
	if unit != nil && *unit == "RMB" {
		return float64(value) / 1e8
	}
	return float64(value)
}
