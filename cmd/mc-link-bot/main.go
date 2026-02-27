package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kyoh86/minecraft-server/cmd/mc-link-bot/internal/mclink"
	"github.com/pelletier/go-toml/v2"
	"github.com/redis/go-redis/v9"
)

const defaultDiscordSecretPath = "/run/secrets/mc_link_discord"

type config struct {
	Token         string
	GuildID       string
	AllowedRoleID map[string]struct{}
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	AllowlistPath string
}

type discordSecret struct {
	BotToken       string   `toml:"bot_token"`
	GuildID        string   `toml:"guild_id"`
	AllowedRoleIDs []string `toml:"allowed_role_ids"`
}

type claimCodeStatus int

const (
	claimCodeSuccess claimCodeStatus = iota
	claimCodeNotFound
	claimCodeAlreadyClaimed
	claimCodeExpired
)

var (
	errCodeNotFound       = errors.New("code not found")
	errCodeAlreadyClaimed = errors.New("code already claimed")
	errCodeExpired        = errors.New("code expired")
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Token == "" {
		log.Fatal("bot token is empty")
	}

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal(err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("discord connected as %s", r.User.String())
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		if i.ApplicationCommandData().Name != "mc" {
			return
		}
		handleCommand(s, i, cfg)
	})

	if err := dg.Open(); err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	if err := registerCommands(dg, cfg.GuildID); err != nil {
		log.Fatal(err)
	}

	log.Printf("mc-link-bot started")
	select {}
}

func loadConfig() (config, error) {
	cfg := config{
		RedisAddr:     env("MC_LINK_REDIS_ADDR", "redis:6379"),
		RedisPassword: env("MC_LINK_REDIS_PASSWORD", ""),
		AllowlistPath: env("MC_LINK_ALLOWLIST_PATH", "/allowlist/allowlist.yml"),
		AllowedRoleID: map[string]struct{}{},
	}
	secretPath := env("MC_LINK_DISCORD_SECRET_FILE", defaultDiscordSecretPath)
	secret, err := loadDiscordSecret(secretPath)
	if err != nil {
		return config{}, err
	}
	cfg.Token = secret.BotToken
	cfg.GuildID = secret.GuildID
	for _, roleID := range secret.AllowedRoleIDs {
		roleID = strings.TrimSpace(roleID)
		if roleID != "" {
			cfg.AllowedRoleID[roleID] = struct{}{}
		}
	}
	db, err := strconv.Atoi(env("MC_LINK_REDIS_DB", "0"))
	if err != nil {
		return config{}, errors.New("MC_LINK_REDIS_DB must be integer")
	}
	cfg.RedisDB = db
	if cfg.GuildID == "" {
		return config{}, errors.New("guild_id is missing in mc_link_discord secret")
	}
	return cfg, nil
}

func loadDiscordSecret(path string) (discordSecret, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return discordSecret{}, err
	}
	var cfg discordSecret
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return discordSecret{}, fmt.Errorf("failed to parse mc_link_discord secret: %w", err)
	}
	cfg.BotToken = strings.TrimSpace(cfg.BotToken)
	cfg.GuildID = strings.TrimSpace(cfg.GuildID)
	if cfg.BotToken == "" || cfg.BotToken == "REPLACE_WITH_DISCORD_BOT_TOKEN" {
		return discordSecret{}, errors.New("discord bot token is missing in mc_link_discord secret")
	}
	if cfg.GuildID == "" || cfg.GuildID == "REPLACE_WITH_DISCORD_GUILD_ID" {
		return discordSecret{}, errors.New("discord guild id is missing in mc_link_discord secret")
	}
	return cfg, nil
}

func registerCommands(s *discordgo.Session, guildID string) error {
	cmd := &discordgo.ApplicationCommand{
		Name:        "mc",
		Description: "Minecraft link operations",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "link",
				Description: "Link with one-time code",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "code",
						Description: "One-time code",
						Required:    true,
					},
				},
			},
		},
	}

	appID := s.State.User.ID
	_, err := s.ApplicationCommandCreate(appID, guildID, cmd)
	return err
}

func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, cfg config) {
	if !isRoleAllowed(i, cfg.AllowedRoleID) {
		respond(s, i, "このコマンドを実行する権限がありません。")
		return
	}

	opt := i.ApplicationCommandData().Options
	if len(opt) == 0 || opt[0].Name != "link" {
		respond(s, i, "サポートされていないコマンドです。")
		return
	}
	sub := opt[0].Options
	if len(sub) == 0 || sub[0].Name != "code" {
		respond(s, i, "code が必要です。")
		return
	}
	code := normalizeLinkCodeInput(sub[0].StringValue())
	if code == "" {
		respond(s, i, "code が空です。")
		return
	}

	ctx := context.Background()
	rdb := mclink.NewRedisClient(mclink.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	entry, status, err := claimCodeAtomically(ctx, rdb, code, interactionUserID(i), time.Now().UTC())
	if err != nil {
		log.Printf("claim code failed: code=%q err=%v", code, err)
		respond(s, i, "内部エラー: code storage を読めませんでした。")
		return
	}
	switch status {
	case claimCodeNotFound:
		respond(s, i, "無効なコードです。")
		return
	case claimCodeAlreadyClaimed:
		respond(s, i, "このコードは既に使用済みです。")
		return
	case claimCodeExpired:
		respond(s, i, "コードの有効期限が切れています。")
		return
	}

	if err := mclink.AddAllowlistEntry(ctx, rdb, cfg.AllowlistPath, entry.Type, entry.Value); err != nil {
		log.Printf(
			"allowlist update failed: path=%q type=%q value=%q err=%v",
			cfg.AllowlistPath, entry.Type, entry.Value, err,
		)
		_ = rollbackClaimIfOwned(ctx, rdb, entry)
		respond(s, i, "内部エラー: allowlist 更新に失敗しました。")
		return
	}

	msg := strings.Join([]string{
		fmt.Sprintf("リンク完了: `%s:%s` を allowlist に追加しました。", entry.Type, entry.Value),
		"ゲーム画面に戻って一度切断し、再接続してください。",
	}, "\n")
	respond(s, i, msg)
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func isRoleAllowed(i *discordgo.InteractionCreate, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	if i == nil || i.Member == nil {
		return false
	}
	for _, roleID := range i.Member.Roles {
		if _, ok := allowed[roleID]; ok {
			return true
		}
	}
	return false
}

func interactionUserID(i *discordgo.InteractionCreate) string {
	if i == nil {
		return ""
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func env(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func normalizeLinkCodeInput(raw string) string {
	code := strings.TrimSpace(raw)
	code = strings.TrimPrefix(code, "/mc link ")
	code = strings.TrimPrefix(code, "code:")
	code = strings.TrimSpace(code)
	return strings.ToUpper(code)
}

func claimCodeAtomically(ctx context.Context, rdb *redis.Client, code, claimer string, claimedAt time.Time) (mclink.CodeEntry, claimCodeStatus, error) {
	key := "mc-link:code:" + strings.ToUpper(strings.TrimSpace(code))
	claimer = strings.TrimSpace(claimer)
	const maxRetries = 5

	for range maxRetries {
		var entry mclink.CodeEntry
		err := rdb.Watch(ctx, func(tx *redis.Tx) error {
			raw, err := tx.HGetAll(ctx, key).Result()
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				return errCodeNotFound
			}

			entry, err = parseCodeEntryHash(raw)
			if err != nil {
				return err
			}
			if entry.Claimed {
				return errCodeAlreadyClaimed
			}
			if claimedAt.After(entry.ExpiresAt) {
				return errCodeExpired
			}

			entry.Claimed = true
			entry.ClaimedBy = claimer
			entry.ClaimedAt = claimedAt

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.HSet(ctx, key,
					"claimed", "true",
					"claimed_by", entry.ClaimedBy,
					"claimed_at_unix", entry.ClaimedAt.Unix(),
				)
				pipe.ExpireAt(ctx, key, entry.ExpiresAt)
				return nil
			})
			return err
		}, key)

		switch {
		case err == nil:
			return entry, claimCodeSuccess, nil
		case errors.Is(err, redis.TxFailedErr):
			continue
		case errors.Is(err, errCodeNotFound):
			return mclink.CodeEntry{}, claimCodeNotFound, nil
		case errors.Is(err, errCodeAlreadyClaimed):
			return mclink.CodeEntry{}, claimCodeAlreadyClaimed, nil
		case errors.Is(err, errCodeExpired):
			return mclink.CodeEntry{}, claimCodeExpired, nil
		default:
			return mclink.CodeEntry{}, claimCodeNotFound, err
		}
	}
	return mclink.CodeEntry{}, claimCodeNotFound, fmt.Errorf("claim retry exceeded")
}

func rollbackClaimIfOwned(ctx context.Context, rdb *redis.Client, entry mclink.CodeEntry) error {
	key := "mc-link:code:" + strings.ToUpper(strings.TrimSpace(entry.Code))
	return rdb.Watch(ctx, func(tx *redis.Tx) error {
		raw, err := tx.HGetAll(ctx, key).Result()
		if err != nil {
			return err
		}
		if len(raw) == 0 {
			return nil
		}
		current, err := parseCodeEntryHash(raw)
		if err != nil {
			return err
		}
		if !current.Claimed || current.ClaimedBy != entry.ClaimedBy || current.ClaimedAt.Unix() != entry.ClaimedAt.Unix() {
			return nil
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, key,
				"claimed", "false",
				"claimed_by", "",
				"claimed_at_unix", "0",
			)
			pipe.ExpireAt(ctx, key, entry.ExpiresAt)
			return nil
		})
		return err
	}, key)
}

func parseCodeEntryHash(raw map[string]string) (mclink.CodeEntry, error) {
	code := strings.ToUpper(strings.TrimSpace(raw["code"]))
	if code == "" {
		return mclink.CodeEntry{}, errors.New("invalid code entry: code is empty")
	}
	expiresUnix, err := strconv.ParseInt(strings.TrimSpace(raw["expires_unix"]), 10, 64)
	if err != nil {
		return mclink.CodeEntry{}, fmt.Errorf("invalid code entry expires_unix: %w", err)
	}
	claimed, err := strconv.ParseBool(strings.TrimSpace(raw["claimed"]))
	if err != nil {
		return mclink.CodeEntry{}, fmt.Errorf("invalid code entry claimed: %w", err)
	}
	claimedAtUnix, err := strconv.ParseInt(strings.TrimSpace(raw["claimed_at_unix"]), 10, 64)
	if err != nil {
		claimedAtUnix = 0
	}
	return mclink.CodeEntry{
		Code:      code,
		Type:      mclink.EntryType(strings.TrimSpace(raw["type"])),
		Value:     strings.TrimSpace(raw["value"]),
		ExpiresAt: time.Unix(expiresUnix, 0).UTC(),
		Claimed:   claimed,
		ClaimedBy: strings.TrimSpace(raw["claimed_by"]),
		ClaimedAt: time.Unix(claimedAtUnix, 0).UTC(),
	}, nil
}
