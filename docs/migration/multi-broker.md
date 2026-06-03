# Multi-Broker 추상화 방향

`tossctl` 의 장기 생존 전략은 **"토스 비공식 CLI"** 에서 **"한국 증권사를 AI
에이전트에 통일된 인터페이스로 연결하는 CLI"** 로 일반화하는 것이다. 이 문서는
그 추상화 설계 방향을 정리한 living document 다. [open-api.md](./open-api.md)
의 Phase 계획과 함께 읽는다.

> **status:** 설계 단계 (코드 착수 전) · 마지막 업데이트 2026-06-03
>
> **⚠ 계획·인터페이스·우선순위는 사전 공지 없이 바뀔 수 있습니다.** 실제 백엔드
> (토스 공식 토큰, KIS/키움 키)를 손에 쥐기 전엔 추상화를 확정하지 않는다 —
> 표면을 보기 전 설계는 잘못 잡을 확률이 크다. 여기 적힌 건 "현 시점 stance".

## 왜 multi-broker 인가

토스가 공식 Open API (`developers.tossinvest.com`, issue #31) 를 내면서 단일
백엔드 의존의 위험이 드러났다:

- 토스가 `llms.txt` 까지 제공 = **AI 친화 레이어를 직접 노림**. 우리 해자의 정면.
- 토스 하나에만 묶이면 토스 공식이 우리를 흡수할 수 있다.
- "한국 증권 통합 AI 인터페이스"가 되면 **어느 한 증권사도 우리를 대체 못 한다.**

핵심 통찰: 우리 가치는 "토스에 접근하는 것"이 아니라 **"여러 증권사를 AI·사람이
가장 쉽게 쓰게 만드는 레이어 + 공식이 안 주는 데이터"** 다. 백엔드는 교체 가능한
부품이어야 한다.

## 목표 아키텍처

```
        cmd/tossctl (명령 레이어 — 백엔드 무관)
                     │
                     ▼
            internal/broker.Broker  (interface)
            ┌────────┼─────────┬──────────────┐
            ▼        ▼         ▼              ▼
   TossSessionBroker  TossOfficialBroker  KISBroker  KiwoomBroker
   (web reverse-eng)  (OAuth2 공식)       (한투)     (키움)
            │              │
   wts-*.tossinvest   openapi.tossinvest
```

- **명령 레이어는 broker 를 모른다.** `tossctl portfolio positions` 는 어떤
  백엔드든 동일하게 동작.
- 백엔드는 config 또는 `--broker` 플래그로 선택. 미지정 시 사용 가능한 것 자동.
- 현재 `internal/client.Client` 가 사실상 단일 구현 (`TossSessionBroker`).
  이것을 interface 뒤로 옮긴다.

## Broker interface (초안 — 미확정)

현재 `internal/client.Client` 의 공개 메서드에서 도출. 모든 백엔드가 전부 구현할
필요는 없다 — 미지원은 `ErrUnsupported` 를 반환하고 명령 레이어가 친절히 처리.

```go
// internal/broker/broker.go (제안)
type Broker interface {
    // 조회 (대부분 백엔드 공통)
    ListAccounts(ctx) ([]domain.Account, error)
    ListPositions(ctx, market) ([]domain.Position, error)
    GetQuote(ctx, symbol) (domain.Quote, error)
    GetChart(ctx, symbol, interval, count) (domain.Chart, error)
    ListPendingOrders(ctx) ([]domain.Order, error)
    ListCompletedOrders(ctx, market) ([]domain.Order, error)

    // 거래 (권한 게이트는 명령 레이어가 유지)
    PlaceOrder(ctx, intent) (domain.OrderResult, error)
    CancelOrder(ctx, id) error
    AmendOrder(ctx, id, changes) error

    // 메타 — 백엔드 능력 선언
    Capabilities() BrokerCapabilities
}

type BrokerCapabilities struct {
    Name        string   // "toss-session" | "toss-official" | "kis" | ...
    Markets     []string // "kr", "us"
    Trading     bool
    Realtime    bool     // push/SSE 지원
    Extras      []string // "transactions", "watchlist", "warnings" 등 비표준 표면
}
```

**`Capabilities()` 가 핵심.** 백엔드마다 표면이 다르므로 (공식 API 엔 watchlist
없음, KIS 엔 또 다른 것), 명령 레이어가 능력을 질의해 미지원 기능은 명확한
메시지로 막는다. (예: `quote limits` 가 US 종목을 거부하는 현재 패턴과 동일 철학)

## 백엔드별 매핑 (현재까지 파악)

| 기능 | toss-session (현재) | toss-official | KIS / 키움 |
|---|---|---|---|
| 계좌·보유·주문·시세 | ✅ | ✅ (OAuth2) | ✅ (각자 OAuth/키) |
| transactions ledger | ✅ | ❌ (공식 표면 없음) | 미확인 |
| watchlist | ✅ | ❌ | 미확인 |
| push (실시간 SSE) | ✅ | ❌ (REST only) | KIS 는 WebSocket 있음 |
| 매수유의·상하한가·체결·환율·장운영 | ✅ (web) | ✅ (공식 endpoint) | 미확인 |
| rate limit | 없음 | 있음 (ACCOUNT 1/s 등) | 있음 |
| 진입 마찰 | `auth login` 1회 | 사전신청→승인→토큰 | 계좌+키 발급 |

## 단계적 도입 (open-api.md Phase 와 정합)

1. **Phase 1 진입 시 (토스 공식 토큰 확보 후)** — `internal/broker` 패키지 신설,
   현재 `Client` 를 `TossSessionBroker` 로 래핑 (동작 변화 0, 순수 리팩토링).
   `TossOfficialBroker` 추가. config 에 `broker` 필드.
2. **안정화** — `--broker` 플래그 + 자동 선택. doctor 가 사용 가능 백엔드 표시.
3. **확장** — KIS broker 평가 (한투 open-trading-api 는 성숙·문서 풍부).
   키움 REST 도 후보.

## 설계 원칙

- **명령 레이어는 절대 백엔드를 직접 알지 않는다.** broker interface 만 의존.
- **미지원은 에러로, 침묵하지 않는다.** `Capabilities()` + `ErrUnsupported`.
- **거래 안전 게이트(config·permission)는 명령 레이어에 유지** — 백엔드로 내리지
  않는다. 어떤 백엔드든 동일한 안전 모델.
- **AI ergonomics 레이어(JSON/CSV, AGENTS.md, monitor)는 백엔드 무관 공통자산** —
  이게 우리 차별점이므로 broker 와 직교하게 유지.
- **추상화는 두 번째 구현이 생길 때 한다.** 지금 `Client` 하나뿐일 때 interface
  먼저 파면 잘못 추상화한다 ([YAGNI](https://martinfowler.com/bliki/Yagni.html)).
  `TossOfficialBroker` 라는 실제 두 번째 구현이 손에 잡힐 때 interface 를 추출.

## 위험·미해결

1. **토스 공식의 AI 레이어 확장** — `llms.txt` 가 신호. 토스가 공식 SDK + agent
   통합까지 내면 우리 ergonomics 해자 약화. → multi-broker 일반화로 방어.
2. **각 증권사 ToS** — KIS/키움 공식 API 는 합법이나 약관·계좌 요건 상이. broker
   별로 문서화 필요.
3. **interface 조기 확정 risk** — 위 "두 번째 구현 생길 때" 원칙으로 회피.
4. **거래 권한 모델 차이** — 공식 API 가 거래 scope 별도 신청이면 Capabilities 에
   반영. KIS 는 모의투자/실전 분리.

## 결정 log

각 항목은 *그 시점의* stance. 뒤집힐 때 사전 공지 없이 새 항목만 추가.

- **2026-06-03** — multi-broker 방향을 장기 생존 전략으로 채택 (현 시점 선호).
  공식 API 출시 후에도 우리 우위(넓은 데이터 표면·rate limit 없음·AI ergonomics)
  를 지키려면 백엔드를 교체 가능 부품으로 만들어야 한다는 판단. **단 코드 착수는
  토스 공식 토큰을 실제 확보(Phase 1)한 뒤** — interface 는 두 번째 구현이 손에
  잡힐 때 추출. 그 전까진 본 문서가 방향 anchor.
