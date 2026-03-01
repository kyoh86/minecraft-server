package mclink

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type EntryType string

const (
	EntryTypeNick EntryType = "nick"
	EntryTypeUUID EntryType = "uuid"
)

type CodeEntry struct {
	Code      string
	Type      EntryType
	Value     string
	ExpiresAt time.Time
	Claimed   bool
	ClaimedBy string
	ClaimedAt time.Time
}

type Store struct {
	Codes map[string]CodeEntry
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisClient(cfg RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     strings.TrimSpace(cfg.Addr),
		Password: strings.TrimSpace(cfg.Password),
		DB:       cfg.DB,
	})
}

func keyForCode(code string) string {
	return "mc-link:code:" + strings.ToUpper(strings.TrimSpace(code))
}

func SaveCode(ctx context.Context, cli *redis.Client, entry CodeEntry) error {
	if cli == nil {
		return fmt.Errorf("redis client is nil")
	}
	code := strings.ToUpper(strings.TrimSpace(entry.Code))
	if code == "" {
		return fmt.Errorf("code is required")
	}
	fields := map[string]any{
		"code":            code,
		"type":            string(entry.Type),
		"value":           entry.Value,
		"expires_unix":    entry.ExpiresAt.Unix(),
		"claimed":         strconv.FormatBool(entry.Claimed),
		"claimed_by":      entry.ClaimedBy,
		"claimed_at_unix": unixOrZero(entry.ClaimedAt),
	}
	key := keyForCode(code)
	if err := cli.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	if !entry.ExpiresAt.IsZero() {
		if err := cli.ExpireAt(ctx, key, entry.ExpiresAt).Err(); err != nil {
			return err
		}
	}
	return nil
}

func unixOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func LoadCode(ctx context.Context, cli *redis.Client, code string) (CodeEntry, bool, error) {
	if cli == nil {
		return CodeEntry{}, false, fmt.Errorf("redis client is nil")
	}
	raw, err := cli.HGetAll(ctx, keyForCode(code)).Result()
	if err != nil {
		return CodeEntry{}, false, err
	}
	if len(raw) == 0 {
		return CodeEntry{}, false, nil
	}
	entry, ok := parseHash(raw)
	if !ok {
		return CodeEntry{}, false, nil
	}
	return entry, true, nil
}

func parseHash(raw map[string]string) (CodeEntry, bool) {
	expiresUnix, err := strconv.ParseInt(raw["expires_unix"], 10, 64)
	if err != nil {
		return CodeEntry{}, false
	}
	claimed, err := strconv.ParseBool(raw["claimed"])
	if err != nil {
		return CodeEntry{}, false
	}
	claimedAtUnix := int64(0)
	if claimed {
		claimedAtUnix, err = strconv.ParseInt(raw["claimed_at_unix"], 10, 64)
		if err != nil {
			return CodeEntry{}, false
		}
	}
	code := strings.ToUpper(strings.TrimSpace(raw["code"]))
	if code == "" {
		return CodeEntry{}, false
	}
	return CodeEntry{
		Code:      code,
		Type:      EntryType(raw["type"]),
		Value:     raw["value"],
		ExpiresAt: time.Unix(expiresUnix, 0).UTC(),
		Claimed:   claimed,
		ClaimedBy: raw["claimed_by"],
		ClaimedAt: time.Unix(claimedAtUnix, 0).UTC(),
	}, true
}
