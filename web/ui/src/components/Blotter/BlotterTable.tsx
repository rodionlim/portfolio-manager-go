import React, { useMemo, useState } from "react";
import { Box, Button, FileInput } from "@mantine/core";
import { IconUpload } from "@tabler/icons-react";
import {
  MantineReactTable,
  MRT_ColumnDef,
  MRT_TableInstance,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { useLocation, useNavigate } from "react-router-dom";
import { getUrl } from "../../utils/url";
import { Trade } from "../../types/blotter";
import BlotterBulkUpdateModal from "./BlotterBulkUpdateModal";

const fetchTrades = async (): Promise<Trade[]> => {
  return fetch(getUrl("/api/v1/blotter/trade"))
    .then((resp) => resp.json())
    .then(
      (data: Trade[]) => {
        return data.sort(
          (x, y) =>
            new Date(y.TradeDate).getTime() - new Date(x.TradeDate).getTime()
        );
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
  return fetch(getUrl("api/v1/blotter/trade"), {
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
        throw new Error("An error occurred while deleting trades");
      }
    );
};

const uploadTradesCSV = async (file: File): Promise<{ message: string }> => {
  const formData = new FormData();
  formData.append("file", file);

  return fetch(getUrl("api/v1/blotter/upload"), {
    method: "POST",
    body: formData,
  })
    .then((resp) => resp.json())
    .then(
      (data) => {
        return data;
      },
      (error) => {
        console.error("error", error);
        throw new Error("An error occurred while uploading trades");
      }
    );
};

const BlotterTable: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [bulkUpdateModalOpened, setBulkUpdateModalOpened] = useState(false);
  const [selectedTrades, setSelectedTrades] = useState<Trade[]>([]);

  // Extract ticker filter from location state or search params
  const searchParams = new URLSearchParams(location.search);
  const filterTicker =
    searchParams.get("ticker") || (location.state as any)?.ticker || "";

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
      { accessorKey: "Side", header: "Side" },
      { accessorKey: "Book", header: "Book" },
      // { accessorKey: "Broker", header: "Broker" },
      { accessorKey: "Account", header: "Account" },
      { accessorKey: "Quantity", header: "Quantity" },
      { accessorKey: "Price", header: "Price" },
      {
        accessorKey: "Fx",
        header: "Fx",
        Cell: ({ cell }) => Number(cell.getValue()).toFixed(4),
      },
      // { accessorKey: "TradeType", header: "Trade Type" },
      // { accessorKey: "SeqNum", header: "Seq Num" },
    ],
    []
  );

  const table = useMantineReactTable({
    columns,
    data: trades,
    initialState: {
      showGlobalFilter: true,
      showColumnFilters: true,
      columnFilters: filterTicker
        ? [{ id: "Ticker", value: filterTicker }]
        : [],
    },
    state: { density: "xs" },
    enableRowSelection: true,
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: ({ table }) => (
      <Box
        style={{
          display: "flex",
          gap: "16px",
          padding: "4px",
          alignItems: "center",
        }}
      >
        <Button color="teal" onClick={handleAddTrade(table)} variant="filled">
          Add Trade
        </Button>
        <Button
          color="red"
          disabled={table.getSelectedRowModel().rows.length === 0}
          onClick={handleDeleteTrades(table)}
          variant="filled"
        >
          Delete Selected Trades
        </Button>
        <Button
          color="blue"
          disabled={table.getSelectedRowModel().rows.length === 0}
          onClick={handleUpdateTrade(table)}
          variant="filled"
        >
          {table.getSelectedRowModel().rows.length === 1
            ? "Update Trade"
            : "Bulk Update"}
        </Button>
        <Button color="gray" variant="outline" onClick={handleExportCSV}>
          Export CSV
        </Button>
        <FileInput
          placeholder="Upload CSV"
          accept=".csv"
          onChange={handleFileUpload}
          leftSection={<IconUpload size={16} />}
          style={{ width: "200px" }}
          clearable
        />
      </Box>
    ),
  });

  // handle add trade allows routing to the add trade page
  const handleAddTrade = (table: MRT_TableInstance<Trade>): (() => void) => {
    return () => {
      // first check if there is any selections
      const selection = table.getSelectedRowModel().rows;
      if (selection.length > 0) {
        const ticker = selection[0];
        navigate("/blotter/add_trade", {
          state: {
            ticker: ticker.original.Ticker,
            broker: ticker.original.Broker,
            account: ticker.original.Account,
            book: ticker.original.Book,
          },
        });
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

  // handle update trade - either single trade update (navigation) or bulk update (modal)
  const handleUpdateTrade = (table: MRT_TableInstance<Trade>): (() => void) => {
    return () => {
      const selectedRows = table.getSelectedRowModel().rows;

      if (selectedRows.length === 1) {
        // Single trade update - navigate to form
        const selection = selectedRows[0].original;
        navigate("/blotter/update_trade", {
          state: {
            tradeId: selection.TradeID,
            date: new Date(selection.TradeDate),
            ticker: selection.Ticker,
            book: selection.Book,
            broker: selection.Broker,
            account: selection.Account,
            qty: selection.Quantity,
            price: selection.Price,
            fx: selection.Fx,
            tradeType: selection.TradeType,
            seqNum: selection.SeqNum,
          },
        });
      } else if (selectedRows.length > 1) {
        // Bulk update - open modal
        const trades = selectedRows.map((row) => row.original);
        setSelectedTrades(trades);
        setBulkUpdateModalOpened(true);
      }
    };
  };

  // Handle file upload
  const handleFileUpload = (file: File | null) => {
    if (!file) return;

    uploadTradesCSV(file)
      .then(
        (resp: { message: string }) => {
          notifications.show({
            title: "Trades successfully uploaded",
            message: `${resp.message}`,
            autoClose: 10000,
          });
        },
        (error) => {
          notifications.show({
            color: "red",
            title: "Error",
            message: `Unable to upload trades to the blotter\n ${error}`,
          });
        }
      )
      .finally(() => {
        refetch();
      });
  };

  // Add the export handler function
  const handleExportCSV = () => {
    const url = getUrl("/api/v1/blotter/export");

    // Create a hidden link and click it to trigger download
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", "trades.csv"); // filename is determined by api server
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading trades</div>;

  return (
    <div>
      <MantineReactTable table={table} />
      <BlotterBulkUpdateModal
        opened={bulkUpdateModalOpened}
        onClose={() => setBulkUpdateModalOpened(false)}
        selectedTrades={selectedTrades}
        onSuccess={() => refetch()}
      />
    </div>
  );
};

export default BlotterTable;
