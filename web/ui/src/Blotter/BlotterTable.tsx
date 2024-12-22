// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo, useState } from "react";
import { Button } from "@mantine/core";
import { MantineReactTable, MRT_ColumnDef } from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";

interface Trade {
  id: number;
  ticker: string;
  assetClass: string;
  quantity: number;
  price: number;
}

const fetchTrades = async (): Promise<Trade[]> => {
  // Hardcoded trades data for now
  return [
    { id: 1, ticker: "AAPL", assetClass: "Equity", quantity: 100, price: 150 },
    { id: 2, ticker: "GOOGL", assetClass: "Equity", quantity: 50, price: 2800 },
    { id: 3, ticker: "TSLA", assetClass: "Equity", quantity: 30, price: 700 },
    { id: 4, ticker: "BTC", assetClass: "Crypto", quantity: 2, price: 45000 },
  ];
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
      { accessorKey: "id", header: "Trade ID" },
      { accessorKey: "ticker", header: "Ticker" },
      { accessorKey: "assetClass", header: "Asset Class" },
      { accessorKey: "quantity", header: "Quantity" },
      { accessorKey: "price", header: "Price" },
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
      <Button onClick={handleAddTrade}>Add Trade</Button>
      <MantineReactTable columns={columns} data={trades} />
    </div>
  );
};

export default BlotterTable;
