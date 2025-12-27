import { describe, it, expect } from "vitest";
import { withRollingVolatility } from "./volatility";
import type { TimestampedMetrics } from "./types";

function sampleStdDev(values: number[]): number | undefined {
  if (values.length < 2) return undefined;
  const mean = values.reduce((acc, v) => acc + v, 0) / values.length;
  const variance =
    values.reduce((acc, v) => acc + (v - mean) * (v - mean), 0) /
    (values.length - 1);
  return Math.sqrt(variance);
}

describe("withRollingVolatility (cash-flow-adjusted)", () => {
  it("computes SMA volatility from cash-flow-adjusted returns (ΔpricePaid)", () => {
    const points: TimestampedMetrics[] = [
      {
        timestamp: "2025-01-01",
        metrics: { irr: 0, mv: 100, pricePaid: 100, totalDividends: 0 },
      },
      {
        timestamp: "2025-01-02",
        metrics: { irr: 0, mv: 102, pricePaid: 101, totalDividends: 0 },
      },
      {
        timestamp: "2025-01-03",
        metrics: { irr: 0, mv: 101, pricePaid: 101, totalDividends: 0 },
      },
    ];

    const out = withRollingVolatility(points, {
      method: "sma",
      window: 2,
      tradingDaysPerYear: 252,
      maxGapDays: 5,
    });

    // Day 2 has only one return in the window => SD undefined
    expect(out[1].metrics.standardDeviation).toBeUndefined();

    const r1 = (102 - 100 + (101 - 100)) / (100 + (101 - 100)); // 3/101
    const r2 = (101 - 102 + (101 - 101)) / (102 + (101 - 101)); // -1/102

    const sdDaily = sampleStdDev([r1, r2]);
    expect(sdDaily).toBeDefined();

    const expectedAnnualized = (sdDaily as number) * Math.sqrt(252);
    expect(out[2].metrics.standardDeviation).toBeDefined();
    expect(out[2].metrics.standardDeviation as number).toBeCloseTo(
      expectedAnnualized,
      10
    );
  });

  it("computes EWMA volatility from cash-flow-adjusted returns (ΔpricePaid)", () => {
    const points: TimestampedMetrics[] = [
      {
        timestamp: "2025-01-01",
        metrics: { irr: 0, mv: 100, pricePaid: 100, totalDividends: 0 },
      },
      {
        timestamp: "2025-01-02",
        metrics: { irr: 0, mv: 102, pricePaid: 101, totalDividends: 0 },
      },
      {
        timestamp: "2025-01-03",
        metrics: { irr: 0, mv: 101, pricePaid: 101, totalDividends: 0 },
      },
    ];

    const window = 2;
    const alpha = 2 / (window + 1);
    const tradingDaysPerYear = 252;

    const out = withRollingVolatility(points, {
      method: "ewma",
      window,
      tradingDaysPerYear,
      maxGapDays: 5,
    });

    const r1 = (102 - 100 + (101 - 100)) / (100 + (101 - 100)); // 3/101
    const r2 = (101 - 102 + (101 - 101)) / (102 + (101 - 101)); // -1/102

    // Implementation initializes ewmaVar from first observation as r1^2 (since window has <2 samples).
    const expectedDay2 = Math.abs(r1) * Math.sqrt(tradingDaysPerYear);
    expect(out[1].metrics.standardDeviation).toBeDefined();
    expect(out[1].metrics.standardDeviation as number).toBeCloseTo(
      expectedDay2,
      10
    );

    const ewmaVarDay3 = (1 - alpha) * (r1 * r1) + alpha * (r2 * r2);
    const expectedDay3 = Math.sqrt(ewmaVarDay3) * Math.sqrt(tradingDaysPerYear);
    expect(out[2].metrics.standardDeviation).toBeDefined();
    expect(out[2].metrics.standardDeviation as number).toBeCloseTo(
      expectedDay3,
      10
    );
  });

  it("skips return when denominator (MV_{t-1} + CF_t) is zero", () => {
    const points: TimestampedMetrics[] = [
      {
        timestamp: "2025-01-01",
        metrics: { irr: 0, mv: 0, pricePaid: 0, totalDividends: 0 },
      },
      {
        timestamp: "2025-01-02",
        metrics: { irr: 0, mv: 10, pricePaid: 0, totalDividends: 0 },
      },
      {
        timestamp: "2025-01-03",
        metrics: { irr: 0, mv: 11, pricePaid: 0, totalDividends: 0 },
      },
    ];

    const out = withRollingVolatility(points, {
      method: "sma",
      window: 2,
      tradingDaysPerYear: 252,
      maxGapDays: 5,
    });

    // First return skipped due to denom 0 => still no SD on day 2
    expect(out[1].metrics.standardDeviation).toBeUndefined();
    // Only one valid return (day2->day3) => SD still undefined
    expect(out[2].metrics.standardDeviation).toBeUndefined();
  });
});
