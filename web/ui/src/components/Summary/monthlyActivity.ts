import type { TimestampedMetrics } from "../Analytics/types";
import type { Trade } from "../../types/blotter";

export interface MonthlyTickerActivity {
  ticker: string;
  earliestTradeDate: string;
  quantityBought: number;
  quantitySold: number;
  netQuantity: number;
  averagePrice: number;
  pricePaid: number;
}

export interface MonthlyPortfolioActivity {
  monthKey: string;
  monthLabel: string;
  mtdPnl?: number;
  dividends?: number;
  netCashFlow: number;
  marketValue?: number;
  trades: MonthlyTickerActivity[];
}

interface NormalizedSnapshot {
  timestamp: number;
  pnl: number;
  dividends: number;
  marketValue: number;
}

const monthNames = [
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
];

const dateOnlyTimestamp = (timestamp: string) => {
  const [year, month, day] = timestamp.slice(0, 10).split("-").map(Number);
  if (![year, month, day].every(Number.isInteger)) {
    return Number.NaN;
  }
  return Date.UTC(year, month - 1, day);
};

const normalizeSnapshots = (
  historicalMetrics: TimestampedMetrics[],
): NormalizedSnapshot[] =>
  historicalMetrics
    .map((snapshot) => ({
      timestamp: dateOnlyTimestamp(snapshot.timestamp),
      pnl:
        snapshot.metrics.mv -
        snapshot.metrics.pricePaid +
        snapshot.metrics.totalDividends,
      dividends: snapshot.metrics.totalDividends,
      marketValue: snapshot.metrics.mv,
    }))
    .filter((snapshot) =>
      Object.values(snapshot).every((value) => Number.isFinite(value)),
    )
    .sort((left, right) => left.timestamp - right.timestamp);

const aggregateTrades = (
  trades: Trade[],
  monthStart: number,
  nextMonthStart: number,
) => {
  const tickerTrades = new Map<
    string,
    {
      quantityBought: number;
      quantitySold: number;
      totalQuantity: number;
      weightedPrice: number;
      pricePaid: number;
      earliestTradeTimestamp: number;
    }
  >();

  trades.forEach((trade) => {
    const tradeTimestamp = dateOnlyTimestamp(trade.TradeDate);
    if (
      !["buy", "sell"].includes(trade.Side.toLowerCase()) ||
      tradeTimestamp < monthStart ||
      tradeTimestamp >= nextMonthStart ||
      !Number.isFinite(trade.Quantity) ||
      !Number.isFinite(trade.Price)
    ) {
      return;
    }

    const current = tickerTrades.get(trade.Ticker) || {
      quantityBought: 0,
      quantitySold: 0,
      totalQuantity: 0,
      weightedPrice: 0,
      pricePaid: 0,
      earliestTradeTimestamp: tradeTimestamp,
    };
    const fx = Number.isFinite(trade.Fx) && trade.Fx > 0 ? trade.Fx : 1;
    const isBuy = trade.Side.toLowerCase() === "buy";
    if (isBuy) {
      current.quantityBought += trade.Quantity;
    } else {
      current.quantitySold += trade.Quantity;
    }
    current.totalQuantity += trade.Quantity;
    current.weightedPrice += trade.Price * trade.Quantity;
    current.pricePaid +=
      trade.Price * trade.Quantity * fx * (isBuy ? 1 : -1);
    current.earliestTradeTimestamp = Math.min(
      current.earliestTradeTimestamp,
      tradeTimestamp,
    );
    tickerTrades.set(trade.Ticker, current);
  });

  return Array.from(tickerTrades.entries())
    .map(([ticker, purchase]) => ({
      ticker,
      earliestTradeDate: new Date(purchase.earliestTradeTimestamp)
        .toISOString()
        .slice(0, 10),
      quantityBought: purchase.quantityBought,
      quantitySold: purchase.quantitySold,
      netQuantity: purchase.quantityBought - purchase.quantitySold,
      averagePrice: purchase.totalQuantity
        ? purchase.weightedPrice / purchase.totalQuantity
        : 0,
      pricePaid: purchase.pricePaid,
    }))
    .sort(
      (left, right) =>
        right.pricePaid - left.pricePaid ||
        left.ticker.localeCompare(right.ticker),
    );
};

export const buildMonthlyPortfolioActivity = (
  historicalMetrics: TimestampedMetrics[],
  trades: Trade[],
  now = new Date(),
  monthCount = 12,
): MonthlyPortfolioActivity[] => {
  const snapshots = normalizeSnapshots(historicalMetrics);
  const currentYear = now.getFullYear();
  const currentMonth = now.getMonth();

  return Array.from({ length: monthCount }, (_, index) => {
    const monthOffset = -index;
    const monthStartDate = new Date(
      Date.UTC(currentYear, currentMonth + monthOffset, 1),
    );
    const monthStart = monthStartDate.getTime();
    const nextMonthStart = Date.UTC(
      monthStartDate.getUTCFullYear(),
      monthStartDate.getUTCMonth() + 1,
      1,
    );
    const snapshotsInMonth = snapshots.filter(
      (snapshot) =>
        snapshot.timestamp >= monthStart &&
        snapshot.timestamp < nextMonthStart,
    );
    const latest = snapshotsInMonth.at(-1);
    const baseline = latest
      ? [...snapshots]
          .reverse()
          .find(
            (snapshot) =>
              snapshot.timestamp <= monthStart &&
              snapshot.timestamp < latest.timestamp,
          )
      : undefined;
    const year = monthStartDate.getUTCFullYear();
    const month = monthStartDate.getUTCMonth();
    const monthlyTrades = aggregateTrades(trades, monthStart, nextMonthStart);

    return {
      monthKey: `${year}-${String(month + 1).padStart(2, "0")}`,
      monthLabel: `${monthNames[month]}-${String(year).slice(-2)}`,
      mtdPnl: latest && baseline ? latest.pnl - baseline.pnl : undefined,
      dividends:
        latest && baseline
          ? latest.dividends - baseline.dividends
          : undefined,
      netCashFlow: monthlyTrades.reduce(
        (sum, trade) => sum + trade.pricePaid,
        0,
      ),
      marketValue: latest?.marketValue,
      trades: monthlyTrades,
    };
  });
};
