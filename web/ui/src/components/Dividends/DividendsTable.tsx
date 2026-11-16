import React, { useMemo, useState, useEffect } from "react";
import {
  ActionIcon,
  Badge,
  Box,
  Button,
  FileInput,
  Group,
  NumberInput,
  Select,
  Text,
  Tooltip,
  useMantineColorScheme,
} from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import {
  IconUpload,
  IconDownload,
  IconPencil,
  IconCheck,
  IconX,
} from "@tabler/icons-react";
import { useDispatch } from "react-redux";
import { AppDispatch } from "../../store";
import { fetchReferenceData } from "../../slices/referenceDataSlice";
import type { ReferenceData } from "../../types";

interface Position {
  Ticker: string;
}

interface Dividend {
  ExDate: string;
  Amount: number;
  AmountPerShare: number;
  Qty: number;
  Source: string;
}

interface DividendsMetadata {
  Ticker: string;
  ExDate: string;
  Amount: number;
  Interest: number;
  AvgInterest: number;
  WithholdingTax: number;
  Source: string;
}

interface DividendsTableProps {
  initialTicker?: string | null;
}

const DividendsTable: React.FC<DividendsTableProps> = ({
  initialTicker = null,
}) => {
  const { colorScheme } = useMantineColorScheme();
  const queryClient = useQueryClient();
  const dispatch = useDispatch<AppDispatch>();
  const [savingCompletion, setSavingCompletion] = useState(false);
  const { data: referenceData } = useQuery<ReferenceData>({
    queryKey: ["dividendReferenceData"],
    queryFn: async () => {
      const resp = await fetch(getUrl("/api/v1/refdata"));
      if (!resp.ok) throw new Error("Unable to load ticker settings");
      return resp.json();
    },
  });
  const [selectedTicker, setSelectedTicker] = useState<string | null>(
    initialTicker,
  );
  const [editingExDate, setEditingExDate] = useState<string | null>(null);
  const [editingAmountPerShare, setEditingAmountPerShare] = useState<
    number | string
  >("");
  const [isSavingEdit, setIsSavingEdit] = useState<boolean>(false);
  const selectedReference = selectedTicker ? referenceData?.[selectedTicker] : undefined;
  const historyComplete = selectedReference?.dividend_history_complete === true;
  const toggleCompletion = async () => {
    if (!selectedReference) return;
    setSavingCompletion(true);
    try {
      const resp = await fetch(getUrl("/api/v1/refdata"), {
        method: "PUT", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ...selectedReference, dividend_history_complete: !historyComplete }),
      });
      if (!resp.ok) {
        const error = await resp.json();
        throw new Error(error.message || "Unable to update dividend history status");
      }
      // Recalculate any cached totals using the updated refresh policy.
      const reset = await fetch(getUrl("/api/v1/dividends/cache/reset"), { method: "POST" }).catch(() => null);
      await queryClient.invalidateQueries({ queryKey: ["dividendReferenceData"] });
      await queryClient.invalidateQueries({ queryKey: ["refData"] });
      dispatch(fetchReferenceData());
      await queryClient.invalidateQueries({ queryKey: ["dividends"] });
      notifications.show({
        title: "Dividend history status saved",
        message: reset?.ok
          ? (historyComplete ? "Automatic dividend refresh resumed." : "History marked complete. Only saved dividends will be used.")
          : "Status saved, but cached totals could not be cleared. They may remain until the cache expires.",
        color: reset?.ok ? "green" : "yellow",
      });
    } catch (error) {
      notifications.show({ color: "red", title: "Unable to update dividend history", message: String(error) });
    } finally { setSavingCompletion(false); }
  };

  // Effect to update selectedTicker when initialTicker changes (e.g., from navigation)
  useEffect(() => {
    if (initialTicker) {
      setSelectedTicker(initialTicker);
    }
  }, [initialTicker]);

  // Reset inline edit state when ticker changes
  useEffect(() => {
    setEditingExDate(null);
    setEditingAmountPerShare("");
  }, [selectedTicker]);

  // Fetch all positions to populate the dropdown of tickers
  const fetchPositions = async (): Promise<Position[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions/lite"));
      if (!resp.ok) throw new Error("Unable to load tickers");
      const tickers: Position[] = await resp.json();
      return tickers;
    } catch (error: any) {
      console.error("Error fetching positions:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch positions: ${error.message}`,
      });
      throw error;
    }
  };

  // Fetch dividends for a specific ticker
  const fetchDividends = async (ticker: string | null): Promise<Dividend[]> => {
    if (!ticker) return [];

    try {
      const resp = await fetch(getUrl(`/api/v1/dividends/${ticker}`));
      if (!resp.ok) {
        const error = await resp.json().catch(() => ({ message: "Unable to fetch dividends" }));
        throw new Error(error.message || "Unable to fetch dividends");
      }
      const dividends = await resp.json();
      return dividends;
    } catch (error: any) {
      console.error("Error fetching dividends:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch dividends for ${ticker}: ${error.message}`,
      });
      throw error;
    }
  };

  const uploadDividendsCSV = async (
    file: File,
  ): Promise<{ message: string }> => {
    const formData = new FormData();
    formData.append("file", file);

    return fetch(getUrl("api/v1/mdata/dividends/upload"), {
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
          throw new Error("An error occurred while uploading custom dividends");
        },
      );
  };

  // Handle customs dividends file upload
  const handleFileUpload = async (file: File | null) => {
    if (!file) return;

    uploadDividendsCSV(file).then(
      (resp: { message: string }) => {
        notifications.show({
          title: "Custom dividends successfully uploaded",
          message: `${resp.message}`,
          autoClose: 10000,
        });
        queryClient.invalidateQueries({ queryKey: ["dividends"] });
      },
      (error) => {
        notifications.show({
          color: "red",
          title: "Error",
          message: `Unable to upload custom dividends to the market data service\n ${error}`,
        });
      },
    );
  };

  // Handle exporting dividends to CSV
  const handleExportCSV = async () => {
    if (!selectedTicker) return;

    try {
      const resp = await fetch(
        getUrl(`/api/v1/mdata/dividends/${selectedTicker}`),
      );
      const data: DividendsMetadata[] = await resp.json();

      if (!data || data.length === 0) {
        notifications.show({
          color: "yellow",
          title: "No data",
          message: `No dividend metadata found for ${selectedTicker} to export.`,
        });
        return;
      }

      // Create CSV content
      const headers = [
        "Ticker",
        "ExDate",
        "Amount",
        "Interest",
        "AvgInterest",
        "WithholdingTax",
      ];
      const csvRows = [
        headers.join(","),
        ...data.map((row) =>
          [
            selectedTicker,
            new Date(row.ExDate).toISOString().split("T")[0],
            row.Amount,
            row.Interest,
            row.AvgInterest,
            row.WithholdingTax,
          ].join(","),
        ),
      ];
      const csvContent = csvRows.join("\n");

      // Trigger download
      const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.setAttribute("href", url);
      link.setAttribute("download", `${selectedTicker}_dividends.csv`);
      link.style.visibility = "hidden";
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error: any) {
      console.error("Error exporting dividends:", error);
      notifications.show({
        color: "red",
        title: "Export Error",
        message: `Unable to export dividends for ${selectedTicker}: ${error.message}`,
      });
    }
  };

  const handleStartInlineEdit = (dividend: Dividend) => {
    setEditingExDate(dividend.ExDate);
    setEditingAmountPerShare(dividend.AmountPerShare);
  };

  const handleCancelInlineEdit = () => {
    setEditingExDate(null);
    setEditingAmountPerShare("");
  };

  const handleSaveInlineEdit = async (dividend: Dividend) => {
    if (!selectedTicker) return;

    const parsedAmount = Number(editingAmountPerShare);
    if (!Number.isFinite(parsedAmount) || parsedAmount <= 0) {
      notifications.show({
        color: "red",
        title: "Invalid Amount",
        message: "Amount per share must be a valid positive number.",
      });
      return;
    }

    setIsSavingEdit(true);
    try {
      const metaResp = await fetch(
        getUrl(`/api/v1/mdata/dividends/${selectedTicker}`),
      );
      if (!metaResp.ok) {
        throw new Error(
          `Failed to load dividend metadata (${metaResp.status})`,
        );
      }
      const metadata: DividendsMetadata[] = (await metaResp.json()) || [];
      const existingMeta = metadata.find((m) => m.ExDate === dividend.ExDate);

      const fallbackWithholdingTax =
        dividend.Qty > 0 && dividend.AmountPerShare > 0
          ? Math.max(
              0,
              Math.round(
                (1 - dividend.Amount / (dividend.Qty * dividend.AmountPerShare)) *
                  10000,
              ) / 10000,
            )
          : 0;

      // Preserve existing Custom dividends for other dates while keeping unedited Official rows untouched
      const customDividends: DividendsMetadata[] = metadata.filter(
        (m) => m.Source === "Custom" && m.ExDate !== dividend.ExDate,
      );

      customDividends.push({
        Ticker: existingMeta?.Ticker || selectedTicker,
        ExDate: dividend.ExDate,
        Amount: parsedAmount,
        Interest: existingMeta?.Interest ?? 0,
        AvgInterest: existingMeta?.AvgInterest ?? 0,
        WithholdingTax: existingMeta?.WithholdingTax ?? fallbackWithholdingTax,
        Source: "Custom",
      });

      customDividends.sort((a, b) => a.ExDate.localeCompare(b.ExDate));

      const storeResp = await fetch(
        getUrl(`/api/v1/mdata/dividends/${selectedTicker}`),
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(customDividends),
        },
      );

      if (!storeResp.ok) {
        const errText = await storeResp.text();
        throw new Error(
          errText || `Failed to save custom dividend (${storeResp.status})`,
        );
      }

      // Reset aggregated dividends cache so summary tabs also reflect the custom change
      fetch(getUrl("/api/v1/dividends/cache/reset"), {
        method: "POST",
      }).catch(() => {});

      await queryClient.invalidateQueries({ queryKey: ["dividends"] });
      queryClient.invalidateQueries({ queryKey: ["allDividends"] });

      setEditingExDate(null);
      setEditingAmountPerShare("");
      notifications.show({
        color: "green",
        title: "Dividend Updated",
        message: `Saved custom dividend for ${selectedTicker} on ${new Date(dividend.ExDate).toLocaleDateString()}.`,
      });
    } catch (error: any) {
      console.error("Error saving custom dividend:", error);
      notifications.show({
        color: "red",
        title: "Save Error",
        message: `Unable to save custom dividend for ${selectedTicker}: ${error.message}`,
      });
    } finally {
      setIsSavingEdit(false);
    }
  };

  // Query to fetch positions for dropdown
  const {
    data: positions = [],
    isLoading: isLoadingPositions,
    error: positionsError,
  } = useQuery({
    queryKey: ["dividendTickerPositions"],
    queryFn: fetchPositions,
  });

  // Query to fetch dividends when ticker is selected
  const {
    data: dividends = [],
    isLoading: isLoadingDividends,
    error: dividendsError,
  } = useQuery({
    queryKey: ["dividends", selectedTicker],
    queryFn: () => fetchDividends(selectedTicker),
    enabled: !!selectedTicker, // Only run query when ticker is selected
  });

  // Calculate totals
  const totalAmount = useMemo(() => {
    return (dividends || []).reduce((sum, div) => sum + div.Amount, 0);
  }, [dividends]);

  // Define table columns
  const columns = useMemo<MRT_ColumnDef<Dividend>[]>(
    () => [
      {
        accessorKey: "ExDate",
        header: "Ex-Dividend Date",
        Cell: ({ cell }) => {
          const date = new Date(cell.getValue<string>());
          return date.toLocaleDateString();
        },
        sortingFn: "datetime",
      },
      {
        accessorKey: "Qty",
        header: "Quantity",
        Cell: ({ cell }) => {
          return cell.getValue<number>().toLocaleString();
        },
      },
      {
        accessorKey: "Source",
        header: "Source",
        Cell: ({ cell }) => {
          const source = cell.getValue<string>() || "Unknown";
          const color =
            source === "Custom"
              ? "orange"
              : source === "Official"
                ? "blue"
                : "gray";

          return (
            <Badge color={color} variant="light">
              {source}
            </Badge>
          );
        },
      },
      {
        accessorKey: "AmountPerShare",
        header: "Amount Per Share",
        Cell: ({ cell, row }) => {
          if (editingExDate === row.original.ExDate) {
            return (
              <NumberInput
                value={editingAmountPerShare}
                onChange={setEditingAmountPerShare}
                decimalScale={4}
                min={0}
                step={0.0001}
                size="xs"
                w={130}
                disabled={isSavingEdit}
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleSaveInlineEdit(row.original);
                  } else if (e.key === "Escape") {
                    e.preventDefault();
                    handleCancelInlineEdit();
                  }
                }}
              />
            );
          }
          return cell.getValue<number>().toFixed(4);
        },
      },
      {
        accessorKey: "Amount",
        header: "Total Amount (After withholding tax)",
        Cell: ({ cell, row }) => {
          if (editingExDate === row.original.ExDate) {
            const parsed = Number(editingAmountPerShare);
            if (Number.isFinite(parsed) && parsed >= 0) {
              const netMultiplier =
                row.original.AmountPerShare > 0
                  ? row.original.Amount / row.original.AmountPerShare
                  : row.original.Qty;
              return (parsed * netMultiplier).toFixed(2);
            }
          }
          return cell.getValue<number>().toFixed(2);
        },
      },
      {
        id: "actions",
        header: "Actions",
        enableSorting: false,
        size: 90,
        Cell: ({ row }) => {
          const isEditingRow = editingExDate === row.original.ExDate;
          if (isEditingRow) {
            return (
              <Group gap={4} wrap="nowrap">
                <Tooltip label="Save custom dividend">
                  <ActionIcon
                    color="green"
                    variant="subtle"
                    size="sm"
                    loading={isSavingEdit}
                    onClick={() => handleSaveInlineEdit(row.original)}
                    aria-label="Save custom dividend"
                  >
                    <IconCheck size={16} />
                  </ActionIcon>
                </Tooltip>
                <Tooltip label="Cancel">
                  <ActionIcon
                    color="gray"
                    variant="subtle"
                    size="sm"
                    disabled={isSavingEdit}
                    onClick={handleCancelInlineEdit}
                    aria-label="Cancel edit"
                  >
                    <IconX size={16} />
                  </ActionIcon>
                </Tooltip>
              </Group>
            );
          }

          return (
            <Tooltip label="Edit amount per share">
              <ActionIcon
                variant="subtle"
                size="sm"
                disabled={isSavingEdit}
                onClick={() => handleStartInlineEdit(row.original)}
                aria-label="Edit amount per share"
              >
                <IconPencil size={16} />
              </ActionIcon>
            </Tooltip>
          );
        },
      },
    ],
    [editingExDate, editingAmountPerShare, isSavingEdit, selectedTicker],
  );

  // Configure the table
  const table = useMantineReactTable({
    columns,
    data: dividends,
    initialState: {
      sorting: [{ id: "ExDate", desc: true }], // Sort by ex-date descending
    },
    state: { density: "xs" },
    enablePagination: true,
    manualPagination: false,
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: () => (
      <Box
        style={{
          display: "flex",
          gap: "16px",
          padding: "4px",
          alignItems: "center",
        }}
      >
        <Select
          placeholder="Select ticker..."
          data={Array.from(
            new Set(positions.map((position) => position.Ticker)),
          )
            .sort()
            .map((ticker) => ({
              label: ticker,
              value: ticker,
            }))}
          value={selectedTicker}
          onChange={setSelectedTicker}
          style={{ width: "300px" }}
          searchable
          clearable
          nothingFoundMessage="No tickers found"
        />
        {selectedTicker && (
          <Text fw={500} size="sm">
            Total Dividends: ${totalAmount.toFixed(2)}
          </Text>
        )}
        <Tooltip
          multiline
          w={300}
          label={
            <Box>
              <Text size="xs" fw={700}>
                Required fields:
              </Text>
              <Text size="xs" mb="xs">
                Ticker, ExDate, Amount, Interest, AvgInterest, WithholdingTax
              </Text>
              <Text size="xs" c={colorScheme === "dark" ? "red.6" : "yellow.8"}>
                Note: Uploading will replace existing custom dividends for
                tickers in the file. Always upload all custom dividends for a
                single ticker together. Tickers not in the upload will remain
                unchanged.
              </Text>
            </Box>
          }
        >
          <FileInput
            placeholder="Upload CSV"
            accept=".csv"
            onChange={handleFileUpload}
            leftSection={<IconUpload size={16} />}
            style={{ width: "200px" }}
            clearable
          />
        </Tooltip>
        {selectedTicker && (
          <Button
            variant="light"
            leftSection={<IconDownload size={16} />}
            onClick={handleExportCSV}
          >
            Export CSV
          </Button>
        )}
      </Box>
    ),
  });

  // Handle loading states
  if (isLoadingPositions) return <div>Loading positions...</div>;
  if (positionsError) return <div>Error loading positions</div>;

  // Render the component
  return (
    <div>
    {selectedReference?.asset_class === "bond" && (
      <Box mb="md" p="sm" style={{ border: "1px solid var(--mantine-color-default-border)", borderRadius: 8 }}>
        <Group justify="space-between">
          <Badge color={historyComplete ? "green" : "blue"}>
            {historyComplete ? "Dividend history complete" : "Automatic dividend refresh"}
          </Badge>
          <Button variant="light" loading={savingCompletion} onClick={toggleCompletion}>
            {historyComplete ? "Resume dividend refresh" : "Mark dividend history complete"}
          </Button>
        </Group>
        <Text size="sm" mt="xs">
          {historyComplete
            ? "This bond uses saved dividend history only. Resume refreshing if you need to check for corrections."
            : "For expired or redeemed bonds only: mark complete after checking that all coupons are recorded. This stops future dividend fetches for this ticker."}
        </Text>
      </Box>
    )}
      <MantineReactTable table={table} />
      {dividendsError && (
        <Text c="red" ta="center" mt="md">
          Error loading dividends
        </Text>
      )}
      {selectedTicker &&
        (dividends || []).length === 0 &&
        !isLoadingDividends &&
        !dividendsError && (
          <Text c="dimmed" ta="center" mt="md">
            No dividend records found for {selectedTicker}
          </Text>
        )}
    </div>
  );
};

export default DividendsTable;
