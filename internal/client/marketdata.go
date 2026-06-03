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

// GetPriceLimits returns the daily upper/lower price band (상/하한가). This is a
// KRX-specific concept; US markets use intraday circuit breakers (LULD) instead
// of a fixed daily band, so non-KR symbols are rejected with a clear message.
func (c *Client) GetPriceLimits(ctx context.Context, symbol string) (domain.PriceLimits, error) {
	productCode, err := c.resolveProductCode(ctx, symbol)
	if err != nil {
		return domain.PriceLimits{}, err
	}
	if deriveSecurityType(productCode) != "kr-s" {
		return domain.PriceLimits{}, fmt.Errorf("상/하한가는 국내(KRX) 종목만 제공됩니다 (미국장은 일일 가격제한 제도가 없음): %s", symbol)
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

// trading-hours/integrated session block.
type tradingSessionRaw struct {
	Date      string `json:"date"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type tradingMarketRaw struct {
	Today      tradingSessionRaw `json:"today"`
	NextBizDay tradingSessionRaw `json:"nextBizDay"`
}

type tradingHoursRaw struct {
	KR tradingMarketRaw `json:"kr"`
	US tradingMarketRaw `json:"us"`
}

type exchangeRatesRaw struct {
	ExchangeRates []struct {
		Code  string  `json:"code"`
		Name  string  `json:"name"`
		Base  float64 `json:"base"`
		Close float64 `json:"close"`
	} `json:"exchangeRates"`
}

// GetExchangeRates returns FX/index quotes (USD/KRW, DXY 등).
func (c *Client) GetExchangeRates(ctx context.Context) (domain.ExchangeRates, error) {
	var envelope quoteEnvelope[exchangeRatesRaw]
	endpoint := c.infoBaseURL + "/api/v1/dashboard/wts/overview/exchange-rates"
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.ExchangeRates{}, err
	}
	out := domain.ExchangeRates{FetchedAt: time.Now().UTC()}
	out.Rates = make([]domain.ExchangeRate, 0, len(envelope.Result.ExchangeRates))
	for _, r := range envelope.Result.ExchangeRates {
		out.Rates = append(out.Rates, domain.ExchangeRate{Code: r.Code, Name: r.Name, Base: r.Base, Close: r.Close})
	}
	return out, nil
}

type tradingTrendRaw struct {
	Body []struct {
		BaseDate                string  `json:"baseDate"`
		NetIndividualsBuyVolume float64 `json:"netIndividualsBuyVolume"`
		NetForeignerBuyVolume   float64 `json:"netForeignerBuyVolume"`
		NetInstitutionBuyVolume float64 `json:"netInstitutionBuyVolume"`
	} `json:"body"`
}

// GetTradingFlows returns investor-type net flows (수급 — 개인·외국인·기관 순매수).
// KRX 전용 (해외 종목은 데이터 없음). 공식 API 에 없는 web 전용 표면.
func (c *Client) GetTradingFlows(ctx context.Context, symbol string, size int) (domain.TradingFlows, error) {
	productCode, err := c.resolveProductCode(ctx, symbol)
	if err != nil {
		return domain.TradingFlows{}, err
	}
	if deriveSecurityType(productCode) != "kr-s" {
		return domain.TradingFlows{}, fmt.Errorf("수급(투자자별 순매수)은 국내(KRX) 종목만 제공됩니다: %s", symbol)
	}
	info, _ := c.getStockInfo(ctx, productCode)
	if size <= 0 {
		size = 20
	}

	endpoint, err := url.Parse(c.infoBaseURL + "/api/v1/stock-infos/trade/trend/trading-trend")
	if err != nil {
		return domain.TradingFlows{}, err
	}
	q := endpoint.Query()
	q.Set("productCode", productCode)
	q.Set("size", strconv.Itoa(size))
	endpoint.RawQuery = q.Encode()

	var envelope quoteEnvelope[tradingTrendRaw]
	if err := c.getJSON(ctx, endpoint.String(), &envelope); err != nil {
		return domain.TradingFlows{}, err
	}
	out := domain.TradingFlows{
		ProductCode: productCode,
		Symbol:      info.Symbol,
		Name:        info.Name,
		FetchedAt:   time.Now().UTC(),
	}
	out.Flows = make([]domain.TradingFlow, 0, len(envelope.Result.Body))
	for _, r := range envelope.Result.Body {
		out.Flows = append(out.Flows, domain.TradingFlow{
			Date:           r.BaseDate,
			NetIndividuals: r.NetIndividualsBuyVolume,
			NetForeigner:   r.NetForeignerBuyVolume,
			NetInstitution: r.NetInstitutionBuyVolume,
		})
	}
	return out, nil
}

type marketIndexRaw struct {
	MajorIndicatorInfos []struct {
		DisplayName string `json:"displayName"`
		Nation      string `json:"nation"`
		Price       struct {
			LatestPrice float64 `json:"latestPrice"`
			BasePrice   float64 `json:"basePrice"`
		} `json:"price"`
	} `json:"majorIndicatorInfos"`
}

// GetMarketIndices returns major market index quotes (코스피·나스닥·VIX 등).
// 공식 API 에 없는 web 전용 표면.
func (c *Client) GetMarketIndices(ctx context.Context) (domain.MarketIndices, error) {
	var envelope quoteEnvelope[marketIndexRaw]
	endpoint := c.certBaseURL + "/api/v1/dashboard/wts/overview/indicator/index"
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.MarketIndices{}, err
	}
	out := domain.MarketIndices{FetchedAt: time.Now().UTC()}
	out.Indices = make([]domain.MarketIndex, 0, len(envelope.Result.MajorIndicatorInfos))
	for _, r := range envelope.Result.MajorIndicatorInfos {
		idx := domain.MarketIndex{
			Name:   r.DisplayName,
			Nation: r.Nation,
			Latest: r.Price.LatestPrice,
			Base:   r.Price.BasePrice,
		}
		idx.Change = idx.Latest - idx.Base
		if idx.Base != 0 {
			idx.ChangeRate = idx.Change / idx.Base
		}
		out.Indices = append(out.Indices, idx)
	}
	return out, nil
}

type rankingRaw struct {
	Data []struct {
		Code   string `json:"code"`
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
		Market struct {
			DisplayName string `json:"displayName"`
		} `json:"market"`
	} `json:"data"`
}

// GetStockRanking returns the realtime popularity ranking (실시간 인기 순위).
// 공식 API 에 없는 discovery 표면.
func (c *Client) GetStockRanking(ctx context.Context, size int) (domain.StockRanking, error) {
	if size <= 0 {
		size = 20
	}
	endpoint, err := url.Parse(c.infoBaseURL + "/api/v1/rankings/realtime/stock")
	if err != nil {
		return domain.StockRanking{}, err
	}
	q := endpoint.Query()
	q.Set("size", strconv.Itoa(size))
	endpoint.RawQuery = q.Encode()

	var envelope quoteEnvelope[rankingRaw]
	if err := c.getJSON(ctx, endpoint.String(), &envelope); err != nil {
		return domain.StockRanking{}, err
	}
	out := domain.StockRanking{FetchedAt: time.Now().UTC()}
	out.Stocks = make([]domain.RankedStock, 0, len(envelope.Result.Data))
	for i, r := range envelope.Result.Data {
		out.Stocks = append(out.Stocks, domain.RankedStock{
			Rank:        i + 1,
			ProductCode: r.Code,
			Symbol:      r.Symbol,
			Name:        r.Name,
			Market:      r.Market.DisplayName,
		})
	}
	return out, nil
}

// GetTradingHours returns today's KR and US session windows (장 운영 시간).
func (c *Client) GetTradingHours(ctx context.Context) (domain.TradingHours, error) {
	var envelope quoteEnvelope[tradingHoursRaw]
	endpoint := c.apiBaseURL + "/api/v2/system/trading-hours/integrated"
	if err := c.getJSON(ctx, endpoint, &envelope); err != nil {
		return domain.TradingHours{}, err
	}
	raw := envelope.Result
	sess := func(s tradingSessionRaw) domain.MarketSession {
		return domain.MarketSession{Date: s.Date, StartTime: s.StartTime, EndTime: s.EndTime}
	}
	return domain.TradingHours{
		KR:        sess(raw.KR.Today),
		US:        sess(raw.US.Today),
		NextKR:    sess(raw.KR.NextBizDay),
		NextUS:    sess(raw.US.NextBizDay),
		FetchedAt: time.Now().UTC(),
	}, nil
}
