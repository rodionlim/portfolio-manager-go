import type { TimestampedMetrics } from "../Analytics/types";

export interface HeadlinePnlMetrics {
  asOf?: string;
  totalDividends?: number;
  wtd?: number;
  mtd?: number;
  qtd?: number;
  ytd?: number;
}

const snapshotPnl = (snapshot: TimestampedMetrics) =>
  snapshot.metrics.mv -
  snapshot.metrics.pricePaid +
  snapshot.metrics.totalDividends;

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
      totalDividends: snapshot.metrics.totalDividends,
    }))
    .filter(
      (snapshot) =>
        Number.isFinite(snapshot.timestamp) &&
        Number.isFinite(snapshot.pnl) &&
        Number.isFinite(snapshot.marketValue) &&
        Number.isFinite(snapshot.totalDividends),
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
  const weekStart = new Date(latest.timestamp);
  const daysSinceMonday = (latestDate.getUTCDay() + 6) % 7;
  weekStart.setUTCDate(latestDate.getUTCDate() - daysSinceMonday);
  const monthStart = new Date(
    Date.UTC(latestDate.getUTCFullYear(), latestDate.getUTCMonth(), 1),
  );
  const quarterStart = new Date(
    Date.UTC(
      latestDate.getUTCFullYear(),
      Math.floor(latestDate.getUTCMonth() / 3) * 3,
      1,
    ),
  );
  const yearStart = new Date(Date.UTC(latestDate.getUTCFullYear(), 0, 1));

  return {
    asOf: latestDate.toISOString(),
    totalDividends: latestSnapshot.totalDividends,
    wtd: pnlSince(snapshots, latest, weekStart),
    mtd: pnlSince(snapshots, latest, monthStart),
    qtd: pnlSince(snapshots, latest, quarterStart),
    ytd: pnlSince(snapshots, latest, yearStart),
  };
};
