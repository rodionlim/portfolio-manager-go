// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import { Box, Button } from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  MRT_TableInstance,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { useNavigate } from "react-router-dom";

interface Trade {
  TradeID: string;
  TradeDate: string;
  Ticker: string;
  Account: string;
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
        throw new Error(
          `An error occurred while fetching trades ${error.message}`
        );
      }
    );
};

const deleteTrades = async (trades: string[]): Promise<{ message: string }> => {
  return fetch("http://localhost:8080/api/v1/blotter/trade", {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(trades),
  })
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
  const navigate = useNavigate();

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
        <Button color="teal" onClick={handleAddTrade(table)} variant="filled">
          Add Trade
        </Button>
        <Button
          color="red"
          disabled={!table.getIsSomeRowsSelected()}
          onClick={handleDeleteTrades(table)}
          variant="filled"
        >
          Delete Selected Trades
        </Button>
      </Box>
    ),
  });

  // handle add trade allows routing to the add trade page
  const handleAddTrade = (table: MRT_TableInstance<Trade>): (() => void) => {
    return () => {
      // first check if there is any selections
      const selection = table
        .getSelectedRowModel()
        .rows.map((trade) => trade.original.Ticker);
      if (selection.length > 0) {
        const ticker = selection[0];
        navigate("/blotter/add_trade", { state: { ticker } });
      } else {
        navigate("/blotter/add_trade");
      }
    };
  };

  const handleDeleteTrades = (
    table: MRT_TableInstance<Trade>
  ): (() => void) => {
    return () => {
      const deletionTrades = table
        .getSelectedRowModel()
        .rows.map((trade) => trade.original.TradeID);

      deleteTrades(deletionTrades)
        .then(
          (resp: { message: string }) => {
            notifications.show({
              title: "Trades successfully deleted",
              message: `${resp.message}`,
              autoClose: 10000,
            });
          },
          (error) => {
            notifications.show({
              color: "red",
              title: "Error",
              message: `Unable to delete trades from the blotter\n ${error}`,
            });
          }
        )
        .finally(() => {
          refetch();
        });
    };
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
