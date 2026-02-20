package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kyoh86/minecraft-server/internal/mclink"
)

type config struct {
	TokenPath     string
	GuildID       string
	StorePath     string
	WhitelistPath string
	ConsoleFIFO   string
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	tokenBytes, err := os.ReadFile(cfg.TokenPath)
	if err != nil {
		log.Fatal(err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		log.Fatal("bot token is empty")
	}

	dg, err := discordgo.New("Bot " + token)
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
		TokenPath:     env("MCLINK_DISCORD_BOT_TOKEN_FILE", "/run/secrets/mclink_discord_bot_token"),
		GuildID:       strings.TrimSpace(os.Getenv("MCLINK_DISCORD_GUILD_ID")),
		StorePath:     env("MCLINK_STORE_PATH", "/data/mclink/codes.json"),
		WhitelistPath: env("MCLINK_WHITELIST_PATH", "/data/velocity/whitelists/default.toml"),
		ConsoleFIFO:   env("MCLINK_CONSOLE_FIFO_PATH", "/data/velocity/.wslctl/velocity-console-in"),
	}
	if cfg.GuildID == "" {
		return config{}, errors.New("MCLINK_DISCORD_GUILD_ID is required")
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

	store, err := mclink.LoadStore(cfg.StorePath)
	if err != nil {
		respond(s, i, "内部エラー: code storage を読めませんでした。")
		return
	}
	entry, ok := store.Codes[code]
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

	if err := mclink.AddWhitelistEntry(cfg.WhitelistPath, entry.Type, entry.Value); err != nil {
		respond(s, i, "内部エラー: whitelist 更新に失敗しました。")
		return
	}
	if err := sendConsoleCommand(cfg.ConsoleFIFO, "whitelist reload"); err != nil {
		respond(s, i, "内部エラー: whitelist reload に失敗しました。")
		return
	}

	entry.Claimed = true
	entry.ClaimedBy = i.Member.User.ID
	entry.ClaimedAt = time.Now().UTC()
	store.Codes[code] = entry
	if err := mclink.SaveStore(cfg.StorePath, store); err != nil {
		respond(s, i, "内部エラー: code の確定保存に失敗しました。")
		return
	}

	msg := fmt.Sprintf("リンク完了: `%s:%s` を whitelist に追加し、`whitelist reload` を実行しました。", entry.Type, entry.Value)
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

func env(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func sendConsoleCommand(path, cmd string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\n", strings.TrimSpace(cmd))
	return err
}
