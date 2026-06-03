package main

import (
	"github.com/junghoonkye/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newMarketCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market",
		Short: "Market-wide information (trading hours, etc.)",
	}

	hoursCmd := &cobra.Command{
		Use:   "hours",
		Short: "Today's KR and US trading session windows (장 운영 시간)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			th, err := app.client.GetTradingHours(cmd.Context())
			if err != nil {
				return err
			}
			return output.WriteTradingHours(cmd.OutOrStdout(), app.format, th)
		},
	}

	fxCmd := &cobra.Command{
		Use:   "fx",
		Short: "FX/index quotes (달러 환율·달러 인덱스 등)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			er, err := app.client.GetExchangeRates(cmd.Context())
			if err != nil {
				return err
			}
			return output.WriteExchangeRates(cmd.OutOrStdout(), app.format, er)
		},
	}

	cmd.AddCommand(hoursCmd, fxCmd)
	return cmd
}
