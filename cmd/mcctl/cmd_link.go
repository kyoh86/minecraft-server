package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kyoh86/minecraft-server/internal/mclink"
	"github.com/spf13/cobra"
)

func newLinkCmd(a app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "temporary link-code operations for Discord /mc link",
	}
	cmd.AddCommand(newLinkIssueCmd(a))
	return cmd
}

func newLinkIssueCmd(a app) *cobra.Command {
	var nick string
	var uuid string
	var ttl time.Duration

	cmd := &cobra.Command{
		Use:   "issue",
		Short: "issue one-time link code",
		RunE: func(cmd *cobra.Command, args []string) error {
			nick = strings.TrimSpace(nick)
			uuid = strings.TrimSpace(uuid)
			if nick == "" && uuid == "" {
				return fmt.Errorf("either --nick or --uuid is required")
			}
			if nick != "" && uuid != "" {
				return fmt.Errorf("specify only one of --nick or --uuid")
			}

			code, err := mclink.NewCode(8)
			if err != nil {
				return err
			}

			redisDB, err := strconv.Atoi(strings.TrimSpace(envOr("MCLINK_REDIS_DB", "0")))
			if err != nil {
				return fmt.Errorf("invalid MCLINK_REDIS_DB: %w", err)
			}
			redisCfg := mclink.RedisConfig{
				Addr:     envOr("MCLINK_REDIS_ADDR", "127.0.0.1:16379"),
				Password: envOr("MCLINK_REDIS_PASSWORD", ""),
				DB:       redisDB,
			}
			rdb := mclink.NewRedisClient(redisCfg)
			defer rdb.Close()

			entryType := mclink.EntryTypeNick
			value := nick
			if uuid != "" {
				entryType = mclink.EntryTypeUUID
				value = uuid
			}

			expiresAt := time.Now().Add(ttl).UTC()
			entry := mclink.CodeEntry{
				Code:      code,
				Type:      entryType,
				Value:     value,
				ExpiresAt: expiresAt,
			}
			if err := mclink.SaveCode(context.Background(), rdb, entry); err != nil {
				return err
			}

			fmt.Printf("code=%s type=%s value=%s expires=%s\n", code, entryType, value, expiresAt.Format(time.RFC3339))
			return nil
		},
	}

	cmd.Flags().StringVar(&nick, "nick", "", "Minecraft player name")
	cmd.Flags().StringVar(&uuid, "uuid", "", "Minecraft player UUID")
	cmd.Flags().DurationVar(&ttl, "ttl", 10*time.Minute, "code lifetime")
	return cmd
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
