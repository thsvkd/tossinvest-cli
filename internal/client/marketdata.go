package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/junghoonkye/tossinvest-cli/internal/domain"
)

// investModeFor returns the ticks viewType/investMode pair for a product code.
// KR codes use the KRX combined view; everything else falls back to the US view.
func investModeFor(productCode string) (viewType, investMode string) {
	if deriveSecurityType(productCode) == "kr-s" {
		return "krx_all", "krx"
	}
	return "unified", "unified"
}

type tradeRaw struct {
	Time             string  `json:"time"`
	Price            float64 `json:"price"`
	Base             float64 `json:"base"`
	Volume           float64 `json:"volume"`
	TradeType        string  `json:"tradeType"`
	CumulativeVolume float64 `json:"cumulativeVolume"`
}

// GetTrades returns the most recent executed ticks (체결) for a symbol.
func (c *Client) GetTrades(ctx context.Context, symbol string, count int) (domain.TradeList, error) {
	productCode, err := c.resolveProductCode(ctx, symbol)
	if err != nil {
		return domain.TradeList{}, err
	}
	info, _ := c.getStockInfo(ctx, productCode)

	if count <= 0 {
		count = 30
	}
	viewType, investMode := investModeFor(productCode)

	endpoint, err := url.Parse(fmt.Sprintf("%s/api/v2/stock-prices/%s/ticks", c.infoBaseURL, productCode))
	if err != nil {
		return domain.TradeList{}, err
	}
	q := endpoint.Query()
	q.Set("viewType", viewType)
	q.Set("investMode", investMode)
	q.Set("count", strconv.Itoa(count))
	endpoint.RawQuery = q.Encode()

	var envelope quoteEnvelope[[]tradeRaw]
	if err := c.getJSON(ctx, endpoint.String(), &envelope); err != nil {
		return domain.TradeList{}, err
	}

	out := domain.TradeList{
		ProductCode: productCode,
		Symbol:      info.Symbol,
		Name:        info.Name,
		FetchedAt:   time.Now().UTC(),
	}
	out.Trades = make([]domain.Trade, 0, len(envelope.Result))
	for _, r := range envelope.Result {
		out.Trades = append(out.Trades, domain.Trade{
			Time:             r.Time,
			Price:            r.Price,
			Base:             r.Base,
			Volume:           r.Volume,
			TradeType:        r.TradeType,
			CumulativeVolume: r.CumulativeVolume,
		})
	}
	return out, nil
}

// upper-lower returns a bare object (no BFF envelope).
type upperLowerRaw struct {
	Date       string  `json:"date"`
	UpperLimit float64 `json:"upperLimit"`
	LowerLimit float64 `json:"lowerLimit"`
}

// GetPriceLimits returns the daily upper/lower price band (상/하한가).
func (c *Client) GetPriceLimits(ctx context.Context, symbol string) (domain.PriceLimits, error) {
	productCode, err := c.resolveProductCode(ctx, symbol)
	if err != nil {
		return domain.PriceLimits{}, err
	}
	info, _ := c.getStockInfo(ctx, productCode)

	var envelope quoteEnvelope[upperLowerRaw]
	endpoint := fmt.Sprintf("%s/api/v2/stock-prices/%s/upper-lower", c.infoBaseURL, productCode)
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.PriceLimits{}, err
	}

	return domain.PriceLimits{
		ProductCode: productCode,
		Symbol:      info.Symbol,
		Name:        info.Name,
		Date:        envelope.Result.Date,
		UpperLimit:  envelope.Result.UpperLimit,
		LowerLimit:  envelope.Result.LowerLimit,
	}, nil
}

// GetStockWarnings returns buy-caution badges (매수 유의사항). Badge shape is
// dynamic; recognized fields are mapped and the full object is kept as Raw.
func (c *Client) GetStockWarnings(ctx context.Context, symbol string) (domain.StockWarnings, error) {
	productCode, err := c.resolveProductCode(ctx, symbol)
	if err != nil {
		return domain.StockWarnings{}, err
	}
	info, _ := c.getStockInfo(ctx, productCode)

	var envelope quoteEnvelope[[]json.RawMessage]
	endpoint := fmt.Sprintf("%s/api/v1/stock-infos/%s/wts-badges", c.infoBaseURL, productCode)
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.StockWarnings{}, err
	}

	out := domain.StockWarnings{
		ProductCode: productCode,
		Symbol:      info.Symbol,
		Name:        info.Name,
		FetchedAt:   time.Now().UTC(),
	}
	out.Warnings = make([]domain.StockWarning, 0, len(envelope.Result))
	for _, raw := range envelope.Result {
		var fields struct {
			Type  string `json:"type"`
			Title string `json:"title"`
			Name  string `json:"name"`
			Text  string `json:"text"`
			Desc  string `json:"description"`
			Level string `json:"level"`
		}
		_ = json.Unmarshal(raw, &fields)
		title := fields.Title
		if title == "" {
			title = fields.Name
		}
		text := fields.Text
		if text == "" {
			text = fields.Desc
		}
		out.Warnings = append(out.Warnings, domain.StockWarning{
			Type:  fields.Type,
			Title: title,
			Text:  text,
			Level: fields.Level,
			Raw:   raw,
		})
	}
	return out, nil
}

// trading-hours/integrated returns a bare object keyed by market.
type tradingHoursRaw struct {
	KR struct {
		Today struct {
			Date      string `json:"date"`
			StartTime string `json:"startTime"`
			EndTime   string `json:"endTime"`
		} `json:"today"`
	} `json:"kr"`
	US struct {
		Today struct {
			Date      string `json:"date"`
			StartTime string `json:"startTime"`
			EndTime   string `json:"endTime"`
		} `json:"today"`
	} `json:"us"`
}

// GetTradingHours returns today's KR and US session windows (장 운영 시간).
func (c *Client) GetTradingHours(ctx context.Context) (domain.TradingHours, error) {
	var envelope quoteEnvelope[tradingHoursRaw]
	endpoint := c.apiBaseURL + "/api/v2/system/trading-hours/integrated"
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.TradingHours{}, err
	}
	raw := envelope.Result
	return domain.TradingHours{
		KR: domain.MarketSession{
			Date:      raw.KR.Today.Date,
			StartTime: raw.KR.Today.StartTime,
			EndTime:   raw.KR.Today.EndTime,
		},
		US: domain.MarketSession{
			Date:      raw.US.Today.Date,
			StartTime: raw.US.Today.StartTime,
			EndTime:   raw.US.Today.EndTime,
		},
		FetchedAt: time.Now().UTC(),
	}, nil
}
