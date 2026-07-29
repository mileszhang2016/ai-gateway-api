package testutil

import (
	"encoding/json"
	"fmt"
)

// CreateEntityType 创建 Entity-Type，返回 type_name
func CreateEntityType(typeName string, level int) (string, error) {
	resp, err := GetClient().Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name":   typeName,
		"description": "auto-created by test helper",
		"level":       level,
	})
	if err != nil {
		return "", err
	}
	if resp.ErrNum != 200 {
		return "", fmt.Errorf("create entity-type failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	name, err := GetDataField(resp, "type_name")
	if err != nil {
		return "", err
	}
	return name.(string), nil
}

// CreateEntity 创建 Entity，返回 id
func CreateEntity(name, typeName, parentID string) (string, error) {
	body := map[string]interface{}{
		"name": name,
		"type": typeName,
	}
	if parentID != "" {
		body["parent_id"] = parentID
	}
	resp, err := GetClient().Post("/open-api/v1/entities", body)
	if err != nil {
		return "", err
	}
	if resp.ErrNum != 200 {
		return "", fmt.Errorf("create entity failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	id, err := GetDataField(resp, "id")
	if err != nil {
		return "", err
	}
	return id.(string), nil
}

// CreateAPIKey 创建 API-Key，返回 id
func CreateAPIKey(description string, entityID string) (string, error) {
	body := map[string]interface{}{
		"description": description,
	}
	if entityID != "" {
		body["entity_id"] = entityID
	}
	resp, err := GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		return "", err
	}
	if resp.ErrNum != 200 {
		return "", fmt.Errorf("create api-key failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	id, err := GetDataField(resp, "id")
	if err != nil {
		return "", err
	}
	return id.(string), nil
}

// CreateCluster 创建 Cluster，返回 name
func CreateCluster(name string) (string, error) {
	resp, err := GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
		"name": name,
		"instance_pool": []interface{}{
			map[string]interface{}{
				"hostname": "backend-1",
				"ip":       "10.0.0.1",
				"weight":   100,
				"ports": map[string]interface{}{
					"Default": 8080,
				},
			},
		},
		"llm_config": map[string]interface{}{
			"models":        []string{"deepseek-chat"},
			"key":           "sk-test",
			"provider_type": "deepseek",
		},
	})
	if err != nil {
		return "", err
	}
	if resp.ErrNum != 200 {
		return "", fmt.Errorf("create cluster failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return name, nil
}

// CreateCertificate 创建证书，返回 cert_name
func CreateCertificate(certName string, isDefault bool) (string, error) {
	certPEM, keyPEM, err := GenerateTestCert(certName)
	if err != nil {
		return "", err
	}
	resp, err := GetClient().Post("/open-api/v1/certificates", map[string]interface{}{
		"cert_name":         certName,
		"description":       "auto-created by test helper",
		"is_default":        isDefault,
		"cert_file_name":    certName + "-cert.pem",
		"cert_file_content": certPEM,
		"key_file_name":     certName + "-key.pem",
		"key_file_content":  keyPEM,
		"expired_date":      "2026-08-23 16:02:31",
	})
	if err != nil {
		return "", err
	}
	if resp.ErrNum != 200 {
		return "", fmt.Errorf("create certificate failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return certName, nil
}

// CreateUser 创建用户
func CreateUser(userName, password string) error {
	resp, err := GetClient().Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": userName,
		"password":  password,
		"is_admin":  true,
	})
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 {
		return fmt.Errorf("create user failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// CreateToken 创建 Token，返回 token 值
func CreateToken(name, scope string) (string, error) {
	resp, err := GetClient().Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":  name,
		"scope": scope,
	})
	if err != nil {
		return "", err
	}
	if resp.ErrNum != 200 {
		return "", fmt.Errorf("create token failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	token, err := GetDataField(resp, "token")
	if err != nil {
		return "", err
	}
	return token.(string), nil
}

// DeleteAPIKey 删除 API-Key
func DeleteAPIKey(id string) error {
	resp, err := GetClient().Delete("/open-api/v1/api-keys/" + id)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete api-key failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// DeleteEntity 删除 Entity
func DeleteEntity(id string) error {
	resp, err := GetClient().Delete("/open-api/v1/entities/" + id)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete entity failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// DeleteEntityType 删除 Entity-Type
func DeleteEntityType(typeName string) error {
	resp, err := GetClient().Delete("/open-api/v1/entity-types/" + typeName)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete entity-type failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// DeleteCluster 删除 Cluster
func DeleteCluster(name string) error {
	resp, err := GetClient().Delete("/open-api/v1/clusters/" + name)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete cluster failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// DeleteCertificate 删除证书
func DeleteCertificate(certName string) error {
	resp, err := GetClient().Delete("/open-api/v1/certificates/" + certName)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete certificate failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// DeleteUser 删除用户
func DeleteUser(userName string) error {
	resp, err := GetClient().Delete("/open-api/v1/auth/users/" + userName)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete user failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// DeleteToken 删除 Token
func DeleteToken(name string) error {
	resp, err := GetClient().Delete("/open-api/v1/auth/tokens/" + name)
	if err != nil {
		return err
	}
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		return fmt.Errorf("delete token failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	return nil
}

// FieldNotExists 检查 Data 中是否不包含指定字段
func FieldNotExists(resp *APIResponse, field string) (bool, error) {
	if resp == nil || len(resp.Data) == 0 {
		return true, nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return false, err
	}
	_, exists := data[field]
	return !exists, nil
}

// FieldExists 检查 Data 中是否包含指定字段
func FieldExists(resp *APIResponse, field string) (bool, error) {
	ok, err := FieldNotExists(resp, field)
	if err != nil {
		return false, err
	}
	return !ok, nil
}
