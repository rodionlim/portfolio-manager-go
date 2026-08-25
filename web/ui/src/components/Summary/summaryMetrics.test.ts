import { describe, expect, it } from "vitest";
import type { TimestampedMetrics } from "../Analytics/types";
import { calculateHeadlinePnlMetrics } from "./summaryMetrics";

const snapshot = (timestamp: string, pnl: number): TimestampedMetrics => ({
  timestamp,
  metrics: {
    irr: 0,
    mv: 1_000 + pnl,
    pricePaid: 1_000,
    totalDividends: 0,
  },
});

describe("calculateHeadlinePnlMetrics", () => {
  it("calculates WTD, MTD, QTD, and YTD P&L from unsorted snapshots", () => {
    const result = calculateHeadlinePnlMetrics([
      snapshot("2026-08-15T00:00:00Z", 180),
      snapshot("2025-12-31T00:00:00Z", 20),
      snapshot("2026-07-31T00:00:00Z", 130),
      snapshot("2026-06-30T00:00:00Z", 100),
      snapshot("2026-08-08T00:00:00Z", 150),
    ]);

    expect(result.wtd).toBe(30);
    expect(result.totalDividends).toBe(0);
    expect(result.mtd).toBe(50);
    expect(result.qtd).toBe(80);
    expect(result.ytd).toBe(160);
    expect(result.asOf).toBe("2026-08-15T00:00:00.000Z");
  });

  it("leaves periods unavailable when a new portfolio lacks a baseline", () => {
    const result = calculateHeadlinePnlMetrics([
      snapshot("2026-08-11T00:00:00Z", 25),
      snapshot("2026-08-15T00:00:00Z", 40),
    ]);

    expect(result.wtd).toBeUndefined();
    expect(result.mtd).toBeUndefined();
    expect(result.qtd).toBeUndefined();
    expect(result.ytd).toBeUndefined();
  });

  it("returns no values when historical metrics are empty", () => {
    expect(calculateHeadlinePnlMetrics([])).toEqual({});
  });

  it("uses current market value as the ending value for every period", () => {
    const metrics = [
      snapshot("2025-12-31T00:00:00Z", 20),
      snapshot("2026-06-30T00:00:00Z", 100),
      snapshot("2026-07-31T00:00:00Z", 130),
      snapshot("2026-08-08T00:00:00Z", 150),
      snapshot("2026-08-15T00:00:00Z", 180),
    ];

    const result = calculateHeadlinePnlMetrics(
      metrics,
      1_230,
      new Date("2026-08-15T12:00:00Z"),
    );

    expect(result.wtd).toBe(80);
    expect(result.mtd).toBe(100);
    expect(result.qtd).toBe(130);
    expect(result.ytd).toBe(210);
    expect(result.asOf).toBe("2026-08-15T00:00:00.000Z");
  });
});
