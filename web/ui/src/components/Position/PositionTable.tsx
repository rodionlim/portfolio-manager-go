import React, { useMemo, useState, useEffect, useRef } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import {
  Badge,
  Box,
  Button,
  Group,
  Paper,
  SegmentedControl,
  Text,
  Tooltip,
} from "@mantine/core";
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
  InstrumentType?: string;
  UnderlyingTicker?: string;
  UnderlyingGroup?: string;
  Qty: number;
  Mv: number;
  PnL: number;
  Dividends: number;
  AvgPx: number;
  Px: number;
  FxRate: number;
  RawTicker?: string;
  PrevValue?: number;
  PrevPx?: number;
  DailyPnl?: number;
  DailyPct?: number;
}

interface PositionRow extends Position {
  RowKey: string;
  IsGroup?: boolean;
  ChildCount?: number;
  Composition?: string;
  Children?: Position[];
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

const isTBillTicker = (ticker: string) =>
  ticker.length === 8 &&
  /^[A-Za-z]$/.test(ticker[0]) &&
  /^[A-Za-z]$/.test(ticker[1]) &&
  /^[A-Za-z]$/.test(ticker[ticker.length - 1]);

const collapseSpecialTicker = (ticker: string) => {
  if (isTBillTicker(ticker)) {
    return "TBill";
  }

  if (ticker.startsWith("SB") && ticker.length === 7) {
    return "SSB";
  }

  return ticker;
};

const specialTickerName = (ticker: string) => {
  if (ticker === "TBill") {
    return "MAS Bills";
  }

  if (ticker === "SSB") {
    return "SSB";
  }

  return "";
};

const formatInstrumentType = (value?: string) => {
  switch ((value || "").toLowerCase()) {
    case "option":
      return "Option";
    case "future":
      return "Future";
    case "outright":
    default:
      return "Outright";
  }
};

const resolveGroupSummaryChild = (children: Position[]) => {
  if (children.length === 1) {
    return children[0];
  }

  const outrightChild = children.find(
    (child) => (child.InstrumentType || "").toLowerCase() === "outright",
  );

  return outrightChild;
};

const resolveGroupSummaryMetrics = (
  groupTicker: string,
  children: Position[],
) => {
  if (groupTicker === "SSB") {
    const totalQty = children.reduce((sum, child) => sum + child.Qty, 0);
    const weightedAvgPx =
      totalQty > 0
        ? children.reduce((sum, child) => sum + child.AvgPx * child.Qty, 0) /
          totalQty
        : 0;

    return {
      Qty: totalQty,
      Px: 100,
      AvgPx: weightedAvgPx,
      PrevPx: 100,
    };
  }

  const summaryChild = resolveGroupSummaryChild(children);

  return {
    Qty: summaryChild?.Qty ?? 0,
    Px: summaryChild?.Px ?? 0,
    AvgPx: summaryChild?.AvgPx ?? 0,
    PrevPx: summaryChild?.PrevPx,
  };
};

const sameFilteredRows = (left: PositionRow[], right: PositionRow[]) => {
  if (left.length !== right.length) {
    return false;
  }

  return left.every((row, index) => {
    const other = right[index];
    return (
      row.RowKey === other.RowKey &&
      row.Qty === other.Qty &&
      row.Px === other.Px &&
      row.PrevPx === other.PrevPx &&
      row.Mv === other.Mv &&
      row.PnL === other.PnL &&
      row.Dividends === other.Dividends &&
      row.AvgPx === other.AvgPx
    );
  });
};

const PositionTable: React.FC = () => {
  const navigate = useNavigate();
  const refData = useSelector((state: RootState) => state.referenceData.data);
  const [filteredPositions, setFilteredPositions] = useState<PositionRow[]>([]);
  const [columnFilters, setColumnFilters] = useState<
    Array<{ id: string; value: unknown }>
  >([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [priceLag, setPriceLag] = useState<"t-1" | "t-2">("t-1");
  const defaultTitleRef = useRef(document.title);

  const {
    data: rawPositions = [],
    isLoading,
    error,
  } = useQuery<Position[], Error>({
    queryKey: ["positions"],
    queryFn: async () => {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions"));
      if (!resp.ok) {
        const errorPayload = await resp
          .json()
          .catch(() => ({ message: "Failed to get positions" }));
        throw new Error(errorPayload?.message || "Failed to get positions");
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

  const groupedTargetTicker = (position: PositionRow) =>
    position.UnderlyingGroup || position.UnderlyingTicker || position.Ticker;

  const handleViewTrades = () => {
    const selectedRows = table.getSelectedRowModel().rows;
    if (selectedRows.length === 1) {
      const ticker = groupedTargetTicker(selectedRows[0].original);
      navigate(`/blotter?ticker=${encodeURIComponent(ticker)}`);
    }
  };

  const handleViewDividends = () => {
    const selectedRows = table.getSelectedRowModel().rows;
    if (selectedRows.length === 1) {
      const ticker = groupedTargetTicker(selectedRows[0].original);
      navigate(`/dividends`, {
        state: { ticker: ticker },
      });
    }
  };

  const leafPositions = useMemo(() => {
    return rawPositions.map((position) => {
      const rawTicker = position.Ticker;
      const lookupTicker = rawTicker;
      const cachedPrice = activeCachedPriceMap.get(lookupTicker);
      const prevPx = cachedPrice?.price;
      const currentPx = position.Px;
      const fxRate = position.FxRate || 1;
      const underlyingGroupRaw =
        position.UnderlyingGroup || position.UnderlyingTicker || rawTicker;
      const underlyingGroup = collapseSpecialTicker(underlyingGroupRaw);
      const tickerDisplay = collapseSpecialTicker(rawTicker);
      const name =
        specialTickerName(tickerDisplay) ||
        refData?.[rawTicker]?.name ||
        rawTicker;

      let dailyPnl = 0;
      let dailyPct = 0;
      let prevValue = 0;

      if (
        prevPx !== undefined &&
        currentPx !== undefined &&
          prevPx > 0 &&
          currentPx > 0 &&
        position.Qty !== 0
      ) {
        const priceChange = currentPx - prevPx;
        dailyPnl = priceChange * position.Qty * fxRate;
        dailyPct = prevPx ? (priceChange / prevPx) * 100 : 0;
        prevValue = prevPx * position.Qty * fxRate;
      }

      return {
        ...position,
        RawTicker: rawTicker,
        Name: name,
        UnderlyingGroup: underlyingGroup,
        PrevPx: prevPx,
        PrevValue: prevValue,
        DailyPnl: dailyPnl,
        DailyPct: dailyPct,
      };
    });
  }, [activeCachedPriceMap, rawPositions, refData]);

  const groupedPositions = useMemo(() => {
    const grouped = new Map<string, PositionRow>();

    leafPositions.forEach((position) => {
      const groupTicker =
        position.UnderlyingGroup || collapseSpecialTicker(position.Ticker);
      const groupName =
        specialTickerName(groupTicker) ||
        refData?.[groupTicker]?.name ||
        refData?.[position.UnderlyingTicker || ""]?.name ||
        position.Name;
      const groupKey = `${groupTicker}-${position.Book}`;
      const existing = grouped.get(groupKey);

      if (!existing) {
        grouped.set(groupKey, {
          RowKey: groupKey,
          IsGroup: true,
          Ticker: groupTicker,
          Name: groupName,
          Book: position.Book,
          Ccy: position.Ccy,
          AssetClass: position.AssetClass,
          AssetSubClass: position.AssetSubClass,
          InstrumentType: "group",
          UnderlyingTicker: position.UnderlyingTicker || groupTicker,
          UnderlyingGroup: groupTicker,
          Qty: 0,
          Mv: position.Mv * position.FxRate,
          PnL: position.PnL * position.FxRate,
          Dividends: position.Dividends * position.FxRate,
          AvgPx: 0,
          Px: 0,
          FxRate: 1,
          PrevPx: undefined,
          PrevValue: position.PrevValue || 0,
          DailyPnl: position.DailyPnl || 0,
          DailyPct: 0,
          ChildCount: 1,
          Composition: formatInstrumentType(position.InstrumentType),
          Children: [position],
        });
        return;
      }

      existing.Mv += position.Mv * position.FxRate;
      existing.PnL += position.PnL * position.FxRate;
      existing.Dividends += position.Dividends * position.FxRate;
      existing.PrevValue =
        (existing.PrevValue || 0) + (position.PrevValue || 0);
      existing.DailyPnl = (existing.DailyPnl || 0) + (position.DailyPnl || 0);
      existing.ChildCount = (existing.ChildCount || 0) + 1;
      existing.Children = [...(existing.Children || []), position];

      const composition = new Set(
        [...(existing.Children || [])].map((child) =>
          formatInstrumentType(child.InstrumentType),
        ),
      );
      existing.Composition = Array.from(composition).join(", ");
      if (existing.Ccy !== position.Ccy) {
        existing.Ccy = "MULTI";
      }
    });

    return Array.from(grouped.values())
      .map((group) => {
        const sortedChildren = [...(group.Children || [])].sort((left, right) => {
          const instrumentRank = (value?: string) => {
            switch ((value || "").toLowerCase()) {
              case "outright":
                return 0;
              case "future":
                return 1;
              case "option":
                return 2;
              default:
                return 3;
            }
          };

          return (
            instrumentRank(left.InstrumentType) -
              instrumentRank(right.InstrumentType) ||
            left.Ticker.localeCompare(right.Ticker)
          );
        });
        const summaryMetrics = resolveGroupSummaryMetrics(
          group.Ticker,
          sortedChildren,
        );

        return {
          ...group,
          Qty: summaryMetrics.Qty,
          Px: summaryMetrics.Px,
          AvgPx: summaryMetrics.AvgPx,
          PrevPx: summaryMetrics.PrevPx,
          DailyPct:
            group.PrevValue && group.PrevValue > 0
              ? ((group.DailyPnl || 0) / group.PrevValue) * 100
              : 0,
          Children: sortedChildren,
        };
      })
      .sort((left, right) => right.Mv - left.Mv);
  }, [leafPositions, refData]);

  const dailyTotals = useMemo(() => {
    if (!hasCachedMetrics) {
      return { dailyPnl: 0, prevValue: 0, pct: 0 };
    }

    return groupedPositions.reduce(
      (acc, row) => {
        acc.prevValue += row.PrevValue || 0;
        acc.dailyPnl += row.DailyPnl || 0;

        return acc;
      },
      { dailyPnl: 0, prevValue: 0, pct: 0 },
    );
  }, [groupedPositions, hasCachedMetrics]);

  const dailyPnlPct = useMemo(() => {
    if (!dailyTotals.prevValue) return 0;
    return (dailyTotals.dailyPnl / dailyTotals.prevValue) * 100;
  }, [dailyTotals]);

  const hasBookFilter = useMemo(
    () => columnFilters.some((filter) => filter.id === "Book" && filter.value),
    [columnFilters],
  );

  const hasAnyFilter = useMemo(
    () => columnFilters.some((filter) => filter.value) || Boolean(globalFilter),
    [columnFilters, globalFilter],
  );

  const displayDailyTotals = useMemo(() => {
    const positions = hasAnyFilter ? filteredPositions : groupedPositions;
    if (!positions.length) {
      return { dailyPnl: 0, prevValue: 0 };
    }

    return positions.reduce(
      (acc, row) => {
        acc.prevValue += row.PrevValue || 0;
        acc.dailyPnl += row.DailyPnl || 0;

        return acc;
      },
      { dailyPnl: 0, prevValue: 0 },
    );
  }, [filteredPositions, groupedPositions, hasAnyFilter]);

  const totals = useMemo(() => {
    const positions = hasAnyFilter ? filteredPositions : groupedPositions;
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
  }, [filteredPositions, groupedPositions, hasAnyFilter]);

  const columns = useMemo<MRT_ColumnDef<PositionRow>[]>(() => {
    const dailyPctLabel =
      hasCachedMetrics && !hasAnyFilter && dailyTotals.prevValue
        ? ` (${dailyPnlPct.toFixed(2)}%)`
        : "";

    const dailyLabel = priceLag === "t-2" ? "T-2" : "T-1";

    const baseColumns: MRT_ColumnDef<PositionRow>[] = [
      {
        accessorKey: "Ticker",
        header: "Ticker",
        Cell: ({ row, cell }) => {
          const value = cell.getValue<string>();
          if (!row.original.IsGroup) {
            return value;
          }

          return <Text size="sm">{value}</Text>;
        },
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
        Cell: ({ row, cell }) => {
          const name = cell.getValue<string>();
          const displayName =
            name.length > 22 ? name.slice(0, 22) + "..." : name;

          if (row.original.IsGroup) {
            return (
              <Tooltip label={name} withArrow>
                <span>{displayName}</span>
              </Tooltip>
            );
          }

          const suffix = formatInstrumentType(row.original.InstrumentType);
          return (
            <Group gap="xs">
              <Tooltip label={name} withArrow>
                <span>{displayName}</span>
              </Tooltip>
              {suffix ? (
                <Badge
                  variant="outline"
                  color={row.original.IsGroup ? "blue" : "gray"}
                >
                  {suffix}
                </Badge>
              ) : null}
            </Group>
          );
        },
      },
      { accessorKey: "Book", header: "Book" },
      {
        accessorKey: "Px",
        header: "Current Px",
        Cell: ({ row, cell }) => {
          const value = cell.getValue<number>();
          if (row.original.IsGroup && (!value || value <= 0)) {
            return <span>-</span>;
          }
          return (
            <span>
              $
              {value?.toLocaleString(undefined, {
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
        Cell: ({ row, cell }) => {
          const value = cell.getValue<number>();
          if (row.original.IsGroup && (!value || value === 0)) {
            return <span>-</span>;
          }
          return <span>{value.toLocaleString()}</span>;
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
        Cell: ({ row, cell }) => {
          const value = cell.getValue<number>();
          if (row.original.IsGroup && (!value || value <= 0)) {
            return <span>-</span>;
          }
          return (
            <span>
              $
              {value?.toLocaleString(undefined, {
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

    const dailyColumns: MRT_ColumnDef<PositionRow>[] = [
      {
        accessorKey: "PrevPx",
        header: `Prev Px (${dailyLabel})`,
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          if (!value || value <= 0) return <span>-</span>;
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

  const columnOrder = useMemo(
    () =>
      hasCachedMetrics
        ? [
            "mrt-row-select",
            "mrt-row-expand",
            "Ticker",
            "Name",
            "Book",
            "Px",
            "PrevPx",
            "DailyPnl",
            "DailyPct",
            "Qty",
            "Mv",
            "PnL",
            "Dividends",
            "AvgPx",
            "Ccy",
          ]
        : [
            "mrt-row-select",
            "mrt-row-expand",
            "Ticker",
            "Name",
            "Book",
            "Px",
            "Qty",
            "Mv",
            "PnL",
            "Dividends",
            "AvgPx",
            "Ccy",
          ],
    [hasCachedMetrics],
  );

  const table = useMantineReactTable({
    columns,
    data: groupedPositions,
    initialState: {
      columnOrder,
      showGlobalFilter: true,
      showColumnFilters: true,
      sorting: [{ id: "Mv", desc: true }],
    },
    displayColumnDefOptions: {
      "mrt-row-expand": {
        size: 28,
        mantineTableHeadCellProps: {
          style: {
            paddingTop: 3,
            paddingBottom: 3,
            paddingLeft: 4,
            paddingRight: 2,
          },
        },
        mantineTableBodyCellProps: {
          style: {
            paddingTop: 3,
            paddingBottom: 3,
            paddingLeft: 4,
            paddingRight: 2,
          },
        },
      },
      "mrt-row-select": {
        size: 32,
        mantineTableHeadCellProps: {
          style: {
            paddingTop: 3,
            paddingBottom: 3,
            paddingLeft: 2,
            paddingRight: 2,
          },
        },
        mantineTableBodyCellProps: {
          style: {
            paddingTop: 3,
            paddingBottom: 3,
            paddingLeft: 2,
            paddingRight: 2,
          },
        },
      },
    },
    state: {
      columnFilters,
      globalFilter,
      density: "xs",
      isLoading: isLoading,
      showLoadingOverlay: isLoading,
    },
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    mantineTableHeadCellProps: {
      style: {
        fontSize: "var(--mantine-font-size-sm)",
        paddingTop: 6,
        paddingBottom: 6,
        paddingLeft: 6,
        paddingRight: 6,
        lineHeight: 1.1,
      },
    },
    mantineTableBodyCellProps: {
      style: {
        fontSize: "var(--mantine-font-size-sm)",
        paddingTop: 6,
        paddingBottom: 6,
        paddingLeft: 6,
        paddingRight: 6,
        lineHeight: 1.1,
      },
    },
    enableRowSelection: true,
    positionToolbarAlertBanner: "bottom",
    renderDetailPanel: ({ row }) => {
      const children = row.original.Children || [];
      if (!children.length) {
        return null;
      }

      return (
        <Paper withBorder p={3} radius="sm">
          <Group justify="space-between" mb={3}>
            <Text fw={600} size="xs">Underlying Components</Text>
            <Badge variant="light" size="xs">{children.length} instruments</Badge>
          </Group>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: "var(--mantine-font-size-sm)",
              lineHeight: 1.2,
            }}
          >
            <thead>
              <tr>
                <th style={{ textAlign: "left", paddingBottom: 3 }}>Type</th>
                <th style={{ textAlign: "left", paddingBottom: 3 }}>Ticker</th>
                <th style={{ textAlign: "left", paddingBottom: 3 }}>Name</th>
                <th style={{ textAlign: "right", paddingBottom: 3 }}>Qty</th>
                <th style={{ textAlign: "right", paddingBottom: 4 }}>
                  MV (SGD)
                </th>
                <th style={{ textAlign: "right", paddingBottom: 4 }}>
                  PnL (SGD)
                </th>
              </tr>
            </thead>
            <tbody>
              {children.map((child) => {
                const marketValue = child.Mv * child.FxRate;
                const pnlValue = child.PnL * child.FxRate;
                return (
                  <tr key={`${child.Book}-${child.Ticker}`}>
                    <td style={{ padding: "3px 0" }}>
                      <Badge variant="outline" size="xs">
                        {formatInstrumentType(child.InstrumentType)}
                      </Badge>
                    </td>
                    <td style={{ padding: "3px 0" }}>{child.Ticker}</td>
                    <td style={{ padding: "3px 0" }}>{child.Name}</td>
                    <td style={{ padding: "3px 0", textAlign: "right" }}>
                      {child.Qty.toLocaleString()}
                    </td>
                    <td style={{ padding: "3px 0", textAlign: "right" }}>
                      {marketValue.toLocaleString(undefined, {
                        minimumFractionDigits: 0,
                        maximumFractionDigits: 0,
                      })}
                    </td>
                    <td
                      style={{
                        padding: "3px 0",
                        textAlign: "right",
                        color:
                          pnlValue < 0
                            ? "var(--mantine-color-red-6)"
                            : "var(--mantine-color-green-6)",
                      }}
                    >
                      {pnlValue.toLocaleString(undefined, {
                        minimumFractionDigits: 0,
                        maximumFractionDigits: 0,
                      })}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </Paper>
      );
    },
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

  useEffect(() => {
    const nextFilteredPositions = table
      .getFilteredRowModel()
      .rows.map((row) => row.original);

    setFilteredPositions((current) =>
      sameFilteredRows(current, nextFilteredPositions)
        ? current
        : nextFilteredPositions,
    );
  }, [table, groupedPositions, columnFilters, globalFilter]);

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
  if (error) {
    const missingRefDataTicker = error.message.match(/REFDATA:([^:]+)/)?.[1];
    const message = missingRefDataTicker
      ? `Missing reference data for ${missingRefDataTicker}. Update refdata and retry.`
      : error.message;

    return (
      <Box py="md">
        <Text c="red" fw={500}>Error loading positions</Text>
        <Text c="dimmed" size="sm">{message}</Text>
      </Box>
    );
  }

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
