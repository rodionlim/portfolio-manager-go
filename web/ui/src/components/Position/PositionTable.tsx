// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";

interface Position {
  Ticker: string;
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

const fetchPosition = async (): Promise<Position[]> => {
  return fetch("http://localhost:8080/api/v1/portfolio/positions")
    .then((resp) => resp.json())
    .then(
      (data) => {
        return data;
      },
      (error) => {
        console.error("error", error);
        throw new Error(
          `An error occurred while fetching positions ${error.message}`
        );
      }
    );
};

const PositionTable: React.FC = () => {
  const {
    data: positions = [],
    isLoading,
    error,
  } = useQuery({ queryKey: ["positions"], queryFn: fetchPosition });

  const columns = useMemo<MRT_ColumnDef<Position>[]>(
    () => [
      { accessorKey: "Ticker", header: "Ticker" },
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
      },
      {
        accessorKey: "PnL",
        header: "PnL",
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
    initialState: { showGlobalFilter: true, showColumnFilters: true },
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
