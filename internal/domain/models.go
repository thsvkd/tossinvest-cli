package domain

import (
	"encoding/json"
	"time"
)

type Account struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name"`
	Name        string   `json:"name,omitempty"`
	Type        string   `json:"type,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	Markets     []string `json:"markets,omitempty"`
	Primary     bool     `json:"primary,omitempty"`
}

type AccountSummary struct {
	TotalAssetAmount      float64                         `json:"total_asset_amount"`
	EvaluatedProfitAmount float64                         `json:"evaluated_profit_amount"`
	ProfitRate            float64                         `json:"profit_rate"`
	OrderableAmountKRW    float64                         `json:"orderable_amount_krw"`
	OrderableAmountUSD    float64                         `json:"orderable_amount_usd"`
	WithdrawableKR        map[string]any                  `json:"withdrawable_kr,omitempty"`
	WithdrawableUS        map[string]any                  `json:"withdrawable_us,omitempty"`
	Markets               map[string]AccountMarketSummary `json:"markets,omitempty"`
}

type AccountMarketSummary struct {
	Market                string  `json:"market"`
	PendingBuyOrderAmount float64 `json:"pending_buy_order_amount"`
	EvaluatedAmount       float64 `json:"evaluated_amount"`
	PrincipalAmount       float64 `json:"principal_amount"`
	EvaluatedProfitAmount float64 `json:"evaluated_profit_amount"`
	ProfitRate            float64 `json:"profit_rate"`
	TotalAssetAmount      float64 `json:"total_asset_amount"`
	OrderableAmountKRW    float64 `json:"orderable_amount_krw"`
	OrderableAmountUSD    float64 `json:"orderable_amount_usd"`
}

type Position struct {
	ProductCode     string  `json:"product_code,omitempty"`
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name,omitempty"`
	MarketType      string  `json:"market_type,omitempty"`
	MarketCode      string  `json:"market_code,omitempty"`
	Quantity        float64 `json:"quantity"`
	AveragePrice    float64 `json:"average_price,omitempty"`
	CurrentPrice    float64 `json:"current_price,omitempty"`
	MarketValue     float64 `json:"market_value,omitempty"`
	UnrealizedPnL   float64 `json:"unrealized_pnl,omitempty"`
	ProfitRate      float64 `json:"profit_rate,omitempty"`
	DailyProfitLoss float64 `json:"daily_profit_loss,omitempty"`
	DailyProfitRate float64 `json:"daily_profit_rate,omitempty"`

	AveragePriceUSD    float64 `json:"average_price_usd,omitempty"`
	CurrentPriceUSD    float64 `json:"current_price_usd,omitempty"`
	MarketValueUSD     float64 `json:"market_value_usd,omitempty"`
	UnrealizedPnLUSD   float64 `json:"unrealized_pnl_usd,omitempty"`
	ProfitRateUSD      float64 `json:"profit_rate_usd,omitempty"`
	DailyProfitLossUSD float64 `json:"daily_profit_loss_usd,omitempty"`
	DailyProfitRateUSD float64 `json:"daily_profit_rate_usd,omitempty"`
}

type Order struct {
	ID                    string          `json:"id"`
	ResolvedFromID        string          `json:"resolved_from_id,omitempty"`
	Symbol                string          `json:"symbol"`
	Name                  string          `json:"name,omitempty"`
	Market                string          `json:"market,omitempty"`
	Side                  string          `json:"side,omitempty"`
	Status                string          `json:"status,omitempty"`
	Quantity              float64         `json:"quantity,omitempty"`
	FilledQuantity        float64         `json:"filled_quantity,omitempty"`
	Price                 float64         `json:"price,omitempty"`
	AverageExecutionPrice float64         `json:"average_execution_price,omitempty"`
	OrderDate             string          `json:"order_date,omitempty"`
	SubmittedAt           *time.Time      `json:"submitted_at,omitempty"`
	Raw                   json.RawMessage `json:"raw,omitempty"`
}

type WatchlistItem struct {
	Group    string  `json:"group,omitempty"`
	Symbol   string  `json:"symbol"`
	Name     string  `json:"name,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Base     float64 `json:"base,omitempty"`
	Last     float64 `json:"last,omitempty"`
}

// Trade is a single executed tick (체결) from the market-data feed.
type Trade struct {
	Time             string  `json:"time"`
	Price            float64 `json:"price"`
	Base             float64 `json:"base,omitempty"`
	Volume           float64 `json:"volume"`
	TradeType        string  `json:"trade_type,omitempty"`        // BUY / SELL
	CumulativeVolume float64 `json:"cumulative_volume,omitempty"`
}

type TradeList struct {
	ProductCode string    `json:"product_code"`
	Symbol      string    `json:"symbol,omitempty"`
	Name        string    `json:"name,omitempty"`
	Trades      []Trade   `json:"trades"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// PriceLimits is the daily upper/lower price band (상/하한가).
type PriceLimits struct {
	ProductCode string  `json:"product_code"`
	Symbol      string  `json:"symbol,omitempty"`
	Name        string  `json:"name,omitempty"`
	Date        string  `json:"date,omitempty"`
	UpperLimit  float64 `json:"upper_limit"`
	LowerLimit  float64 `json:"lower_limit"`
}

// StockWarning is a buy-caution badge (매수 유의사항). The web feed's badge
// shape is dynamic, so non-core fields are preserved as raw JSON.
type StockWarning struct {
	Type    string          `json:"type,omitempty"`
	Title   string          `json:"title,omitempty"`
	Text    string          `json:"text,omitempty"`
	Level   string          `json:"level,omitempty"`
	Raw     json.RawMessage `json:"raw,omitempty"`
}

type StockWarnings struct {
	ProductCode string         `json:"product_code"`
	Symbol      string         `json:"symbol,omitempty"`
	Name        string         `json:"name,omitempty"`
	Warnings    []StockWarning `json:"warnings"`
	FetchedAt   time.Time      `json:"fetched_at"`
}

// ExchangeRate is one FX/index quote (e.g. USD/KRW, DXY).
type ExchangeRate struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Base  float64 `json:"base,omitempty"`
	Close float64 `json:"close"`
}

type ExchangeRates struct {
	Rates     []ExchangeRate `json:"rates"`
	FetchedAt time.Time      `json:"fetched_at"`
}

// MarketSession is one trading day's session times for a market.
type MarketSession struct {
	Date      string `json:"date,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

// TradingHours holds today's (and next business day's) KR and US session
// windows (장 운영 시간). NextKR/NextUS are useful when today is a holiday.
type TradingHours struct {
	KR        MarketSession `json:"kr"`
	US        MarketSession `json:"us"`
	NextKR    MarketSession `json:"next_kr,omitempty"`
	NextUS    MarketSession `json:"next_us,omitempty"`
	FetchedAt time.Time     `json:"fetched_at"`
}

type Transaction struct {
	Type             string          `json:"type"`
	Category         string          `json:"category"`
	Code             string          `json:"code,omitempty"`
	DisplayName      string          `json:"display_name,omitempty"`
	DisplayType      string          `json:"display_type,omitempty"`
	Summary          string          `json:"summary,omitempty"`
	Market           string          `json:"market"`
	Currency         string          `json:"currency"`
	StockCode        string          `json:"stock_code,omitempty"`
	StockName        string          `json:"stock_name,omitempty"`
	Quantity         float64         `json:"quantity,omitempty"`
	Amount           float64         `json:"amount"`
	AdjustedAmount   float64         `json:"adjusted_amount"`
	CommissionAmount float64         `json:"commission_amount,omitempty"`
	TaxAmount        float64         `json:"tax_amount,omitempty"`
	BalanceAmount    float64         `json:"balance_amount,omitempty"`
	Date             string          `json:"date,omitempty"`
	DateTime         string          `json:"datetime,omitempty"`
	OrderDate        string          `json:"order_date,omitempty"`
	SettlementDate   string          `json:"settlement_date,omitempty"`
	TradeType        string          `json:"trade_type,omitempty"`
	ReferenceType    string          `json:"reference_type,omitempty"`
	ReferenceID      string          `json:"reference_id,omitempty"`
	SortKey          string          `json:"sort_key,omitempty"`
	Raw              json.RawMessage `json:"raw,omitempty"`
}

type TransactionPage struct {
	Market   string        `json:"market"`
	Items    []Transaction `json:"items"`
	LastPage bool          `json:"last_page"`
	Next     *PagingParam  `json:"next,omitempty"`
}

type PagingParam struct {
	Number  int    `json:"number,omitempty"`
	Size    int    `json:"size,omitempty"`
	Key     string `json:"key,omitempty"`
	Filters string `json:"filters,omitempty"`
	Type    string `json:"type,omitempty"`
}

type TransactionOverview struct {
	Market                  string                         `json:"market"`
	OrderableKRW            float64                        `json:"orderable_krw"`
	OrderableUSD            float64                        `json:"orderable_usd"`
	Withdrawable            []SettlementBucket             `json:"withdrawable,omitempty"`
	DisplayWithdrawable     []SettlementBucket             `json:"display_withdrawable,omitempty"`
	Deposit                 []SettlementBucket             `json:"deposit,omitempty"`
	EstimateSettlement      []SettlementEstimate           `json:"estimate_settlement,omitempty"`
	WithdrawableBottomSheet []WithdrawableBottomSheetEntry `json:"withdrawable_bottom_sheet,omitempty"`
}

type SettlementBucket struct {
	Date string  `json:"date,omitempty"`
	KRW  float64 `json:"krw,omitempty"`
	USD  float64 `json:"usd,omitempty"`
}

type SettlementEstimate struct {
	Date       string  `json:"date,omitempty"`
	BuyAmount  float64 `json:"buy_amount,omitempty"`
	SellAmount float64 `json:"sell_amount,omitempty"`
}

type WithdrawableBottomSheetEntry struct {
	Title string  `json:"title"`
	KRW   float64 `json:"krw,omitempty"`
	USD   float64 `json:"usd,omitempty"`
}

type Candle struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume,omitempty"`
}

type Chart struct {
	ProductCode string    `json:"product_code"`
	Symbol      string    `json:"symbol,omitempty"`
	Name        string    `json:"name,omitempty"`
	Interval    string    `json:"interval"`
	Base        float64   `json:"base,omitempty"`
	Candles     []Candle  `json:"candles"`
	FetchedAt   time.Time `json:"fetched_at"`
}

type Quote struct {
	ProductCode    string    `json:"product_code,omitempty"`
	Symbol         string    `json:"symbol"`
	Name           string    `json:"name,omitempty"`
	MarketCode     string    `json:"market_code,omitempty"`
	Market         string    `json:"market,omitempty"`
	Currency       string    `json:"currency,omitempty"`
	ReferencePrice float64   `json:"reference_price,omitempty"`
	Last           float64   `json:"last,omitempty"`
	Change         float64   `json:"change,omitempty"`
	ChangeRate     float64   `json:"change_rate,omitempty"`
	Volume         float64   `json:"volume,omitempty"`
	Open           float64   `json:"open,omitempty"`
	High           float64   `json:"high,omitempty"`
	Low            float64   `json:"low,omitempty"`
	High52w        float64   `json:"high_52w,omitempty"`
	Low52w         float64   `json:"low_52w,omitempty"`
	MarketCap      float64   `json:"market_cap,omitempty"`
	TradingValue   float64   `json:"trading_value,omitempty"`   // 거래대금
	TradingStrength float64  `json:"trading_strength,omitempty"` // 체결강도 (%)
	PrevVolume     float64   `json:"prev_volume,omitempty"`
	UpperLimit     float64   `json:"upper_limit,omitempty"`
	LowerLimit     float64   `json:"lower_limit,omitempty"`
	Status         string    `json:"status,omitempty"`
	BadgeCount     int       `json:"badge_count,omitempty"`
	NoticeCount    int       `json:"notice_count,omitempty"`
	FetchedAt      time.Time `json:"fetched_at"`
}
