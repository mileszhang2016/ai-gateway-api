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
//limitations under the License. All rights reserved.

// Copyright (c) 2021 The BFE Authors.
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

package stateful

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"

	"github.com/bfenetworks/go-lib/log"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xcrypto"
)

// secretRing is the runtime keyring; nil means encryption is not configured
// (writes stay plaintext until a keyring file is configured and reloaded).
var secretRing atomic.Value // stores *xcrypto.Keyring

// SecurityConfig is the [Security] section of ai_gateway_api.toml.
type SecurityConfig struct {
	// MasterKeyFile points to the keyring file (the ONLY injection way; no
	// env fallback by design). Empty disables encryption (gradual enable).
	MasterKeyFile string `toml:"MasterKeyFile"`
	// ActiveKeyID selects which key in the keyring encrypts new values.
	ActiveKeyID int `toml:"ActiveKeyID"`
}

// LoadSecretRing loads (or reloads) the keyring from the current
// DefaultConfig and atomically swaps it in. Failure keeps the previous
// keyring effective.
func LoadSecretRing() error {
	if DefaultConfig == nil {
		return errors.New("security: DefaultConfig not initialized")
	}
	return loadSecretRing(&DefaultConfig.Security)
}

func loadSecretRing(conf *SecurityConfig) error {
	path := conf.MasterKeyFile
	if path == "" {
		secretRing.Store((*xcrypto.Keyring)(nil))
		log.Logger.Info("security: MasterKeyFile not configured, encryption at rest disabled")
		return nil
	}
	kr, err := xcrypto.LoadKeyringFile(path)
	if err != nil {
		return fmt.Errorf("security: load keyring %s: %v", path, err)
	}
	secretRing.Store(kr)
	log.Logger.Info("security: keyring loaded from %s, activeKeyID=%d, keys=%v",
		path, kr.ActiveID(), kr.KeyIDs())
	return nil
}

// SecretRing returns the current keyring; nil when encryption is disabled.
func SecretRing() *xcrypto.Keyring {
	if v := secretRing.Load(); v != nil {
		return v.(*xcrypto.Keyring)
	}
	return nil
}

// EncryptIfEnabled encrypts value with the active key when a keyring is
// configured; otherwise returns value unchanged.
func EncryptIfEnabled(value string) (string, error) {
	kr := SecretRing()
	if kr == nil {
		return value, nil
	}
	return kr.Encrypt(value)
}

// DecryptIfEnabled decrypts envelope values when a keyring is configured;
// plaintext values pass through. Without a keyring values are returned
// as-is (encryption disabled).
func DecryptIfEnabled(value string) (string, error) {
	kr := SecretRing()
	if kr == nil {
		return value, nil
	}
	return kr.Decrypt(value)
}

// ReloadSecurity is the monitor-port /reload/security handler: re-read the
// keyring file and swap it atomically. Any failure keeps the old keyring.
func ReloadSecurity(query url.Values) (string, error) {
	old := SecretRing()
	var oldRules = -1
	if old != nil {
		oldRules = len(old.KeyIDs())
	}
	if err := LoadSecretRing(); err != nil {
		return "", err
	}
	now := SecretRing()
	newKeys := 0
	if now != nil {
		newKeys = len(now.KeyIDs())
	}
	log.Logger.Info("security: keyring reloaded: keys %d -> %d", oldRules, newKeys)
	return fmt.Sprintf("security keyring reloaded: keys %d -> %d", oldRules, newKeys), nil
}

// CheckSecretAtRest enforces the fail-fast discipline AFTER the DB is up:
// if any sensitive column holds ciphertext, a usable keyring is mandatory;
// without ciphertext a missing keyring is allowed (gradual enable).
func CheckSecretAtRest() error {
	db, err := BFEDB()
	if err != nil {
		return err
	}
	has, err := anyCiphertextPresent(db)
	if err != nil {
		return fmt.Errorf("security: check ciphertext presence: %v", err)
	}
	kr := SecretRing()
	switch {
	case has && kr == nil:
		return errors.New("security: database holds encrypted secrets but no keyring is configured ([Security].MasterKeyFile); refuse to start")
	case has && kr != nil:
		log.Logger.Info("security: encrypted secrets present, keyring ready (activeKeyID=%d)", kr.ActiveID())
	case !has:
		log.Logger.Info("security: no encrypted secrets found in database")
	}
	return nil
}

func anyCiphertextPresent(db *sql.DB) (bool, error) {
	ctx := context.Background()
	var one int
	err := db.QueryRowContext(ctx,
		`SELECT 1 FROM providers WHERE api_keys LIKE ? LIMIT 1`, xcrypto.Marker+"%").Scan(&one)
	if err == nil {
		return true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	err = db.QueryRowContext(ctx,
		`SELECT 1 FROM api_keys WHERE api_key LIKE ? LIMIT 1`, xcrypto.Marker+"%").Scan(&one)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}
