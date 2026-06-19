package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/junghoonkye/tossinvest-cli/internal/session"
)

// testClientFor wires a client whose info+cert hosts both point at srv.
func testClientFor(srv *httptest.Server) *Client {
	return New(Config{
		HTTPClient:  srv.Client(),
		APIBaseURL:  srv.URL,
		InfoBaseURL: srv.URL,
		CertBaseURL: srv.URL,
		Session:     &session.Session{Cookies: map[string]string{"SESSION": "x"}, Headers: map[string]string{"X-XSRF-TOKEN": "t"}},
	})
}

func TestGetOrderBook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/api/v3/stock-prices/A005930/quotes"):
			w.Write([]byte(`{"result":{"close":358250,"offerPrices":[359000,358500],"offerVolumes":[100,200],"bidPrices":[358000,357500],"bidVolumes":[300,400],"offerVolume":300,"bidVolume":700}}`))
		case strings.Contains(r.URL.Path, "/stock-infos/"):
			w.Write([]byte(`{"result":{"symbol":"005930","name":"삼성전자"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	ob, err := testClientFor(srv).GetOrderBook(context.Background(), "A005930")
	if err != nil {
		t.Fatalf("GetOrderBook error: %v", err)
	}
	if len(ob.Offers) != 2 || len(ob.Bids) != 2 {
		t.Fatalf("expected 2 offers/2 bids, got %d/%d", len(ob.Offers), len(ob.Bids))
	}
	if ob.Offers[0].Price != 359000 || ob.Offers[0].Volume != 100 {
		t.Fatalf("offer[0] mismatch: %+v", ob.Offers[0])
	}
	if ob.Bids[0].Price != 358000 || ob.Bids[0].Volume != 300 {
		t.Fatalf("bid[0] mismatch: %+v", ob.Bids[0])
	}
	if ob.TotalOffer != 300 || ob.TotalBid != 700 {
		t.Fatalf("totals mismatch: offer=%v bid=%v", ob.TotalOffer, ob.TotalBid)
	}
}

func TestGetOrderBookSkipsEmptyLevels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/quotes") {
			w.Write([]byte(`{"result":{"offerPrices":[359000,0],"offerVolumes":[100,0],"bidPrices":[358000],"bidVolumes":[300]}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	ob, err := testClientFor(srv).GetOrderBook(context.Background(), "A005930")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(ob.Offers) != 1 {
		t.Fatalf("expected empty (0/0) level dropped, got %d offers", len(ob.Offers))
	}
}

func TestGetSellableQuantity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/orderable-quantity/sell") {
			w.Write([]byte(`{"result":7}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	sq, err := testClientFor(srv).GetSellableQuantity(context.Background(), "A005930")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if sq.Quantity != 7 {
		t.Fatalf("expected 7, got %v", sq.Quantity)
	}
}

func TestGetCommission(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cost-basis-elements") {
			w.Write([]byte(`{"result":{"commissionRate":0.00015,"taxRate":0.002}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c, err := testClientFor(srv).GetCommission(context.Background(), "A005930")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if c.CommissionRate != 0.00015 || c.TaxRate != 0.002 {
		t.Fatalf("rate mismatch: %+v", c)
	}
}

func TestGetInvestorRankings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/rankings/by-investors") {
			w.Write([]byte(`{"result":{"rankings":{"foreigner":{"type":"외국인","basedAt":"2026-06-19T06:00:00Z","buyStocks":[{"stockCode":"A000660","name":"SK하이닉스","amount":4.8e11,"base":2685000,"close":2789000},{"stockCode":"A005930","name":"삼성전자","amount":1.0e11,"base":360000,"close":362000}]},"institution":{"type":"기관","basedAt":"x","buyStocks":[{"stockCode":"A000660","name":"SK하이닉스","amount":4.1e10,"base":1,"close":2}]},"individual":{"type":"개인","basedAt":"x","buyStocks":[]}}}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	ir, err := testClientFor(srv).GetInvestorRankings(context.Background(), 1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(ir.Rankings) != 3 {
		t.Fatalf("expected 3 investor types, got %d", len(ir.Rankings))
	}
	if ir.Rankings[0].InvestorType != "외국인" || len(ir.Rankings[0].Stocks) != 1 {
		t.Fatalf("order/size mismatch: %+v", ir.Rankings[0])
	}
	if ir.Rankings[0].Stocks[0].Name != "SK하이닉스" || ir.Rankings[0].Stocks[0].Rank != 1 {
		t.Fatalf("stock mismatch: %+v", ir.Rankings[0].Stocks[0])
	}
}

func TestGetEarningCalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/earning-call/upcoming") {
			w.Write([]byte(`{"result":[{"eventId":1,"eventTitle":"26년 6월 어닝콜","status":"UPCOMING","liveAt":"2026-06-23T23:00:00+09:00","companyCode":"X","companyName":"카니발","subContentText":"주요기업"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	ec, err := testClientFor(srv).GetEarningCalls(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(ec.Events) != 1 || ec.Events[0].CompanyName != "카니발" || ec.Events[0].Category != "주요기업" {
		t.Fatalf("earning-call mismatch: %+v", ec.Events)
	}
}
