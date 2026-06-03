# Toss Securities Open API 마이그레이션 계획

토스증권이 공식 Open API 사전 신청을 시작했습니다 (광고 심의 시작일 2026-05-14).
이 문서는 `tossctl` 이 공식 API 출시 흐름에 맞춰 어떻게 진화할지 정리한 living
document 입니다. issue [#31](https://github.com/JungHoonGhae/tossinvest-cli/issues/31)
이 트래킹 anchor.

> **status:** Phase 0.5 (공식 spec + 문서 공개, 토큰 발급 절차 미검증) · 마지막 업데이트 2026-06-02
>
> **⚠ 본 문서의 계획·timeline·phase 정의·우선순위는 사전 공지 없이 바뀔 수 있습니다.**
> 토스 공식 표면이 실제로 드러나는 시점, 자원/시간 사정, 새로운 정보에 따라 유연하게
> 재조정합니다. 여기 적힌 내용은 commitment 가 아니라 "현 시점 stance" 입니다.

## 토스 Open API 의 윤곽 (공식 문서 기준)

2026-06-02 공식 개발자 문서 + OpenAPI spec 이 공개됐습니다 (제보: @skyisle, issue #31).
아래는 추정이 아니라 **공식 spec 에서 확인된 사실**입니다.

- **개발자 문서:** https://developers.tossinvest.com/docs (+ AI agent 용 `/llms.txt`)
- **OpenAPI spec (source of truth):** `https://openapi.tossinvest.com/openapi-docs/latest/openapi.json` (v1.0.3, 20 endpoints)
- **Base URL:** `https://openapi.tossinvest.com`
- **인증:** OAuth 2.0 **Client Credentials Grant** — `POST /oauth2/token` 에 `client_id` + `client_secret` (form-urlencoded) → `access_token` (Bearer, `expires_in` 86400, refresh token 없음, client 당 유효 토큰 1개). 계좌·자산·주문 API 는 `Authorization: Bearer` 외에 `X-Tossinvest-Account` 헤더 필수.
- **프로토콜:** REST only (마케팅 페이지의 WebSocket 언급은 아직 spec 에 없음)
- **endpoint 표면 (tossctl 명령과의 매핑):**

  | tossctl | 공식 endpoint |
  |---|---|
  | `account list` | `GET /api/v1/accounts` |
  | `portfolio positions` | `GET /api/v1/holdings` |
  | `quote get` | `GET /api/v1/prices` · `/orderbook` · `/trades` |
  | `quote chart` | `GET /api/v1/candles` (1분봉·일봉) |
  | `order place` | `POST /api/v1/orders` |
  | `order amend` | `POST /api/v1/orders/{orderId}/modify` |
  | `order cancel` | `POST /api/v1/orders/{orderId}/cancel` |
  | `orders list` / `order show` | `GET /api/v1/orders` · `/orders/{orderId}` |
  | (신규) | `buying-power` · `sellable-quantity` · `commissions` · `exchange-rate` · `market-calendar/{KR,US}` · `stocks` · `price-limits` |

- **미검증:** client_id/secret 발급 콘솔 절차, 거래 권한 scope 모델, rate limit 실측, candle 이 1분봉·일봉만이라 기존 tossctl 의 3/5/15/30/60분봉과 차이 있음

## 포지셔닝 — 공식이 나와도 tossctl 은 갈아타지 않는다

**핵심 교정 (2026-06-03):** 초기 계획은 "공식 나오면 tossctl 이 공식으로 이주,
session 은 deprecate" 였다. 이는 *"공식 = 상위호환"* 이라는 잘못된 가정에 기반했다.

공식 API 는 상위호환이 아니라 **다른 trade-off** 다:

| | tossctl (비공식 web session) | 토스 공식 API |
|---|---|---|
| 데이터 표면 | **넓음** — transactions ledger·watchlist·push(SSE) 포함 | 좁음 — 시세·계좌·주문만 (ledger/watchlist/push **없음**) |
| rate limit | 사실상 없음 | 있음 (ACCOUNT 초당 1회 등) |
| 진입 | `auth login` 1회 | 사전신청→승인→토큰 24h 재발급 |
| 합법성·안정성 | 약함 (TOS 리스크, 예고 없는 변경) | 강함 (계약된 versioned spec) |

→ tossctl 이 공식으로 이주하면 **우리 해자(넓은 표면 + 무제한)를 스스로 버린다.**
그래서 **이주·deprecation 하지 않는다.** tossctl 은 "비공식·넓은 표면" 도구로 영구 유지.

**토스 공식 adapter 는 tossctl 이 아니라 별도 멀티-브로커 CLI (investctl) 로 간다.**
공식의 강점(합법성·통합)은 거기서, 비공식의 강점(표면·무제한)은 여기서 — 두 제품이
다른 수요를 영구히 분담한다. 자기잠식 없음. 상세: [multi-broker.md](./multi-broker.md).

## tossctl 이 공식 출시에 실제로 할 일 (작음)

이주가 아니므로 할 일은 작다:

| 시점 | tossctl 동작 |
|---|---|
| **지금** | 현행 유지. issue #31 모니터링 (spec version 추적 — daily-monitor.yml) |
| **공식 안정 후** | README 에 "합법성·통합 원하면 investctl" 교차 링크. tossctl 은 그대로 |
| **비공식 endpoint 차단 시** (위험 1) | 그때만 대응 — 공식 adapter 옵션 추가 검토. 그 전엔 불필요 |

## 위험 요소

1. **비공식 endpoint 차단 가능성** — 공식 출시 후 토스가 reverse-engineered 접근을
   정책/기술적으로 막을 수 있음. 이것이 tossctl 의 유일한 실존적 위험 (해자인 넓은
   표면이 통째로 사라짐). 차단되면 그때 공식 adapter 옵션을 검토 — 단 공식은 표면이
   좁아 완전 대체는 안 되므로 일부 기능 손실 감수. 그 전엔 선제 이주 불필요.
2. **investctl 과의 중복 유지비** — 두 레포에 core(auth/output/config) 가 copy-first
   로 갈라져 유지비 발생. 두 번째 구현으로 패턴 검증되면 공유 모듈화 검토 (그 전엔 X).

## 결정 log

각 항목은 *그 시점의* stance. 이후 정보로 뒤집힐 수 있고, 뒤집힐 때 사전 공지하지
않습니다 — 새 항목을 추가할 뿐입니다.

- **2026-05-19** — issue #31 등록 (제보: @DaeHyeoNi). 사전 신청 페이지 확인. 사전 신청 진행. tossctl 의 일반화 (multi-broker) 방향을 *현재로서는* 선호. Phase 1 진입 전까지 코드 추상화는 보류 — 공식 표면을 보기 전 추상화는 잘못 잡을 확률이 크다는 판단
- **2026-06-02** — 공식 개발자 문서 + OpenAPI spec 공개 확인 (제보: @skyisle). `developers.tossinvest.com/docs` + `openapi.tossinvest.com/openapi-docs/.../openapi.json` (v1.0.3, 20 endpoints, OAuth2 Client Credentials). 모니터링 신호를 corp 페이지 chunk hash → spec version + endpoint 목록 hash 로 교체 (훨씬 강한 신호). endpoint 표면이 tossctl 명령과 거의 1:1 매핑 확인.
- **2026-06-03** — **이주·deprecation 계획 철회.** 공식 API 가 상위호환이 아니라 다른 trade-off (표면 좁고 rate limit 있음) 임이 spec 으로 확인됨. tossctl 이 공식으로 갈아타면 해자(transactions·watchlist·push·무제한)를 스스로 버림. → tossctl 은 "비공식·넓은 표면" 으로 영구 유지. 토스 공식 adapter 는 tossctl 이 아니라 별도 멀티-브로커 CLI(investctl, `oss-kr-investctl`)로. 두 제품이 다른 수요를 분담, 자기잠식 없음. 이 문서의 Broker 추상화 범위는 "tossctl 자체가 공식으로 갈아타기"(폐기) 가 아니라 investctl 의 cross-broker 통합 ([multi-broker.md](./multi-broker.md)) 으로 이관.

## 외부 contributor / 사용자에게 부탁

1. 토스 Open API 토큰을 받으셨다면 issue #31 에 댓글로 알려주세요 — phase 1 진입
   판단의 가장 빠른 신호. 토큰/계정번호 같은 민감 정보는 절대 공개 댓글에 붙이지 마세요
2. 공식 API 의 endpoint/스펙 문서를 발견하시면 issue #31 에 링크 공유 환영
3. 이 문서의 timeline/계획에 의견 있으면 issue #31 댓글 또는 별도 PR 환영
