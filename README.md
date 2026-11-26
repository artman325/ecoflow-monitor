# ecoflow-monitor
EcoFlow station charge on Windows desktop (Rainmeter) and in-game OSD (RTSS). Learning Go.

## Known API quirks

Found by testing a DELTA Max 2000 against the official Open API (EU host).
Not documented by EcoFlow.

- `quota/all` has no timestamp. When the station goes offline, the API keeps
  returning the last cached values with `code: "0"`.
- `online` in `device/list` lags 5–8 minutes behind the real state.
  During that window the monitor keeps showing the cached values as fresh,
  and `LastUpdate` keeps moving.
- `pd.remainTime` is always positive: minutes to full while charging,
  minutes to empty while discharging. `5999` = neither (passthrough).
- `pd.sysChgDsgState`: `2` = charging, `1` = not charging
  (both discharging and passthrough). Use input/output watts to tell them apart.
- Error codes seen so far:
  - `8512` no permission to do it (SN not bound to the account)
  - `8513` accessKey is invalid
  - `8521` signature is wrong

## Polling

Polls every `EF_INTERVAL_SECONDS` (default 60, min 10).
On errors it backs off to 1, 2, 3, 5 minutes and stays at 5 until a request
succeeds. If `EF_INTERVAL_SECONDS` is longer, it is used as is.