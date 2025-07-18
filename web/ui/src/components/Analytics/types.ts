// Common interfaces for metrics-related components
export interface Metrics {
  irr: number;
  pricePaid: number;
  mv: number;
  totalDividends: number;
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
