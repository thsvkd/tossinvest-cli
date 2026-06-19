package main

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newPortfolioCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "portfolio",
		Short: "Read portfolio and holdings data",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "positions",
			Short: "List current positions",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}

				positions, err := app.client.ListPositions(cmd.Context())
				if err != nil {
					return userFacingCommandError(err)
				}

				return output.WritePositions(cmd.OutOrStdout(), app.format, positions)
			},
		},
		&cobra.Command{
			Use:   "allocation",
			Short: "Show portfolio allocation",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := newAppContext(opts)
				if err != nil {
					return err
				}

				summary, err := app.client.GetAccountSummary(cmd.Context())
				if err != nil {
					return userFacingCommandError(err)
				}

				return output.WriteAllocation(cmd.OutOrStdout(), app.format, summary.Markets)
			},
		},
		newDividendsCmd(opts),
	)

	return cmd
}

func newDividendsCmd(opts *rootOptions) *cobra.Command {
	var (
		year          int
		byPaymentDate bool
	)
	cmd := &cobra.Command{
		Use:   "dividends",
		Short: "Annual dividend report (연간 배당 내역 — 총액·지역·월별). 공식 API 에 없음",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			d, err := app.client.GetDividends(cmd.Context(), year, byPaymentDate)
			if err != nil {
				return userFacingCommandError(err)
			}
			return output.WriteDividends(cmd.OutOrStdout(), app.format, d)
		},
	}
	cmd.Flags().IntVar(&year, "year", 0, "조회 연도 (기본: 올해)")
	cmd.Flags().BoolVar(&byPaymentDate, "by-payment-date", false, "지급일 기준 (세금·수수료 포함)")
	return cmd
}
