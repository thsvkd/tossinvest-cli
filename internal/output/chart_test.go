package output

import (
	"strings"
	"testing"
	"time"

	"github.com/junghoonkye/tossinvest-cli/internal/domain"
)

func TestTickSizeKR(t *testing.T) {
	cases := []struct {
		price float64
		want  float64
	}{
		{500, 1},
		{1999, 1},
		{2000, 5},
		{4999, 5},
		{5000, 10},
		{19999, 10},
		{20000, 50},
		{49999, 50},
		{50000, 100},
		{199999, 100},
		{200000, 500},
		{499999, 500},
		{500000, 1000},
		{1_000_000, 1000},
	}
	for _, c := range cases {
		got := tickSizeKR(c.price)
		if got != c.want {
			t.Errorf("tickSizeKR(%v): want %v, got %v", c.price, c.want, got)
		}
	}
}

func TestFormatPriceKR(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{-1000, "-1,000"},
		{1500.4, "1,500"}, // rounds to nearest int
		{1500.6, "1,501"},
	}
	for _, c := range cases {
		got := formatPriceKR(c.in)
		if got != c.want {
			t.Errorf("formatPriceKR(%v): want %q, got %q", c.in, c.want, got)
		}
	}
}

func TestRenderSparklineEmpty(t *testing.T) {
	if got := renderSparkline(nil, 10); got != "" {
		t.Errorf("nil candles: want empty, got %q", got)
	}
	if got := renderSparkline([]domain.Candle{}, 10); got != "" {
		t.Errorf("empty candles: want empty, got %q", got)
	}
}

func TestRenderSparklineContainsBlockGlyphs(t *testing.T) {
	// 8 거의 균등한 step 으로 8 가지 블록 글리프 (▁▂▃▄▅▆▇█) 가 나오게.
	candles := make([]domain.Candle, 8)
	for i := range candles {
		candles[i] = domain.Candle{Close: float64(i)}
	}
	out := renderSparkline(candles, 8)
	for _, g := range []string{"▁", "█"} { // 최소·최대만 확인
		if !strings.Contains(out, g) {
			t.Errorf("expected glyph %q in output %q", g, out)
		}
	}
}

func TestRenderSparklineSamplesToWidth(t *testing.T) {
	// 30개 candle 을 width=10 으로 샘플링하면 10개 글리프.
	candles := make([]domain.Candle, 30)
	for i := range candles {
		candles[i] = domain.Candle{Close: float64(i), Time: time.Unix(int64(i), 0)}
	}
	out := renderSparkline(candles, 10)
	// ANSI escape 제거 후 rune count 가 10.
	stripped := stripANSI(out)
	runes := []rune(stripped)
	if len(runes) != 10 {
		t.Errorf("expected 10 glyphs, got %d (%q)", len(runes), stripped)
	}
}

func stripANSI(s string) string {
	for {
		i := strings.Index(s, "\033[")
		if i < 0 {
			return s
		}
		j := strings.IndexByte(s[i:], 'm')
		if j < 0 {
			return s
		}
		s = s[:i] + s[i+j+1:]
	}
}
