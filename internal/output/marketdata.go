package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/junghoonkye/tossinvest-cli/internal/domain"
)

func WriteTrades(w io.Writer, format Format, list domain.TradeList) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(list)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"time", "price", "volume", "trade_type", "cumulative_volume"}); err != nil {
			return err
		}
		for _, t := range list.Trades {
			if err := cw.Write([]string{
				t.Time, formatFloat(t.Price), formatFloat(t.Volume),
				t.TradeType, formatFloat(t.CumulativeVolume),
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"시각", "체결가", "수량", "구분"}
		rows := make([][]string, 0, len(list.Trades))
		for _, t := range list.Trades {
			side := t.TradeType
			switch t.TradeType {
			case "BUY":
				side = "매수"
			case "SELL":
				side = "매도"
			}
			rows = append(rows, []string{t.Time, formatKRW(t.Price), formatFloat(t.Volume), side})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WritePriceLimits(w io.Writer, format Format, pl domain.PriceLimits) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(pl)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"date", "upper_limit", "lower_limit"}); err != nil {
			return err
		}
		if err := cw.Write([]string{pl.Date, formatFloat(pl.UpperLimit), formatFloat(pl.LowerLimit)}); err != nil {
			return err
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		name := pl.Name
		if name == "" {
			name = pl.Symbol
		}
		_, err := fmt.Fprintf(w,
			"%s (%s) · %s\n상한가: %s\n하한가: %s\n",
			name, pl.ProductCode, pl.Date,
			formatKRW(pl.UpperLimit), formatKRW(pl.LowerLimit),
		)
		return err
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteStockWarnings(w io.Writer, format Format, sw domain.StockWarnings) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sw)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"type", "title", "text", "level"}); err != nil {
			return err
		}
		for _, x := range sw.Warnings {
			if err := cw.Write([]string{x.Type, x.Title, x.Text, x.Level}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		name := sw.Name
		if name == "" {
			name = sw.Symbol
		}
		if len(sw.Warnings) == 0 {
			_, err := fmt.Fprintf(w, "%s (%s): 매수 유의사항 없음\n", name, sw.ProductCode)
			return err
		}
		if _, err := fmt.Fprintf(w, "%s (%s) 매수 유의사항 %d건\n", name, sw.ProductCode, len(sw.Warnings)); err != nil {
			return err
		}
		for _, x := range sw.Warnings {
			label := x.Title
			if label == "" {
				label = x.Type
			}
			line := "• " + label
			if x.Text != "" {
				line += " — " + x.Text
			}
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteTradingHours(w io.Writer, format Format, th domain.TradingHours) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(th)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"market", "date", "start_time", "end_time"}); err != nil {
			return err
		}
		for _, row := range [][]string{
			{"KR", th.KR.Date, th.KR.StartTime, th.KR.EndTime},
			{"US", th.US.Date, th.US.StartTime, th.US.EndTime},
		} {
			if err := cw.Write(row); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"시장", "일자", "개장", "마감"}
		rows := [][]string{
			{"KR", th.KR.Date, sessionTime(th.KR.StartTime), sessionTime(th.KR.EndTime)},
			{"US", th.US.Date, sessionTime(th.US.StartTime), sessionTime(th.US.EndTime)},
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// sessionTime trims the "HH:MM:SS.mmm" wire format down to "HH:MM", and shows a
// dash for a closed/holiday session (null → empty string).
func sessionTime(s string) string {
	if s == "" {
		return "휴장"
	}
	if len(s) >= 5 {
		return s[:5]
	}
	return s
}
