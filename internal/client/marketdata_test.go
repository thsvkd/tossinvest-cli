package client

import "testing"

func TestInvestModeFor(t *testing.T) {
	cases := []struct {
		code             string
		wantView, wantIM string
	}{
		{"A005930", "krx_all", "krx"},   // KR stock
		{"A114800", "krx_all", "krx"},   // KR ETF
		{"US20220809012", "unified", "unified"}, // US stock code
		{"AAPL", "unified", "unified"},  // US ticker
	}
	for _, c := range cases {
		gv, gi := investModeFor(c.code)
		if gv != c.wantView || gi != c.wantIM {
			t.Errorf("investModeFor(%q) = (%q,%q), want (%q,%q)", c.code, gv, gi, c.wantView, c.wantIM)
		}
	}
}
