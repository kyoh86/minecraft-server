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
		AllowlistPath: env("MC_LINK_ALLOWLIST_PATH", "/data/velocity/allowlist.yml"),
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
	code := strings.ToUpper(strings.TrimSpace(sub[0].StringValue()))
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

	entry, ok, err := mclink.LoadCode(ctx, rdb, code)
	if err != nil {
		respond(s, i, "内部エラー: code storage を読めませんでした。")
		return
	}
	if !ok {
		respond(s, i, "無効なコードです。")
		return
	}
	if entry.Claimed {
		respond(s, i, "このコードは既に使用済みです。")
		return
	}
	if time.Now().UTC().After(entry.ExpiresAt) {
		respond(s, i, "コードの有効期限が切れています。")
		return
	}

	if err := mclink.AddAllowlistEntry(cfg.AllowlistPath, entry.Type, entry.Value); err != nil {
		respond(s, i, "内部エラー: allowlist 更新に失敗しました。")
		return
	}

	entry.Claimed = true
	entry.ClaimedBy = interactionUserID(i)
	entry.ClaimedAt = time.Now().UTC()
	if err := saveClaimed(ctx, rdb, entry); err != nil {
		respond(s, i, "内部エラー: code の確定保存に失敗しました。")
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

func saveClaimed(ctx context.Context, rdb *redis.Client, entry mclink.CodeEntry) error {
	ttl := time.Until(entry.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Minute
	}
	pipe := rdb.TxPipeline()
	key := "mc-link:code:" + entry.Code
	pipe.HSet(ctx, key,
		"code", entry.Code,
		"type", string(entry.Type),
		"value", entry.Value,
		"expires_unix", entry.ExpiresAt.Unix(),
		"claimed", "true",
		"claimed_by", entry.ClaimedBy,
		"claimed_at_unix", entry.ClaimedAt.Unix(),
	)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}
