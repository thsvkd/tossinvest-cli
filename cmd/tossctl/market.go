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

	var investorsSize int
	investorsCmd := &cobra.Command{
		Use:   "investors",
		Short: "Net-buy ranking by investor type (외국인·기관·개인 순매수 상위)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			ir, err := app.client.GetInvestorRankings(cmd.Context(), investorsSize)
			if err != nil {
				return err
			}
			return output.WriteInvestorRankings(cmd.OutOrStdout(), app.format, ir)
		},
	}
	investorsCmd.Flags().IntVar(&investorsSize, "size", 10, "top stocks per investor type")

	earningsCmd := &cobra.Command{
		Use:   "earnings",
		Short: "Upcoming earnings-call calendar (어닝콜 일정)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			ec, err := app.client.GetEarningCalls(cmd.Context())
			if err != nil {
				return err
			}
			return output.WriteEarningCalls(cmd.OutOrStdout(), app.format, ec)
		},
	}

	var (
		screenerNation string
		screenerSize   int
		screenerFilter string
	)
	screenerCmd := &cobra.Command{
		Use:   "screener [preset-id]",
		Short: "Stock screeners (조건검색: 가치주·배당주·성장주 등). 인자 없으면 프리셋 목록",
		Long: `조건 검색. 인자 없으면 프리셋 목록, preset-id 주면 해당 조건 종목 반환.

커스텀 조건은 --filter 로 raw JSON 배열을 직접 전달 (토스 web 의 필터 스키마):
  tossctl market screener --filter '[{"id":"시가총액","conditions":[...]}]' --nation kr

필터 ID/조건 구조는 프리셋 출력(--output json)을 참고해 변형하면 됩니다.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}
			// --filter (custom raw) 우선
			if screenerFilter != "" {
				res, err := app.client.RunScreenerRaw(cmd.Context(), screenerFilter, screenerNation, screenerSize)
				if err != nil {
					return err
				}
				return output.WriteScreenerResult(cmd.OutOrStdout(), app.format, res)
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
	screenerCmd.Flags().StringVar(&screenerFilter, "filter", "", "custom raw filter JSON array (preset 대신)")

	cmd.AddCommand(hoursCmd, fxCmd, indexCmd, rankingCmd, signalsCmd, investorsCmd, earningsCmd, screenerCmd)
	return cmd
}
