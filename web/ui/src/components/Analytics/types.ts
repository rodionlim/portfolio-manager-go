// Common interfaces for metrics-related components
export interface Metrics {
  irr: number;
  pricePaid: number;
  mv: number;
  totalDividends: number;
  /** Annualized rolling volatility as decimal (e.g. 0.12 = 12%). Computed client-side. */
  standardDeviation?: number;
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
