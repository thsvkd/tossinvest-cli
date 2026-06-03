package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/junghoonkye/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newWatchlistCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watchlist",
		Short: "Read and manage watchlist (관심종목 조회·관리)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List watchlist entries",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				items, err := app.client.ListWatchlist(cmd.Context())
				if err != nil {
					return userFacingCommandError(err)
				}
				return output.WriteWatchlist(cmd.OutOrStdout(), app.format, items)
			},
		},
		&cobra.Command{
			Use:   "groups",
			Short: "List watchlist folders (관심종목 폴더)",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				groups, err := app.client.ListWatchlistGroups(cmd.Context())
				if err != nil {
					return userFacingCommandError(err)
				}
				return output.WriteWatchlistGroups(cmd.OutOrStdout(), app.format, groups)
			},
		},
		newWatchlistGroupCmd(opts),
		newWatchlistAddRemoveCmd(opts, "add", "관심종목에 종목 추가"),
		newWatchlistAddRemoveCmd(opts, "remove", "관심종목에서 종목 제거"),
	)

	return cmd
}

func newWatchlistGroupCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage watchlist folders (폴더 생성·이름변경·삭제)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "create <name>",
			Short: "Create a watchlist folder",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				g, err := app.client.CreateWatchlistGroup(cmd.Context(), strings.Join(args, " "))
				if err != nil {
					return userFacingCommandError(err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "폴더 생성: %s (id=%d)\n", g.Name, g.ID)
				return nil
			},
		},
		&cobra.Command{
			Use:   "rename <id> <new-name>",
			Short: "Rename a watchlist folder",
			Args:  cobra.MinimumNArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				id, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("폴더 id 는 숫자여야 합니다: %s", args[0])
				}
				name := strings.Join(args[1:], " ")
				if err := app.client.RenameWatchlistGroup(cmd.Context(), id, name); err != nil {
					return userFacingCommandError(err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "폴더 이름 변경: id=%d → %s\n", id, name)
				return nil
			},
		},
		&cobra.Command{
			Use:   "delete <id>",
			Short: "Delete a watchlist folder",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}
				id, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("폴더 id 는 숫자여야 합니다: %s", args[0])
				}
				if err := app.client.DeleteWatchlistGroup(cmd.Context(), id); err != nil {
					return userFacingCommandError(err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "폴더 삭제: id=%d\n", id)
				return nil
			},
		},
	)
	return cmd
}

func newWatchlistAddRemoveCmd(opts *rootOptions, verb, short string) *cobra.Command {
	var groupID int64
	c := &cobra.Command{
		Use:   verb + " <symbol or name>",
		Short: short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			if groupID == 0 {
				return fmt.Errorf("--group <폴더id> 필요 (`watchlist groups` 로 확인)")
			}
			symbol := strings.Join(args, " ")
			if verb == "add" {
				err = app.client.AddWatchlistItem(cmd.Context(), groupID, symbol)
			} else {
				err = app.client.RemoveWatchlistItem(cmd.Context(), groupID, symbol)
			}
			if err != nil {
				return userFacingCommandError(err)
			}
			action := "추가"
			if verb == "remove" {
				action = "제거"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "관심종목 %s: %s (폴더 id=%d)\n", action, symbol, groupID)
			return nil
		},
	}
	c.Flags().Int64Var(&groupID, "group", 0, "대상 폴더 id (watchlist groups 로 확인)")
	return c
}
