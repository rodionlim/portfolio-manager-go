import { TimestampedMetrics } from "./types";

export type VolatilityMethod = "sma" | "ewma";

export interface VolatilityOptions {
  method: VolatilityMethod;
  /** Window in trading days (observations). */
  window: number;
  /** Max allowed gap (calendar days) between points to treat as consecutive. */
  maxGapDays?: number;
  /** Annualize using sqrt(tradingDaysPerYear). */
  tradingDaysPerYear?: number;
}

const DEFAULT_MAX_GAP_DAYS = 5; // tolerate weekends/holidays
const DEFAULT_TRADING_DAYS_PER_YEAR = 252;

function daysBetween(a: Date, b: Date): number {
  return (a.getTime() - b.getTime()) / (1000 * 60 * 60 * 24);
}

function sampleStdDev(values: number[]): number | undefined {
  if (values.length < 2) return undefined;
  const mean = values.reduce((acc, v) => acc + v, 0) / values.length;
  const variance =
    values.reduce((acc, v) => acc + (v - mean) * (v - mean), 0) /
    (values.length - 1);
  return Math.sqrt(variance);
}

/**
 * Computes rolling annualized volatility (standard deviation) from cash-flow-adjusted daily returns.
 *
 * Notes:
 * - Uses simple returns adjusted for net cash flow (excluding dividends):
 *   r_t = (MV_t - MV_{t-1} + CF_t) / (MV_{t-1} + CF_t)
 *   where CF_t is the net cash flow for the period (in our app: Δ(pricePaid) between the two days).
 * - Skips returns when the gap between timestamps is > maxGapDays (prevents year-apart jumps)
 * - Resets the rolling window after a gap
 * - Returns volatility as a decimal (e.g. 0.12 = 12% annualized)
 */
export function withRollingVolatility(
  metrics: TimestampedMetrics[],
  options: VolatilityOptions
): TimestampedMetrics[] {
  const maxGapDays = options.maxGapDays ?? DEFAULT_MAX_GAP_DAYS;
  const tradingDaysPerYear =
    options.tradingDaysPerYear ?? DEFAULT_TRADING_DAYS_PER_YEAR;

  const window = Math.max(1, Math.floor(options.window));

  const sorted = [...metrics].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  // rolling state (resets on gap)
  const rollingReturns: number[] = [];

  // EWMA state (resets on gap)
  let ewmaVar: number | undefined;
  const alpha = 2 / (window + 1); // standard EMA-style smoothing

  const out: TimestampedMetrics[] = sorted.map((m) => ({
    ...m,
    metrics: { ...m.metrics, standardDeviation: undefined },
  }));

  for (let i = 1; i < out.length; i++) {
    const prev = out[i - 1];
    const cur = out[i];

    const prevT = new Date(prev.timestamp);
    const curT = new Date(cur.timestamp);
    const gap = daysBetween(curT, prevT);

    // Non-forward or too large gap => reset state
    if (!(gap > 0) || gap > maxGapDays) {
      rollingReturns.length = 0;
      ewmaVar = undefined;
      continue;
    }

    const prevMV = prev.metrics.mv;
    const curMV = cur.metrics.mv;
    const prevPricePaid = prev.metrics.pricePaid;
    const curPricePaid = cur.metrics.pricePaid;

    if (
      !Number.isFinite(prevMV) ||
      !Number.isFinite(curMV) ||
      !Number.isFinite(prevPricePaid) ||
      !Number.isFinite(curPricePaid)
    ) {
      continue;
    }

    // Net cash flow for the period (excluding dividends), inferred from change in cumulative price paid.
    const netCashFlow = curPricePaid - prevPricePaid;
    const denom = prevMV + netCashFlow;
    if (!Number.isFinite(denom) || denom === 0) continue;

    const r = (curMV - prevMV + netCashFlow) / denom;
    if (!Number.isFinite(r)) continue;

    if (options.method === "sma") {
      rollingReturns.push(r);
      if (rollingReturns.length > window) rollingReturns.shift();

      const sdDaily = sampleStdDev(rollingReturns);
      if (sdDaily !== undefined) {
        cur.metrics.standardDeviation = sdDaily * Math.sqrt(tradingDaysPerYear);
      }
    } else {
      // EWMA on squared returns, mean assumed ~0
      const r2 = r * r;
      if (ewmaVar === undefined) {
        // initialize from first observations
        rollingReturns.push(r);
        if (rollingReturns.length > window) rollingReturns.shift();
        const init = sampleStdDev(rollingReturns);
        ewmaVar = init !== undefined ? init * init : r2;
      } else {
        ewmaVar = (1 - alpha) * ewmaVar + alpha * r2;
      }

      if (ewmaVar !== undefined) {
        cur.metrics.standardDeviation =
          Math.sqrt(ewmaVar) * Math.sqrt(tradingDaysPerYear);
      }
    }
  }

  return out;
}
