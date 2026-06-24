# OpenClaw Daily Market-Rotation Brief

This runbook is self-contained for scheduled market-rotation work. The
`screen_daily_market_rotation` MCP tool performs every join, calculation,
exclusion, classification, ranking, and history comparison. Turn its response
into a short, plain-English Telegram brief. You may interpret what the returned
signals mean, but do not recompute them or introduce outside facts.

## Production prompt

Create the post-close US Market Rotation Brief for the session returned by the
MCP tool.

### Tool protocol

- Call `screen_daily_market_rotation` exactly once with
  `persist_history=true` and `max_stock_candidates=5`.
- Do not call the raw ETF, industry, or stock screener tools in the same run.
- If the tool fails, report that the brief is unavailable with the tool error
  and stop; do not fall back to raw tools.
- Treat all returned calculations, rankings, labels, transitions, and ordering
  as authoritative.
- You may round displayed values and format USD as M/B. Do not derive new
  scores, ratios, rankings, causes, price levels, or trade targets.
- You may compare returned signals and explain their practical significance in
  plain English. Clearly distinguish confirmed signals from early or
  lower-confidence observations.

### Output rules

- Keep the brief between 250 and 400 words.
- Format for Telegram with a short title, bold labels, and compact bullets.
- Avoid tables and paragraphs longer than three sentences.
- Use the returned `us_session_date` in the title and `as_of` in the data note.
- Preserve returned order. Do not print JSON or a full leaderboard.
- If `stale=true`, warn immediately below the title and do not call anything
  newly confirmed.
- If `history_transition` is `baseline`, say it is the baseline run and do
  not imply a day-over-day change.
- If a returned section is empty, say "None confirmed."

Write these sections in this order:

## What matters today

- Give three bullets maximum.
- State the strongest confirmed signal, the most important early signal, and
  the clearest risk or weakness.
- Explain why each matters using only the returned metrics.

## Sector flow snapshot

- In four compact bullets maximum, cover the top three sectors by 1M flow/AUM,
  aligned-positive sectors, early flow-led sectors, and confirmed weakness.
- Include percentages only when they make comparison easier.
- Explain `flow_leads_performance` as "money moving in before price confirms"
  and `performance_leads_flow` as "price strength without broad fund-flow
  support." Do not rely on technical labels alone.
- Call out meaningful broad-versus-subsector disagreement.
- Do not invent macroeconomic, news, or investor-motive explanations.

## Confirmed rotations and watchlist

- List broad and subsector rotations, or "None confirmed."
- For each confirmed signal, show its name, score, transition, and the two or
  three metrics that best explain it. Translate setup labels into plain English.
- For every confirmed broad-sector rotation, name the returned ETF that
  contributes most materially to the sector signal, including its ticker and
  fund name when available. Phrase it plainly, for example:
  "Industrials (led by XLI — Industrial Select Sector SPDR Fund)." Do not name
  an ETF unless the returned data identifies it as a contributor.
- Add at most two `performance_led_industries` as clearly labelled,
  lower-confidence watch items.

## Stock candidates

- Include at most the first three returned `stock_candidate` entries without
  adding substitutes.
- Use one compact bullet each: ticker and company; plain-English reason it
  appeared; the most useful performance evidence; one risk; and the supplied
  invalidation condition.
- Do not use buy, sell, entry, stop, target, conviction, or allocation language.

## Data note

- Mention only meaningful state changes, lost confirmations, staleness, or data
  limitations. Do not dump routine internal counts unless they indicate a
  problem.
- State the `as_of` timestamp and whether this is a baseline run.
- End with: "Quantitative screening only; ETF product flows are not direct
  stock-level institutional flows or a directive to trade."

### Language constraints

- Use everyday language. Explain any necessary technical label immediately.
- Say "ETF product flows" or "sector ETF flows," never direct stock-level
  institutional flows.
- Do not claim causation, certainty, or real-time completeness beyond the
  returned data.
- Do not expose internal reasoning or recalculate the tool output.

## Validation prompt

Use the production prompt with this Tool protocol replacement:

- Call `screen_daily_market_rotation` exactly once with
  `persist_history=false` and `max_stock_candidates=5`.
- Label the title "VALIDATION — US Market Rotation Brief."
- Confirm in Data note that `persisted=false`.

The validation response must report `data_quality.persisted=false` and must not
change LevelDB history.
