import React, { useMemo, useState, useEffect, useRef } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import { Button, Text, Tooltip, SegmentedControl, Box } from "@mantine/core";
import { getUrl } from "../../utils/url";
import { useNavigate } from "react-router-dom";
import { IconHistory, IconCoins } from "@tabler/icons-react";

import classes from "../../styles.module.css";

interface Position {
  Ticker: string;
  Name: string;
  Book: string;
  Ccy: string;
  AssetClass: string;
  AssetSubClass: string;
  Qty: number;
  Mv: number;
  PnL: number;
  Dividends: number;
  AvgPx: number;
  Px: number;
  FxRate: number;
  RawTicker?: string;
  PrevPx?: number;
  DailyPnl?: number;
  DailyPct?: number;
}

interface CachedPrice {
  ticker: string;
  price: number;
  timestamp: string;
}

interface CachedPricesResponse {
  metrics?: {
    timestamp: string;
    metrics: {
      irr: number;
      pricePaid: number;
      mv: number;
      totalDividends: number;
    };
  };
  prices: CachedPrice[];
  pricesPrev2?: CachedPrice[];
  missing?: string[];
}

const PositionTable: React.FC = () => {
  const navigate = useNavigate();
  const refData = useSelector((state: RootState) => state.referenceData.data);
  const [filteredPositions, setFilteredPositions] = useState<Position[]>([]);
  const [hasBookFilter, setHasBookFilter] = useState(false);
  const [hasAnyFilter, setHasAnyFilter] = useState(false);
  const [priceLag, setPriceLag] = useState<"t-1" | "t-2">("t-1");
  const defaultTitleRef = useRef(document.title);

  const {
    data: rawPositions = [],
    isLoading,
    error,
  } = useQuery<Position[]>({
    queryKey: ["positions"],
    queryFn: async () => {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions"));
      if (!resp.ok) {
        console.error(await resp.json());
        throw new Error(`Error fetching positions`);
      }
      return resp.json();
    },
    retry: false,
    refetchOnWindowFocus: false,
  });

  const uniqueTickers = useMemo(() => {
    const tickers = new Set<string>();
    rawPositions.forEach((position) => {
      if (position.Ticker && position.Qty !== 0) {
        tickers.add(position.Ticker);
      }
    });
    return Array.from(tickers);
  }, [rawPositions]);

  const { data: cachedPricesData } = useQuery<CachedPricesResponse | null>({
    queryKey: ["cachedDailyPrices", uniqueTickers],
    queryFn: async () => {
      if (uniqueTickers.length === 0) return null;
      const resp = await fetch(getUrl("/api/v1/historical/prices/cached"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ tickers: uniqueTickers }),
      });
      if (!resp.ok) {
        return null;
      }
      return resp.json();
    },
    retry: false,
    refetchOnWindowFocus: false,
    enabled: uniqueTickers.length > 0,
  });

  const hasCachedMetrics = Boolean(cachedPricesData?.metrics);

  const cachedPriceMapT1 = useMemo(() => {
    const map = new Map<string, CachedPrice>();
    if (!hasCachedMetrics || !cachedPricesData?.prices) {
      return map;
    }
    cachedPricesData.prices.forEach((price) => {
      map.set(price.ticker, price);
    });
    return map;
  }, [cachedPricesData, hasCachedMetrics]);

  const cachedPriceMapT2 = useMemo(() => {
    const map = new Map<string, CachedPrice>();
    if (!hasCachedMetrics || !cachedPricesData?.pricesPrev2) {
      return map;
    }
    cachedPricesData.pricesPrev2.forEach((price) => {
      map.set(price.ticker, price);
    });
    return map;
  }, [cachedPricesData, hasCachedMetrics]);

  const activeCachedPriceMap =
    priceLag === "t-2" ? cachedPriceMapT2 : cachedPriceMapT1;

  const hasPrev2Prices = Boolean(cachedPricesData?.pricesPrev2?.length);

  // Add this function to handle navigation
  const handleViewTrades = () => {
    const selectedRows = table.getSelectedRowModel().rows;
    if (selectedRows.length === 1) {
      const ticker = selectedRows[0].original.Ticker;
      // Navigate to blotter with ticker filter
      navigate(`/blotter?ticker=${encodeURIComponent(ticker)}`);
    }
  };

  // Add this function to handle navigation to dividends view
  const handleViewDividends = () => {
    const selectedRows = table.getSelectedRowModel().rows;
    if (selectedRows.length === 1) {
      const ticker = selectedRows[0].original.Ticker;
      // Navigate to dividends with ticker filter
      navigate(`/dividends`, {
        state: { ticker: ticker },
      });
    }
  };

  // Use useMemo to aggregate positions using the latest refData as well.
  const aggregatedPositions = useMemo(() => {
    if (!rawPositions) return [];
    return Object.values(
      rawPositions.reduce(
        (acc: Record<string, Position>, curr: Position) => {
          let tickerKey = curr.Ticker; // Normalize ticker key to refdata
          let tickerName: string;

          // If it's a mas tbill, set key to "TBill".
          if (
            tickerKey.length === 8 &&
            /^[A-Za-z]$/.test(tickerKey[0]) &&
            /^[A-Za-z]$/.test(tickerKey[1]) &&
            /^[A-Za-z]$/.test(tickerKey[tickerKey.length - 1])
          ) {
            tickerKey = "TBill";
            tickerName = "MAS Bills";
          } else if (tickerKey.startsWith("SB") && tickerKey.length === 7) {
            // If ticker starts with "SB" and has 7 characters, set key to "SSB".
            tickerKey = "SSB";
            tickerName = "SSB";
          } else {
            // Use updated refData here.
            tickerName = refData?.[tickerKey]?.name ?? "";
          }

          const key = `${tickerKey}-${curr.Book}`; // Aggregate by Ticker and Book
          const rawTicker = curr.Ticker;

          if (acc[key]) {
            acc[key].Qty += curr.Qty;
            acc[key].Mv += curr.Mv * curr.FxRate;
            acc[key].PnL += curr.PnL;
            acc[key].Dividends += curr.Dividends;
            acc[key].Name = tickerName;
            acc[key].Px = curr.Px;
          } else {
            acc[key] = {
              ...curr,
              Ticker: tickerKey,
              Name: tickerName,
              RawTicker: rawTicker,
            };
          }

          return acc;
        },
        {} as Record<string, Position>,
      ),
    );
  }, [rawPositions, refData]);

  const positionsWithDaily = useMemo(() => {
    if (!hasCachedMetrics) return aggregatedPositions;

    return aggregatedPositions.map((position) => {
      const lookupTicker = position.RawTicker || position.Ticker;
      const cachedPrice = activeCachedPriceMap.get(lookupTicker);
      const prevPx = cachedPrice?.price;
      const currentPx = position.Px;
      const fxRate = position.FxRate || 1;

      let dailyPnl = 0;
      let dailyPct = 0;

      if (
        prevPx !== undefined &&
        currentPx !== undefined &&
        prevPx > 1 &&
        currentPx > 1 &&
        position.Qty !== 0
      ) {
        const priceChange = currentPx - prevPx;
        dailyPnl = priceChange * position.Qty * fxRate;
        dailyPct = prevPx ? (priceChange / prevPx) * 100 : 0;
      }

      return {
        ...position,
        PrevPx: prevPx,
        DailyPnl: dailyPnl,
        DailyPct: dailyPct,
      };
    });
  }, [aggregatedPositions, activeCachedPriceMap, hasCachedMetrics]);

  const dailyTotals = useMemo(() => {
    if (!hasCachedMetrics) {
      return { dailyPnl: 0, prevValue: 0, pct: 0 };
    }

    return positionsWithDaily.reduce(
      (acc, row) => {
        const prevPx = row.PrevPx;
        const currentPx = row.Px;
        const fxRate = row.FxRate || 1;

        if (
          prevPx !== undefined &&
          currentPx !== undefined &&
          prevPx > 1 &&
          currentPx > 1 &&
          row.Qty !== 0
        ) {
          const prevValue = prevPx * row.Qty * fxRate;
          acc.prevValue += prevValue;
          acc.dailyPnl += (currentPx - prevPx) * row.Qty * fxRate;
        }

        return acc;
      },
      { dailyPnl: 0, prevValue: 0, pct: 0 },
    );
  }, [positionsWithDaily, hasCachedMetrics]);

  const dailyPnlPct = useMemo(() => {
    if (!dailyTotals.prevValue) return 0;
    return (dailyTotals.dailyPnl / dailyTotals.prevValue) * 100;
  }, [dailyTotals]);

  const displayDailyTotals = useMemo(() => {
    const positions = hasAnyFilter ? filteredPositions : positionsWithDaily;
    if (!positions.length) {
      return { dailyPnl: 0, prevValue: 0 };
    }

    return positions.reduce(
      (acc, row) => {
        const prevPx = row.PrevPx;
        const currentPx = row.Px;
        const fxRate = row.FxRate || 1;

        if (
          prevPx !== undefined &&
          currentPx !== undefined &&
          prevPx > 1 &&
          currentPx > 1 &&
          row.Qty !== 0
        ) {
          acc.prevValue += prevPx * row.Qty * fxRate;
          acc.dailyPnl += (currentPx - prevPx) * row.Qty * fxRate;
        }

        return acc;
      },
      { dailyPnl: 0, prevValue: 0 },
    );
  }, [filteredPositions, positionsWithDaily, hasAnyFilter]);

  // Calculate totals based on filtered positions
  const totals = useMemo(() => {
    const positions =
      filteredPositions.length > 0 ? filteredPositions : positionsWithDaily;
    const res = positions.reduce(
      (acc, row) => {
        acc.Mv += row.Mv * row.FxRate;
        acc.Pnl += row.PnL * row.FxRate;
        acc.Dividends += row.Dividends * row.FxRate;

        if (row.AssetSubClass !== "govies") {
          acc.MvLessGovies += row.Mv * (row.FxRate ? row.FxRate : 1);
        }

        return acc;
      },
      { Mv: 0, MvLessGovies: 0, Pnl: 0, Dividends: 0 },
    );
    return res;
  }, [filteredPositions, positionsWithDaily]);

  const columns = useMemo<MRT_ColumnDef<Position>[]>(() => {
    const dailyPctLabel =
      hasCachedMetrics && !hasAnyFilter && dailyTotals.prevValue
        ? ` (${dailyPnlPct.toFixed(2)}%)`
        : "";

    const dailyLabel = priceLag === "t-2" ? "T-2" : "T-1";

    const baseColumns: MRT_ColumnDef<Position>[] = [
      {
        accessorKey: "Ticker",
        header: "Ticker",
        Footer: () =>
          hasAnyFilter ? null : (
            <Text
              size="sm"
              className={classes["default-xs-font-size"]}
              fw={500}
              c={totals.Pnl > 0 ? "green" : "blue"}
            >
              {"P&L: $" +
                totals.Pnl.toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                }) +
                dailyPctLabel}
            </Text>
          ),
      },
      {
        accessorKey: "Name",
        header: "Name",
        Cell: ({ cell }) => {
          const name = cell.getValue<string>();
          const displayName =
            name.length > 22 ? name.slice(0, 22) + "..." : name;
          return (
            <Tooltip label={name} withArrow>
              <span>{displayName}</span>
            </Tooltip>
          );
        },
      },
      { accessorKey: "Book", header: "Book" },
      {
        accessorKey: "Px",
        header: "Current Px",
        Cell: ({ cell }) => {
          return (
            <span>
              $
              {cell.getValue<number>()?.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 2,
              }) ?? ""}
            </span>
          );
        },
      },
      {
        accessorKey: "Qty",
        header: "Qty",
        Cell: ({ cell }) => {
          return <span>{cell.getValue<number>().toLocaleString()}</span>;
        },
      },
      {
        accessorKey: "Mv",
        header: "Mv (SGD)",
        Cell: ({ cell }) => {
          const fxRate = cell.row.original.FxRate;
          const value = cell.getValue<number>() * fxRate;
          const percentage = totals.Mv ? (value / totals.Mv) * 100 : 0;
          return (
            <span>
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}{" "}
              <Text size="sm" c="dimmed" component="span">
                ({percentage.toFixed(0)}%)
              </Text>
            </span>
          );
        },
        Footer: () => (
          <div>
            <div>
              {"All: $" +
                totals.Mv.toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
            </div>
            <div>
              {"Ex. govies: $" +
                totals.MvLessGovies.toLocaleString(undefined, {
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
            </div>
          </div>
        ),
      },
      {
        accessorKey: "PnL",
        header: "PnL (SGD)",
        Footer: () => (
          <div>
            {"$" +
              totals.Pnl.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
          </div>
        ),
        Cell: ({ cell }) => {
          const fxRate = cell.row.original.FxRate;
          const value = cell.getValue<number>() * fxRate;
          const color = value < 0 ? "red" : "green";

          return (
            <span style={{ color }}>
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 2,
              })}
            </span>
          );
        },
      },
      {
        accessorKey: "Dividends",
        header: "Dividends (SGD)",
        Cell: ({ cell }) => {
          const fxRate = cell.row.original.FxRate;
          return (
            <span>
              $
              {(cell.getValue<number>() * fxRate).toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </span>
          );
        },
        Footer: () => (
          <div>
            {"$" +
              totals.Dividends.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
          </div>
        ),
      },
      {
        accessorKey: "AvgPx",
        header: "AvgPx",
        Cell: ({ cell }) => {
          return (
            <span>
              $
              {cell.getValue<number>()?.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 2,
              }) ?? ""}
            </span>
          );
        },
      },
      { accessorKey: "Ccy", header: "Ccy" },
    ];

    if (!hasCachedMetrics) {
      return baseColumns;
    }

    const dailyColumns: MRT_ColumnDef<Position>[] = [
      {
        accessorKey: "PrevPx",
        header: `Prev Px (${dailyLabel})`,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          if (!value || value <= 1) return <span>-</span>;
          return (
            <span>
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 2,
              })}
            </span>
          );
        },
      },
      {
        accessorKey: "DailyPnl",
        header: `${dailyLabel} P&L (SGD)`,
        Footer: () => {
          const value = displayDailyTotals.dailyPnl || 0;
          const color = value < 0 ? "red" : "green";
          return (
            <span style={{ color }}>
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </span>
          );
        },
        Cell: ({ cell }) => {
          const value = cell.getValue<number>() || 0;
          const color = value < 0 ? "red" : "green";
          return (
            <span style={{ color }}>
              $
              {value.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
            </span>
          );
        },
      },
      {
        accessorKey: "DailyPct",
        header: `${dailyLabel} %`,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>() || 0;
          const color = value < 0 ? "red" : "green";
          return <span style={{ color }}>{value.toFixed(2)}%</span>;
        },
      },
    ];

    return [
      baseColumns[0],
      baseColumns[1],
      baseColumns[2],
      baseColumns[3],
      ...dailyColumns,
      ...baseColumns.slice(4),
    ];
  }, [
    totals,
    refData,
    hasCachedMetrics,
    hasBookFilter,
    hasAnyFilter,
    dailyTotals.prevValue,
    dailyPnlPct,
    displayDailyTotals.dailyPnl,
    priceLag,
  ]);

  const table = useMantineReactTable({
    columns,
    data: positionsWithDaily,
    initialState: {
      showGlobalFilter: true,
      showColumnFilters: true,
      sorting: [{ id: "Mv", desc: true }],
    },
    state: {
      density: "xs",
      isLoading: isLoading,
      showLoadingOverlay: isLoading,
    },
    enableRowSelection: true,
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: ({ table }) => {
      const selectedRows = table.getSelectedRowModel().rows;
      const isOneRowSelected = selectedRows.length === 1;

      return (
        <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
          <Button
            leftSection={<IconHistory size={16} />}
            onClick={handleViewTrades}
            disabled={!isOneRowSelected}
            variant="filled"
            color="blue"
            size="sm"
          >
            View Trade History
          </Button>
          <Button
            leftSection={<IconCoins size={16} />}
            onClick={handleViewDividends}
            disabled={!isOneRowSelected}
            variant="filled"
            color="green"
            size="sm"
          >
            View Dividends
          </Button>
          <SegmentedControl
            value={priceLag}
            onChange={(value) => setPriceLag(value as "t-1" | "t-2")}
            data={[
              { label: "T-1", value: "t-1" },
              { label: "T-2", value: "t-2", disabled: !hasPrev2Prices },
            ]}
            size="xs"
          />
        </div>
      );
    },
  });

  // Update filtered positions when table filters change
  useEffect(() => {
    if (table) {
      const filtered = table
        .getFilteredRowModel()
        .rows.map((row) => row.original);
      setFilteredPositions(filtered);
      const { columnFilters, globalFilter } = table.getState();
      const bookFilterApplied = columnFilters.some(
        (filter) => filter.id === "Book" && filter.value,
      );
      const anyFilterApplied =
        columnFilters.some((filter) => filter.value) || Boolean(globalFilter);
      setHasBookFilter(bookFilterApplied);
      setHasAnyFilter(anyFilterApplied);
    }
  }, [
    table,
    positionsWithDaily,
    table?.getState().columnFilters,
    table?.getState().globalFilter,
  ]);

  useEffect(() => {
    if (!hasCachedMetrics || hasBookFilter || !dailyTotals.prevValue) {
      document.title = defaultTitleRef.current;
      return;
    }

    const titlePrefix = dailyPnlPct >= 0 ? "+" : "";
    const titleLabel = priceLag === "t-2" ? "T-2" : "T-1";
    document.title = `${titlePrefix}${dailyPnlPct.toFixed(2)}% ${titleLabel} P&L`;
  }, [
    hasCachedMetrics,
    hasBookFilter,
    dailyTotals.prevValue,
    dailyPnlPct,
    priceLag,
  ]);

  // Remove the separate loading check since the table handles it now
  if (error) return <div>Error loading positions</div>;

  return (
    <div>
      <MantineReactTable table={table} />
      <Box
        style={{
          display: "flex",
          justifyContent: "center",
          padding: "12px 0",
        }}
      >
        <SegmentedControl
          value={
            (table.getState().columnFilters.find((f) => f.id === "Ccy")
              ?.value as string) || "all"
          }
          onChange={(value) => {
            if (value === "all") {
              table.setColumnFilters(
                table.getState().columnFilters.filter((f) => f.id !== "Ccy"),
              );
            } else {
              const otherFilters = table
                .getState()
                .columnFilters.filter((f) => f.id !== "Ccy");
              table.setColumnFilters([
                ...otherFilters,
                { id: "Ccy", value: value },
              ]);
            }
          }}
          data={[
            { label: "All", value: "all" },
            { label: "SGD", value: "SGD" },
            { label: "USD", value: "USD" },
          ]}
          size="xs"
        />
      </Box>
    </div>
  );
};

export default PositionTable;
