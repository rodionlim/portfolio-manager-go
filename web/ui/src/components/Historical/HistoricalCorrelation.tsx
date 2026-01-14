import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  ActionIcon,
  Button,
  Card,
  Collapse,
  Group,
  LoadingOverlay,
  NumberInput,
  Select,
  Stack,
  Switch,
  Table,
  Text,
  Title,
  Tooltip,
  useMantineColorScheme,
  useMantineTheme,
} from "@mantine/core";
import { useSelector } from "react-redux";
import { DateInput } from "@mantine/dates";
import { notifications } from "@mantine/notifications";
import dayjs from "dayjs";
import {
  ColorType,
  createChart,
  LineSeries,
  UTCTimestamp,
} from "lightweight-charts";
import {
  IconChevronDown,
  IconChevronUp,
  IconRefresh,
} from "@tabler/icons-react";
import { getUrl } from "../../utils/url";
import { RootState } from "../../store";

type CorrelationOptions = {
  frequency: "D";
  is_price_series: boolean;
  date_method: "rolling" | "in_sample" | "expanding";
  rollyears: number;
  interval_frequency: "12M";
  using_exponent: boolean;
  ew_lookback: number;
  min_periods: number;
  floor_at_zero: boolean;
};

type CorrelationRequest = {
  from: string;
  resync: boolean;
  options: CorrelationOptions;
};

type CorrelationMatrix = {
  columns: string[];
  values: number[][];
};

type CorrelationPeriod = {
  fit_start: string;
  fit_end: string;
  period_start: string;
  period_end: string;
  no_data: boolean;
  correlation: CorrelationMatrix;
};

type CorrelationResponse = {
  columns: string[];
  periods: CorrelationPeriod[];
};

const clamp = (value: number, min: number, max: number) =>
  Math.min(max, Math.max(min, value));

const correlationToColor = (corr: number, mode: "dark" | "light"): string => {
  const v = clamp(corr, -1, 1);
  const hue = (v + 1) * 60; // -1 -> 0 (red), 0 -> 60 (yellow), 1 -> 120 (green)
  const sat = mode === "dark" ? 65 : 75;
  const light = mode === "dark" ? 28 : 78;
  return `hsl(${hue} ${sat}% ${light}%)`;
};

const textColorForBg = (corr: number, mode: "dark" | "light") => {
  if (mode === "dark") return "#ffffff";
  // In light mode, very light backgrounds benefit from dark text.
  // Near extremes we still keep dark text for readability.
  return Math.abs(corr) > 0.7 ? "#111" : "#111";
};

const formatPeriodLabel = (p: CorrelationPeriod) => {
  const start = dayjs(p.period_start).format("YYYY-MM-DD");
  const end = dayjs(p.period_end).format("YYYY-MM-DD");
  return `${start} → ${end}`;
};

const periodKey = (p: CorrelationPeriod) => `${p.period_start}|${p.period_end}`;

const makePalette = (
  theme: ReturnType<typeof useMantineTheme>,
  mode: "dark" | "light"
) => {
  const idx = mode === "dark" ? 4 : 7;
  return [
    theme.colors.blue[idx],
    theme.colors.green[idx],
    theme.colors.orange[idx],
    theme.colors.red[idx],
    theme.colors.grape[idx],
    theme.colors.cyan[idx],
    theme.colors.violet[idx],
    theme.colors.teal[idx],
    theme.colors.pink[idx],
    theme.colors.lime[idx],
  ].filter(Boolean);
};

const HistoricalCorrelation: React.FC = () => {
  const theme = useMantineTheme();
  const { colorScheme } = useMantineColorScheme();
  const mode: "dark" | "light" = colorScheme === "dark" ? "dark" : "light";

  const refData = useSelector((state: RootState) => state.referenceData.data);
  const tickerNameMap = useMemo(() => {
    const map = new Map<string, string>();
    if (!refData) return map;
    for (const item of Object.values(refData)) {
      const name = (item as any)?.name;
      const yahooTicker = (item as any)?.yahoo_ticker;
      const id = (item as any)?.id;
      if (yahooTicker && name) map.set(yahooTicker, name);
      if (id && name) map.set(id, name);
    }
    return map;
  }, [refData]);

  const tickerFullName = (ticker: string) => tickerNameMap.get(ticker) ?? null;

  const [fromDate, setFromDate] = useState<Date | null>(
    dayjs().subtract(5, "year").toDate()
  );

  // For now, only allow the parameters in the sample request.
  const [dateMethod, setDateMethod] = useState<
    "rolling" | "in_sample" | "expanding"
  >("rolling");
  const [rollYears, setRollYears] = useState(1);
  const [ewLookback, setEwLookback] = useState(100);
  const [minPeriods, setMinPeriods] = useState(200);
  const [floorAtZero, setFloorAtZero] = useState(true);

  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<CorrelationResponse | null>(null);
  const [hasRun, setHasRun] = useState(false);
  const [matrixCollapsed, setMatrixCollapsed] = useState(false);
  const [timeSeriesCollapsed, setTimeSeriesCollapsed] = useState(false);

  const periods = useMemo(() => data?.periods ?? [], [data]);

  const sortedPeriods = useMemo(() => {
    return [...periods].sort(
      (a, b) => dayjs(b.period_end).valueOf() - dayjs(a.period_end).valueOf()
    );
  }, [periods]);

  const periodOptions = useMemo(() => {
    return sortedPeriods.map((p) => ({
      value: periodKey(p),
      label: formatPeriodLabel(p) + (p.no_data ? " (no data)" : ""),
    }));
  }, [sortedPeriods]);

  const [selectedPeriodKey, setSelectedPeriodKey] = useState<string | null>(
    null
  );

  const selectedPeriod = useMemo(() => {
    if (!sortedPeriods.length) return null;
    if (!selectedPeriodKey) return sortedPeriods[0];
    return (
      sortedPeriods.find((p) => periodKey(p) === selectedPeriodKey) ??
      sortedPeriods[0]
    );
  }, [sortedPeriods, selectedPeriodKey]);

  const columns = useMemo(() => {
    return (
      selectedPeriod?.correlation?.columns ??
      data?.columns ??
      data?.periods?.[0]?.correlation?.columns ??
      []
    );
  }, [data, selectedPeriod]);

  const [baseTicker, setBaseTicker] = useState<string>("");

  useEffect(() => {
    if (!columns.length) return;
    setBaseTicker((prev) => (prev ? prev : columns[0]));
  }, [columns]);

  const requestBody: CorrelationRequest | null = useMemo(() => {
    if (!fromDate) return null;
    return {
      from: dayjs(fromDate).format("YYYY-MM-DD"),
      resync: false,
      options: {
        frequency: "D",
        is_price_series: true,
        date_method: dateMethod,
        rollyears: rollYears,
        interval_frequency: "12M",
        using_exponent: true,
        ew_lookback: ewLookback,
        min_periods: minPeriods,
        floor_at_zero: floorAtZero,
      },
    };
  }, [fromDate, dateMethod, rollYears, ewLookback, minPeriods, floorAtZero]);

  const fetchCorrelation = async () => {
    if (!requestBody) return;

    setHasRun(true);
    setLoading(true);
    try {
      const resp = await fetch(getUrl("/api/v1/historical/correlation"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(requestBody),
      });
      const json = await resp.json();
      if (!resp.ok)
        throw new Error(json?.error || "Failed to fetch correlation");

      setData(json);
      const respPeriods: CorrelationPeriod[] =
        json?.periods.filter((x: any) => !x.no_data) ?? [];
      const sorted = [...respPeriods].sort(
        (a, b) => dayjs(b.period_end).valueOf() - dayjs(a.period_end).valueOf()
      );
      setSelectedPeriodKey(sorted.length ? periodKey(sorted[0]) : null);

      const initialCols: string[] =
        json?.periods?.[0]?.correlation?.columns ?? json?.columns ?? [];
      const sortedTickers = [...initialCols].sort((a, b) => a.localeCompare(b));
      setBaseTicker(sortedTickers[0] ?? "");
    } catch (e: any) {
      notifications.show({
        title: "Error",
        message: e?.message || "Failed to fetch correlation",
        color: "red",
      });
    } finally {
      setLoading(false);
    }
  };

  // ---- Line chart (correlation over time) ----
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<ReturnType<typeof createChart> | null>(null);
  const seriesByTickerRef = useRef<Record<string, any>>({});

  const correlationSeries = useMemo(() => {
    if (!data || !baseTicker)
      return [] as Array<{
        ticker: string;
        points: Array<{ time: UTCTimestamp; value: number }>;
      }>;

    const tickers = (data.columns ?? []).filter((t) => t && t !== baseTicker);
    const pointsByTicker: Record<
      string,
      Array<{ time: UTCTimestamp; value: number }>
    > = {};
    for (const t of tickers) pointsByTicker[t] = [];

    for (const p of data.periods ?? []) {
      if (p.no_data) continue;

      const periodCols: string[] = p.correlation?.columns ?? data.columns ?? [];
      const baseIdx = periodCols.indexOf(baseTicker);
      if (baseIdx === -1) continue;

      const tEnd = dayjs(p.period_end);
      const ts = Math.floor(tEnd.valueOf() / 1000) as UTCTimestamp;

      for (const t of tickers) {
        const otherIdx = periodCols.indexOf(t);
        if (otherIdx === -1) continue;
        const value = p.correlation?.values?.[baseIdx]?.[otherIdx];
        if (typeof value !== "number" || Number.isNaN(value)) continue;
        pointsByTicker[t].push({ time: ts, value });
      }
    }

    return tickers
      .map((ticker) => {
        const points = pointsByTicker[ticker] ?? [];
        points.sort((a, b) => a.time - b.time);
        return { ticker, points };
      })
      .filter((s) => s.points.length > 0);
  }, [data, baseTicker]);

  useEffect(() => {
    if (!hasRun) return;
    if (timeSeriesCollapsed) return;
    if (!chartContainerRef.current) return;

    // create once
    if (!chartRef.current) {
      const bg = mode === "dark" ? theme.colors.dark[7] : "#ffffff";
      const fg = mode === "dark" ? theme.colors.gray[2] : "#333333";
      const grid = mode === "dark" ? theme.colors.dark[5] : "#f0f0f0";

      const chart = createChart(chartContainerRef.current, {
        layout: {
          background: { type: ColorType.Solid, color: bg },
          textColor: fg,
          attributionLogo: false,
        },
        grid: {
          vertLines: { color: grid },
          horzLines: { color: grid },
        },
        width: chartContainerRef.current.clientWidth,
        height: 320,
        timeScale: {
          timeVisible: true,
          secondsVisible: false,
        },
        rightPriceScale: {
          scaleMargins: { top: 0.15, bottom: 0.15 },
        },
      });

      chartRef.current = chart;
      seriesByTickerRef.current = {};

      const resizeObserver = new ResizeObserver(() => {
        if (!chartContainerRef.current || !chartRef.current) return;
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
        });
        chartRef.current.timeScale().fitContent();
      });
      resizeObserver.observe(chartContainerRef.current);

      return () => {
        resizeObserver.disconnect();
        chart.remove();
        chartRef.current = null;
        seriesByTickerRef.current = {};
      };
    }

    return;
  }, [hasRun, timeSeriesCollapsed, mode, theme]);

  // Update chart data and theme when dependencies change.
  useEffect(() => {
    if (!hasRun) return;
    if (timeSeriesCollapsed) return;
    if (!chartRef.current) return;

    const bg = mode === "dark" ? theme.colors.dark[7] : "#ffffff";
    const fg = mode === "dark" ? theme.colors.gray[2] : "#333333";
    const grid = mode === "dark" ? theme.colors.dark[5] : "#f0f0f0";
    const palette = makePalette(theme, mode);

    chartRef.current.applyOptions({
      layout: {
        background: { type: ColorType.Solid, color: bg },
        textColor: fg,
        attributionLogo: false,
      },
      grid: {
        vertLines: { color: grid },
        horzLines: { color: grid },
      },
    });

    const desiredTickers = new Set(correlationSeries.map((s) => s.ticker));
    for (const existing of Object.keys(seriesByTickerRef.current)) {
      if (!desiredTickers.has(existing)) {
        try {
          chartRef.current.removeSeries(seriesByTickerRef.current[existing]);
        } catch {
          // ignore
        }
        delete seriesByTickerRef.current[existing];
      }
    }

    correlationSeries.forEach((s, i) => {
      if (!seriesByTickerRef.current[s.ticker]) {
        const series = chartRef.current!.addSeries(LineSeries, {
          color:
            palette[i % palette.length] ||
            (mode === "dark" ? theme.colors.blue[4] : theme.colors.blue[7]),
          lineWidth: 2,
          title: s.ticker,
        });
        seriesByTickerRef.current[s.ticker] = series;
      }

      const series = seriesByTickerRef.current[s.ticker];
      series.setData(s.points);
      series.applyOptions({
        color:
          palette[i % palette.length] ||
          (mode === "dark" ? theme.colors.blue[4] : theme.colors.blue[7]),
      });
    });

    chartRef.current.timeScale().fitContent();
  }, [hasRun, timeSeriesCollapsed, correlationSeries, mode, theme]);

  const sortedTickers = useMemo(() => {
    return [...columns].sort((a, b) => a.localeCompare(b));
  }, [columns]);

  const matrixIndex = useMemo(() => {
    const map = new Map<string, number>();
    const cols = selectedPeriod?.correlation?.columns ?? [];
    cols.forEach((c, i) => map.set(c, i));
    return map;
  }, [selectedPeriod]);

  const valueAtTickers = (
    matrix: CorrelationMatrix,
    rowTicker: string,
    colTicker: string
  ) => {
    const r = matrixIndex.get(rowTicker);
    const c = matrixIndex.get(colTicker);
    if (r === undefined || c === undefined) return null;
    const v = matrix?.values?.[r]?.[c];
    return typeof v === "number" ? v : null;
  };

  const content = (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end" wrap="wrap">
        <div>
          <Title order={3}>Correlation</Title>
          <Text size="sm" c="dimmed">
            Heatmap correlation matrix + correlation over time.
          </Text>
        </div>

        <Tooltip label="Fetch correlation" withArrow>
          <ActionIcon
            variant="light"
            color="blue"
            onClick={fetchCorrelation}
            loading={loading}
            size="lg"
          >
            <IconRefresh size={18} />
          </ActionIcon>
        </Tooltip>
      </Group>

      <Card withBorder shadow="sm">
        <Stack gap="sm">
          <Group align="flex-end" wrap="wrap">
            <DateInput
              label="From"
              value={fromDate}
              onChange={setFromDate}
              clearable={false}
            />

            <NumberInput
              label="Rolling years"
              min={1}
              max={10}
              value={rollYears}
              onChange={(v) => setRollYears(Number(v) || 1)}
              style={{ width: 140 }}
            />

            <Select
              label="Interval"
              value="12M"
              data={[{ value: "12M", label: "12M" }]}
              disabled
              style={{ width: 110 }}
            />

            <Select
              label="Date Method"
              value={dateMethod}
              onChange={(v) =>
                setDateMethod(
                  (v as "rolling" | "in_sample" | "expanding") || "rolling"
                )
              }
              data={[
                { value: "rolling", label: "rolling" },
                { value: "in_sample", label: "in_sample" },
                { value: "expanding", label: "expanding" },
              ]}
              style={{ width: 140 }}
            />

            <Select
              label="Frequency"
              value="D"
              data={[{ value: "D", label: "D" }]}
              disabled
              style={{ width: 110 }}
            />

            <NumberInput
              label="EW lookback"
              min={1}
              max={5000}
              value={ewLookback}
              onChange={(v) => setEwLookback(Number(v) || 100)}
              style={{ width: 140 }}
            />

            <NumberInput
              label="Min periods"
              min={1}
              max={100000}
              value={minPeriods}
              onChange={(v) => setMinPeriods(Number(v) || 200)}
              style={{ width: 140 }}
            />

            <Switch
              label="Floor at zero"
              checked={floorAtZero}
              onChange={(e) => setFloorAtZero(e.currentTarget.checked)}
            />

            <Button onClick={fetchCorrelation} loading={loading}>
              Run
            </Button>
          </Group>

          <Text size="xs" c="dimmed">
            Note: not all correlation parameters are exposed in the UI.
          </Text>
        </Stack>
      </Card>

      {hasRun && (
        <>
          <Card withBorder shadow="sm" style={{ position: "relative" }}>
            <LoadingOverlay visible={loading} />
            <Stack gap="md">
              <Group justify="space-between" align="flex-end" wrap="wrap">
                <Title order={4}>Correlation Matrix (Heatmap)</Title>

                <Group>
                  <ActionIcon
                    variant="subtle"
                    onClick={() => setMatrixCollapsed((v) => !v)}
                    aria-label={
                      matrixCollapsed ? "Expand matrix" : "Collapse matrix"
                    }
                  >
                    {matrixCollapsed ? (
                      <IconChevronDown size={18} />
                    ) : (
                      <IconChevronUp size={18} />
                    )}
                  </ActionIcon>

                  <Select
                    label="Period"
                    placeholder="Select period"
                    value={selectedPeriodKey}
                    onChange={setSelectedPeriodKey}
                    data={periodOptions}
                    style={{ minWidth: 260 }}
                    disabled={!periodOptions.length}
                  />
                </Group>
              </Group>

              {!selectedPeriod && (
                <Text size="sm" c="dimmed">
                  No results. Try adjusting inputs and click Run.
                </Text>
              )}

              {selectedPeriod && selectedPeriod.no_data && (
                <Text size="sm" c="orange">
                  This period is marked as no-data.
                </Text>
              )}

              <Collapse in={!matrixCollapsed}>
                {selectedPeriod && (
                  <Table.ScrollContainer minWidth={900}>
                    <div style={{ maxHeight: 520, overflowY: "auto" }}>
                      <Table withTableBorder withColumnBorders>
                        <Table.Thead>
                          <Table.Tr>
                            <Table.Th>Asset</Table.Th>
                            {sortedTickers.map((t) => {
                              const fullName = tickerFullName(t);
                              return (
                                <Table.Th key={t}>
                                  {fullName ? (
                                    <Tooltip label={fullName} withArrow>
                                      <Text fw={600}>{t}</Text>
                                    </Tooltip>
                                  ) : (
                                    <Text fw={600}>{t}</Text>
                                  )}
                                </Table.Th>
                              );
                            })}
                          </Table.Tr>
                        </Table.Thead>
                        <Table.Tbody>
                          {sortedTickers.map((rowTicker) => {
                            const fullName = tickerFullName(rowTicker);
                            return (
                              <Table.Tr key={rowTicker}>
                                <Table.Td>
                                  {fullName ? (
                                    <Tooltip label={fullName} withArrow>
                                      <Text fw={600}>{rowTicker}</Text>
                                    </Tooltip>
                                  ) : (
                                    <Text fw={600}>{rowTicker}</Text>
                                  )}
                                </Table.Td>
                                {sortedTickers.map((colTicker) => {
                                  const v = valueAtTickers(
                                    selectedPeriod.correlation,
                                    rowTicker,
                                    colTicker
                                  );
                                  const display =
                                    typeof v === "number" ? v.toFixed(3) : "-";
                                  const bg =
                                    typeof v === "number"
                                      ? correlationToColor(v, mode)
                                      : mode === "dark"
                                      ? theme.colors.dark[6]
                                      : theme.colors.gray[1];
                                  const tc =
                                    typeof v === "number"
                                      ? textColorForBg(v, mode)
                                      : mode === "dark"
                                      ? theme.colors.gray[2]
                                      : theme.colors.dark[9];

                                  return (
                                    <Table.Td
                                      key={`${rowTicker}-${colTicker}`}
                                      style={{
                                        backgroundColor: bg,
                                        color: tc,
                                        textAlign: "center",
                                        fontVariantNumeric: "tabular-nums",
                                      }}
                                    >
                                      {display}
                                    </Table.Td>
                                  );
                                })}
                              </Table.Tr>
                            );
                          })}
                        </Table.Tbody>
                      </Table>
                    </div>
                  </Table.ScrollContainer>
                )}
              </Collapse>
            </Stack>
          </Card>

          <Card withBorder shadow="sm">
            <Stack gap="md">
              <Group justify="space-between" align="flex-end" wrap="wrap">
                <Title order={4}>Correlation Over Time</Title>
                <Group>
                  <ActionIcon
                    variant="subtle"
                    onClick={() => setTimeSeriesCollapsed((v) => !v)}
                    aria-label={
                      timeSeriesCollapsed
                        ? "Expand time series"
                        : "Collapse time series"
                    }
                  >
                    {timeSeriesCollapsed ? (
                      <IconChevronDown size={18} />
                    ) : (
                      <IconChevronUp size={18} />
                    )}
                  </ActionIcon>
                  <Select
                    label="Base ticker"
                    value={baseTicker}
                    onChange={(v) => setBaseTicker(v || "")}
                    data={sortedTickers.map((c) => ({ value: c, label: c }))}
                    disabled={!columns.length}
                    style={{ minWidth: 240 }}
                  />
                </Group>
              </Group>

              <Text size="sm" c="dimmed">
                Shows correlation of every ticker against the base.
              </Text>

              {data && baseTicker && correlationSeries.length === 0 && (
                <Text size="sm" c="dimmed">
                  No data points available for the selected base ticker.
                </Text>
              )}

              <Collapse in={!timeSeriesCollapsed}>
                <div
                  ref={chartContainerRef}
                  style={{
                    width: "100%",
                    borderRadius: theme.radius.md,
                    overflow: "hidden",
                  }}
                />
              </Collapse>

              {data && (
                <Text size="xs" c="dimmed">
                  X-axis uses period end dates (yearly rolling windows).
                </Text>
              )}
            </Stack>
          </Card>
        </>
      )}
    </Stack>
  );

  return <div style={{ marginTop: 24 }}>{content}</div>;
};

export default HistoricalCorrelation;
