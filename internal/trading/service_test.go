package trading

import (
	"context"
	"errors"
	"testing"

	"github.com/junghoonkye/tossinvest-cli/internal/config"
	"github.com/junghoonkye/tossinvest-cli/internal/orderintent"
)

type brokerStub struct {
	placeCalled  bool
	cancelCalled bool
	amendCalled  bool
	lastOrderID  string
	placeResult  MutationResult
	cancelResult MutationResult
	amendResult  MutationResult
}

func (b *brokerStub) PlacePendingOrder(_ context.Context, intent orderintent.PlaceIntent) (MutationResult, error) {
	b.placeCalled = true
	b.lastOrderID = intent.Symbol
	if b.placeResult.Kind == "" {
		b.placeResult = MutationResult{Kind: "place", Status: "accepted_pending", OrderID: intent.Symbol}
	}
	return b.placeResult, nil
}

func (b *brokerStub) GetOrderAvailableActions(_ context.Context, orderID string) (map[string]any, error) {
	b.lastOrderID = orderID
	return map[string]any{"cancelSupported": true}, nil
}

func (b *brokerStub) CancelPendingOrder(_ context.Context, intent orderintent.CancelIntent) (MutationResult, error) {
	b.cancelCalled = true
	b.lastOrderID = intent.OrderID
	if b.cancelResult.Kind == "" {
		b.cancelResult = MutationResult{Kind: "cancel", Status: "canceled", OrderID: intent.OrderID}
	}
	return b.cancelResult, nil
}

func (b *brokerStub) AmendPendingOrder(_ context.Context, intent orderintent.AmendIntent) (MutationResult, error) {
	b.amendCalled = true
	b.lastOrderID = intent.OrderID
	if b.amendResult.Kind == "" {
		b.amendResult = MutationResult{Kind: "amend", Status: "amended_pending", OrderID: intent.OrderID, CurrentOrderID: intent.OrderID}
	}
	return b.amendResult, nil
}

func TestPlaceRequiresExecutionFlags(t *testing.T) {
	service := NewService(config.Trading{
		Place:                 true,
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	// No flags → --execute required.
	if _, err := service.Place(context.Background(), intent, ExecuteOptions{}); !errors.Is(err, ErrExecuteRequired) {
		t.Fatalf("expected ErrExecuteRequired, got %v", err)
	}
	// --execute but no confirm token → confirm mismatch (empty token).
	if _, err := service.Place(context.Background(), intent, ExecuteOptions{Execute: true}); !errors.Is(err, ErrConfirmMismatch) {
		t.Fatalf("expected ErrConfirmMismatch for empty token, got %v", err)
	}
	// --execute with a wrong token → confirm mismatch.
	if _, err := service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: "badtoken",
	}); !errors.Is(err, ErrConfirmMismatch) {
		t.Fatalf("expected ErrConfirmMismatch, got %v", err)
	}
	// --execute + valid token → gates pass; broker is nil so mutation is pending.
	if _, err := service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	}); !errors.Is(err, ErrLiveMutationPending) {
		t.Fatalf("expected ErrLiveMutationPending, got %v", err)
	}
}

func TestPlaceCallsBrokerForSupportedIntent(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Place:                 true,
		AllowLiveOrderActions: true,
	}, broker)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	result, err := service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}
	if !broker.placeCalled {
		t.Fatal("expected broker place to be called")
	}
	if result.Status != "accepted_pending" {
		t.Fatalf("expected accepted_pending, got %q", result.Status)
	}
}

func TestCancelExecutesBrokerAndReconciles(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Cancel:                true,
		AllowLiveOrderActions: true,
	}, broker)

	intent, err := orderintent.NormalizeCancel("5", "TSLL")
	if err != nil {
		t.Fatalf("NormalizeCancel returned error: %v", err)
	}

	result, err := service.Cancel(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewCancel(intent).ConfirmToken,
	})
	if err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	if !broker.cancelCalled {
		t.Fatal("expected broker cancel to be called")
	}
	if broker.lastOrderID != "5" {
		t.Fatalf("expected order id 5, got %q", broker.lastOrderID)
	}
	if result.Status != "canceled" {
		t.Fatalf("expected canceled result, got %q", result.Status)
	}
}

func TestAmendCallsBrokerAfterGate(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Amend:                 true,
		AllowLiveOrderActions: true,
	}, broker)
	price := 700.0
	intent, err := orderintent.NormalizeAmend("13", nil, &price)
	if err != nil {
		t.Fatalf("NormalizeAmend returned error: %v", err)
	}

	result, err := service.Amend(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewAmend(intent).ConfirmToken,
	})
	if err != nil {
		t.Fatalf("Amend returned error: %v", err)
	}
	if !broker.amendCalled {
		t.Fatal("expected broker amend to be called")
	}
	if result.Status != "amended_pending" {
		t.Fatalf("expected amended_pending, got %q", result.Status)
	}
}

func TestPlaceFailsWhenActionDisabledInConfig(t *testing.T) {
	service := NewService(config.Trading{
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	_, err = service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	var disabled *DisabledActionError
	if !errors.As(err, &disabled) || disabled.Action != ActionPlace {
		t.Fatalf("expected place action to be disabled, got %v", err)
	}
}

func TestPlaceIntentSupportedAcceptsSell(t *testing.T) {
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}
	if !placeIntentSupported(intent) {
		t.Fatal("expected placeIntentSupported to return true for sell intent")
	}
}

func TestSellPlaceFailsWhenSellDisabledInConfig(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Place:                 true,
		Sell:                  false,
		AllowLiveOrderActions: true,
	}, broker)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	_, err = service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	var disabled *DisabledActionError
	if !errors.As(err, &disabled) || disabled.Action != "sell" {
		t.Fatalf("expected sell action to be disabled, got %v", err)
	}
	if broker.placeCalled {
		t.Fatal("broker should not have been called when sell is disabled")
	}
}

func TestSellPlaceCallsBrokerWhenSellEnabled(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Place:                 true,
		Sell:                  true,
		AllowLiveOrderActions: true,
	}, broker)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	result, err := service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}
	if !broker.placeCalled {
		t.Fatal("expected broker place to be called for sell")
	}
	if result.Status != "accepted_pending" {
		t.Fatalf("expected accepted_pending, got %q", result.Status)
	}
}

func TestPreviewPlaceSellDisabled(t *testing.T) {

	service := NewService(config.Trading{
		Place:                 true,
		Sell:                  false,
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	preview := service.PreviewPlace(intent)
	if preview.MutationReady {
		t.Fatal("expected MutationReady to be false when sell is disabled")
	}
	found := false
	for _, w := range preview.Warnings {
		if w == "Config currently disables `order place --side sell`." {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected sell-disabled warning in preview, got %v", preview.Warnings)
	}
}

func TestPreviewPlaceSellEnabled(t *testing.T) {

	service := NewService(config.Trading{
		Place:                 true,
		Sell:                  true,
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	preview := service.PreviewPlace(intent)
	if !preview.LiveReady {
		t.Fatal("expected LiveReady to be true for sell")
	}
	if !preview.MutationReady {
		t.Fatal("expected MutationReady to be true when sell is enabled")
	}
}

func TestPlaceIntentSupportedAcceptsFractionalMarketUS(t *testing.T) {
	intent := orderintent.PlaceIntent{Symbol: "TSLL", Market: "us", Side: "buy", OrderType: "market", Amount: 18000, CurrencyMode: "KRW", Fractional: true}
	if !placeIntentSupported(intent) {
		t.Fatal("expected true for US fractional market order")
	}
}

func TestPlaceIntentSupportedRejectsFractionalLimit(t *testing.T) {
	intent := orderintent.PlaceIntent{Symbol: "TSLL", Market: "us", Side: "buy", OrderType: "limit", Price: 500, CurrencyMode: "KRW", Fractional: true}
	if placeIntentSupported(intent) {
		t.Fatal("expected false for fractional limit order")
	}
}

func TestPlaceIntentSupportedAcceptsUSLimitUSD(t *testing.T) {
	intent := orderintent.PlaceIntent{Symbol: "MRVL", Market: "us", Side: "buy", OrderType: "limit", Quantity: 1, Price: 158.01, CurrencyMode: "USD"}
	if !placeIntentSupported(intent) {
		t.Fatal("expected true for US limit order with USD price input")
	}
}

func TestPlaceIntentSupportedRejectsKRLimitUSD(t *testing.T) {
	intent := orderintent.PlaceIntent{Symbol: "290080", Market: "kr", Side: "buy", OrderType: "limit", Quantity: 1, Price: 10, CurrencyMode: "USD"}
	if placeIntentSupported(intent) {
		t.Fatal("expected false for KR limit order with USD price input")
	}
}

func TestPlaceIntentSupportedRejectsFractionalKR(t *testing.T) {
	intent := orderintent.PlaceIntent{Symbol: "290080", Market: "kr", Side: "buy", OrderType: "market", Amount: 8000, CurrencyMode: "KRW", Fractional: true}
	if placeIntentSupported(intent) {
		t.Fatal("expected false for KR fractional order")
	}
}

func TestFractionalPlaceFailsWhenDisabled(t *testing.T) {
	svc := NewService(config.Trading{Place: true, Fractional: false, AllowLiveOrderActions: true}, nil)
	intent := orderintent.PlaceIntent{Symbol: "TSLL", Market: "us", Side: "buy", OrderType: "market", Amount: 18000, CurrencyMode: "KRW", Fractional: true}
	_, err := svc.Place(context.Background(), intent, ExecuteOptions{Execute: true, Confirm: svc.PreviewPlace(intent).ConfirmToken})
	var disabled *DisabledActionError
	if !errors.As(err, &disabled) || disabled.Action != "fractional" {
		t.Fatalf("expected fractional disabled, got %v", err)
	}
}

func TestFractionalPlaceCallsBroker(t *testing.T) {
	broker := &brokerStub{}
	svc := NewService(config.Trading{Place: true, Fractional: true, AllowLiveOrderActions: true}, broker)
	intent := orderintent.PlaceIntent{Symbol: "TSLL", Market: "us", Side: "buy", OrderType: "market", Amount: 18000, CurrencyMode: "KRW", Fractional: true}
	result, err := svc.Place(context.Background(), intent, ExecuteOptions{Execute: true, Confirm: svc.PreviewPlace(intent).ConfirmToken})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !broker.placeCalled {
		t.Fatal("expected broker called")
	}
	if result.Status != "accepted_pending" {
		t.Fatalf("expected accepted_pending, got %q", result.Status)
	}
}

func TestPreviewPlaceFractionalDisabled(t *testing.T) {
	svc := NewService(config.Trading{Place: true, Fractional: false, AllowLiveOrderActions: true}, nil)
	intent := orderintent.PlaceIntent{Symbol: "TSLL", Market: "us", Side: "buy", OrderType: "market", Amount: 18000, CurrencyMode: "KRW", Fractional: true}
	preview := svc.PreviewPlace(intent)
	if preview.MutationReady {
		t.Fatal("expected MutationReady false")
	}
}

func TestPlaceIntentSupportedAcceptsKR(t *testing.T) {
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "290080",
		Market:       "kr",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        8000,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}
	if !placeIntentSupported(intent) {
		t.Fatal("expected placeIntentSupported to return true for kr intent")
	}
}

func TestKRPlaceFailsWhenKRDisabledInConfig(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Place:                 true,
		KR:                    false,
		AllowLiveOrderActions: true,
	}, broker)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "290080",
		Market:       "kr",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        8000,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	_, err = service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	var disabled *DisabledActionError
	if !errors.As(err, &disabled) || disabled.Action != "kr" {
		t.Fatalf("expected kr action to be disabled, got %v", err)
	}
	if broker.placeCalled {
		t.Fatal("broker should not have been called when kr is disabled")
	}
}

func TestKRPlaceCallsBrokerWhenKREnabled(t *testing.T) {
	broker := &brokerStub{}
	service := NewService(config.Trading{
		Place:                 true,
		KR:                    true,
		AllowLiveOrderActions: true,
	}, broker)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "290080",
		Market:       "kr",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        8000,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	result, err := service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	if err != nil {
		t.Fatalf("Place returned error: %v", err)
	}
	if !broker.placeCalled {
		t.Fatal("expected broker place to be called for kr")
	}
	if result.Status != "accepted_pending" {
		t.Fatalf("expected accepted_pending, got %q", result.Status)
	}
}

func TestPlacePolicyChecksBeforeGuard(t *testing.T) {
	// NO grant — guard() would fail with ErrExecuteRequired etc.
	// But policy check should catch it first.

	service := NewService(config.Trading{
		Place:                 true,
		KR:                    false,
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "290080",
		Market:       "kr",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        8000,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	// Without --execute, guard would return ErrExecuteRequired.
	// But kr policy should be checked BEFORE guard, so we get DisabledActionError.
	_, err = service.Place(context.Background(), intent, ExecuteOptions{})
	var disabled *DisabledActionError
	if !errors.As(err, &disabled) || disabled.Action != "kr" {
		t.Fatalf("expected kr disabled error BEFORE guard, got %v", err)
	}
}

func TestPreviewPlaceKRDisabled(t *testing.T) {

	service := NewService(config.Trading{
		Place:                 true,
		KR:                    false,
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "290080",
		Market:       "kr",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        8000,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	preview := service.PreviewPlace(intent)
	if preview.MutationReady {
		t.Fatal("expected MutationReady to be false when kr is disabled")
	}
	found := false
	for _, w := range preview.Warnings {
		if w == "Config currently disables `order place --market kr`." {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected kr-disabled warning in preview, got %v", preview.Warnings)
	}
}

func TestPreviewPlaceKREnabled(t *testing.T) {

	service := NewService(config.Trading{
		Place:                 true,
		KR:                    true,
		AllowLiveOrderActions: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "290080",
		Market:       "kr",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        8000,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	preview := service.PreviewPlace(intent)
	if !preview.LiveReady {
		t.Fatal("expected LiveReady to be true for kr")
	}
	if !preview.MutationReady {
		t.Fatal("expected MutationReady to be true when kr is enabled")
	}
}

func TestPlaceFailsWhenLiveActionsDisabledInConfig(t *testing.T) {
	service := NewService(config.Trading{
		Place: true,
	}, nil)
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{
		Symbol:       "TSLL",
		Market:       "us",
		Side:         "buy",
		OrderType:    "limit",
		Quantity:     1,
		Price:        500,
		CurrencyMode: "KRW",
	})
	if err != nil {
		t.Fatalf("NormalizePlace returned error: %v", err)
	}

	_, err = service.Place(context.Background(), intent, ExecuteOptions{
		Execute: true,
		Confirm: service.PreviewPlace(intent).ConfirmToken,
	})
	if !errors.Is(err, ErrLiveActionsDisabled) {
		t.Fatalf("expected ErrLiveActionsDisabled, got %v", err)
	}
}
