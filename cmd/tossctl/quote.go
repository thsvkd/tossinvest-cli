package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/junghoonkye/tossinvest-cli/internal/domain"
	"github.com/junghoonkye/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

func newQuoteCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quote",
		Short: "Read quote data",
	}

	getCmd := &cobra.Command{
		Use:   "get <symbol or name>",
		Short: "Fetch quote data for a symbol or stock name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}

			symbol := strings.Join(args, " ")
			quote, err := app.client.GetQuote(cmd.Context(), symbol)
			if err != nil {
				return err
			}

			return output.WriteQuote(cmd.OutOrStdout(), app.format, quote)
		},
	}

	var (
		batchChart    bool
		batchLive     bool
		batchInterval int
	)
	batchCmd := &cobra.Command{
		Use:   "batch <symbol>[,symbol,...] [...]",
		Short: "Fetch quotes for multiple symbols at once (comma-separated supported)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}

			symbols := parseBatchSymbols(args)

			fetchAndRender := func(ctx context.Context, w io.Writer) error {
				var quotes []domain.Quote
				for _, symbol := range symbols {
					quote, err := app.client.GetQuote(ctx, symbol)
					if err != nil {
						return err
					}
					quotes = append(quotes, quote)
				}

				if !batchChart {
					return output.WriteQuotes(w, app.format, quotes)
				}

				warnW := io.Writer(cmd.ErrOrStderr())
				if batchLive {
					warnW = w
				}
				var charts []domain.Chart
				for _, q := range quotes {
					chart, err := app.client.GetChart(ctx, q.ProductCode, "3m", 30)
					if err != nil {
						fmt.Fprintf(warnW, "warning: chart unavailable for %s: %v\n", q.Symbol, err)
						charts = append(charts, domain.Chart{})
						continue
					}
					charts = append(charts, chart)
				}
				return output.WriteQuotesWithCharts(w, app.format, quotes, charts)
			}

			if !batchLive {
				return fetchAndRender(cmd.Context(), cmd.OutOrStdout())
			}

			if !isTerminal(cmd.OutOrStdout()) {
				return fmt.Errorf("--live requires an interactive terminal")
			}

			if batchInterval < 1 {
				batchInterval = 1
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			w := cmd.OutOrStdout()
			fmt.Fprint(w, "\033[?1049h\033[?25l")
			defer fmt.Fprint(w, "\033[?25h\033[?1049l")

			fmt.Fprintf(w, "\033[H\033[JEvery %ds | Fetching...\n", batchInterval)

			var lastGood string
			interval := time.Duration(batchInterval) * time.Second
			for {
				var buf strings.Builder
				fmt.Fprintf(&buf, "Every %ds | %s\n\n",
					batchInterval, time.Now().Local().Format("2006-01-02 15:04:05"))

				if err := fetchAndRender(ctx, &buf); err != nil {
					if ctx.Err() != nil {
						return nil
					}
					fmt.Fprintf(&buf, "%s\n\033[31merror: %v\033[0m\n", lastGood, err)
				} else {
					lastGood = buf.String()
				}

				fmt.Fprint(w, "\033[H"+buf.String()+"\033[J")

				select {
				case <-ctx.Done():
					return nil
				case <-time.After(interval):
				}
			}
		},
	}
	batchCmd.Flags().BoolVar(&batchChart, "chart", false, "show sparkline chart for each symbol")
	batchCmd.Flags().BoolVar(&batchLive, "live", false, "continuously refresh (like watch/viddy)")
	batchCmd.Flags().IntVar(&batchInterval, "interval", 2, "refresh interval in seconds (used with --live)")

	var (
		chartInterval string
		chartCount    int
	)
	chartCmd := &cobra.Command{
		Use:   "chart <symbol or name>",
		Short: "Fetch candle chart for a symbol or stock name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := newAppContext(opts)
			if err != nil {
				return err
			}

			symbol := strings.Join(args, " ")
			chart, err := app.client.GetChart(cmd.Context(), symbol, chartInterval, chartCount)
			if err != nil {
				return err
			}

			return output.WriteChart(cmd.OutOrStdout(), app.format, chart)
		},
	}
	chartCmd.Flags().StringVar(&chartInterval, "interval", "3m", "candle interval: 1m, 3m, 5m, 10m, 15m, 30m, 60m")
	chartCmd.Flags().IntVar(&chartCount, "count", 30, "number of candles to fetch")

	cmd.AddCommand(getCmd, batchCmd, chartCmd)

	return cmd
}

func parseBatchSymbols(args []string) []string {
	var symbols []string
	for _, arg := range args {
		for _, s := range strings.Split(arg, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				symbols = append(symbols, s)
			}
		}
	}
	return symbols
}
