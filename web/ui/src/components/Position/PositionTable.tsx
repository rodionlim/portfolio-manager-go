import React, { useMemo, useState, useEffect } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import { Button, Text, Tooltip } from "@mantine/core";
import { getUrl } from "../../utils/url";
import { useNavigate } from "react-router-dom";
import { IconHistory, IconCoins } from "@tabler/icons-react";

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
}

const PositionTable: React.FC = () => {
  const navigate = useNavigate();
  const refData = useSelector((state: RootState) => state.referenceData.data);
  const [filteredPositions, setFilteredPositions] = useState<Position[]>([]);

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
      rawPositions.reduce((acc: Record<string, Position>, curr: Position) => {
        let tickerKey = curr.Ticker;
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

        if (acc[tickerKey]) {
          acc[tickerKey].Qty += curr.Qty;
          acc[tickerKey].Mv += curr.Mv * curr.FxRate;
          acc[tickerKey].PnL += curr.PnL;
          acc[tickerKey].Dividends += curr.Dividends;
          acc[tickerKey].Name = tickerName;
        } else {
          acc[tickerKey] = { ...curr, Ticker: tickerKey, Name: tickerName };
        }

        return acc;
      }, {} as Record<string, Position>)
    );
  }, [rawPositions, refData]);

  // Calculate totals based on filtered positions
  const totals = useMemo(() => {
    const positions =
      filteredPositions.length > 0 ? filteredPositions : aggregatedPositions;
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
      { Mv: 0, MvLessGovies: 0, Pnl: 0, Dividends: 0 }
    );
    return res;
  }, [filteredPositions, aggregatedPositions]);

  const columns = useMemo<MRT_ColumnDef<Position>[]>(
    () => [
      {
        accessorKey: "Ticker",
        header: "Ticker",
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
    ],
    [totals, refData]
  );

  const table = useMantineReactTable({
    columns,
    data: aggregatedPositions,
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
        <div style={{ display: "flex", gap: "8px" }}>
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
        </div>
      );
    },
  });

  // Update filtered positions when table filters change
  useEffect(() => {
    const filtered = table
      .getFilteredRowModel()
      .rows.map((row) => row.original);
    setFilteredPositions(filtered);
  }, [table.getFilteredRowModel().rows]);

  // Remove the separate loading check since the table handles it now
  if (error) return <div>Error loading positions</div>;

  return (
    <div>
      <MantineReactTable table={table} />
    </div>
  );
};

export default PositionTable;
