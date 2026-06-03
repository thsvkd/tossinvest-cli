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

## 우리 포지셔닝 변화

현재: "토스증권 웹 세션을 reverse-engineer 한 비공식 CLI"

공식 API 출시 후: **"한국 증권사를 AI 에이전트에 통일된 인터페이스로 연결하는 CLI"** —
백엔드 plugin 으로 추상화해서 official Toss / 비공식 Toss / (장기적으로) KIS · 키움 등을
같은 명령어로 다룸. 사용자가 토큰을 받았든 안 받았든 `tossctl portfolio positions` 의
표면은 동일. 추상화 설계 상세는 [multi-broker.md](./multi-broker.md).

## Phase 별 계획 (잠정)

표 안의 phase 정의·동작·작업 모두 잠정. 공식 표면이 드러나면 항목이 합쳐지거나
분리되거나 순서가 바뀔 수 있습니다.

| Phase | 트리거 | tossctl 동작 | 작업 |
|---|---|---|---|
| **0** *(지금)* | 사전 신청만 가능, 토큰 발급 0 | 현행 — session-based 만 | issue #31 트래킹, 사전 신청, 본 문서 유지 |
| **1** | 일부 사용자 토큰 발급 시작 | `tossctl auth login --official-token <token>` 추가. config 에 토큰 있으면 official, 없으면 session. **명령어 표면 동일** | `Broker` interface 추상화 (`TossSessionBroker` / `TossOfficialBroker`), `OAuthBearer` 인증, doctor 안내 |
| **2** | 대부분 토큰 발급, official 안정 | default 가 official, session 은 fallback. doctor 가 자동 전환 권장 | 거래 권한 모델 정리 (official 의 trading scope 가 별도 신청이라면 분기) |
| **3** | 정착 | session-based deprecation. KIS/키움 broker plugin 검토 | `tossctl --broker toss|kis|kiwoom` 가능성 평가 |

## Phase 1 의 UX 원칙

- 토큰 받은 사용자: `tossctl auth login --official-token ...` 한 번. 이후 끝
- 토큰 못 받은 사용자: 기존 흐름 그대로. 아무것도 안 변함
- **두 그룹이 같은 README, 같은 명령어, 같은 AGENTS.md** 를 봄. tossctl 이 매개

doctor 출력 예시:
```
Backend: toss-session (active)
Official API: not yet (waitlist) — apply at https://corp.tossinvest.com/ko/open-api
```

토큰 발급 후 (정확한 필드명/만료 정책은 실제 토큰 받은 후 확정):
```
Backend: toss-official (token expires ...)
Session fallback: configured
```

## 위험 요소

1. **비공식 endpoint 차단 가능성** — 공식 출시 후 토스가 reverse-engineered 접근을
   정책/기술적으로 막을 수 있음. 그 시점에 session 백엔드는 빠르게 deprecate 가 강제될
   수 있음
2. **공식 API 의 거래 권한** — official 이 거래 권한을 별도 신청해야 할 가능성 (대부분
   증권사가 그럼). 이 경우 tossctl 의 거래 기능은 official 백엔드에서 분기 처리 필요
3. **추상화 over-engineering** — 공식 표면을 직접 보기 전에 interface 를 짜면 잘못
   잡을 확률 ↑. (관련 결정은 아래 *결정 log* 참조)

## 결정 log

각 항목은 *그 시점의* stance. 이후 정보로 뒤집힐 수 있고, 뒤집힐 때 사전 공지하지
않습니다 — 새 항목을 추가할 뿐입니다.

- **2026-05-19** — issue #31 등록 (제보: @DaeHyeoNi). 사전 신청 페이지 확인. 사전 신청 진행. tossctl 의 일반화 (multi-broker) 방향을 *현재로서는* 선호. Phase 1 진입 전까지 코드 추상화는 보류 — 공식 표면을 보기 전 추상화는 잘못 잡을 확률이 크다는 판단
- **2026-06-02** — 공식 개발자 문서 + OpenAPI spec 공개 확인 (제보: @skyisle). `developers.tossinvest.com/docs` + `openapi.tossinvest.com/openapi-docs/.../openapi.json` (v1.0.3, 20 endpoints, OAuth2 Client Credentials). 모니터링 신호를 corp 페이지 chunk hash → spec version + endpoint 목록 hash 로 교체 (훨씬 강한 신호). endpoint 표면이 tossctl 명령과 거의 1:1 매핑 확인 → Phase 1 Broker 추상화 설계가 이제 *가능*. 단 client_id/secret 발급 콘솔 절차 미검증이라 실제 구현 착수는 토큰 직접 확보 후로 유지

## 외부 contributor / 사용자에게 부탁

1. 토스 Open API 토큰을 받으셨다면 issue #31 에 댓글로 알려주세요 — phase 1 진입
   판단의 가장 빠른 신호. 토큰/계정번호 같은 민감 정보는 절대 공개 댓글에 붙이지 마세요
2. 공식 API 의 endpoint/스펙 문서를 발견하시면 issue #31 에 링크 공유 환영
3. 이 문서의 timeline/계획에 의견 있으면 issue #31 댓글 또는 별도 PR 환영
