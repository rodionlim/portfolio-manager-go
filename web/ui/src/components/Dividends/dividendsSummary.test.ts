import { describe, expect, it } from "vitest";
import {
  aggregateMonthlyDividends,
  getLastTwelveMonthKeys,
  getMonthKeysForYears,
  getYearsFromMonthlyDividends,
  pivotMonthlyDividendsByYear,
  removeFutureMonthKeys,
} from "./dividendsSummary";
import { ReferenceData } from "../../types";
import { Trade } from "../../types/blotter";

const refData: ReferenceData = {
  AAPL: {
    id: "AAPL",
    name: "Apple",
    underlying_ticker: "",
    yahoo_ticker: "",
    google_ticker: "",
    dividends_sg_ticker: "",
    asset_class: "equity",
    asset_sub_class: "stock",
    category: "",
    category_sgx: "",
    sub_category: "",
    ccy: "USD",
    domicile: "",
    coupon_rate: 0,
    maturity_date: "",
    strike_price: 0,
    call_put: "",
  },
  SBFEB50: {
    id: "SBFEB50",
    name: "Singapore Savings Bond",
    underlying_ticker: "",
    yahoo_ticker: "",
    google_ticker: "",
    dividends_sg_ticker: "",
    asset_class: "bond",
    asset_sub_class: "govies",
    category: "",
    category_sgx: "",
    sub_category: "",
    ccy: "SGD",
    domicile: "",
    coupon_rate: 0,
    maturity_date: "",
    strike_price: 0,
    call_put: "",
  },
  BS24124Z: {
    id: "BS24124Z",
    name: "MAS Bill",
    underlying_ticker: "",
    yahoo_ticker: "",
    google_ticker: "",
    dividends_sg_ticker: "",
    asset_class: "bond",
    asset_sub_class: "govies",
    category: "",
    category_sgx: "",
    sub_category: "",
    ccy: "SGD",
    domicile: "",
    coupon_rate: 0,
    maturity_date: "",
    strike_price: 0,
    call_put: "",
  },
};

const trade = (
  tradeDate: string,
  ticker: string,
  side: "buy" | "sell",
  quantity: number,
  price: number
): Trade => ({
  TradeID: `${tradeDate}-${ticker}-${side}`,
  TradeDate: tradeDate,
  Ticker: ticker,
  Book: "default",
  Broker: "",
  Account: "",
  Quantity: quantity,
  Price: price,
  Fx: 1,
  TradeType: false,
  Side: side,
  SeqNum: 1,
});

describe("dividends monthly summary", () => {
  it("returns the current month and previous 11 months", () => {
    expect(getLastTwelveMonthKeys(new Date(2026, 5, 15))).toEqual([
      "2025-07",
      "2025-08",
      "2025-09",
      "2025-10",
      "2025-11",
      "2025-12",
      "2026-01",
      "2026-02",
      "2026-03",
      "2026-04",
      "2026-05",
      "2026-06",
    ]);
  });

  it("returns all calendar months for selected years", () => {
    const monthKeys = getMonthKeysForYears([2026, 2024]);

    expect(monthKeys).toHaveLength(24);
    expect(monthKeys[0]).toBe("2024-01");
    expect(monthKeys[11]).toBe("2024-12");
    expect(monthKeys[12]).toBe("2026-01");
    expect(monthKeys[23]).toBe("2026-12");
  });

  it("removes future months from selected years", () => {
    const monthKeys = removeFutureMonthKeys(
      getMonthKeysForYears([2025, 2026]),
      new Date(2026, 5, 15)
    );

    expect(monthKeys).toHaveLength(18);
    expect(monthKeys[0]).toBe("2025-01");
    expect(monthKeys[17]).toBe("2026-06");
    expect(monthKeys).not.toContain("2026-07");
  });

  it("aggregates only the past 12 months while carrying opening net forward", () => {
    const monthly = aggregateMonthlyDividends({
      dividends: {
        AAPL: [
          {
            ExDate: "2025-06-15",
            Amount: 100,
            AmountPerShare: 1,
            Qty: 100,
          },
          {
            ExDate: "2026-06-15",
            Amount: 100,
            AmountPerShare: 1,
            Qty: 100,
          },
        ],
        SBFEB50: [
          {
            ExDate: "2026-06-01",
            Amount: 20,
            AmountPerShare: 1,
            Qty: 20,
          },
        ],
        BS24124Z: [
          {
            ExDate: "2026-05-01",
            Amount: 30,
            AmountPerShare: 1,
            Qty: 30,
          },
        ],
      },
      trades: [
        trade("2025-06-01", "AAPL", "buy", 10, 10),
        trade("2026-05-01", "AAPL", "buy", 2, 25),
        trade("2026-06-01", "AAPL", "sell", 1, 20),
      ],
      fx: {
        USD: 1.35,
        SGD: 1,
      },
      refData,
      now: new Date(2026, 5, 15),
    });

    expect(monthly).toHaveLength(12);
    expect(monthly[0].Month).toBe("2026-06");
    expect(monthly[11].Month).toBe("2025-07");

    const june2026 = monthly[0];
    expect(june2026.Dividends).toBeCloseTo(155);
    expect(june2026.DividendsEquity).toBeCloseTo(135);
    expect(june2026.DividendsSSB).toBeCloseTo(20);
    expect(june2026.Sales).toBeCloseTo(20);
    expect(june2026.CumulativeNet).toBeCloseTo(130);
    expect(june2026.CumulativeNetExclGov).toBeCloseTo(130);

    const may2026 = monthly[1];
    expect(may2026.DividendsTBill).toBeCloseTo(30);
    expect(may2026.Purchases).toBeCloseTo(50);
    expect(may2026.CumulativeNet).toBeCloseTo(150);

    const july2025 = monthly[11];
    expect(july2025.Dividends).toBe(0);
    expect(july2025.CumulativeNet).toBeCloseTo(100);
  });

  it("aggregates explicit year selections without future months", () => {
    const monthly = aggregateMonthlyDividends({
      dividends: {
        AAPL: [
          {
            ExDate: "2025-01-15",
            Amount: 10,
            AmountPerShare: 1,
            Qty: 10,
          },
          {
            ExDate: "2026-12-15",
            Amount: 20,
            AmountPerShare: 1,
            Qty: 20,
          },
          {
            ExDate: "2024-12-15",
            Amount: 30,
            AmountPerShare: 1,
            Qty: 30,
          },
        ],
      },
      trades: [trade("2024-12-01", "AAPL", "buy", 10, 10)],
      fx: {
        USD: 1,
      },
      refData,
      years: [2025, 2026],
      now: new Date(2026, 5, 15),
    });

    expect(monthly).toHaveLength(18);
    expect(monthly[0].Month).toBe("2026-06");
    expect(monthly[17].Month).toBe("2025-01");
    expect(monthly.find((row) => row.Month === "2026-12")).toBeUndefined();
    expect(monthly.find((row) => row.Month === "2026-07")).toBeUndefined();
    expect(monthly.find((row) => row.Month === "2025-01")?.Dividends).toBe(10);
    expect(monthly.find((row) => row.Month === "2024-12")).toBeUndefined();
    expect(
      monthly.find((row) => row.Month === "2025-01")?.CumulativeNet
    ).toBe(100);
  });

  it("pivots monthly dividends into month rows and year metric columns", () => {
    const monthly = aggregateMonthlyDividends({
      dividends: {
        AAPL: [
          {
            ExDate: "2025-01-15",
            Amount: 10,
            AmountPerShare: 1,
            Qty: 10,
          },
          {
            ExDate: "2026-01-15",
            Amount: 20,
            AmountPerShare: 1,
            Qty: 20,
          },
          {
            ExDate: "2026-02-15",
            Amount: 30,
            AmountPerShare: 1,
            Qty: 30,
          },
        ],
      },
      trades: [trade("2024-12-01", "AAPL", "buy", 100, 1)],
      fx: {
        USD: 1,
      },
      refData,
      years: [2025, 2026],
      now: new Date(2026, 1, 15),
    });

    const pivoted = pivotMonthlyDividendsByYear(monthly);

    expect(getYearsFromMonthlyDividends(monthly)).toEqual([2026, 2025]);
    expect(pivoted.map((row) => row.Month)).toEqual([
      "Jan",
      "Feb",
      "Mar",
      "Apr",
      "May",
      "Jun",
      "Jul",
      "Aug",
      "Sep",
      "Oct",
      "Nov",
      "Dec",
    ]);
    expect(pivoted[0]["2025Dividends"]).toBe(10);
    expect(pivoted[0]["2026Dividends"]).toBe(20);
    expect(pivoted[1]["2026Dividends"]).toBe(30);
    expect(pivoted[2]["2026Dividends"]).toBeUndefined();
  });
});
