// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import { Tooltip } from "@mantine/core";

interface Position {
  Ticker: string;
  Name: string;
  Trader: string;
  Ccy: string;
  AssetClass: string;
  AssetSubClass: string;
  Qty: number;
  Mv: number;
  PnL: number;
  Dividends: number;
  AvgPx: number;
}

const PositionTable: React.FC = () => {
  const refData = useSelector((state: RootState) => state.referenceData.data);

  const {
    data: rawPositions = [],
    isLoading,
    error,
  } = useQuery<Position[]>({
    queryKey: ["positions"],
    queryFn: async () => {
      const resp = await fetch(
        "http://localhost:8080/api/v1/portfolio/positions"
      );
      return resp.json();
    },
  });

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
          acc[tickerKey].Mv += curr.Mv;
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

  const totals = useMemo(() => {
    const res = rawPositions.reduce(
      (acc, row) => {
        acc.Mv += row.Mv;
        acc.Pnl += row.PnL;
        acc.Dividends += row.Dividends;

        if (row.AssetSubClass !== "govies") {
          acc.MvLessGovies += row.Mv;
        }

        return acc;
      },
      { Mv: 0, MvLessGovies: 0, Pnl: 0, Dividends: 0 }
    );
    return res;
  }, [rawPositions]);

  const columns = useMemo<MRT_ColumnDef<Position>[]>(
    () => [
      { accessorKey: "Ticker", header: "Ticker" },
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
      { accessorKey: "Ccy", header: "Ccy" },
      {
        accessorKey: "Qty",
        header: "Qty",
        Cell: ({ cell }) => {
          return <span>{cell.getValue<number>().toLocaleString()}</span>;
        },
      },
      {
        accessorKey: "Mv",
        header: "Mv",
        Cell: ({ cell }) => {
          return (
            <span>
              $
              {cell.getValue<number>().toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
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
        header: "PnL",
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
          const value = cell.getValue<number>();
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
        header: "Dividends",
        Cell: ({ cell }) => {
          return (
            <span>
              $
              {cell.getValue<number>().toLocaleString(undefined, {
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
              {cell.getValue<number>().toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 2,
              })}
            </span>
          );
        },
      },
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
    state: { density: "xs" },
    enableRowSelection: true,
    positionToolbarAlertBanner: "bottom",
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading positions</div>;

  return (
    <div>
      <MantineReactTable table={table} />
    </div>
  );
};

export default PositionTable;
