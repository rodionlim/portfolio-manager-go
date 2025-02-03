// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
// import { useSelector } from "react-redux";
// import { RootState } from "../../store";

interface Position {
  Ticker: string;
  //   Name: string;
  Trader: string;
  Ccy: string;
  AssetClass: string;
  AssetSubClass: number;
  Qty: number;
  Mv: number;
  PnL: number;
  Dividends: number;
  AvgPx: number;
}

const PositionTable: React.FC = () => {
  //   const refData = useSelector((state: RootState) => state.referenceData.data);

  const fetchPosition = async (): Promise<Position[]> => {
    return fetch("http://localhost:8080/api/v1/portfolio/positions")
      .then((resp) => resp.json())
      .then(
        (data: Position[]) => {
          // collapse SSB and Mas Bills in data to a single position
          // TODO: allow fully uncollapsed positions
          const aggregatedPositions = Object.values(
            data.reduce((acc: Record<string, Position>, curr: Position) => {
              let tickerKey = curr.Ticker;
              let tickerName: string;

              // If it's a mas tbill (8 characters, first two and last are letters), set key to "TBill".
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
                // If it's not a mas tbill or SSB, set key to the original ticker.
                // tickerName = refData?.[tickerKey]?.name ?? "";
              }

              // If the key already exists, sum the values.
              if (acc[tickerKey]) {
                acc[tickerKey].Qty += curr.Qty;
                acc[tickerKey].Mv += curr.Mv;
                acc[tickerKey].PnL += curr.PnL;
                acc[tickerKey].Dividends += curr.Dividends;
                // acc[tickerKey].Name = tickerName;
              } else {
                // Create a new entry with the updated tickerKey.
                acc[tickerKey] = {
                  ...curr,
                  Ticker: tickerKey,
                  //   Name: tickerName,
                };
              }
              return acc;
            }, {} as Record<string, Position>)
          );
          return aggregatedPositions;
        },
        (error) => {
          console.error("error", error);
          throw new Error(
            `An error occurred while fetching positions ${error.message}`
          );
        }
      );
  };

  const {
    data: positions = [],
    isLoading,
    error,
  } = useQuery({ queryKey: ["positions"], queryFn: fetchPosition });

  const totals = useMemo(() => {
    return positions.reduce(
      (acc, row) => {
        acc.Mv += row.Mv;
        acc.Pnl += row.PnL;
        acc.Dividends += row.Dividends;
        return acc;
      },
      { Mv: 0, Pnl: 0, Dividends: 0 }
    );
  }, [positions]);

  const columns = useMemo<MRT_ColumnDef<Position>[]>(
    () => [
      { accessorKey: "Ticker", header: "Ticker" },
      //   { accessorKey: "Name", header: "Name" },
      { accessorKey: "Ccy", header: "Ccy" },
      { accessorKey: "Qty", header: "Qty" },
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
            {"$" +
              totals.Mv.toLocaleString(undefined, {
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              })}
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
    []
  );

  const table = useMantineReactTable({
    columns,
    data: positions,
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
