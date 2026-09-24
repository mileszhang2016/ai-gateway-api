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

package ioperlog

import (
	"strings"
)

const (
	maskPlaceholder = "******"
	updatedMarker   = "[已更新]"
)

// MaskAPIKeyToken masks an API key token, keeping the first and last 4 chars.
func MaskAPIKeyToken(token string) string {
	if len(token) <= 8 {
		return maskPlaceholder
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// MaskString returns a masked placeholder for any sensitive string.
func MaskString(_ string) string {
	return maskPlaceholder
}

// MaskSensitiveFields recursively masks sensitive fields in a map.
// It mutates the input map in place.
func MaskSensitiveFields(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	for key, val := range data {
		lowerKey := strings.ToLower(key)

		switch lowerKey {
		case "password", "secret", "session_key", "sessionkey", "private_key", "privatekey",
			"token", "access_token", "refresh_token", "session_token", "id_token",
			"secret_key", "secretkey", "client_secret", "clientsecret",
			"access_key", "accesskey":
			data[key] = maskPlaceholder
		case "api_key", "apikey", "key":
			if s, ok := val.(string); ok {
				data[key] = MaskAPIKeyToken(s)
			}
		case "certificate", "cert", "cert_body", "private_key_body":
			data[key] = updatedMarker
		default:
			// Recurse into nested maps and maps inside arrays.
			switch typed := val.(type) {
			case map[string]interface{}:
				data[key] = MaskSensitiveFields(typed)
			case []interface{}:
				data[key] = maskSlice(typed)
			}
		}
	}

	return data
}

// maskSlice masks sensitive fields in maps nested inside a slice.
func maskSlice(items []interface{}) []interface{} {
	for i, item := range items {
		switch v := item.(type) {
		case map[string]interface{}:
			items[i] = MaskSensitiveFields(v)
		case []interface{}:
			items[i] = maskSlice(v)
		}
	}
	return items
}

// MaskErrorMessage redacts known sensitive values (e.g. API-Key values) from a
// free-text error message by replacing each occurrence with the masked token
// form (first 4 + "****" + last 4, "******" for short values), mirroring the
// change_summary masking contract. Empty values are skipped.
func MaskErrorMessage(msg string, values ...string) string {
	for _, v := range values {
		if v == "" {
			continue
		}
		msg = strings.ReplaceAll(msg, v, MaskAPIKeyToken(v))
	}
	return msg
}
