import React, { useMemo, useState, useEffect } from "react";
import { Box, Button, FileInput, ActionIcon, Tooltip } from "@mantine/core";
import { IconUpload, IconPaperclip, IconDownload } from "@tabler/icons-react";
import { useMediaQuery } from "@mantine/hooks";
import {
  MantineReactTable,
  MRT_ColumnDef,
  MRT_TableInstance,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { useSelector } from "react-redux";
import { notifications } from "@mantine/notifications";
import { useLocation, useNavigate } from "react-router-dom";
import { getUrl } from "../../utils/url";
import { Trade } from "../../types/blotter";
import { RootState } from "../../store";
import BlotterBulkUpdateModal from "./BlotterBulkUpdateModal";

const fetchTrades = async (): Promise<Trade[]> => {
  return fetch(getUrl("/api/v1/blotter/trade"))
    .then((resp) => resp.json())
    .then(
      (data: Trade[]) => {
        return data.sort(
          (x, y) =>
            new Date(y.TradeDate).getTime() - new Date(x.TradeDate).getTime(),
        );
      },
      (error) => {
        console.error("error", error);
        throw new Error(
          `An error occurred while fetching trades ${error.message}`,
        );
      },
    );
};

const deleteTrades = async (trades: string[]): Promise<{ message: string }> => {
  return fetch(getUrl("/api/v1/blotter/trade"), {
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
      },
    );
};

const uploadTradesCSV = async (file: File): Promise<{ message: string }> => {
  const formData = new FormData();
  formData.append("file", file);

  return fetch(getUrl("/api/v1/blotter/upload"), {
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
      },
    );
};

const fetchConfirmationsMetadata = async (): Promise<{
  [tradeId: string]: boolean;
}> => {
  return fetch(getUrl("/api/v1/blotter/confirmations/metadata"))
    .then((resp) => resp.json())
    .then(
      (data: Array<{ tradeId: string }>) => {
        const map: { [tradeId: string]: boolean } = {};
        data.forEach((item) => {
          map[item.tradeId] = true;
        });
        return map;
      },
      (error) => {
        console.error("error fetching confirmations", error);
        return {};
      },
    );
};

const exportConfirmations = async (tradeIds: string[]): Promise<Blob> => {
  return fetch(getUrl("/api/v1/blotter/confirmations/export"), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(tradeIds),
  }).then((resp) => {
    if (!resp.ok) {
      throw new Error("Failed to export confirmations");
    }
    return resp.blob();
  });
};

const BlotterTable: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [bulkUpdateModalOpened, setBulkUpdateModalOpened] = useState(false);
  const [selectedTrades, setSelectedTrades] = useState<Trade[]>([]);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const refData = useSelector((state: RootState) => state.referenceData.data);
  const [confirmationsMap, setConfirmationsMap] = useState<{
    [tradeId: string]: boolean;
  }>({});

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

  // Fetch confirmations metadata whenever trades change
  useEffect(() => {
    fetchConfirmationsMetadata().then(setConfirmationsMap);
  }, [trades]);

  const columns = useMemo<MRT_ColumnDef<Trade>[]>(
    () => [
      // Conditionally include TradeID column only on non-mobile devices
      ...(isMobile
        ? []
        : [{ accessorKey: "TradeID" as keyof Trade, header: "Trade ID" }]),
      {
        accessorKey: "TradeDate",
        header: "Date",
        filterVariant: "date",
        sortingFn: "datetime",
        accessorFn: (row) => new Date(row.TradeDate),
        Cell: ({ cell }) => {
          const date = cell.getValue<Date>();
          return date instanceof Date && !isNaN(date.getTime())
            ? date.toLocaleDateString()
            : "";
        },
        columnFilterModeOptions: [
          "equals",
          "greaterThan",
          "greaterThanOrEqualTo",
          "lessThan",
          "lessThanOrEqualTo",
        ],
      },
      { accessorKey: "Ticker", header: "Ticker" },
      {
        id: "tickerName", // Use a unique id instead of accessorKey
        header: "Name",
        accessorFn: (row) => refData?.[row.Ticker]?.name || "", // This enables filtering
        Cell: ({ row }) => {
          const ticker = row.original.Ticker;
          return refData?.[ticker]?.name || "";
        },
      },
      {
        id: "asset_class",
        header: "Asset Class",
        accessorFn: (row) => refData?.[row.Ticker]?.asset_class || "",
      },
      {
        id: "asset_sub_class",
        header: "Asset Sub Class",
        accessorFn: (row) => refData?.[row.Ticker]?.asset_sub_class || "",
      },
      {
        id: "ccy",
        header: "CCY",
        accessorFn: (row) => refData?.[row.Ticker]?.ccy || "",
      },
      {
        id: "category",
        header: "Category",
        accessorFn: (row) => refData?.[row.Ticker]?.category || "",
      },
      {
        id: "sub_category",
        header: "Sub Category",
        accessorFn: (row) => refData?.[row.Ticker]?.sub_category || "",
      },
      {
        id: "domicile",
        header: "Domicile",
        accessorFn: (row) => refData?.[row.Ticker]?.domicile || "",
      },
      { accessorKey: "Side", header: "Side" },
      { accessorKey: "Book", header: "Book" },
      // { accessorKey: "Broker", header: "Broker" },
      { accessorKey: "Account", header: "Account" },
      { accessorKey: "Quantity", header: "Quantity" },
      { accessorKey: "Price", header: "Price" },
      {
        id: "Value",
        header: "Value",
        accessorFn: (row) => (row.Price * row.Quantity).toFixed(2),
        Cell: ({ cell }) => cell.getValue<string>(),
      },
      {
        accessorKey: "Fx",
        header: "Fx",
        Cell: ({ cell }) => Number(cell.getValue()).toFixed(4),
      },
      {
        id: "confirmation",
        header: "Confirmation",
        enableColumnFilter: false,
        enableSorting: false,
        Cell: ({ row }) => {
          const hasConfirmation = confirmationsMap[row.original.TradeID];
          if (!hasConfirmation) return null;

          return (
            <Tooltip label="Has confirmation">
              <ActionIcon
                variant="subtle"
                color="blue"
                onClick={() => {
                  // Download confirmation
                  window.open(
                    getUrl(
                      `/api/v1/blotter/confirmation/${row.original.TradeID}`,
                    ),
                    "_blank",
                  );
                }}
              >
                <IconPaperclip size={18} />
              </ActionIcon>
            </Tooltip>
          );
        },
      },
      // { accessorKey: "TradeType", header: "Trade Type" },
      // { accessorKey: "SeqNum", header: "Seq Num" },
    ],
    [isMobile, refData, confirmationsMap],
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
      columnVisibility: {
        asset_class: false,
        asset_sub_class: false,
        ccy: false,
        category: false,
        sub_category: false,
        domicile: false,
        confirmation: false,
      },
    },
    state: { density: "xs" },
    enableColumnFilterModes: true,
    enableDensityToggle: false,
    enableRowSelection: true,
    mantineTopToolbarProps: {
      style: {
        overflowX: "auto",
        paddingBottom: "4px",
      },
    },
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: ({ table }) => (
      <Box
        style={{
          display: "flex",
          gap: "16px",
          padding: "4px",
          alignItems: "center",
          flexWrap: "nowrap",
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
        <Button
          color="gray"
          variant="outline"
          onClick={() => handleExportCSV(table)}
        >
          Export CSV
        </Button>
        <Button
          color="grape"
          variant="outline"
          onClick={() => handleExportConfirmations(table)}
          leftSection={<IconDownload size={16} />}
        >
          Export Confirmations
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
    table: MRT_TableInstance<Trade>,
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
          },
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
        },
      )
      .finally(() => {
        refetch();
      });
  };

  // Add the export handler function
  const handleExportCSV = (table: MRT_TableInstance<Trade>) => {
    const visibleColumns = table.getVisibleLeafColumns();
    // Exclude the selection column and any internal MRT columns if they exist
    const columnsToExport = visibleColumns.filter(
      (col) => col.id !== "mrt-row-select" && col.id !== "mrt-row-actions",
    );

    const headers = columnsToExport
      .map((col) => col.columnDef.header)
      .join(",");

    const rows = table.getFilteredRowModel().rows;
    const csvData = rows
      .map((row) => {
        return columnsToExport
          .map((col) => {
            let stringValue = "";
            const value = row.getValue(col.id);

            if (col.id === "TradeDate") {
              // Preserve original ISO format for TradeDate
              stringValue = row.original.TradeDate;
            } else if (value instanceof Date) {
              stringValue = value.toISOString();
            } else {
              stringValue =
                value !== null && value !== undefined ? String(value) : "";
            }

            // Escape quotes and wrap in quotes if contains comma or newline
            if (
              stringValue.includes(",") ||
              stringValue.includes('"') ||
              stringValue.includes("\n")
            ) {
              return `"${stringValue.replace(/"/g, '""')}"`;
            }
            return stringValue;
          })
          .join(",");
      })
      .join("\n");

    const blob = new Blob([headers + "\n" + csvData], {
      type: "text/csv;charset=utf-8;",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", "trades_blotter.csv");
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  // Handle export confirmations
  const handleExportConfirmations = (table: MRT_TableInstance<Trade>) => {
    // Get filtered rows
    const filteredRows = table.getFilteredRowModel().rows;

    // Get trade IDs from filtered rows
    const tradeIds = filteredRows.map((row) => row.original.TradeID);

    // Filter to only include trades that have confirmations
    const tradeIdsWithConfirmations = tradeIds.filter(
      (id) => confirmationsMap[id],
    );

    if (tradeIdsWithConfirmations.length === 0) {
      notifications.show({
        color: "yellow",
        title: "No Confirmations",
        message: "No confirmations found for the current selection",
      });
      return;
    }

    // Export confirmations
    exportConfirmations(tradeIdsWithConfirmations)
      .then((blob) => {
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        const dateString = new Date()
          .toISOString()
          .split("T")[0]
          .replace(/-/g, "");
        link.setAttribute("download", `confirmations_${dateString}.zip`);
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);

        notifications.show({
          title: "Export Successful",
          message: `Exported ${tradeIdsWithConfirmations.length} confirmation(s)`,
          autoClose: 5000,
        });
      })
      .catch((error) => {
        notifications.show({
          color: "red",
          title: "Export Failed",
          message: `Failed to export confirmations: ${error.message}`,
        });
      });
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
