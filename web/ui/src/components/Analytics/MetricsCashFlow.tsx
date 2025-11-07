import React, { useMemo, useState, useEffect } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import {
  Paper,
  Title,
  Text,
  Group,
  Stack,
  Button,
  Tooltip,
} from "@mantine/core";
import { IconListDetails, IconListTree } from "@tabler/icons-react";
import { getUrl } from "../../utils/url";
import { useSelector } from "react-redux";
import { RootState } from "../../store";

interface CashFlow {
  date: string;
  cash: number;
  ticker: string;
  description: string;
}

interface AggregatedCashFlow {
  ticker: string;
  totalCash: number;
  count: number;
  buyCash: number;
  sellCash: number;
  dividendCash: number;
  marketValue?: number;
  pnl?: number;
}

interface Position {
  Ticker: string;
  Mv: number;
  FxRate: number;
}

interface Metrics {
  irr: number;
  pricePaid: number;
  mv: number;
  totalDividends: number;
}

interface MetricsResponse {
  metrics: Metrics;
  cashFlows: CashFlow[];
  label: string;
}

const MetricsCashFlow: React.FC = () => {
  const [groupByTicker, setGroupByTicker] = useState(false);
  const [aggregateBonds, setAggregateBonds] = useState(false);
  const refData = useSelector((state: RootState) => state.referenceData.data);

  const {
    data: metricsData,
    isLoading,
    error,
  } = useQuery<MetricsResponse>({
    queryKey: ["metrics"],
    queryFn: async () => {
      const resp = await fetch(getUrl("/api/v1/metrics"));
      if (!resp.ok) {
        throw new Error(`Error fetching metrics: ${resp.statusText}`);
      }
      return resp.json();
    },
    retry: false,
    refetchOnWindowFocus: false,
  });

  // Fetch positions data
  const { data: positions = [] } = useQuery<Position[]>({
    queryKey: ["positions"],
    queryFn: async () => {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions"));
      if (!resp.ok) {
        throw new Error(`Error fetching positions`);
      }
      return resp.json();
    },
    retry: false,
    refetchOnWindowFocus: false,
  });

  // Aggregate cash flows by ticker
  const aggregatedData = useMemo(() => {
    if (!metricsData?.cashFlows) return [];

    // Create a map of positions by ticker for quick lookup
    const positionMap = positions.reduce((acc, pos) => {
      if (!acc[pos.Ticker]) {
        acc[pos.Ticker] = { mv: 0 };
      }
      acc[pos.Ticker].mv += pos.Mv * pos.FxRate;
      return acc;
    }, {} as Record<string, { mv: number }>);

    const aggregated = metricsData.cashFlows.reduce((acc, flow) => {
      // Skip the final portfolio value entry
      if (flow.ticker === "Portfolio") return acc;

      let tickerKey = flow.ticker;

      // Aggregate T-Bills and SSB bonds if enabled
      if (aggregateBonds) {
        if (
          tickerKey.length === 8 &&
          /^[A-Za-z]$/.test(tickerKey[0]) &&
          /^[A-Za-z]$/.test(tickerKey[1]) &&
          /^[A-Za-z]$/.test(tickerKey[tickerKey.length - 1])
        ) {
          tickerKey = "TBill";
        } else if (tickerKey.startsWith("SB") && tickerKey.length === 7) {
          tickerKey = "SSB";
        }
      }

      if (!acc[tickerKey]) {
        acc[tickerKey] = {
          ticker: tickerKey,
          totalCash: 0,
          count: 0,
          buyCash: 0,
          sellCash: 0,
          dividendCash: 0,
          marketValue: positionMap[tickerKey]?.mv || 0,
          pnl: 0,
        };
      }

      acc[tickerKey].totalCash += flow.cash;
      acc[tickerKey].count += 1;

      // Aggregate market values for bonds
      if (aggregateBonds && (tickerKey === "TBill" || tickerKey === "SSB")) {
        acc[tickerKey].marketValue =
          (acc[tickerKey].marketValue || 0) +
          (positionMap[flow.ticker]?.mv || 0);
      }

      if (flow.description === "buy") {
        acc[tickerKey].buyCash += flow.cash;
      } else if (flow.description === "sell") {
        acc[tickerKey].sellCash += flow.cash;
      } else if (flow.description === "dividend") {
        acc[tickerKey].dividendCash += flow.cash;
      }

      return acc;
    }, {} as Record<string, AggregatedCashFlow>);

    // Calculate P&L for each ticker (Market Value + Total Cash Flow)
    Object.values(aggregated).forEach((item) => {
      item.pnl = (item.marketValue || 0) + item.totalCash;
    });

    return Object.values(aggregated).sort((a, b) =>
      a.ticker.localeCompare(b.ticker)
    );
  }, [metricsData?.cashFlows, positions, aggregateBonds]);

  const columns = useMemo<MRT_ColumnDef<CashFlow>[]>(
    () => [
      {
        accessorKey: "ticker",
        header: "Ticker",
        size: 120,
      },
      {
        accessorKey: "date",
        header: "Date",
        size: 150,
        Cell: ({ cell }) => {
          const date = new Date(cell.getValue<string>());
          return date.toLocaleDateString("en-GB");
        },
      },
      {
        accessorKey: "description",
        header: "Type",
        size: 100,
      },
      {
        accessorKey: "cash",
        header: "Cash Flow",
        size: 150,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          const color = value < 0 ? "red" : "green";
          return (
            <Text c={color} fw={500} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              })}
            </Text>
          );
        },
      },
    ],
    []
  );

  const aggregatedColumns = useMemo<MRT_ColumnDef<AggregatedCashFlow>[]>(
    () => [
      {
        accessorKey: "ticker",
        header: "Ticker",
        size: 100,
        Cell: ({ cell }) => {
          const ticker = cell.getValue<string>();
          let name = "";

          if (ticker === "TBill") {
            name = "MAS Bills";
          } else if (ticker === "SSB") {
            name = "SSB";
          } else {
            name = refData?.[ticker]?.name || ticker;
          }

          return (
            <Tooltip label={name} withArrow>
              <span>{ticker}</span>
            </Tooltip>
          );
        },
      },
      {
        accessorKey: "count",
        header: "# Transactions",
        size: 100,
      },
      {
        accessorKey: "buyCash",
        header: "Buy",
        size: 120,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return (
            <Text c="red" fw={400} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </Text>
          );
        },
      },
      {
        accessorKey: "sellCash",
        header: "Sell",
        size: 120,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return (
            <Text c="green" fw={400} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </Text>
          );
        },
      },
      {
        accessorKey: "dividendCash",
        header: "Dividends",
        size: 120,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return (
            <Text c="blue" fw={400} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </Text>
          );
        },
      },
      {
        accessorKey: "totalCash",
        header: "Net Cash Flow",
        size: 130,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          const color = value < 0 ? "red" : "green";
          return (
            <Text c={color} fw={500} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </Text>
          );
        },
      },
      {
        accessorKey: "marketValue",
        header: "Market Value",
        size: 130,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return (
            <Text fw={500} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </Text>
          );
        },
      },
      {
        accessorKey: "pnl",
        header: "P&L",
        size: 130,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          const color = value < 0 ? "red" : "green";
          return (
            <Text c={color} fw={700} size="sm">
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </Text>
          );
        },
      },
    ],
    [refData]
  );

  const table = useMantineReactTable({
    columns: (groupByTicker ? aggregatedColumns : columns) as any,
    data: (groupByTicker
      ? aggregatedData
      : (metricsData?.cashFlows || []).filter(
          (flow) => flow.ticker !== "Portfolio"
        )) as any,
    initialState: {
      sorting: [{ id: "ticker", desc: false }],
      density: "xs",
    },
    state: {
      isLoading,
    },
    enableRowSelection: false,
    enableColumnFilters: true,
    enableGlobalFilter: true,
    autoResetAll: false, // Prevent auto-reset
  });

  // Reset table state when switching between views
  useEffect(() => {
    table.resetColumnFilters();
    table.resetGlobalFilter();
    table.resetSorting();
  }, [groupByTicker, table]);

  if (error) {
    return (
      <Paper p="md" withBorder>
        <Text c="red">Error loading metrics: {error.message}</Text>
      </Paper>
    );
  }

  return (
    <Stack gap="md">
      {metricsData?.metrics && (
        <Paper p="md" withBorder>
          <Title order={3} mb="md">
            Portfolio Metrics
          </Title>
          <Group gap="xl">
            <div>
              <Text size="sm" c="dimmed">
                IRR
              </Text>
              <Text size="xl" fw={700}>
                {(metricsData.metrics.irr * 100).toFixed(2)}%
              </Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">
                Price Paid
              </Text>
              <Text size="xl" fw={700}>
                $
                {metricsData.metrics.pricePaid.toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
              </Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">
                Market Value
              </Text>
              <Text size="xl" fw={700}>
                $
                {metricsData.metrics.mv.toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
              </Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">
                Total Dividends
              </Text>
              <Text size="xl" fw={700}>
                $
                {metricsData.metrics.totalDividends.toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
              </Text>
            </div>
            <div>
              <Text size="sm" c="dimmed">
                Total Return
              </Text>
              <Text size="xl" fw={700} c="green">
                $
                {(
                  metricsData.metrics.mv +
                  metricsData.metrics.totalDividends -
                  metricsData.metrics.pricePaid
                ).toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
              </Text>
            </div>
          </Group>
        </Paper>
      )}

      <Paper p="md" withBorder>
        <Group justify="space-between" mb="md">
          <Title order={3}>Cash Flows</Title>
          <Group gap="xs">
            {groupByTicker && (
              <Button
                onClick={() => setAggregateBonds(!aggregateBonds)}
                variant={aggregateBonds ? "filled" : "light"}
                size="sm"
              >
                {aggregateBonds ? "Show All Bonds" : "Aggregate Bonds"}
              </Button>
            )}
            <Button
              leftSection={
                groupByTicker ? (
                  <IconListDetails size={16} />
                ) : (
                  <IconListTree size={16} />
                )
              }
              onClick={() => setGroupByTicker(!groupByTicker)}
              variant="light"
              size="sm"
            >
              {groupByTicker ? "Show Details" : "Group by Ticker"}
            </Button>
          </Group>
        </Group>
        <MantineReactTable table={table} />
      </Paper>
    </Stack>
  );
};

export default MetricsCashFlow;
