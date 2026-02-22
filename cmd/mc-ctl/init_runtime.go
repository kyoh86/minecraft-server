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

	"github.com/pelletier/go-toml/v2"
)

const (
	discordBotTokenPlaceholder  = "REPLACE_WITH_DISCORD_BOT_TOKEN"
	discordGuildIDPlaceholder   = "REPLACE_WITH_DISCORD_GUILD_ID"
	discordGuildNamePlaceholder = "REPLACE_WITH_DISCORD_GUILD_NAME"
)

type mcLinkDiscordSecret struct {
	BotToken       string   `toml:"bot_token"`
	GuildID        string   `toml:"guild_id"`
	AllowedRoleIDs []string `toml:"allowed_role_ids"`
}

func (a app) ensureSecrets() error {
	secretsDir := filepath.Join(a.baseDir, "secrets")
	if err := os.MkdirAll(secretsDir, 0o700); err != nil {
		return err
	}

	mcLinkDiscordPath := filepath.Join(secretsDir, "mc_link_discord.toml")
	if err := ensureMcLinkDiscordSecret(mcLinkDiscordPath); err != nil {
		return err
	}

	forwardingPath := filepath.Join(secretsDir, "mc_forwarding_secret.txt")
	if err := ensureForwardingSecret(forwardingPath); err != nil {
		return err
	}

	guildNamePath := filepath.Join(secretsDir, "mc_link_discord_guild_name.txt")
	if err := ensureDiscordGuildName(guildNamePath); err != nil {
		return err
	}
	return nil
}

func ensureMcLinkDiscordSecret(path string) error {
	if !fileExists(path) {
		token, err := promptSecret("Discord Bot tokenを入力してください（未設定のままにする場合はEnter）: ")
		if err != nil {
			return err
		}
		if token == "" {
			token = discordBotTokenPlaceholder
		}
		guildID, err := promptSecret("Discord Guild IDを入力してください（未設定のままにする場合はEnter）: ")
		if err != nil {
			return err
		}
		if guildID == "" {
			guildID = discordGuildIDPlaceholder
		}
		roleIDsText, err := promptSecret("許可ロールID（複数はカンマ区切り、未設定ならEnter）: ")
		if err != nil {
			return err
		}
		secret := mcLinkDiscordSecret{
			BotToken:       token,
			GuildID:        guildID,
			AllowedRoleIDs: parseCommaSeparated(roleIDsText),
		}
		return writeMcLinkDiscordSecret(path, secret)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var secret mcLinkDiscordSecret
	if err := toml.Unmarshal(b, &secret); err != nil {
		return fmt.Errorf("parse mc_link_discord.toml: %w", err)
	}

	changed := false
	if strings.TrimSpace(secret.BotToken) == "" || secret.BotToken == discordBotTokenPlaceholder {
		token, err := promptSecret("Discord Bot tokenが未設定です。入力して更新しますか？（空Enterでスキップ）: ")
		if err != nil {
			return err
		}
		if token != "" {
			secret.BotToken = token
			changed = true
		}
	}
	if strings.TrimSpace(secret.GuildID) == "" || secret.GuildID == discordGuildIDPlaceholder {
		guildID, err := promptSecret("Discord Guild IDが未設定です。入力して更新しますか？（空Enterでスキップ）: ")
		if err != nil {
			return err
		}
		if guildID != "" {
			secret.GuildID = guildID
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return writeMcLinkDiscordSecret(path, secret)
}

func writeMcLinkDiscordSecret(path string, secret mcLinkDiscordSecret) error {
	if strings.TrimSpace(secret.BotToken) == "" {
		secret.BotToken = discordBotTokenPlaceholder
	}
	if strings.TrimSpace(secret.GuildID) == "" {
		secret.GuildID = discordGuildIDPlaceholder
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("bot_token = %q\n", secret.BotToken))
	b.WriteString(fmt.Sprintf("guild_id = %q\n", secret.GuildID))
	b.WriteString("allowed_role_ids = [")
	for i, roleID := range secret.AllowedRoleIDs {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", strings.TrimSpace(roleID)))
	}
	b.WriteString("]\n")
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func parseCommaSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func ensureDiscordGuildName(path string) error {
	if !fileExists(path) {
		name, err := promptSecret("Discordサーバー名を入力してください（未設定のままにする場合はEnter）: ")
		if err != nil {
			return err
		}
		if name == "" {
			name = discordGuildNamePlaceholder
		}
		return os.WriteFile(path, []byte(name+"\n"), 0o600)
	}
	current, err := readTrimmed(path)
	if err != nil {
		return err
	}
	if current != discordGuildNamePlaceholder {
		return nil
	}
	name, err := promptSecret("Discordサーバー名が未設定です。入力して更新しますか？（空Enterでスキップ）: ")
	if err != nil {
		return err
	}
	if name == "" {
		return nil
	}
	return os.WriteFile(path, []byte(name+"\n"), 0o600)
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
	guildName, err := a.readDiscordGuildName()
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
	if err := tpl.Execute(&out, map[string]string{
		"ForwardingSecret": secret,
		"DiscordGuildName": guildName,
	}); err != nil {
		return fmt.Errorf("render template %s: %w", src, err)
	}

	dst := filepath.Join(a.baseDir, "secrets", "limbo", "server.toml")
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	return os.WriteFile(dst, out.Bytes(), 0o600)
}

func (a app) readDiscordGuildName() (string, error) {
	path := filepath.Join(a.baseDir, "secrets", "mc_link_discord_guild_name.txt")
	name, err := readTrimmed(path)
	if err != nil {
		return "", err
	}
	if name == "" || name == discordGuildNamePlaceholder {
		return "your Discord server", nil
	}
	return name, nil
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
