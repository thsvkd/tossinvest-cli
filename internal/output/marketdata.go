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
		if err := cw.Write([]string{"market", "session", "date", "start_time", "end_time"}); err != nil {
			return err
		}
		for _, row := range [][]string{
			{"KR", "today", th.KR.Date, th.KR.StartTime, th.KR.EndTime},
			{"US", "today", th.US.Date, th.US.StartTime, th.US.EndTime},
			{"KR", "next", th.NextKR.Date, th.NextKR.StartTime, th.NextKR.EndTime},
			{"US", "next", th.NextUS.Date, th.NextUS.StartTime, th.NextUS.EndTime},
		} {
			if err := cw.Write(row); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"시장", "구분", "일자", "개장", "마감"}
		rows := [][]string{
			{"KR", "오늘", th.KR.Date, sessionTime(th.KR.StartTime), sessionTime(th.KR.EndTime)},
			{"US", "오늘", th.US.Date, sessionTime(th.US.StartTime), sessionTime(th.US.EndTime)},
		}
		// 오늘 휴장 시 다음 영업일도 표시
		if th.KR.StartTime == "" && th.NextKR.Date != "" {
			rows = append(rows, []string{"KR", "다음", th.NextKR.Date, sessionTime(th.NextKR.StartTime), sessionTime(th.NextKR.EndTime)})
		}
		if th.US.StartTime == "" && th.NextUS.Date != "" {
			rows = append(rows, []string{"US", "다음", th.NextUS.Date, sessionTime(th.NextUS.StartTime), sessionTime(th.NextUS.EndTime)})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteExchangeRates(w io.Writer, format Format, er domain.ExchangeRates) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(er)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"code", "name", "base", "close"}); err != nil {
			return err
		}
		for _, r := range er.Rates {
			if err := cw.Write([]string{r.Code, r.Name, formatFloat(r.Base), formatFloat(r.Close)}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"이름", "기준", "현재"}
		rows := make([][]string, 0, len(er.Rates))
		for _, r := range er.Rates {
			rows = append(rows, []string{r.Name, formatFloat(r.Base), formatFloat(r.Close)})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteScreenerPresets(w io.Writer, format Format, sp domain.ScreenerPresets) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sp)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"id", "name", "description"}); err != nil {
			return err
		}
		for _, p := range sp.Presets {
			if err := cw.Write([]string{p.ID, p.Name, p.Description}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"ID", "이름", "설명"}
		rows := make([][]string, 0, len(sp.Presets))
		for _, p := range sp.Presets {
			rows = append(rows, []string{p.ID, p.Name, p.Description})
		}
		if err := renderTable(w, headers, rows); err != nil {
			return err
		}
		_, err := fmt.Fprintln(w, "\n실행: tossctl market screener <ID> [--nation kr|us]")
		return err
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteScreenerResult(w io.Writer, format Format, sr domain.ScreenerResult) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sr)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"product_code", "name", "close", "change", "change_rate"}); err != nil {
			return err
		}
		for _, s := range sr.Stocks {
			if err := cw.Write([]string{
				s.ProductCode, s.Name, formatFloat(s.Close), formatFloat(s.Change), formatFloat(s.ChangeRate),
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		if _, err := fmt.Fprintf(w, "%s (%s) — %d종목 중 상위 %d\n",
			sr.PresetName, sr.Nation, sr.TotalCount, len(sr.Stocks)); err != nil {
			return err
		}
		headers := []string{"종목", "이름", "현재가", "변동률"}
		rows := make([][]string, 0, len(sr.Stocks))
		for _, s := range sr.Stocks {
			rows = append(rows, []string{s.ProductCode, s.Name, formatFloat(s.Close), formatPct(s.ChangeRate)})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteAISignals(w io.Writer, format Format, sg domain.AISignals) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sg)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"asset_name", "title", "keyword", "fluctuation", "stock_code"}); err != nil {
			return err
		}
		for _, s := range sg.Signals {
			if err := cw.Write([]string{s.AssetName, s.Title, s.Keyword, s.Fluctuation, s.StockCode}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		label := sg.Label
		if label == "" {
			label = "AI 시그널"
		}
		if _, err := fmt.Fprintf(w, "%s\n", label); err != nil {
			return err
		}
		headers := []string{"종목", "시그널", "키워드", "등락"}
		rows := make([][]string, 0, len(sg.Signals))
		for _, s := range sg.Signals {
			rows = append(rows, []string{s.AssetName, s.Title, s.Keyword, s.Fluctuation})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteTradingFlows(w io.Writer, format Format, tf domain.TradingFlows) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(tf)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"date", "net_individuals", "net_foreigner", "net_institution"}); err != nil {
			return err
		}
		for _, f := range tf.Flows {
			if err := cw.Write([]string{
				f.Date, formatFloat(f.NetIndividuals), formatFloat(f.NetForeigner), formatFloat(f.NetInstitution),
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		name := tf.Name
		if name == "" {
			name = tf.Symbol
		}
		if _, err := fmt.Fprintf(w, "%s (%s) 수급 — 순매수(주)\n", name, tf.ProductCode); err != nil {
			return err
		}
		headers := []string{"일자", "개인", "외국인", "기관"}
		rows := make([][]string, 0, len(tf.Flows))
		for _, f := range tf.Flows {
			rows = append(rows, []string{
				f.Date, signedInt(f.NetIndividuals), signedInt(f.NetForeigner), signedInt(f.NetInstitution),
			})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// signedInt formats a net volume with explicit +/- sign and thousands commas.
func signedInt(v float64) string {
	s := formatKRW(v)
	if v > 0 {
		return "+" + s
	}
	return s
}

func WriteMarketIndices(w io.Writer, format Format, mi domain.MarketIndices) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(mi)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"name", "nation", "latest", "base", "change", "change_rate"}); err != nil {
			return err
		}
		for _, x := range mi.Indices {
			if err := cw.Write([]string{
				x.Name, x.Nation, formatFloat(x.Latest), formatFloat(x.Base),
				formatFloat(x.Change), formatFloat(x.ChangeRate),
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"지수", "현재", "변동", "변동률"}
		rows := make([][]string, 0, len(mi.Indices))
		for _, x := range mi.Indices {
			sign := ""
			if x.Change > 0 {
				sign = "+"
			}
			rows = append(rows, []string{
				x.Name,
				fmt.Sprintf("%.2f", x.Latest),
				sign + fmt.Sprintf("%.2f", x.Change),
				formatPct(x.ChangeRate),
			})
		}
		return renderTable(w, headers, rows)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func WriteStockRanking(w io.Writer, format Format, sr domain.StockRanking) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(sr)
	case FormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"rank", "symbol", "name", "market", "product_code"}); err != nil {
			return err
		}
		for _, x := range sr.Stocks {
			if err := cw.Write([]string{
				fmt.Sprintf("%d", x.Rank), x.Symbol, x.Name, x.Market, x.ProductCode,
			}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case FormatTable:
		headers := []string{"순위", "종목", "이름", "시장"}
		rows := make([][]string, 0, len(sr.Stocks))
		for _, x := range sr.Stocks {
			rows = append(rows, []string{fmt.Sprintf("%d", x.Rank), x.Symbol, x.Name, x.Market})
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
