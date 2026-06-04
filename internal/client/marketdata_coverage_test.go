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
