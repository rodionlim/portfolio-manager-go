// Common interfaces for metrics-related components
export interface Metrics {
  irr: number;
  pricePaid: number;
  mv: number;
  totalDividends: number;
  /** Annualized rolling volatility as decimal (e.g. 0.12 = 12%). Computed client-side. */
  standardDeviation?: number;
  /** Debug: adjusted portfolio value = mv + totalDividends. Computed client-side. */
  adjustedValue?: number;
  /** Debug: daily P&L in currency, cash-flow-adjusted and dividend-inclusive. Computed client-side. */
  dailyPnl?: number;
  /** Debug: per-period adjusted return used for volatility. Computed client-side. */
  adjustedReturn?: number;
}

export interface TimestampedMetrics {
  timestamp: string;
  metrics: Metrics;
}

export interface MetricsJob {
  BookFilter: string;
  CronExpr: string;
  TaskId: string;
}
