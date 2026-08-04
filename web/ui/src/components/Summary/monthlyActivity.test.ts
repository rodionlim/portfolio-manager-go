import { describe, expect, it } from "vitest";
import type { TimestampedMetrics } from "../Analytics/types";
import type { Trade } from "../../types/blotter";
import { buildMonthlyPortfolioActivity } from "./monthlyActivity";

const metric = (
  timestamp: string,
  pnl: number,
  dividends: number,
  marketValue: number,
): TimestampedMetrics => ({
  timestamp,
  metrics: {
    irr: 0,
    mv: marketValue,
    pricePaid: marketValue + dividends - pnl,
    totalDividends: dividends,
  },
});

const trade = (overrides: Partial<Trade>): Trade => ({
  TradeID: "trade-1",
  TradeDate: "2026-06-10T00:00:00Z",
  Ticker: "AAPL",
  Book: "Main",
  Broker: "Broker",
  Account: "Account",
  Quantity: 10,
  Price: 100,
  Fx: 1.3,
  TradeType: false,
  Side: "buy",
  SeqNum: 1,
  ...overrides,
});

describe("buildMonthlyPortfolioActivity", () => {
  it("builds twelve calendar months of portfolio metrics and ticker-level trades", () => {
    const rows = buildMonthlyPortfolioActivity(
      [
        metric("2026-05-31T00:00:00Z", 10, 1, 900),
        metric("2026-06-30T00:00:00Z", 30, 5, 1_100),
        metric("2026-07-01T00:00:00Z", 30, 5, 1_100),
        metric("2026-07-31T00:00:00Z", 20, 8, 1_050),
        metric("2026-08-01T00:00:00Z", 25, 8, 1_075),
        metric("2026-08-15T00:00:00Z", 50, 10, 1_200),
      ],
      [
        trade({}),
        trade({
          TradeID: "trade-2",
          TradeDate: "2026-06-20T00:00:00Z",
          Quantity: 5,
          Price: 120,
          SeqNum: 2,
        }),
        trade({ TradeID: "trade-3", Side: "sell", Quantity: 2, SeqNum: 3 }),
        trade({
          TradeID: "trade-4",
          TradeDate: "2026-07-12T00:00:00Z",
          Ticker: "MSFT",
          Price: 200,
          Fx: 1.2,
          SeqNum: 4,
        }),
      ],
      undefined,
      new Date("2026-08-15T00:00:00Z"),
    );

    expect(rows.map((row) => row.monthLabel)).toEqual([
      "Aug-26",
      "Jul-26",
      "Jun-26",
      "May-26",
      "Apr-26",
      "Mar-26",
      "Feb-26",
      "Jan-26",
      "Dec-25",
      "Nov-25",
      "Oct-25",
      "Sep-25",
    ]);
    expect(rows[2]).toMatchObject({
      mtdPnl: 20,
      dividends: 4,
      netCashFlow: 1_820,
      marketValue: 1_100,
    });
    expect(rows[2].trades).toHaveLength(1);
    expect(rows[2].trades[0].earliestTradeDate).toBe("2026-06-10");
    expect(rows[2].trades[0].quantityBought).toBe(15);
    expect(rows[2].trades[0].quantitySold).toBe(2);
    expect(rows[2].trades[0].netQuantity).toBe(13);
    expect(rows[2].trades[0].averagePrice).toBeCloseTo(105.88, 2);
    expect(rows[2].trades[0].pricePaid).toBe(1_820);
    expect(rows[1].trades[0]).toMatchObject({
      ticker: "MSFT",
      earliestTradeDate: "2026-07-12",
      quantityBought: 10,
      quantitySold: 0,
      netQuantity: 10,
      averagePrice: 200,
      pricePaid: 2_400,
    });
  });

  it("keeps monthly metrics unavailable when there is no baseline history", () => {
    const rows = buildMonthlyPortfolioActivity(
      [metric("2026-08-15T00:00:00Z", 50, 10, 1_200)],
      [],
      undefined,
      new Date("2026-08-15T00:00:00Z"),
    );

    expect(rows[0]).toMatchObject({
      monthLabel: "Aug-26",
      mtdPnl: undefined,
      dividends: undefined,
      netCashFlow: 0,
      marketValue: 1_200,
      trades: [],
    });
  });

  it("uses the current market value for the latest month only", () => {
    const rows = buildMonthlyPortfolioActivity(
      [
        metric("2026-06-30T00:00:00Z", 10, 1, 900),
        metric("2026-07-31T00:00:00Z", 30, 5, 1_100),
        metric("2026-08-15T00:00:00Z", 50, 10, 1_200),
      ],
      [],
      1_250,
      new Date("2026-08-15T00:00:00Z"),
    );

    expect(rows[0]).toMatchObject({ marketValue: 1_250, mtdPnl: 70 });
    expect(rows[1]).toMatchObject({ marketValue: 1_100, mtdPnl: 20 });
  });
});
