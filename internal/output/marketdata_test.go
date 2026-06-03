package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/junghoonkye/tossinvest-cli/internal/domain"
)

func TestSessionTime(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "휴장"},
		{"09:00:00.000", "09:00"},
		{"22:30:00.000", "22:30"},
		{"9:0", "9:0"}, // too short, passthrough
	}
	for _, c := range cases {
		if got := sessionTime(c.in); got != c.want {
			t.Errorf("sessionTime(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWriteStockWarningsEmpty(t *testing.T) {
	var buf bytes.Buffer
	sw := domain.StockWarnings{ProductCode: "A005930", Name: "삼성전자"}
	if err := WriteStockWarnings(&buf, FormatTable, sw); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "매수 유의사항 없음") {
		t.Errorf("expected empty-warning notice, got %q", buf.String())
	}
}

func TestWriteTradesCSVHeader(t *testing.T) {
	var buf bytes.Buffer
	list := domain.TradeList{Trades: []domain.Trade{{Time: "09:00:01", Price: 360500, Volume: 10, TradeType: "BUY"}}}
	if err := WriteTrades(&buf, FormatCSV, list); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "time,price,volume,trade_type,cumulative_volume") {
		t.Errorf("unexpected CSV header: %q", out)
	}
	if !strings.Contains(out, "360500") {
		t.Errorf("expected price in CSV, got %q", out)
	}
}
