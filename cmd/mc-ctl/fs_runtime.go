package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

func (a app) ensureRuntimeLayout() error {
	for _, dir := range []string{
		filepath.Join(a.baseDir, "runtime"),
		filepath.Join(a.baseDir, "runtime", "mc-link"),
		filepath.Join(a.baseDir, "runtime", "limbo"),
		filepath.Join(a.baseDir, "runtime", "redis"),
		filepath.Join(a.baseDir, "runtime", "world"),
		filepath.Join(a.baseDir, "runtime", "world", "config"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "ClickMobs"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "WorldGuard", "worlds"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Core"),
		filepath.Join(a.baseDir, "runtime", "world", "plugins", "Multiverse-Portals"),
		filepath.Join(a.baseDir, "runtime", "velocity"),
		filepath.Join(a.baseDir, "runtime", "velocity", "plugins"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		probePath := filepath.Join(dir, ".writecheck")
		if err := os.WriteFile(probePath, []byte("ok"), 0o644); err != nil {
			return err
		}
		_ = os.Remove(probePath)
	}
	allowlistPath := filepath.Join(a.baseDir, "runtime", "velocity", "allowlist.yml")
	if !fileExists(allowlistPath) {
		const defaultAllowlist = "uuids: []\nnicks: []\n"
		if err := os.WriteFile(allowlistPath, []byte(defaultAllowlist), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (a app) checkRuntimeOwnership() error {
	root := filepath.Join(a.baseDir, "runtime")
	wantUID := uint32(os.Getuid())
	wantGID := uint32(os.Getgid())

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return nil
		}
		if st.Uid != wantUID || st.Gid != wantGID {
			return fmt.Errorf(
				"ownership mismatch: %s is %d:%d, expected %d:%d (fix with: sudo chown -R %d:%d runtime)",
				path, st.Uid, st.Gid, wantUID, wantGID, wantUID, wantGID,
			)
		}
		return nil
	})
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		info, err := d.Info()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		return err
	})
}
