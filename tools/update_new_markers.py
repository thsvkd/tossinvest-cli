#!/usr/bin/env python3
"""Maintain 🆕 markers on README comparison-table rows.

Policy: a feature row is marked 🆕 if it was first released within the last
NEW_WINDOW_DAYS (rolling, from today). Each taggable row carries its release
date inline as an HTML comment, e.g.

    | **배당 내역** | `portfolio dividends` ... | ❌ | ✅ | <!--since:2026-06-19-->

This script reads those dates and adds/removes the leading "🆕 " in the first
cell so the markers never go stale. Run it at every release (and it's safe to
run anytime — it's idempotent).

    python3 tools/update_new_markers.py          # uses today
    NEW_MARKER_DATE=2026-06-19 python3 tools/update_new_markers.py   # pin date

stdlib only. Exits non-zero if any file changed when --check is passed (for CI).
"""
import datetime
import os
import re
import sys

NEW_WINDOW_DAYS = 30
FILES = ["README.md", "README.en.md"]

since_re = re.compile(r"<!--since:(\d{4}-\d{2}-\d{2})-->")
# first table cell: | [**]['🆕 ']<name>[**] |
cell_re = re.compile(r"^(\| )(\*\*)?(?:🆕 )?(.*?)(\*\*)?( \|)")


def main():
    check = "--check" in sys.argv
    today = datetime.date.fromisoformat(
        os.environ.get("NEW_MARKER_DATE") or datetime.date.today().isoformat()
    )
    cutoff = today - datetime.timedelta(days=NEW_WINDOW_DAYS)

    changed_any = False
    for path in FILES:
        if not os.path.exists(path):
            continue
        src = open(path, encoding="utf-8").read()
        lines = src.split("\n")
        new_count = 0
        for i, ln in enumerate(lines):
            m = since_re.search(ln)
            if not m or not ln.startswith("|"):
                continue
            since = datetime.date.fromisoformat(m.group(1))
            is_new = since >= cutoff
            if is_new:
                new_count += 1

            def repl(mm, is_new=is_new):
                pre, b1, name, b2, post = (
                    mm.group(1), mm.group(2) or "", mm.group(3), mm.group(4) or "", mm.group(5),
                )
                tag = "🆕 " if is_new else ""
                return f"{pre}{b1}{tag}{name}{b2}{post}"

            lines[i] = cell_re.sub(repl, ln, count=1)
        out = "\n".join(lines)
        if out != src:
            changed_any = True
            if not check:
                open(path, "w", encoding="utf-8").write(out)
        print(f"{path}: {new_count} rows marked 🆕 (since >= {cutoff})"
              + (" [WOULD CHANGE]" if check and out != src else ""))

    if check and changed_any:
        print("ERROR: 🆕 markers are stale — run tools/update_new_markers.py", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
