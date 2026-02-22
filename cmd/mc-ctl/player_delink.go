package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/goccy/go-yaml"
)

type velocityAllowlist struct {
	UUIDs []string `yaml:"uuids"`
	Nicks []string `yaml:"nicks"`
}

type allowlistEntry struct {
	typ   string
	value string
}

func (a app) playerDelink(uuid string) error {
	path := filepath.Join(a.baseDir, "runtime", "velocity", "allowlist.yml")
	if strings.TrimSpace(uuid) != "" {
		removed, err := removeAllowlistEntry(path, allowlistEntry{typ: "uuid", value: uuid})
		if err != nil {
			return err
		}
		if !removed {
			return fmt.Errorf("指定UUIDは allowlist にありません: %s", uuid)
		}
		fmt.Printf("削除しました: uuid:%s\n", uuid)
		return nil
	}

	if !isInteractiveStdin() {
		return errors.New("player delink は対話端末で実行するか、UUID引数を指定してください")
	}

	cfg, err := loadVelocityAllowlist(path)
	if err != nil {
		return err
	}
	entries := flattenAllowlistEntries(cfg)
	if len(entries) == 0 {
		fmt.Printf("allowlist は空です: %s\n", path)
		return nil
	}

	fmt.Println("削除するエントリを選択してください。")
	for i, e := range entries {
		fmt.Printf("%2d) %s:%s\n", i+1, e.typ, e.value)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("番号を入力 (qで中止): ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "q") {
			fmt.Println("キャンセルしました。")
			return nil
		}

		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(entries) {
			fmt.Println("無効な番号です。表示された番号を入力してください。")
			continue
		}

		target := entries[n-1]
		removed, err := removeAllowlistEntry(path, target)
		if err != nil {
			return err
		}
		if !removed {
			return fmt.Errorf("削除対象が見つかりませんでした: %s:%s", target.typ, target.value)
		}
		fmt.Printf("削除しました: %s:%s\n", target.typ, target.value)
		return nil
	}
}

func loadVelocityAllowlist(path string) (velocityAllowlist, error) {
	cfg := velocityAllowlist{}
	if !fileExists(path) {
		return cfg, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return velocityAllowlist{}, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return velocityAllowlist{}, fmt.Errorf("allowlist の読み込みに失敗: %w", err)
	}
	return cfg, nil
}

func flattenAllowlistEntries(cfg velocityAllowlist) []allowlistEntry {
	entries := make([]allowlistEntry, 0, len(cfg.UUIDs)+len(cfg.Nicks))
	for _, v := range cfg.UUIDs {
		v = strings.TrimSpace(v)
		if v != "" {
			entries = append(entries, allowlistEntry{typ: "uuid", value: v})
		}
	}
	for _, v := range cfg.Nicks {
		v = strings.TrimSpace(v)
		if v != "" {
			entries = append(entries, allowlistEntry{typ: "nick", value: v})
		}
	}
	return entries
}

func removeAllowlistEntry(path string, target allowlistEntry) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	lockPath := path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return false, err
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return false, err
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

	cfg, err := loadVelocityAllowlist(path)
	if err != nil {
		return false, err
	}

	removed := false
	switch target.typ {
	case "uuid":
		cfg.UUIDs, removed = removeOneFold(cfg.UUIDs, target.value)
	case "nick":
		cfg.Nicks, removed = removeOneFold(cfg.Nicks, target.value)
	default:
		return false, fmt.Errorf("unknown entry type: %s", target.typ)
	}
	if !removed {
		return false, nil
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return false, err
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".allowlist-*.yml")
	if err != nil {
		return false, err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(out); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return false, err
	}
	if err := tmpFile.Chmod(0o644); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return false, err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return false, err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return false, err
	}
	return true, nil
}

func removeOneFold(values []string, needle string) ([]string, bool) {
	out := make([]string, 0, len(values))
	removed := false
	for _, v := range values {
		if !removed && strings.EqualFold(strings.TrimSpace(v), strings.TrimSpace(needle)) {
			removed = true
			continue
		}
		out = append(out, v)
	}
	return out, removed
}
