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

	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Major market indices (코스피·코스닥·나스닥·S&P500·VIX 등)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			mi, err := app.client.GetMarketIndices(cmd.Context())
			if err != nil {
				return err
			}
			return output.WriteMarketIndices(cmd.OutOrStdout(), app.format, mi)
		},
	}

	var rankingSize int
	rankingCmd := &cobra.Command{
		Use:   "ranking",
		Short: "Realtime popularity ranking (실시간 인기 종목)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			sr, err := app.client.GetStockRanking(cmd.Context(), rankingSize)
			if err != nil {
				return err
			}
			return output.WriteStockRanking(cmd.OutOrStdout(), app.format, sr)
		},
	}
	rankingCmd.Flags().IntVar(&rankingSize, "size", 20, "number of ranked stocks")

	signalsCmd := &cobra.Command{
		Use:   "signals",
		Short: "Toss AI market signals (토스증권 AI 시그널)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			sg, err := app.client.GetAISignals(cmd.Context())
			if err != nil {
				return err
			}
			return output.WriteAISignals(cmd.OutOrStdout(), app.format, sg)
		},
	}

	var (
		screenerNation string
		screenerSize   int
	)
	screenerCmd := &cobra.Command{
		Use:   "screener [preset-id]",
		Short: "Stock screeners (조건검색: 가치주·배당주·성장주 등). 인자 없으면 프리셋 목록",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				presets, err := app.client.GetScreenerPresets(cmd.Context())
				if err != nil {
					return err
				}
				return output.WriteScreenerPresets(cmd.OutOrStdout(), app.format, presets)
			}
			res, err := app.client.RunScreener(cmd.Context(), args[0], screenerNation, screenerSize)
			if err != nil {
				return err
			}
			return output.WriteScreenerResult(cmd.OutOrStdout(), app.format, res)
		},
	}
	screenerCmd.Flags().StringVar(&screenerNation, "nation", "kr", "market: kr | us")
	screenerCmd.Flags().IntVar(&screenerSize, "size", 30, "max stocks to return")

	cmd.AddCommand(hoursCmd, fxCmd, indexCmd, rankingCmd, signalsCmd, screenerCmd)
	return cmd
}
