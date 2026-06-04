# Changelog

All notable changes to this project will be documented in this file.

## [0.5.2] - 2026-06-04

불필요한 마찰·비대칭 게이트 정리 + 성능(병렬화) 묶음.

### Changed
- **`trading.kr` config 게이트 제거 (schema_version 2→3)** — 비대칭이던 시장 게이트 삭제. KR 주문은 US 주문보다 위험하지 않으므로 시장을 **대칭 취급**: `trading.place` + `allow_live_order_actions` 가 US/KR 양쪽을 동일하게 게이트. (기존엔 KR만 별도 opt-in 필요 — 한국 증권사 CLI에서 거꾸로였음) 기존 config 의 `kr` 필드는 `grant` 처럼 무시되며 legacy 로 감지·경고. 거래 차단 강도 불변(`place`/`sell`/`fractional` + 마스터 + `--execute` + `--confirm`).
- **`monitor api` probe 병렬 실행** — 15개 probe 를 순차(최악 ~N×10s)에서 bounded 병렬(동시 8개)로 전환. daily-monitor cron wall-clock 대폭 단축. 결과는 probe 순서 그대로 반환.
- **`quote batch` 멀티 종목 병렬 fetch** — 종목별 순차 호출(각 다중 HTTP)에서 bounded 병렬(동시 6개)로 전환, 입력 순서 보존. `--live` 갱신 모드 응답성 개선. fail-fast 동작 유지.
- **`quote get` enrichment 병렬화** — productCode 해석 후 info/price/detail/v3-details 4개 호출을 순차에서 동시 실행(~4 RTT→1). info/price 에러는 fatal, detail/v3 는 non-fatal 유지.

### Fixed (UX)
- **KR 종목코드 시장 자동 판별** — `order place/preview` 에서 6자리 한국 종목코드(예: `005930`)를 넣으면 `--market kr` 를 안 줘도 자동으로 KR 시장으로 라우팅. (기존엔 에러로 거부) KR 코드는 US 티커와 겹칠 수 없어 안전. 거래 게이트(`--execute` + `--confirm <token>` + config)는 그대로.

### Migration
- 기존 `config.json` 에 `trading.kr` 이 있어도 자동 무시됩니다. 일반 명령 실행 시 stderr 경고 1줄로 안내되고 `config status`/`doctor` `legacy_config` 에서 감지됩니다. 제거하려면 해당 줄을 지우고 `schema_version` 을 3 으로 올리면 됩니다 (`config init` 후 재설정도 가능).

## [0.5.1] - 2026-06-04

0.5.0 직후 같은 날 진행한 후속 정리 모음 — 안전 모델 용어/잉여 게이트 정리 + config 경고 + 문서 일치화.

### Changed
- **거래 런타임 게이트 정리 (Tier 1 + Tier 2).** live mutation 의 런타임 게이트를 `--execute` + `--confirm <token>` **2개로 축소**. 거짓 이름이던 `--dangerously-skip-permissions` 를 **은퇴** — 0.5.0 에서 `internal/permissions` 패키지를 지웠으므로 가리킬 permissions 가 없었고, 의미도 "보호 skip" 이 아니라 "위험 opt-in" 으로 역방향이었으며, `--execute` 와 중복이었음. 실제 안전장치는 preview 를 봐야만 얻는 주문별 `--confirm <token>`. 영속 게이트(config: per-action + scope + `allow_live_order_actions`)는 그대로. 거래 차단 강도 동일.
- 내부 식별자 정리: `ErrDangerousExecuteDisabled` → `ErrLiveActionsDisabled`, `ErrDangerousFlagRequired` 제거, `ExecuteOptions.DangerouslySkipPermissions` 제거.
- README "Safety Model" 을 **mermaid flowchart** 로 도식화 (영속/런타임 게이트 분리 + 거부 경로).

### Added
- **config legacy 자동 경고** — config 에 더 이상 쓰이지 않는 legacy 필드(예: `trading.grant`)가 있거나 스키마 버전이 바이너리보다 낮으면, `config status`/`doctor` 를 직접 돌리지 않아도 일반 명령 실행 시 stderr 에 1줄 경고. update notice 와 동일한 24h backoff 로 매 호출 반복 방지. JSON 출력·`config`/`doctor`/`version`/`help` 명령에서는 무음. 단위 테스트 포함.

### Deprecated
- `--dangerously-skip-permissions` 플래그 — no-op 으로 동작하며 사용 시 deprecation notice 출력. 한 릴리즈 동안 기존 스크립트/agent 호환을 위해 받아들임. `--execute` + `--confirm <token>` 으로 대체.

### Docs
- README·architecture·configuration·examples 문서를 새 게이트 체인과 0.5.0 스키마 변경(제거된 `trading-permission.json`, `order permissions` 명령, grant 참조)에 맞춰 일치화.
- 전략 문서에서 "해자" 표현 제거 — 중립적 표현(고유 범위/강점)으로 교체.

## [0.5.0] - 2026-06-04

공식 Open API 에 없는 web 전용 표면 대거 확장 + 첫 mutation 기능(관심종목 관리)
+ 거래 안전 모델 간소화. minor 버전 bump.

### Added
- **`market index`** — 주요 시장 지수 (코스피·코스닥·나스닥·S&P500·필라델피아 반도체·VIX·다우 등) 현재가·변동·변동률. **공식 API 에 없음.**
- **`market ranking --size N`** — 실시간 인기 종목 순위. 공식 API 에 없는 discovery 표면.
- **`quote flows <symbol>`** — 수급 (개인·외국인·기관 일별 순매수, KR 전용). 공식 API 에 없는 프리미엄 표면. US 종목은 친절한 거부.
- **`market signals`** — 토스증권 AI 시그널 (종목별 AI 분석 시그널·키워드·등락). 공식 API 에 없고 repo hero(AI 연결)와 정합하는 차별 표면.
- **`market screener [id] --nation kr\|us`** — 조건 검색 (가치주·배당주·성장주·쌍끌이 매수 등 프리셋). 인자 없으면 프리셋 목록, id 주면 해당 조건의 종목 반환. `--filter '<json>'` 으로 커스텀 raw 필터도 지원. 공식 API 에 없는 discovery 표면. (POST 필터 body 리버스 엔지니어링)
- **관심종목 관리 (첫 mutation 기능)** — `watchlist groups` (폴더 목록), `watchlist group create\|rename\|delete`, `watchlist add\|remove <symbol> --group <id>`. 공식 API 에 없는 web 전용 표면. 비금융·되돌림 가능이라 거래 권한 게이트와 별개로 가볍게 동작. `new-watchlists` namespace 의 POST/PATCH/DELETE body 를 실제 Chrome(channel=chrome) 캡처 + 자기계좌 경험적 검증으로 리버싱. X-XSRF-TOKEN 자동 적용. 계약 잠금 httptest 포함.
- `monitor api` 에 public/auth probe 추가 (market-index, stock-ranking, trading-flows, ai-signals, screener-presets, watchlist-groups).

### Changed
- README "지원 범위" 를 **`공식 API` / `tossctl` 칼럼 ✅/❌/🔸 매트릭스**로 재편 (조회 + 거래 모두). 토스 공식 Open API 가 사전 신청 단계(미출시)임을 명시. 공식 API 가 ❌ 인 행 = tossctl 고유 범위: 수급·지수·인기순위·AI시그널·스크리너·watchlist 관리·ledger·overview·CSV·push·멀티시세·소수점주문·dry-run preview 등. 앞으로 표면 추가 시 이 매트릭스 유지.
- **거래 안전 모델 간소화** — 중복이던 TTL grant 레이어 제거. `allow_live_order_actions` 마스터 스위치가 동일 보호를 이미 하므로, 남은 게이트(per-action 토글 + 마스터 스위치 + `--execute` + `--dangerously-skip-permissions` + confirm token)로 충분. 거래 차단 강도는 동일, 표면만 −500 라인.

### Removed
- `tossctl order permissions` (grant/status/revoke) 명령 + `internal/permissions` 패키지 + `trading-permission.json`. `allow_live_order_actions` 와 중복.

### Internal
- `internal/client/marketdata.go` 에 `GetMarketIndices` / `GetStockRanking` / `GetTradingFlows` / `GetAISignals` / screener 메서드 추가. `internal/client/watchlist_manage.go` (mutation). output formatter + 단위 테스트 (US 거부, signed/comma, 계약 잠금 httptest).

## [0.4.19] - 2026-06-03

### Added
- **`quote get` 정보 대폭 확장** — 기존 현재가/변동/거래량에 더해 당일 OHLC, 52주 고저, 시가총액, 거래대금, 체결강도, 상/하한가 추가. web `v3/stock-prices/details` 를 enrichment 용으로 별도 호출 (실패해도 기본 quote 는 동작 — graceful). agent 가 "삼성 52주 고저/시총 알려줘" 류 질의에 직접 응답 가능.
- **`market fx`** — 달러 환율·달러 인덱스 등 FX/지수 조회. (공식 Open API 의 exchange-rate 대응)
- **`market hours` 가 다음 영업일도 표시** — 오늘 휴장(예: 선거일)일 때 KR/US 다음 개장일·시간 자동 노출.

### Internal
- `quote get` 의 주문 흐름(trading)은 기존 v1 endpoint 유지하고 enrichment 만 v3 로 분리 — blast radius 최소화. 신규 fixture + 단위 테스트 추가.

## [0.4.18] - 2026-06-03

### Fixed
- **`quote limits` 가 미국 종목에 raw 400 에러** 대신 친절한 메시지. 상/하한가는 KRX 전용 제도라 (미국장은 일일 가격제한 없이 LULD circuit breaker 사용) 비-KR 종목은 네트워크 호출 전에 명확한 안내로 거부. 단위 테스트 추가.

## [0.4.17] - 2026-06-03

### Added
- **`quote trades <symbol>`** — 최근 체결 틱 조회 (시각·체결가·수량·매수/매도 구분). `--count` 로 개수 조절. (web `/api/v2/stock-prices/{code}/ticks`)
- **`quote limits <symbol>`** — 당일 상/하한가 조회. (web `/api/v2/stock-prices/{code}/upper-lower`)
- **`quote warnings <symbol>`** — 매수 유의사항 badge (정리매매·투자경고/위험·VI 등). badge shape 이 동적이라 type/title/text/level 매핑 + raw 보존. (web `/api/v1/stock-infos/{code}/wts-badges`)
- **`market hours`** — 오늘 KR·US 장 운영 시간 (개장/마감, 휴장 표시). (web `/api/v2/system/trading-hours/integrated`)
- 위 4개는 모두 토스 공식 Open API (issue #31) 의 Market Data 카테고리에 대응하는 기능 — 공식 토큰 발급 없이 기존 web session 으로 동작. table / json / csv 지원.
- `monitor api` 에 3개 public probe 추가 (quote-trades, quote-price-limits, market-trading-hours) — 회귀 조기 감지.

### Internal
- `internal/client/marketdata.go` 신설 (4개 메서드 + `investModeFor` KR/US 분기). `internal/output/marketdata.go` 신설 (formatter). 단위 테스트 추가.

## [0.4.16] - 2026-05-28

### Added
- **`tossctl quote batch --live`** — `watch`/`viddy` 대체용 갱신 모드. alternate screen buffer (`\033[?1049h`) 로 깔끔한 화면 전환, 더블 버퍼링으로 깜빡임 방지, `--interval` 으로 갱신 주기 (기본 2초, 최소 1초 클램핑). `signal.NotifyContext` 로 Ctrl+C 안전 종료. 에러 시 마지막 성공 결과 보존 + 빨간 에러. **non-tty 환경에서 거부** — cron/agent pipe 에서 ANSI escape 오염 방지. (PR #33, author: @Castor103)
- **`quote batch` 의 comma-separated symbol 인자** — `tossctl quote batch "삼성전자,KB금융,현대차"` 처럼 한 인자에 여러 종목 가능. agent 가 array 를 join 해서 한 번에 넘기기 좋음.

### Internal
- `parseBatchSymbols` 단위 테스트 추가 (single/space-sep/comma-sep/mixed/whitespace/empty 케이스).

## [0.4.15] - 2026-05-20

### Added
- **`tossctl quote chart <종목>`** — 1/3/5/10/15/30/60분봉 ASCII 캔들 차트. 반블록 문자로 세로 해상도 2배, 한국 호가단위 자동 적용, ANSI 색상 (양봉 빨강 / 음봉 파랑). table · json · csv 모두 지원. (PR #32, author: @Castor103)
- **`tossctl quote batch ... --chart`** — 종목별 스파크라인을 테이블 우측 컬럼으로 표시. 차트 API 실패 시 warning 만 출력하고 나머지는 정상 표시.
- `quote get` / `quote chart` 가 복수 단어 종목명 지원 (`"KODEX 인버스"` 등).

### Changed
- 차트/스파크라인 ANSI 색상이 stdout 이 tty 가 아니거나 `NO_COLOR` 환경변수가 설정되어 있으면 자동 비활성화. cron · agent pipe · redirect 환경에서 raw escape sequence 가 출력에 섞이지 않음.

### Internal
- `internal/client/chart.go` 의 `normalizeChartInterval`, `deriveSecurityType` 단위 테스트 추가.
- `internal/output/chart.go` 의 `tickSizeKR`, `formatPriceKR`, `renderSparkline` 단위 테스트 추가.

## [0.4.14] - 2026-05-14

### Changed
- **자동 업데이트 알림이 AI 에이전트에도 surface 되도록 확장.** v0.4.13 의 tty-only 게이트가 모든 비-tty 호출 (Claude Code, Codex, OpenClaw 등 모든 agent 통한 호출 포함) 에서 알림을 숨겨, repo description 인 "AI 에이전트를 토스증권에 연결하는 비공식 CLI" 와 모순이었음. 이제 tty 검사 제거 + `internal/updatecheck` cache 에 `update_notified_at` 추가로 24h 1회만 출력 — cron/loop 호출에서도 노이즈 누적 없음. JSON/CSV 출력은 여전히 자동 skip (구조화 파이프라인 보호).
- **세션 만료 경고에 1시간 backoff 추가.** 동일 원칙. 24h 이내 만료 윈도우 안에서 매 명령 호출마다 stderr 한 줄이 떴던 것을 1시간 1회로. `internal/updatecheck` cache 의 `expiry_notified_at` 활용 — 별도 cache 파일 추가 없음.

### Added
- **`tossctl version` 출력에 `latest` 표시.** JSON 모드의 `latest` + `update_available` 필드, table 모드의 `latest:` 줄. agent 가 `tossctl version --output json` 만 parse 해도 새 버전 사실을 사용자에게 surface 가능. `update_check.enabled=false` 와 무관하게 version 명령은 항상 표시 (사용자가 명시적으로 버전 컨텍스트를 요청한 것이므로).
- **`tossctl config show` 의 `Update Check` 줄 + JSON `update_check`.**  CSV 헤더에 `update_check_enabled` 추가.

## [0.4.13] - 2026-05-14

### Added
- **자동 업데이트 알림** — 명령 실행 후 (성공 시) GitHub Releases 의 latest stable tag 를 조회해 새 버전이 있으면 stderr 에 한 줄 안내. 24h 디스크 캐시 (`<cache>/tossctl/update-check.json`), 네트워크 실패는 silent (다음 실행 때 재시도). config `update_check.enabled` (기본 `true`) 로 끌 수 있음. JSON/CSV 출력, non-tty (cron 등), dev 빌드, prerelease tag 에서는 자동 skip — 자동화 출력은 절대 오염시키지 않음.
- `internal/updatecheck` 패키지 + 단위 테스트 (semver 비교, 캐시 hit/miss, 네트워크 실패 시 stale fallback).

### Changed
- `config show` / `doctor` 의 status 출력에 `update_check` 표시 (schema_version 은 그대로 2 — 신규 필드는 additive 라 마이그레이션 불필요).

## [0.4.12] - 2026-05-14

### Fixed
- **Windows 에서 `tossctl auth login` 이 `AttributeError: module 'os' has no attribute 'fchmod'` 로 끊기던 회귀 (#26)** — `os.fchmod` 는 Unix 전용이라 Windows Python 에서는 노출되지 않음. auth-helper 의 권한 설정을 `_set_private_permissions` 헬퍼로 통일하고, fchmod 가 없으면 chmod 로 fallback. chmod 의 fd 인자 지원도 platform-dependent 라 (`os.supports_fd`), 일부 Windows 환경에서 `TypeError`/`OSError` 가 발생하면 best-effort 로 silently skip — 권한 설정은 못 해도 storage-state 파일 저장은 정상 진행. 회귀 검증 단위 테스트 3개 추가. (제보 + 수정: @netics01)

## [0.4.11] - 2026-05-13

### Changed
- `monitor` 모듈 잔여 over-abstraction 정리. `expectStatus` 의 사용 안 되는 `body` 파라미터 제거 (호출자 6곳 정리), `printResults` 의 인라인 익명 인터페이스를 `io.Writer` 로 치환, `monitor api --help` 의 Discord-specific 예시 제거 (alert recipe 는 `AGENTS.md` 한 곳으로 단일화). `Probe` 필드 인라인 주석 정리. 동작 변화 없음 — 코드만 unix 철학에 맞게 더 좁힘.

## [0.4.10] - 2026-05-13

### Changed
- `monitor api` 시그니처 단순화: `--webhook` / `TOSSCTL_MONITOR_WEBHOOK` 제거. exit 0/1 만 반환하고 알림 채널 (Discord · Slack · ntfy · macOS notification · 이메일 등) 은 cron 라인의 `|| <command>` 우항에서 사용자가 합성. 합성 recipe 는 신규 `AGENTS.md`.

### Removed
- `internal/monitor/discord.go` (Discord webhook 송출 헬퍼). 동등한 효과는 한 줄 `curl` 합성으로 충분.

### Added
- `AGENTS.md` — OpenClaw · Claude Code · Codex · Cursor 등 AI 에이전트가 `monitor api` 와 알림 채널을 cron 으로 묶을 때 참고할 짧은 recipe 모음.

## [0.4.9] - 2026-05-13

### Added
- **`tossctl monitor api`** — 6개 read-only endpoint (account-list, summary-overview, positions, watchlist, quote, pending-orders) 에 schema-invariant probe 실행. 핵심 JSON 경로/타입 검증이 실패하면 exit 1 + 선택적 Discord webhook 알림 (`--webhook` 또는 `TOSSCTL_MONITOR_WEBHOOK`). 본인 세션으로 본인 머신에서 본인이 설정한 webhook 한 곳에만 보고. #29 같은 토스 서버측 body 계약 변경을 cron 으로 조기 감지하기 위한 도구. 설정 가이드: `docs/operations.md`.

### Changed
- `expectStatus` 가 webhook 알림에 status code + 기대값만 노출 (응답 본문 fragment 미포함). 단위 테스트 `TestProbeChecksDoNotLeakResponseBodyOnFailure` 로 회귀 방지.

## [0.4.8] - 2026-05-13

### Fixed
- **`portfolio positions` · `watchlist list` 회귀 (#29)** — 2026-05-13 16:10 KST 즈음 토스 서버가 `/api/v2/dashboard/asset/sections/all` 의 body 계약을 변경. 기존 빈 `{}` body는 여전히 200을 반환하지만 sections가 빈 배열 + `pollIntervalMillis: 3000` 만 내려와, CLI가 "SORTED_OVERVIEW section not found" / "WATCHLIST section not found"로 실패했음. 새 계약은 `{"types":[<section-name>]}` 필터 필수. 두 커맨드 모두 해당 필터를 송출하도록 수정. (제보: kwakmu18, #29)
- 라이브 캡처 결과 `docs/reverse-engineering/rpc-catalog.md` 에 body 형식 명시.

### Notes
- 다른 read-only API (`account`, `orders`, `quote`, `transactions`, `account summary` 등)는 이번 토스 변경에 영향 없음이 같은 세션으로 확인됨.

## [0.4.7] - 2026-05-06

### Fixed
- **`order place --fractional --currency-mode USD` 거부 (#28)** — 와이어 페이로드는 항상 `currencyMode="KRW"`인데 (토스 스펙) Fractional 분기는 `intent.Amount`를 USD/KRW 구분 없이 그대로 `orderAmount`에 실어 보내고 있었음. 결과적으로 `--amount 100 --currency-mode USD`가 서버에 ₩100으로 전달되어 "금액주문은 $1 또는 1,000원 이상" 오류가 났음. 이제 USD 모드일 때는 `intent.Amount * meta.ExchangeRate`로 KRW 변환 후 송출. `--currency-mode KRW`(기본) 동작은 그대로. (제보: @leesj10147)

## [0.4.6] - 2026-05-06

### Added
- **`tossctl auth extend`** — 폰 토스 앱 푸시 승인을 통해 서버 측 세션 만료를 연장합니다. 블로킹 + Braille 스피너 UX, `--timeout` (기본 120초), `Ctrl+C` 처리. 사용 endpoint chain: `POST /api/v1/wts-login-extend/doc/request` → polling `GET /doc/{txId}/status` (`REQUESTED` → `COMPLETED`) → `POST /api/v1/wts-login-extend/{txId}/state` (필수 finalize) → `GET /api/v1/session/expired-at`. 약 7일 연장됩니다. 기여: @skyisle (PR #27)
- **24h 만료 경고** — 모든 명령 시작 시 서버 측 세션 만료가 24시간 미만으로 남았으면 stderr 한 줄로 경고 (`⚠ session expires in ~21h 58m; run \`tossctl auth extend\` to renew`). `--output json` 모드와 `auth/version/help` skip-list에서는 silent.
- **`session.json` 에 `server_expires_at` 필드 추가** — 서버 측 ~7일 활성 만료 시계 (기존 `expires_at` 의 1년 쿠키 만료와 별개). `auth status` table에 `Server Expiry: YYYY-MM-DD HH:MM KST` 줄 추가, JSON 출력에 `server_expires_at`.

### Changed
- `ExtensionTimeoutError{Elapsed}` 타입 도입 — 기존 `fmt.Errorf("%w (waited %s)", ...)` + caller-side `extractParenDetail` 문자열 파싱을 typed wrapper로 대체. 에러 메시지 포맷이 바뀌어도 elapsed 추출이 안 깨짐.
- `formatKST(t)` helper 추출 — `auth status` / `auth extend` 출력의 `2006-01-02 15:04 KST` 포맷이 한 곳에서 관리됨.

### Fixed
- v0.4.0의 persistent SESSION 쿠키(1년)는 정상이어도 토스 서버 측 ~7일 활성 timer가 별도로 흐르는 사실이 그동안 묵묵히 401을 일으키고 있었음. `auth extend`로 매번 `auth login`(QR + 폰 풀 플로우)을 다시 돌지 않고 폰 푸시 한 번으로 연장 가능. (참고: `docs/reverse-engineering/auth-notes.md` Session Lifetime 섹션)

## [0.4.5] - 2026-04-29

### Added
- **`tossctl push listen` — SSE 푸시 리스너** — 토스증권 웹이 `GET https://sse-message.tossinvest.com/api/v1/wts-notification` 로 제공하는 Server-Sent Events 채널 구독. 세션 쿠키 재사용, JSONL 로 stdout 출력, 기본 exponential backoff 재연결 (`--retry=false` 로 비활성 가능), 토스 서버의 graceful `event: connection-close` 핸드오프 즉시 재연결. Toss 는 thin notification 방식이라 payload 대신 `{"type":"pending-order-refresh",...}` 같은 "재조회하라" 신호만 내려줌 (관찰된 이벤트 타입: `pending-order-refresh`, `purchase-price-refresh`, `share-holdings`, `web-push` 등). 구독 가능한 이벤트 타입과 의미는 `docs/reverse-engineering/push-events.md` 에 정리. 기여: @skyisle (PR #25)

### Changed
- **공통 User-Agent 상수 통합** — `client.defaultBrowserUserAgent` (private) 를 `client.DefaultBrowserUserAgent` (export) 로 승격하고 `internal/push` 도 동일 상수 재사용. 기존엔 두 패키지에 동일 Chrome UA 문자열이 중복돼 있어 향후 Chrome 버전 bump 시 두 곳을 동기화해야 했음 — 이제 한 군데만 수정하면 `wts-api`/`wts-cert-api`/`wts-info-api`/`sse-message` 전 채널 fingerprint가 같이 갱신됨.

## [0.4.4] - 2026-04-23

### Added
- **US 지정가 주문에서 USD 가격 입력 허용** — `order place --currency-mode USD --price 158.01 ...`. 기존엔 비 fractional 경로가 `CurrencyMode=KRW`만 허용해 `Live Ready=false`로 떨어졌음. 입력만 USD로 받고 와이어 페이로드는 기존과 동일하게 `currencyMode="KRW"`+USD 가격 필드 (캡처된 토스 웹 UI 스펙과 일치). `--currency-mode KRW`(기본)는 환율 변환 경로 그대로 유지. 기여: @skyisle (PR #24)

### Changed
- `buildPlaceBody`의 USD branch 가드를 `case intent.Market == "us" && intent.CurrencyMode == "USD"`로 명시화 — 향후 market 추가 시 암묵적 fall-through 방지 (현재 `placeIntentSupported` 화이트리스트로 안전하지만 defense-in-depth)

## [0.4.3] - 2026-04-23

### Removed
- **죽은 config 토글 3개 제거** — 모두 코드 어디서도 참조되지 않아 UX 오해만 유발하던 필드. 기존 config에 남아있어도 무시되며 `doctor`의 `legacy_config`에서 감지됨:
  - `trading.grant` — `order permissions grant` 커맨드 게이트처럼 보였지만 실제 매매 게이트는 `place`/`cancel`/`amend` + `allow_live_order_actions`만 확인. `grant=false, place=true` 설정 시 매매는 되면서 grant만 막히는 기묘한 조합 가능하던 문제
  - `dangerous_automation.complete_trade_auth` — 어떤 handler에도 연결되지 않은 dead code. `true`로 켜면 doctor에서 경고만 뜨고 기능은 무변화
  - `dangerous_automation.accept_product_ack` — 동일 (dead code)

### Changed
- `order permissions grant` 실행 조건 단순화 — `place` 또는 `cancel` 또는 `amend` 중 하나라도 허용되어 있으면 실행 가능 (기존: 별도의 `grant` 토글 추가 요구)
- README / `docs/configuration.md` 토글 섹션 재구성 — "경로 게이트" (`place`/`cancel`/`amend`, broker API 분기) vs "스코프 선언" (`sell`/`kr`/`fractional`, 유저 자가 제한) 두 범주로 명확히 구분해서 설명. 기존에 "안전 게이트"로 묶여있어 오해 소지 있던 부분 정리
- `config.schema.json` 업데이트 — 제거된 필드들 `required`에서 빠짐

### Migration
- 별도 조치 불필요. v0.4.2에서 오던 config를 그대로 두면 됨. `doctor` 돌리면 `legacy_config` info에 "trading.grant / complete_trade_auth / accept_product_ack는 v0.4.3부터 무시됨"이 표시됨
- 혹시 `trading.grant=true`에만 의존하고 실제 매매 토글(`place` 등)은 꺼둔 상태였다면, 동작상 차이 없음 — 예전에도 실제 매매는 불가능했음

## [0.4.2] - 2026-04-23

### Added
- **`tossctl doctor --report`** — JSON 진단 번들. 기존 doctor 정보 + `wts-api` · `wts-cert-api` · `wts-info-api` 각 1회 실시간 probe (200/401/403 + 응답시간) + 파일 권한 audit (`session.json`/`config.json`/... 가 0600, 디렉토리 0700인지) + orphan intermediate 파일 탐지를 한 덩어리로 출력. 홈 디렉토리 경로는 `~`로 자동 redact되어 사용자명이 노출되지 않으므로 GitHub 이슈에 그대로 첨부 가능.
- Issue 템플릿 `bug_report.yml`에 `tossctl doctor --report` 필드 추가 — 제보자가 실행 결과를 JSON으로 붙이면 유지보수자가 한눈에 환경·세션·엔드포인트 상태를 파악할 수 있음. 기존 `--report` 없이는 각 이슈마다 "어떤 endpoint가 403?" 을 개별로 물어봐야 했던 오버헤드 제거.

### Notes
- v0.4.1 이전에 생성된 tossctl 디렉토리(`~/Library/Application Support/tossctl/` 등)는 여전히 `0755` 모드로 남아있을 수 있음. `doctor --report`의 `file_modes` 항목에서 `expected: "0700"` 과 비교해 `ok: false`로 표시되며, 원하는 사용자는 `chmod 0700 <dir>` 로 수동 정리 가능. 기능에 영향 없음.

## [0.4.1] - 2026-04-23

보안 하드닝 릴리즈. 전체 시스템을 점검하여 기능 영향 없이 좁힐 수 있는 부분만 적용.

### Security
- **Intermediate storage-state 파일 권한 0o600 고정** — Python helper가 저장하는 `~/Library/Caches/tossctl/auth/playwright-storage-state.json`은 이전엔 default umask(보통 0o644)를 따라 쓰였음. `os.open(..., O_CREAT|O_WRONLY|O_TRUNC, 0o600)` + `os.fchmod(0o600)`로 기존 파일 유무와 관계없이 항상 소유자만 읽도록 변경. Go CLI가 session.json으로 복사하기 전 찰나의 창에 같은 호스트의 다른 로컬 사용자가 전체 쿠키 세트를 읽을 수 있던 공격면 차단.
- **`--qr-output` PNG 권한 강제 0o600** — 동일 `fchmod` 패턴 적용. 기존 파일이 이미 0o644로 존재해도 overwrite 시 명시적으로 좁힘.
- **Intermediate storage-state 파일 로그인 성공 후 자동 삭제** — `LoginWith`가 `ImportPlaywrightState` 성공 직후 `os.Remove(result.StorageStatePath)` 호출. 같은 쿠키의 중복 사본이 cache dir에 무기한 남던 문제 해결 (`auth logout`은 session.json만 지웠음). 사용자가 직접 부르는 `auth import-playwright-state <path>`는 경유하지 않으므로 외부 파일은 그대로 유지.
- **tossctl 상태 디렉토리 권한 0o755 → 0o700** — `session/store.go`, `config/service.go`, `permissions/service.go`, `orderlineage/service.go`의 `os.MkdirAll` 모드 좁힘. macOS `~/Library/Application Support`는 부모 디렉토리가 이미 0o700이라 영향 미미하지만 Linux/CI 환경에선 같은 호스트의 다른 사용자가 `tossctl/` 디렉토리 목록을 열람 가능하던 문제 차단.
- **`AuthError`에서 응답 본문(Body) 필드 제거** — `wts-api` / `wts-cert-api`의 401/403 응답 본문엔 CSRF 진단이나 세션 식별자 조각이 포함될 수 있음. 현재 어떤 caller도 `AuthError.Body`를 읽지 않지만, 향후 `%+v` 디버그 로그나 에러 값 직렬화 시 유출될 수 있어 필드 자체를 제거. `StatusError.Body`는 trading broker 메시지 분류에 실제로 사용되므로 유지.

## [0.4.0] - 2026-04-23

### Added
- **거래내역 ledger + cash overview** — `tossctl transactions list --market kr|us`로 매매, 입출금, 배당, 주식 입출고를 조회. `--from/--to`, `--filter all|trade|cash|inout|cash-alt`, `--all` 페이지네이션 지원. `tossctl transactions overview --market kr|us`는 주문가능/출금가능/예정입금 요약. table/JSON/CSV 출력 지원 (Toss 200일 단일쿼리 캡 반영). 기여: @skyisle (PR #20)
- **영속(Persistent) 세션 캡처** — `auth login` 이 폰의 "이 기기 로그인 유지" 2차 인증까지 기다린 뒤 storage state 를 저장. 2차 인증 완료 시 Toss가 장기 SESSION 쿠키를 발급하므로 서버 idle timeout 면제 (≈1시간 후에도 401 안 남). `auth status` / `auth import-playwright-state` 출력에 `Persistence` 필드 추가 (`persistent (expires ...)` 또는 `session-scoped (≈1h idle timeout)`). JSON 출력에 `expires_at`, `persistent` 필드. 기여: @skyisle (PR #23)
- **원격/헤드리스 로그인** — `tossctl auth login --headless`. QR 탭 자동 활성화 + `/api/v2/login/wts/toss/{qr,status}` 응답 인터셉트로 QR URL과 확인 문자(answerLetter)를 stderr 출력. 텔레그램 등으로 URL만 폰에 보내 탭하면 Toss 앱이 열려 카메라 없이 인증. PNG 파일 저장은 `--qr-output <path>` (0600 권한). 기여: @skyisle (PR #22, 보안 강화 후 merge)
- **uv-managed Python 우선 탐지** — `auth login`이 helper Python을 찾을 때 `$TOSSCTL_AUTH_HELPER_PYTHON` → uv tool 관리 Python (`$UV_TOOL_DIR`, `$XDG_DATA_HOME/uv/tools`, `~/.local/share/uv/tools`, Windows `%APPDATA%/uv/tools`) → PATH의 `python3` 순서로 선택. `uv tool install ./auth-helper`로 전역 Python 오염 없이 helper 실행 가능. 기여: @keenranger (PR #21)

### Fixed
- **1시간 뒤 401 재발** — 과거 `auth login` 이 QR 1차 인증 직후 종료하여 session-scoped SESSION 만 저장했고, 약 1시간 idle 후 서버가 세션을 invalidate 하던 문제. "이 기기 로그인 유지" 2차 확인까지 기다려 persistent SESSION 저장하도록 변경. (참고: `docs/reverse-engineering/auth-notes.md` — Session Lifetime 섹션)

### Security
- 헤드리스 로그인의 `--qr-output` 파일을 `0o600` 권한으로 배타 쓰기 — 공유 머신에서 다른 사용자가 PNG 읽고 먼저 로그인 탭을 완료하는 시나리오 차단
- QR 응답 인터셉트가 path뿐 아니라 host(`wts-api.tossinvest.com`)까지 검증 — 동일 path suffix의 타 origin 응답 파싱 방지

## [0.3.6] - 2026-04-17

### Fixed
- **auth login 무한 대기 해결** — Python helper가 `DEVICE_INFO` localStorage 키를 필수로 기다리던 체크 제거. 토스 웹이 해당 키를 더 이상 보장하지 않아 로그인 성공 감지가 실패하던 회귀 수정 (Fixes #17, thanks to @pinion05)
- **`wts-api` 403 차단 해결** — `applySession`에 브라우저형 기본 `User-Agent` 설정. 기본 Go HTTP User-Agent(`Go-http-client/1.1`)가 토스 서버에서 핑거프린팅으로 차단되어 `account/*`, `portfolio/*`, `quote/*` 호출이 403을 받던 문제 해결. `auth login` 직후/10분 후 모두 정상 동작 확인 (Fixes #15, #17)

### Notes
- 명시적으로 `User-Agent`가 설정된 요청은 override되지 않고 그대로 유지됨

## [0.3.5] - 2026-03-30

### Added
- **테이블 출력 개선** — `portfolio positions`, `orders list`, `watchlist`, `quotes` 명령의 table 출력을 정렬된 컬럼 형식으로 변경. 종목명 좌측 정렬, 숫자 우측 정렬, 천단위 쉼표 적용
- `CONTRIBUTING.md` 추가

## [0.3.4] - 2026-03-28

### Fixed
- **auth login 브라우저 차단 해결** — Playwright 번들 Chromium 대신 시스템 Google Chrome 사용 (`channel="chrome"`). 토스증권이 `Sec-Ch-Ua` 헤더에서 `"Google Chrome"` 브랜드를 확인하도록 변경되어 Chromium이 차단됨 (Fixes #13)

### Changed
- `tossctl doctor` 브라우저 체크가 Chromium 대신 Chrome 감지로 변경
- `playwright install chromium` 불필요 — 시스템에 Google Chrome만 설치되어 있으면 됨

## [0.3.3] - 2026-03-24

### Added
- **USD 표시** — US 포지션에 매입가/현재가/평가금/손익을 USD로 병기 (by @seilk, PR #11)
- **설치 스크립트** — `curl -fsSL .../install.sh | sh` 한 줄 설치 (macOS/Linux)
- Issue/PR 템플릿, GitHub Sponsors 지원

### Fixed
- install.sh가 auth-helper를 누락하여 Linux에서 `auth login` 실패하던 문제 (Fixes #12)

## [0.3.2] - 2026-03-23

### Added
- **Cross-platform release builds** — Windows (amd64), Linux (amd64/arm64) 바이너리 자동 빌드
- Quick Start에 Windows/Linux 설치 가이드 추가

### Changed
- README Quick Start를 macOS/Linux/Windows 플랫폼별로 재구성
- 설치 섹션 중복 제거, Quick Start로 통합

### Docs
- architecture.md 갭 목록 업데이트 — sell, KR, fractional 구현 완료 반영
- README disclaimer 강화 (TOS 위반 가능성 명시)

## [0.3.1] - 2026-03-21

### Fixed
- US stock price rounding: `round4` → `round2` — prices now round to $0.01 (cent) precision instead of $0.0001, fixing `invalid.limit.price.scale` API errors
- `placeIntentSupported()` now accepts USD currency mode for fractional orders

### Changed
- README rewritten — restructured around feature tables, added fractional/KR examples, removed outdated sections, cleaner config reference

## [0.3.0] - 2026-03-21

### Added
- **Fractional (소수점) order support** — `tossctl order place --symbol TSLL --fractional --amount 18000`
  - US market only, market orders (시장가), amount-based (금액 기반)
  - `trading.fractional` config toggle (default: false)
  - `--amount` flag for specifying KRW amount
  - `--fractional` flag auto-selects market order type
- Fractional policy gate in `Place()` with "disabled by config" error
- `buildPlaceBody` fractional branch: `price=0, quantity=0, orderAmount=<KRW>, orderPriceType=01, isFractionalOrder=true`
- `placeIntentSupported()` now accepts fractional orders (US + market only)
- `NormalizePlace` validates fractional constraints (US only, amount required, auto market order)
- 10 new tests: fractional capability, policy, preview, payload, orderintent validation
- API compatibility verified via prepare dry-run (422 = payload accepted, insufficient balance)

## [0.2.3] - 2026-03-21

### Removed
- **MCP server** (`tossctl-mcp`) — CLI 자체가 AI 에이전트에서 직접 실행 가능하므로 불필요한 추상화 제거
- `make build-mcp` Makefile 타겟
- Release workflow에서 tossctl-mcp 바이너리

## [0.2.2] - 2026-03-21

### Added
- `tossctl quote batch <symbol> [symbol...]` — fetch multiple stock quotes at once
- `tossctl export positions --market kr|us|all` — filter exported positions by market
- `tossctl export orders --market kr|us|all` — filter exported orders by market
- Quote output tests (6 test cases)

### Fixed
- Floating point display artifacts in quote batch table output (e.g., `-0.8500000000000014` → `-0.85`)

## [0.2.1] - 2026-03-21

### Added
- MCP server unit tests (10 test cases) covering initialize, tools/list, tool calls, error handling
- Refactored MCP server to testable pure functions (handleMethod, buildInitializeResponse, buildToolsList)

### Removed
- Unused `stub.go` command helper (export commands now fully implemented)

## [0.2.0] - 2026-03-21

### Added
- **MCP server** (`tossctl-mcp`) — read-only Model Context Protocol server for AI agent integration
  - `get_portfolio_positions` — 보유 포지션 조회
  - `get_account_summary` — 계좌 요약 조회
  - `get_quote` — 종목 시세 조회 (US/KR)
  - `list_pending_orders` — 미체결 주문 조회
  - `list_completed_orders` — 체결 완료 내역 조회 (market filter 지원)
  - `list_watchlist` — 관심 종목 조회
- `tossctl export positions` — CSV 포지션 내보내기 (stub에서 실제 구현으로 전환)
- `tossctl export orders` — CSV 체결 내역 내보내기 (stub에서 실제 구현으로 전환)
- `make build-mcp` Makefile 타겟
- Release workflow에 `tossctl-mcp` 바이너리 포함

## [0.1.7] - 2026-03-21

### Added
- Korean stock (국내주식) trading support — `tossctl order place --symbol 005930 --market kr`
- `trading.kr` config toggle (default: false) — KR requires `trading.place` and `trading.kr`
- KR branch in `buildPlaceBody`: raw KRW price, no USD conversion, no `allowAutoExchange`
- KR branch in `PlacePendingOrder`: skip USD exchange rate fetch and FX consent
- `InferMarketFromStockCode()` for cancel/amend market recovery from stock code pattern
- KR symbol detection in `NormalizePlace`: numeric 6-digit + market=us → error with guidance
- 13 new test cases (T1-T13) for KR gate, preview, config, payload, reconciliation, symbol detection
- KR cancel/amend live verification TODO

### Changed
- `placeIntentSupported()` now allows both "us" and "kr" markets (was us-only)
- `Place()` reordered: capability → market policy → side policy → execution guard (was guard-first)
- Reconciliation functions parameterized: `Market: "us"` hardcoding (8 places) → market parameter
- TODOS.md: reconciliation parameterization marked as completed
- README, architecture.md, configuration.md updated with KR documentation

## [0.1.6] - 2026-03-21

### Added
- Sell order support for US limit / KRW / non-fractional — `tossctl order place --side sell`
- `trading.sell` config toggle (default: false) — sell requires both `trading.place` and `trading.sell` to be enabled
- Sell policy gate in `Place()` with distinct "disabled by config" error (not "unsupported")
- Sell state reflected in `PreviewPlace` warnings and `MutationReady`
- Sell toggle visible in `config show`, `doctor`, and `EnabledActions()`
- 10 new test cases (T1-T10) covering sell gate, preview, config parsing, payload, and error messages
- `TODOS.md` for tracking deferred work (reconciliation parameterization, sell live verification)

### Changed
- `placeIntentSupported()` no longer restricts by side — both buy and sell are broker-capable
- Warning message updated: "US buy/sell limit orders" (was "US buy limit orders")
- `ErrPlaceUnsupported` message no longer includes `--side buy`
- Doctor trading_scope check updated to reflect buy/sell support
- README, architecture.md, configuration.md updated with sell documentation
