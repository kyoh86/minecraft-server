package mclink

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/redis/go-redis/v9"
)

type Allowlist struct {
	UUIDs []string `yaml:"uuids"`
}

func AddAllowlistEntry(ctx context.Context, cli *redis.Client, path string, typ EntryType, value string) error {
	if cli == nil {
		return errors.New("redis client is nil")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("value must not be empty")
	}
	unlock, err := acquireAllowlistRedisLock(ctx, cli, path)
	if err != nil {
		return err
	}
	defer unlock()
	return addAllowlistEntryUnlocked(path, typ, value)
}

func addAllowlistEntryUnlocked(path string, typ EntryType, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		b = nil
	}

	cfg := Allowlist{}
	if len(b) > 0 {
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			return err
		}
	}

	switch typ {
	case EntryTypeUUID:
		if !containsFold(cfg.UUIDs, value) {
			cfg.UUIDs = append(cfg.UUIDs, value)
		}
	default:
		return errors.New("unsupported entry type")
	}
	slices.Sort(cfg.UUIDs)

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return writeAllowlist(path, out)
}

func containsFold(values []string, needle string) bool {
	for _, v := range values {
		if strings.EqualFold(v, needle) {
			return true
		}
	}
	return false
}

func allowlistRedisLockKey(path string) string {
	sum := sha256.Sum256([]byte(path))
	return "mc-link:allowlist:lock:" + hex.EncodeToString(sum[:8])
}

func acquireAllowlistRedisLock(ctx context.Context, cli *redis.Client, path string) (func(), error) {
	const (
		lockTTL      = 5 * time.Second
		retryCount   = 20
		retryInterval = 50 * time.Millisecond
	)
	key := allowlistRedisLockKey(path)
	token, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	for i := 0; i < retryCount; i++ {
		ok, err := cli.SetNX(ctx, key, token, lockTTL).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			return func() { releaseAllowlistRedisLock(ctx, cli, key, token) }, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
	return nil, fmt.Errorf("failed to acquire allowlist redis lock: key=%s", key)
}

func releaseAllowlistRedisLock(ctx context.Context, cli *redis.Client, key, token string) {
	const script = `
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`
	_, _ = cli.Eval(ctx, script, []string{key}, token).Result()
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeAllowlist(path string, out []byte) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
