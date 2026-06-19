package main

import (
	"github.com/junghoonkye/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCommunityCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community",
		Short: "Toss community data (커뮤니티). 공식 API 에 없음",
	}

	var rankType string
	rankingsCmd := &cobra.Command{
		Use:   "rankings",
		Short: "Community leaderboards (인플루언서·수익률·팔로워 급증)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			r, err := app.client.GetCommunityRankings(cmd.Context(), rankType)
			if err != nil {
				return userFacingCommandError(err)
			}
			return output.WriteCommunityRanking(cmd.OutOrStdout(), app.format, r)
		},
	}
	rankingsCmd.Flags().StringVar(&rankType, "type", "influencer", "ranking type: influencer | profit | followers")

	cmd.AddCommand(rankingsCmd)
	return cmd
}
