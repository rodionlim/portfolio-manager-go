import { Trade } from "../../types/blotter";
import { ReferenceData } from "../../types";

export interface Dividend {
  ExDate: string;
  Amount: number;
  AmountPerShare: number;
  Qty: number;
}

export interface FxRates {
  [currency: string]: number;
}

export interface MonthlyDividends {
  Month: string;
  MonthLabel: string;
  Dividends: number;
  DividendsSSB: number;
  DividendsTBill: number;
  DividendsEquity: number;
  Purchases: number;
  Sales: number;
  Net: number;
  CumulativeNet: number;
  DividendYield: number;
  PurchasesExclGov: number;
  SalesExclGov: number;
  NetExclGov: number;
  CumulativeNetExclGov: number;
  DividendYieldExclGov: number;
}

export interface PivotedMonthlyDividends {
  Month: string;
  MonthIndex: number;
  [key: string]: string | number;
}

interface AggregateMonthlyDividendsParams {
  dividends: Record<string, Dividend[]>;
  trades: Trade[];
  fx: FxRates;
  refData: ReferenceData;
  years?: number[];
  now?: Date;
}

const MONTH_COUNT = 12;

const formatMonthKey = (year: number, monthIndex: number): string => {
  return `${year}-${String(monthIndex + 1).padStart(2, "0")}`;
};

const getMonthKey = (dateString: string): string | null => {
  const match = /^(\d{4})-(\d{2})/.exec(dateString);
  if (match) {
    return `${match[1]}-${match[2]}`;
  }

  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) {
    return null;
  }

  return formatMonthKey(date.getFullYear(), date.getMonth());
};

const getMonthDate = (monthKey: string): Date => {
  const [year, month] = monthKey.split("-").map(Number);
  return new Date(year, month - 1, 1);
};

const getMonthName = (monthIndex: number): string => {
  return new Date(2000, monthIndex - 1, 1).toLocaleDateString("en-US", {
    month: "short",
  });
};

const createEmptyMonth = (monthKey: string): MonthlyDividends => {
  return {
    Month: monthKey,
    MonthLabel: getMonthDate(monthKey).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
    }),
    Dividends: 0,
    DividendsSSB: 0,
    DividendsTBill: 0,
    DividendsEquity: 0,
    Purchases: 0,
    Sales: 0,
    Net: 0,
    CumulativeNet: 0,
    DividendYield: 0,
    PurchasesExclGov: 0,
    SalesExclGov: 0,
    NetExclGov: 0,
    CumulativeNetExclGov: 0,
    DividendYieldExclGov: 0,
  };
};

export const getLastTwelveMonthKeys = (now = new Date()): string[] => {
  const months: string[] = [];
  const currentMonth = new Date(now.getFullYear(), now.getMonth(), 1);

  for (let offset = MONTH_COUNT - 1; offset >= 0; offset -= 1) {
    const date = new Date(
      currentMonth.getFullYear(),
      currentMonth.getMonth() - offset,
      1
    );
    months.push(formatMonthKey(date.getFullYear(), date.getMonth()));
  }

  return months;
};

export const getMonthKeysForYears = (years: number[]): string[] => {
  const uniqueYears = [...new Set(years)]
    .filter((year) => Number.isInteger(year))
    .sort((a, b) => a - b);

  return uniqueYears.flatMap((year) =>
    Array.from({ length: 12 }, (_, monthIndex) =>
      formatMonthKey(year, monthIndex)
    )
  );
};

export const removeFutureMonthKeys = (
  monthKeys: string[],
  now = new Date()
): string[] => {
  const currentMonthKey = formatMonthKey(now.getFullYear(), now.getMonth());
  return monthKeys.filter((monthKey) => monthKey <= currentMonthKey);
};

export const isSgGov = (
  refData: ReferenceData,
  ticker: string
): boolean => {
  return (
    refData[ticker]?.ccy === "SGD" &&
    refData[ticker]?.asset_sub_class === "govies" &&
    refData[ticker]?.asset_class === "bond"
  );
};

export const pivotMonthlyDividendsByYear = (
  monthlyData: MonthlyDividends[]
): PivotedMonthlyDividends[] => {
  const pivotedMonths = new Map<number, PivotedMonthlyDividends>();

  monthlyData.forEach((row) => {
    const [year, month] = row.Month.split("-").map(Number);
    if (!Number.isInteger(year) || !Number.isInteger(month)) return;

    const pivotedRow =
      pivotedMonths.get(month) ??
      ({
        Month: getMonthName(month),
        MonthIndex: month,
      } as PivotedMonthlyDividends);

    pivotedRow[`${year}DividendYield`] = row.DividendYield;
    pivotedRow[`${year}Dividends`] = row.Dividends;
    pivotedMonths.set(month, pivotedRow);
  });

  return [...pivotedMonths.values()].sort(
    (a, b) => a.MonthIndex - b.MonthIndex
  );
};

export const getYearsFromMonthlyDividends = (
  monthlyData: MonthlyDividends[]
): number[] => {
  return [
    ...new Set(
      monthlyData
        .map((row) => Number(row.Month.slice(0, 4)))
        .filter((year) => Number.isInteger(year))
    ),
  ].sort((a, b) => b - a);
};

export const aggregateMonthlyDividends = ({
  dividends,
  trades,
  fx,
  refData,
  years,
  now = new Date(),
}: AggregateMonthlyDividendsParams): MonthlyDividends[] => {
  const monthKeys =
    years && years.length > 0
      ? removeFutureMonthKeys(getMonthKeysForYears(years), now)
      : getLastTwelveMonthKeys(now);
  const firstMonth = monthKeys[0];
  const monthlyData = monthKeys.reduce<Record<string, MonthlyDividends>>(
    (acc, monthKey) => {
      acc[monthKey] = createEmptyMonth(monthKey);
      return acc;
    },
    {}
  );

  Object.entries(dividends).forEach(([ticker, tickerDividends]) => {
    tickerDividends?.forEach((dividend) => {
      const monthKey = getMonthKey(dividend.ExDate);
      if (!monthKey || !monthlyData[monthKey]) return;

      const tickerRef = refData[ticker];
      const dividendAmountInSGD =
        dividend.Amount * (fx[tickerRef?.ccy || "SGD"] ?? 1);
      const monthData = monthlyData[monthKey];
      const sgGov = isSgGov(refData, ticker);

      monthData.Dividends += dividendAmountInSGD;

      if (sgGov && ticker.startsWith("SB")) {
        monthData.DividendsSSB += dividendAmountInSGD;
      } else if (sgGov) {
        monthData.DividendsTBill += dividendAmountInSGD;
      } else {
        monthData.DividendsEquity += dividendAmountInSGD;
      }
    });
  });

  let openingCumulativeNet = 0;
  let openingCumulativeNetExclGov = 0;

  trades?.forEach((trade) => {
    const monthKey = getMonthKey(trade.TradeDate);
    if (!monthKey) return;

    const tradeValue = trade.Quantity * trade.Price * trade.Fx;
    const isBuy = trade.Side.toLowerCase() === "buy";
    const isSell = trade.Side.toLowerCase() === "sell";
    if (!isBuy && !isSell) return;

    const netValue = isBuy ? tradeValue : -tradeValue;
    const sgGov = isSgGov(refData, trade.Ticker);

    if (monthKey < firstMonth) {
      openingCumulativeNet += netValue;
      if (!sgGov) {
        openingCumulativeNetExclGov += netValue;
      }
      return;
    }

    const monthData = monthlyData[monthKey];
    if (!monthData) return;

    if (isBuy) {
      monthData.Purchases += tradeValue;
      if (!sgGov) {
        monthData.PurchasesExclGov += tradeValue;
      }
    } else {
      monthData.Sales += tradeValue;
      if (!sgGov) {
        monthData.SalesExclGov += tradeValue;
      }
    }
  });

  let cumulativeNet = openingCumulativeNet;
  let cumulativeNetExclGov = openingCumulativeNetExclGov;

  monthKeys.forEach((monthKey) => {
    const monthData = monthlyData[monthKey];

    monthData.Net = monthData.Purchases - monthData.Sales;
    cumulativeNet += monthData.Net;
    monthData.CumulativeNet = cumulativeNet;

    monthData.NetExclGov =
      monthData.PurchasesExclGov - monthData.SalesExclGov;
    cumulativeNetExclGov += monthData.NetExclGov;
    monthData.CumulativeNetExclGov = cumulativeNetExclGov;

    monthData.DividendYield =
      monthData.CumulativeNet > 0
        ? (monthData.Dividends / monthData.CumulativeNet) * 100
        : 0;
    monthData.DividendYieldExclGov =
      monthData.CumulativeNetExclGov > 0
        ? (monthData.DividendsEquity / monthData.CumulativeNetExclGov) * 100
        : 0;
  });

  return monthKeys.map((monthKey) => monthlyData[monthKey]).reverse();
};
