import type { TimestampedMetrics } from "../Analytics/types";

export interface HeadlinePnlMetrics {
  asOf?: string;
  oneWeek?: number;
  mtd?: number;
  threeMonth?: number;
  sixMonth?: number;
  ytd?: number;
}

const snapshotPnl = (snapshot: TimestampedMetrics) =>
  snapshot.metrics.mv -
  snapshot.metrics.pricePaid +
  snapshot.metrics.totalDividends;

const subtractCalendarMonths = (date: Date, months: number) => {
  const targetMonth = date.getUTCMonth() - months;
  const lastDayOfTargetMonth = new Date(
    Date.UTC(date.getUTCFullYear(), targetMonth + 1, 0),
  ).getUTCDate();

  return new Date(
    Date.UTC(
      date.getUTCFullYear(),
      targetMonth,
      Math.min(date.getUTCDate(), lastDayOfTargetMonth),
    ),
  );
};

const snapshotDate = (timestamp: string) => {
  const dateParts = timestamp.slice(0, 10).split("-").map(Number);
  if (
    dateParts.length !== 3 ||
    dateParts.some((part) => !Number.isInteger(part))
  ) {
    return Number.NaN;
  }

  const [year, month, day] = dateParts;
  return Date.UTC(year, month - 1, day);
};

const pnlSince = (
  snapshots: Array<{ timestamp: number; pnl: number }>,
  latest: { timestamp: number; pnl: number },
  periodStart: Date,
) => {
  const startTime = periodStart.getTime();
  const baseline = [...snapshots]
    .reverse()
    .find(
      (snapshot) =>
        snapshot.timestamp <= startTime &&
        snapshot.timestamp < latest.timestamp,
    );

  return baseline ? latest.pnl - baseline.pnl : undefined;
};

export const calculateHeadlinePnlMetrics = (
  historicalMetrics: TimestampedMetrics[],
  currentMarketValue?: number,
  now = new Date(),
): HeadlinePnlMetrics => {
  const snapshots = historicalMetrics
    .map((snapshot) => ({
      timestamp: snapshotDate(snapshot.timestamp),
      pnl: snapshotPnl(snapshot),
      marketValue: snapshot.metrics.mv,
    }))
    .filter(
      (snapshot) =>
        Number.isFinite(snapshot.timestamp) &&
        Number.isFinite(snapshot.pnl) &&
        Number.isFinite(snapshot.marketValue),
    )
    .sort((left, right) => left.timestamp - right.timestamp);

  const latestSnapshot = snapshots.at(-1);
  if (!latestSnapshot) {
    return {};
  }

  const latest =
    currentMarketValue !== undefined &&
    Number.isFinite(currentMarketValue)
      ? {
          ...latestSnapshot,
          timestamp: Date.UTC(
            now.getFullYear(),
            now.getMonth(),
            now.getDate(),
          ),
          pnl:
            latestSnapshot.pnl +
            currentMarketValue -
            latestSnapshot.marketValue,
        }
      : latestSnapshot;

  const latestDate = new Date(latest.timestamp);
  const monthStart = new Date(
    Date.UTC(latestDate.getUTCFullYear(), latestDate.getUTCMonth(), 1),
  );
  const yearStart = new Date(Date.UTC(latestDate.getUTCFullYear(), 0, 1));
  const oneWeekStart = new Date(latest.timestamp - 7 * 24 * 60 * 60 * 1_000);
  const threeMonthStart = subtractCalendarMonths(latestDate, 3);
  const sixMonthStart = subtractCalendarMonths(latestDate, 6);

  return {
    asOf: latestDate.toISOString(),
    oneWeek: pnlSince(snapshots, latest, oneWeekStart),
    mtd: pnlSince(snapshots, latest, monthStart),
    threeMonth: pnlSince(snapshots, latest, threeMonthStart),
    sixMonth: pnlSince(snapshots, latest, sixMonthStart),
    ytd: pnlSince(snapshots, latest, yearStart),
  };
};
