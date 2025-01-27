// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import { Box, Button } from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";

interface Trade {
  TradeID: number;
  TradeDate: string;
  Ticker: string;
  Account: string;
  // assetClass: string; // to add back reference data
  Quantity: number;
  Price: number;
}

const fetchTrades = async (): Promise<Trade[]> => {
  return fetch("http://localhost:8080/api/v1/blotter/trade")
    .then((resp) => resp.json())
    .then(
      (data) => {
        return data;
      },
      (error) => {
        console.error("error", error);
        throw new Error("An error occurred while fetching trades");
      }
    );
};

const BlotterTable: React.FC = () => {
  const {
    data: trades = [],
    isLoading,
    error,
    refetch,
  } = useQuery({ queryKey: ["trades"], queryFn: fetchTrades });

  const columns = useMemo<MRT_ColumnDef<Trade>[]>(
    () => [
      { accessorKey: "TradeID", header: "Trade ID" },
      { accessorKey: "TradeDate", header: "Date" },
      { accessorKey: "Ticker", header: "Ticker" },
      { accessorKey: "Account", header: "Account" },
      { accessorKey: "Quantity", header: "Quantity" },
      { accessorKey: "Price", header: "Price" },
    ],
    []
  );

  const table = useMantineReactTable({
    columns,
    data: trades,
    initialState: { showGlobalFilter: true },
    state: { density: "xs" },
    enableRowSelection: true,
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: ({ table }) => (
      <Box style={{ display: "flex", gap: "16px", padding: "4px" }}>
        <Button color="teal" onClick={handleAddTrade} variant="filled">
          Add Trade
        </Button>
        <Button
          color="red"
          disabled={!table.getIsSomeRowsSelected()}
          onClick={handleDeleteTrades}
          variant="filled"
        >
          Delete Selected Trades
        </Button>
      </Box>
    ),
  });

  const handleAddTrade = () => {
    alert("Add Trade");
    refetch();
  };

  const handleDeleteTrades = () => {
    alert("Delete Selected Trades");
    refetch();
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading trades</div>;

  return (
    <div>
      <MantineReactTable table={table} />
    </div>
  );
};

export default BlotterTable;
