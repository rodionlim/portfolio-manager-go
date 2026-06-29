# Daily Market-Rotation Strategy

This is the canonical, lazily loaded design and memory document for the TradingView-to-OpenClaw market brief. Keep `AGENTS.md` and memory entries as short pointers to this file rather than copying the strategy into always-on context.

Tracking: [GitHub issue #183](https://github.com/rodionlim/portfolio-manager-go/issues/183), following the raw TradingView integration in [issue #181](https://github.com/rodionlim/portfolio-manager-go/issues/181).

## Objective

Produce a daily, post-US-close brief that explains one- to three-month sector ETF product flows, identifies broad and subsector rotations, drills into matching USA stock industries, and returns zero to five liquid stock candidates. Deterministic Go code owns data joins, calculations, exclusions, classifications, rankings, and history comparisons. The LLM owns concise interpretation only.

The strategy targets weeks-to-months holding periods and prefers confirmed turns over anticipatory entries. It must never fill a candidate quota when no setup qualifies.

## Data and universe

The source is TradingView's [Sector ETFs page](https://www.tradingview.com/markets/etfs/funds-sector-etfs/) via table ID `etfs_funds.sector_etfs`, using its Overview, Performance, and Fund Flows tabs. The raw MCP tools retain the top-100 global screen. The rotation scorer restricts aggregates to US-listed, USD-denominated equity ETFs.

Exclude foreign listings, non-USD funds, non-equity funds, leveraged/inverse products, and single-stock ETFs from sector aggregates. Include eligible thematic ETFs by mapping recognizable themes to their economic sectors, with an unmapped Thematic fallback bucket. Return exclusion counts by reason.

Classify remaining ETFs as:

- Broad sector: the explicit pairs XLC/VOX, XLY/VCR, XLP/VDC, XLE/VDE, XLF/VFH, XLV/VHT, XLI/VIS, XLB/VAW, XLRE/VNQ, XLK/VGT, and XLU/VPU.
- Subsector: other eligible, unleveraged sector, industry, or thematic ETFs.

Subsector and thematic ETFs can create independent secondary signals without claiming that an entire broad sector is confirmed.

Map common themes deterministically from ticker and fund name into economic sectors and stock-industry hints: semiconductors and AI/robotics/software themes → Information Technology, defense/space → Industrials, metals/uranium/lithium → Materials, clean energy → Energy, biotech/genomics → Health Care, crypto/fintech → Financials, gaming/media → Communication Services, and retail/EV themes → Consumer Discretionary.

## One- to three-month sector flow analysis

For each sector, precompute:

- Combined AUM and 1M/3M ETF product flows.
- 1M and 3M flows as percentages of combined AUM.
- Sector flow landscape ordering by an equal-weight blend of 1M and 3M flow/AUM percentages.
- Prior two-month average monthly flow: `(Flow3M - Flow1M) / 2`.
- Flow acceleration: current 1M normalized flow minus the prior two-month normalized monthly pace.
- Breadth: positive-flow ETF count and percentage.
- Broad and subsector flow totals and acceleration separately.
- Broad-versus-subsector agreement or divergence.
- Up to three genuine positive contributors and three genuine negative detractors.
- Flow state: accelerating inflow, decelerating inflow, improving outflow, or worsening outflow.
- History state: baseline, new, newly accelerating, strengthening, weakening, or unchanged.

These are ETF product flows, not direct stock-level institutional flows.

## Performance shape and signal lanes

Convert percentage fields to decimals for compounding calculations, then report percent units. Derive the embedded prior two-month return pace as:

`((1 + R3M) / (1 + R1M))^(1/2) - 1`

Return acceleration is current 1M return minus that prior monthly pace.

Classify setups as:

- Recovery: 6M or 1Y performance is non-positive, followed by positive 1W/1M performance and positive acceleration.
- Ignition: longer-term performance is below the 60th percentile while recent acceleration is in the top quartile.
- Continuation: 3M and 6M performance are in the top quartile with positive supporting flow.
- Unconfirmed: insufficient multi-period confirmation.

Reject actionable setups driven primarily by one top-decile trading day. Maintain three lanes:

1. Broad-sector flow and performance confirmed: highest confidence.
2. Independent subsector rotation: medium-high confidence and explicitly labeled.
3. Performance-led industry without flow confirmation: eligible but lower confidence.

Score broad-sector and subsector ETF signals with 20% 1M flow/AUM percentile, 20% 3M flow/AUM percentile, 25% flow acceleration percentile, and 35% ETF performance acceleration percentile.

Flow disagreement reduces confidence; it does not automatically eliminate strong industry performance.

## Industry mapping

Map at the TradingView industry level rather than treating its 20 stock sectors as equivalent to the 11 ETF sectors. Important splits include:

- Finance → Financials, except REITs and real-estate development → Real Estate.
- Electronic Technology → Information Technology, except Aerospace & Defense → Industrials.
- Technology Services → Information Technology, except Internet Software/Services → Communication Services.
- Consumer Services → Consumer Discretionary, with media, broadcasting, publishing, and cable → Communication Services.
- Industrial Services → Industrials, with pipelines, drilling, and oilfield services → Energy.
- Process Industries → Materials, with agricultural commodities → Consumer Staples.
- Distribution Services → Industrials, with medical distribution → Health Care and food distribution → Consumer Staples.
- Consumer Non-Durables → Consumer Staples, except apparel/footwear → Consumer Discretionary.
- TradingView's generic Miscellaneous industries remain unmapped when no defensible assignment exists.

Common explicit subsector hints include KBWB/KRE/KBE → banks, XBI/IBB → biotechnology, SMH/SOXX → semiconductors, ITA/PPA/XAR → aerospace and defense, GDX/GDXJ → precious metals, and COPX/XME → metals/mining industries.

## Stock eligibility and ranking

Drill into at most three selected industries, using only the top three sorted industry signals. A stock must have:

- USD pricing.
- Market capitalization of at least $2 billion.
- Price of at least $5.
- Current daily dollar turnover of at least $20 million.
- Monthly volatility outside the highest peer quintile.
- Positive 1W, 1M, and 3M performance with positive return acceleration.
- No disqualifying one-day spike.

Do not use relative volume. Rank qualified stocks using individual performance/acceleration (60%), inherited sector-industry confidence (30%), and history state (10%). Fundamentals are tie-breakers and risk flags, not primary gates.

## Deterministic MCP contract

Use `screen_daily_market_rotation` with:

- `persist_history`: defaults to true; set false for dry runs.
- `max_stock_candidates`: 1–5, default 5.

The compact response contains sector flow summaries, broad rotations, subsector rotations, performance-led industries, stock candidates, rejection counts, lost confirmations, data-quality notes, and narration instructions. OpenClaw must preserve returned ordering and must not recompute or rerank results.

Persist at most 20 US sessions under the `MARKET_ROTATION` LevelDB prefix. Do not persist unchanged/stale fingerprints. Detect new, strengthening, weakening, newly accelerating, and lost-confirmation states.

## Brief and operations

The full narration prompt, validation prompt, and suggested 6:30 a.m. Asia/Singapore Tuesday–Saturday schedule are in [openclaw-market-rotation.md](openclaw-market-rotation.md).

The brief order is:

1. 1M/3M sector fund-flow landscape.
2. Performance alignment and divergences.
3. Broad-sector rotations.
4. Independent subsector rotations.
5. Performance-led opportunities.
6. Zero to five stock candidates with evidence and invalidation conditions.
7. Data-quality and exclusion notes.

## Validation requirements

- Unit tests cover exclusions, aggregation, breadth, acceleration, crosswalk exceptions, setup classification, spike rejection, history transitions, and no-candidate behavior.
- The full Go suite and vet must pass.
- Live validation uses `persist_history=false` and verifies a compact response without writing state.
- Perform a completed post-close dry run before enabling scheduled delivery; intraday runs are diagnostic only.

This feature is a quantitative screen, not a directive to trade.
