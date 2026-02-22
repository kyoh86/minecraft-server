package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func (a app) ensureSecrets() error {
	secretsDir := filepath.Join(a.baseDir, "secrets")
	if err := os.MkdirAll(secretsDir, 0o700); err != nil {
		return err
	}

	tokenPath := filepath.Join(secretsDir, "mc_link_discord_bot_token.txt")
	if err := ensureDiscordBotToken(tokenPath); err != nil {
		return err
	}

	forwardingPath := filepath.Join(secretsDir, "mc_forwarding_secret.txt")
	if err := ensureForwardingSecret(forwardingPath); err != nil {
		return err
	}
	return nil
}

func ensureDiscordBotToken(path string) error {
	placeholder := "REPLACE_WITH_DISCORD_BOT_TOKEN"
	if !fileExists(path) {
		token, err := promptSecret("Discord Bot tokenを入力してください（未設定のままにする場合はEnter）: ")
		if err != nil {
			return err
		}
		if token == "" {
			token = placeholder
		}
		return os.WriteFile(path, []byte(token+"\n"), 0o600)
	}
	current, err := readTrimmed(path)
	if err != nil {
		return err
	}
	if current != placeholder {
		return nil
	}
	token, err := promptSecret("Discord Bot tokenが未設定です。入力して更新しますか？（空Enterでスキップ）: ")
	if err != nil {
		return err
	}
	if token == "" {
		return nil
	}
	return os.WriteFile(path, []byte(token+"\n"), 0o600)
}

func ensureForwardingSecret(path string) error {
	if !fileExists(path) {
		secret, err := promptSecret("forwarding secretを入力してください（空Enterで自動生成）: ")
		if err != nil {
			return err
		}
		if secret == "" {
			secret, err = generateForwardingSecret()
			if err != nil {
				return err
			}
		}
		return os.WriteFile(path, []byte(secret+"\n"), 0o600)
	}
	current, err := readTrimmed(path)
	if err != nil {
		return err
	}
	if current != "" {
		return nil
	}
	secret, err := promptSecret("forwarding secretが空です。入力してください（空Enterで自動生成）: ")
	if err != nil {
		return err
	}
	if secret == "" {
		secret, err = generateForwardingSecret()
		if err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(secret+"\n"), 0o600)
}

func promptSecret(prompt string) (string, error) {
	if !isInteractiveStdin() {
		return "", nil
	}
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	s, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func isInteractiveStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func readTrimmed(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func (a app) readForwardingSecret() (string, error) {
	path := filepath.Join(a.baseDir, "secrets", "mc_forwarding_secret.txt")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return "", fmt.Errorf("forwarding secret is empty: %s", path)
	}
	return secret, nil
}

func (a app) renderLimboConfig() error {
	secret, err := a.readForwardingSecret()
	if err != nil {
		return err
	}

	src := filepath.Join(a.baseDir, "infra", "limbo", "config", "server.toml.tmpl")
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	tpl, err := template.New(filepath.Base(src)).Option("missingkey=error").Parse(string(in))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", src, err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, map[string]string{"ForwardingSecret": secret}); err != nil {
		return fmt.Errorf("render template %s: %w", src, err)
	}

	dst := filepath.Join(a.baseDir, "runtime", "limbo", "server.toml")
	dst = filepath.Join(a.baseDir, "secrets", "limbo", "server.toml")
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	return os.WriteFile(dst, out.Bytes(), 0o600)
}

func (a app) renderWorldPaperGlobal() error {
	secret, err := a.readForwardingSecret()
	if err != nil {
		return err
	}

	src := filepath.Join(a.baseDir, "infra", "world", "config", "paper-global.yml.tmpl")
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	tpl, err := template.New(filepath.Base(src)).Option("missingkey=error").Parse(string(in))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", src, err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, map[string]string{"ForwardingSecret": secret}); err != nil {
		return fmt.Errorf("render template %s: %w", src, err)
	}

	dst := filepath.Join(a.baseDir, "secrets", "world", "paper-global.yml")
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	return os.WriteFile(dst, out.Bytes(), 0o600)
}

func generateForwardingSecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
