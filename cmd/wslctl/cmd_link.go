package main

import (
	"fmt"
	"path/filepath"
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

			storePath := filepath.Join(a.baseDir, "runtime", "velocity", ".wslctl", "mclink-codes.tsv")
			store, err := mclink.LoadStore(storePath)
			if err != nil {
				return err
			}

			entryType := mclink.EntryTypeNick
			value := nick
			if uuid != "" {
				entryType = mclink.EntryTypeUUID
				value = uuid
			}

			expiresAt := time.Now().Add(ttl).UTC()
			store.Codes[code] = mclink.CodeEntry{
				Code:      code,
				Type:      entryType,
				Value:     value,
				ExpiresAt: expiresAt,
			}
			if err := mclink.SaveStore(storePath, store); err != nil {
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
