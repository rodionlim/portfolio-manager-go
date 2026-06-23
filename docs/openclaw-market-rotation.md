# OpenClaw Daily Market-Rotation Brief

The canonical strategy and scoring rationale live in [market-rotation-strategy.md](market-rotation-strategy.md). Load that document only for relevant market-rotation work.

The `screen_daily_market_rotation` MCP tool performs every join, calculation, exclusion, classification, ranking, and history comparison in Go. OpenClaw should narrate its compact response without recomputing it.

## Scheduled prompt

```text
Create today's US market-rotation brief.

Call screen_daily_market_rotation with persist_history=true and max_stock_candidates=5. Treat the returned calculations, rankings, lane labels, and ordering as authoritative: do not recalculate, rerank, or fill a candidate quota.

Write a concise brief in this exact order:
1. 1M sector fund-flow landscape: explain where ETF capital is concentrating, which sectors have accelerating inflows, which remain in outflow but are improving, and which are worsening. Mention breadth and the largest ETF contributors or detractors when they materially explain a sector total.
2. Market-performance alignment and divergences: identify where broad ETF performance confirms flow and where performance leads or conflicts with flow. Call out broad-versus-subsector disagreement.
3. Broad-sector rotations.
4. Independent subsector rotations.
5. Performance-led, lower-confidence industries.
6. Zero to five stock candidates, preserving the returned order. For each, state its lane, setup, evidence, primary risk, and supplied invalidation condition.
7. Data-quality and exclusion notes.

Mention newly accelerating, strengthening, weakening, and lost-confirmation states when returned.

Use the phrase "ETF product flows" or "sector ETF flows". Never describe these values as direct stock-level institutional flows. If no stocks qualify, say so plainly. This is a quantitative screening brief, not a directive to trade.
```

## Validation prompt

Use the scheduled prompt with `persist_history=false` for dry runs. The response must report `data_quality.persisted=false` and must not change LevelDB history.

## Suggested schedule

Run after the completed US session at 6:30 a.m. Singapore time, Tuesday through Saturday:

```sh
openclaw cron create "30 6 * * 2-6" \
  "Create today's US market-rotation brief using screen_daily_market_rotation and the workspace market-rotation prompt." \
  --name "US market rotation brief" \
  --session isolated \
  --tz Asia/Singapore
```

Add the desired OpenClaw `--agent`, delivery channel, and destination for the local installation before enabling delivery.
