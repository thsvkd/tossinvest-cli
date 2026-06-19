# README 🆕 마커 정책

비교표(README.md / README.en.md 의 "지원 범위" 표)에서 **최근 한 달 안에 처음
출시된 기능**을 `🆕` 로 표시한다.

## 규칙

- **기준 기간:** 오늘로부터 **30일**(롤링 윈도우). `tools/update_new_markers.py`
  의 `NEW_WINDOW_DAYS` 로 조정.
- **기준 날짜:** 그 기능이 **처음 릴리즈된 날**(기능을 강화·확장한 날이 아니라
  최초 도입일). CHANGELOG 의 해당 버전 날짜를 쓴다.
- 표의 각 행은 출시일을 인라인 HTML 주석으로 들고 있다 (렌더링에는 안 보임):

  ```
  | **배당 내역** | `portfolio dividends` ... | ❌ | ✅ | <!--since:2026-06-19-->
  ```

- `🆕` 표시는 손으로 붙이지 않는다. 스크립트가 `<!--since:-->` 날짜와 오늘을
  비교해 첫 칸의 `🆕 ` 를 자동으로 붙이거나 뗀다 (idempotent).

## 운영

```bash
# 마커 갱신 (오늘 기준)
python3 tools/update_new_markers.py

# CI/검증: 마커가 낡았으면 비정상 종료
python3 tools/update_new_markers.py --check
```

- **새 기능 추가 시:** 비교표에 행을 넣고 줄 끝에 `<!--since:YYYY-MM-DD-->`
  (= 릴리즈 날짜)를 붙인 뒤 `update_new_markers.py` 를 돌린다.
- **릴리즈할 때마다** 한 번 돌린다. 시간이 지나 30일이 지난 기능은 자동으로
  `🆕` 가 떨어진다 — 수동 정리 불필요.
- ko/en 두 README 모두 같은 스크립트가 처리한다.

## 메모

- 날짜는 결정적 실행을 위해 `NEW_MARKER_DATE=YYYY-MM-DD` 로 고정할 수 있다.
- 표의 범례에 `🆕 최근 한 달 내 새로 추가된 기능` 항목이 있어야 한다.
