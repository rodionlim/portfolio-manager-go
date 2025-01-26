// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import { Button } from "@mantine/core";
import { MantineReactTable, MRT_ColumnDef } from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";

interface Trade {
  TradeID: number;
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
      { accessorKey: "Ticker", header: "Ticker" },
      { accessorKey: "Account", header: "Account" },
      { accessorKey: "Quantity", header: "Quantity" },
      { accessorKey: "Price", header: "Price" },
    ],
    []
  );

  const handleAddTrade = () => {
    refetch();
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading trades</div>;

  return (
    <div>
      <Button mb="sm" onClick={handleAddTrade}>
        Add Trade
      </Button>
      <MantineReactTable
        columns={columns}
        data={trades}
        initialState={{ showGlobalFilter: true }}
        state={{ density: "xs" }}
      />
    </div>
  );
};

export default BlotterTable;
