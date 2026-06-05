# Security Policy

[한국어](#보안-정책) · English

## Supported Versions

Only the latest released version receives fixes. Please reproduce any issue on
the most recent release (or `main`) before reporting.

| Version | Supported |
| ------- | --------- |
| latest release / `main` | ✅ |
| older releases | ❌ |

## Reporting a Vulnerability

**Do not open a public issue for security problems.**

Preferred: GitHub **private vulnerability reporting** —
repo **Security** tab → **Report a vulnerability**
(<https://github.com/JungHoonGhae/tossinvest-cli/security/advisories/new>).

Alternative: email **lucas.ghae@remodule.dev** with `[tossinvest-cli security]` in
the subject.

Please include: affected version/commit, environment (OS, install method),
reproduction steps, and impact. We aim to acknowledge within a few days. This is
a volunteer-maintained project with no bug bounty.

## Scope

In scope — issues in **this CLI's own code**, e.g.:

- Local credential/session handling (`session.json`, cookie/XSRF storage, file permissions)
- Leaking secrets or account data into logs, output, or diagnostics
- Command/argument injection, unsafe file writes, path traversal
- CI/release supply-chain issues (workflows, build, Homebrew formula)

Out of scope:

- Vulnerabilities in **Toss Securities' own servers or web API** — report those
  to Toss, not here. This is an **unofficial** project that reuses Toss's
  internal web API; it is not affiliated with or endorsed by Toss Securities.
- Risks inherent to the project's documented design (e.g. trading is gated
  behind explicit config + `--execute` + `--confirm`).

## When reporting

- **Never include real credentials, cookies, tokens, or account data.** Redact
  them or use dummy values — the same rule the project follows everywhere.
- `tossctl doctor --report` redacts your home directory automatically; still
  review its output before sharing.

---

## 보안 정책

### 지원 버전

최신 릴리즈만 수정 대상입니다. 신고 전 최신 릴리즈(또는 `main`)에서 재현해 주세요.

### 취약점 신고

**보안 문제는 공개 이슈로 올리지 마세요.**

- 권장: 레포 **Security** 탭 → **Report a vulnerability**
  (<https://github.com/JungHoonGhae/tossinvest-cli/security/advisories/new>)
- 또는: **lucas.ghae@remodule.dev** 으로 메일 (제목에 `[tossinvest-cli security]`)

영향 버전/커밋, 환경(OS·설치 방법), 재현 절차, 영향 범위를 포함해 주세요. 며칠 내
확인을 목표로 하며, 자원봉사로 운영되는 프로젝트라 버그 바운티는 없습니다.

### 범위

대상 — **이 CLI 자체 코드**의 문제 (로컬 자격증명/세션 처리, 시크릿·계좌 데이터
유출, 명령/인자 인젝션·경로 조작, CI/릴리즈 공급망 등).

대상 아님 — **토스증권 자체 서버·웹 API**의 취약점(토스로 신고). 본 프로젝트는
토스 내부 웹 API를 재사용하는 **비공식** 도구로, 토스증권과 무관합니다.

### 신고 시

- **실제 자격증명·쿠키·토큰·계좌 데이터를 절대 포함하지 마세요.** 가리거나 더미
  값을 사용하세요. `tossctl doctor --report` 는 홈 디렉터리를 자동으로 가립니다.
