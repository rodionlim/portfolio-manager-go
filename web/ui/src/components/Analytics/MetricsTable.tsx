import React, { useMemo, useState } from "react";
import {
  Text,
  Box,
  Button,
  Group,
  Modal,
  Select,
  Stack,
  useMantineTheme,
} from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { NumberInput } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import { IconDownload, IconUpload, IconTrash } from "@tabler/icons-react";
import { TimestampedMetrics, MetricsJob } from "./types";
import { withRollingVolatility, VolatilityMethod } from "./volatility";

type MetricsTableProps = {
  volatilityMethod: VolatilityMethod;
  setVolatilityMethod: (method: VolatilityMethod) => void;
  volatilityWindow: number;
  setVolatilityWindow: (window: number) => void;
};

interface DeleteMetricsResponse {
  deleted: number;
  failed: number;
  failures: string[];
}

const MetricsTable: React.FC<MetricsTableProps> = ({
  volatilityMethod,
  setVolatilityMethod,
  volatilityWindow,
  setVolatilityWindow,
}) => {
  const theme = useMantineTheme();

  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedMetric, setSelectedMetric] =
    useState<TimestampedMetrics | null>(null);
  const [selectedMetrics, setSelectedMetrics] = useState<TimestampedMetrics[]>(
    []
  );
  const [isBatchDelete, setIsBatchDelete] = useState(false);
  const [selectedBookFilter, setSelectedBookFilter] = useState<string | null>(
    "None"
  );

  // Fetch metrics jobs to populate book filter dropdown
  const fetchMetricsJobs = async (): Promise<MetricsJob[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/historical/metrics/jobs"));
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching metrics jobs:", error);
      return [];
    }
  };

  // Query to fetch metrics jobs
  const { data: metricsJobs = [] } = useQuery({
    queryKey: ["metricsJobs"],
    queryFn: fetchMetricsJobs,
  });

  // Create book filter options
  const bookFilterOptions = useMemo(() => {
    const options = [{ value: "None", label: "None" }];
    metricsJobs.forEach((job) => {
      if (!options.find((opt) => opt.value === job.BookFilter)) {
        options.push({ value: job.BookFilter, label: job.BookFilter });
      }
    });
    return options;
  }, [metricsJobs]);

  // Fetch all historical metrics
  const fetchHistoricalMetrics = async (): Promise<TimestampedMetrics[]> => {
    try {
      const bookFilter =
        selectedBookFilter === "None" ? "" : selectedBookFilter || "";
      const url = bookFilter
        ? getUrl(
            `/api/v1/historical/metrics?book_filter=${encodeURIComponent(
              bookFilter
            )}`
          )
        : getUrl("/api/v1/historical/metrics");
      const resp = await fetch(url);
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching historical metrics:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch historical metrics: ${error.message}`,
      });
      return [];
    }
  };

  // Export metrics to CSV
  const exportMetricsCSV = async () => {
    try {
      const bookFilter =
        selectedBookFilter === "None" ? "" : selectedBookFilter || "";
      const queryUrl = bookFilter
        ? getUrl(
            `/api/v1/historical/metrics/export?book_filter=${encodeURIComponent(
              bookFilter
            )}`
          )
        : getUrl("/api/v1/historical/metrics/export");

      const response = await fetch(queryUrl);

      if (!response.ok) {
        throw new Error(`Export failed with status: ${response.status}`);
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "historical_metrics_export.csv";
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);

      notifications.show({
        title: "Success",
        message: "Historical metrics exported successfully",
        color: "green",
      });
    } catch (error: any) {
      console.error("Error exporting metrics:", error);
      notifications.show({
        color: "red",
        title: "Export Failed",
        message: error.message,
      });
    }
  };

  // Simplified delete process to use a single endpoint
  const deleteMetrics = async (timestamps: string[]) => {
    try {
      const bookFilter =
        selectedBookFilter === "None" ? "" : selectedBookFilter || "";
      const url = bookFilter
        ? getUrl(
            `/api/v1/historical/metrics/delete?book_filter=${encodeURIComponent(
              bookFilter
            )}`
          )
        : getUrl(`/api/v1/historical/metrics/delete`);

      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ timestamps }),
      });

      if (!response.ok) {
        throw new Error(`Delete failed with status: ${response.status}`);
      }

      const result = (await response.json()) as DeleteMetricsResponse;

      if (result.failed > 0) {
        notifications.show({
          title: "Partial Success",
          message: `Deleted ${result.deleted} record(s), failed to delete ${result.failed} record(s)`,
          color: "yellow",
        });
      } else {
        notifications.show({
          title: "Success",
          message: `Successfully deleted ${result.deleted} record(s)`,
          color: "green",
        });
      }

      // Refetch the data after deleting
      refetch();
    } catch (error: any) {
      console.error("Error deleting metrics:", error);
      notifications.show({
        color: "red",
        title: "Delete Failed",
        message: error.message,
      });
    }
  };

  const confirmDelete = () => {
    if (isBatchDelete) {
      const timestamps = selectedMetrics.map((metric) => metric.timestamp);
      deleteMetrics(timestamps);
    } else if (selectedMetric) {
      deleteMetrics([selectedMetric.timestamp]);
    }
    setDeleteModalOpen(false);
  };

  // Upload metrics CSV file
  const uploadFileRef = React.useRef<HTMLInputElement>(null);

  const handleUploadClick = () => {
    uploadFileRef.current?.click();
  };

  const handleFileUpload = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const files = event.target.files;
    if (!files || files.length === 0) return;

    const file = files[0];
    const formData = new FormData();
    formData.append("file", file);

    try {
      const response = await fetch(
        getUrl("/api/v1/historical/metrics/import"),
        {
          method: "POST",
          body: formData,
        }
      );

      if (!response.ok) {
        if (uploadFileRef.current) uploadFileRef.current.value = "";
        throw new Error(`Upload failed with status: ${response.status}`);
      }

      const result = await response.json();
      notifications.show({
        title: "Success",
        message: `Imported ${result.imported} historical metrics records`,
        color: "green",
      });

      // Reset file input
      if (uploadFileRef.current) uploadFileRef.current.value = "";

      // Refetch the data
      refetch();
    } catch (error: any) {
      console.error("Error uploading metrics:", error);
      notifications.show({
        color: "red",
        title: "Upload Failed",
        message: error.message,
      });
    }
  };

  // Query to fetch historical metrics
  const {
    data: historicalMetrics = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ["historicalMetrics", selectedBookFilter],
    queryFn: fetchHistoricalMetrics,
  });

  const metricsWithVolatility = useMemo(() => {
    return withRollingVolatility(historicalMetrics, {
      method: volatilityMethod,
      window: volatilityWindow,
    });
  }, [historicalMetrics, volatilityMethod, volatilityWindow]);

  const getVolatilityHeatStyles = (volatilityAnnDecimal: number) => {
    const volPct = volatilityAnnDecimal * 100;
    const normalized = Math.max(0, Math.min(1, volPct / 60));
    const bucket = Math.min(
      6,
      Math.max(0, Math.floor(normalized * theme.colors.red.length))
    );

    const shade =
      theme.colors.red[Math.min(bucket + 2, theme.colors.red.length - 1)];
    const alpha = 0.05 + normalized * 0.25;

    return {
      backgroundColor: `color-mix(in srgb, ${shade} ${Math.round(
        alpha * 100
      )}%, transparent)`,
      borderRadius: theme.radius.sm,
      paddingInline: theme.spacing.xs,
      paddingBlock: 2,
      display: "inline-block",
      minWidth: 72,
      textAlign: "right" as const,
      fontVariantNumeric: "tabular-nums" as const,
    };
  };

  // Define table columns
  const columns = useMemo<MRT_ColumnDef<TimestampedMetrics>[]>(
    () => [
      {
        accessorKey: "timestamp",
        header: "Date",
        Cell: ({ cell }) => {
          const date = new Date(cell.getValue<string>());
          return date.toLocaleDateString();
        },
        sortingFn: "datetime",
      },
      {
        accessorKey: "metrics.mv",
        header: "Market Value",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          })}`;
        },
      },
      {
        accessorKey: "metrics.pricePaid",
        header: "Price Paid",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          })}`;
        },
      },
      {
        accessorKey: "metrics.totalDividends",
        header: "Total Dividends",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toLocaleString(undefined, {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          })}`;
        },
      },
      {
        id: "pnl",
        header: "P&L",
        accessorFn: (row) => {
          const mv = row.metrics.mv;
          const pricePaid = row.metrics.pricePaid;
          const totalDividends = row.metrics.totalDividends;
          return mv - pricePaid + totalDividends;
        },
        Cell: ({ cell }) => {
          const pnl = cell.getValue<number>();
          const color = pnl >= 0 ? "green" : "red";
          return (
            <span style={{ color }}>
              $
              {pnl.toLocaleString(undefined, {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
              })}
            </span>
          );
        },
      },
      {
        accessorKey: "metrics.irr",
        header: "IRR",
        Cell: ({ cell }) => {
          const irr = cell.getValue<number>();
          const color = irr >= 0 ? "green" : "red";
          return <span style={{ color }}>{(irr * 100).toFixed(2)}%</span>;
        },
      },
      {
        id: "standardDeviation",
        header: "Volatility (Ann.)",
        accessorFn: (row) => row.metrics.standardDeviation,
        Cell: ({ cell }) => {
          const v = cell.getValue<number | undefined>();
          if (v === undefined || !Number.isFinite(v)) {
            return (
              <Text size="xs" c="dimmed">
                -
              </Text>
            );
          }

          return (
            <span style={getVolatilityHeatStyles(v)}>
              <Text size="xs" component="span">
                {(v * 100).toFixed(2)}%
              </Text>
            </span>
          );
        },
        size: 140,
      },
    ],
    []
  );

  // Configure the table
  const table = useMantineReactTable({
    columns,
    data: metricsWithVolatility,
    initialState: {
      sorting: [{ id: "timestamp", desc: true }], // Sort by date descending
    },
    state: {
      density: "xs",
    },
    mantineTableHeadCellProps: {
      style: {
        fontSize: "var(--mantine-font-size-xs)",
      },
    },
    mantineTableBodyCellProps: {
      style: {
        fontSize: "var(--mantine-font-size-xs)",
        paddingTop: 6,
        paddingBottom: 6,
      },
    },
    enablePagination: true,
    manualPagination: false,
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
        <Stack>
          <Group align="flex-end">
            <Select
              description="Book Filter"
              placeholder="Select book filter"
              data={bookFilterOptions}
              value={selectedBookFilter}
              onChange={setSelectedBookFilter}
              clearable={false}
              size="xs"
            />
            <Select
              description="Volatility Method"
              data={[
                { value: "sma", label: "SMA" },
                { value: "ewma", label: "EWMA" },
              ]}
              value={volatilityMethod}
              onChange={(v) =>
                setVolatilityMethod((v as VolatilityMethod) || "sma")
              }
              size="xs"
              w={160}
              clearable={false}
            />
            <NumberInput
              description="Window (days)"
              value={volatilityWindow}
              onChange={(v) => setVolatilityWindow(Number(v) || 1)}
              min={2}
              step={1}
              size="xs"
              w={140}
            />
          </Group>
          <Group>
            <Button
              onClick={exportMetricsCSV}
              leftSection={<IconDownload size={18} />}
              variant="outline"
            >
              Export CSV
            </Button>
            <Button
              onClick={handleUploadClick}
              leftSection={<IconUpload size={18} />}
              variant="outline"
            >
              Import CSV
            </Button>
            <Button
              onClick={() => {
                const selectedRows = table.getSelectedRowModel().rows;
                if (selectedRows.length === 1) {
                  // Single selection
                  setSelectedMetric(selectedRows[0].original);
                  setSelectedMetrics([]);
                  setIsBatchDelete(false);
                  setDeleteModalOpen(true);
                } else if (selectedRows.length > 1) {
                  // Multiple selection
                  setSelectedMetric(null);
                  setSelectedMetrics(selectedRows.map((row) => row.original));
                  setIsBatchDelete(true);
                  setDeleteModalOpen(true);
                }
              }}
              leftSection={<IconTrash size={18} />}
              variant="outline"
              color="red"
              disabled={table.getSelectedRowModel().rows.length === 0}
            >
              Delete{" "}
              {table.getSelectedRowModel().rows.length > 1
                ? `Selected (${table.getSelectedRowModel().rows.length})`
                : "Record"}
            </Button>
            <input
              ref={uploadFileRef}
              type="file"
              accept=".csv"
              onChange={handleFileUpload}
              style={{ display: "none" }}
            />
          </Group>
        </Stack>
      </Box>
    ),
  });

  // Handle loading states
  if (isLoading) return <div>Loading historical metrics...</div>;
  if (error) return <div>Error loading historical metrics</div>;

  // Render the component
  return (
    <div>
      <MantineReactTable table={table} />
      {historicalMetrics.length === 0 && !isLoading && !error && (
        <Text c="dimmed" ta="center" mt="md">
          No historical metrics records found
        </Text>
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        opened={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Confirm Deletion"
        size="sm"
      >
        {isBatchDelete ? (
          <Text mb="md">
            Are you sure you want to delete {selectedMetrics.length} historical
            metrics records?
          </Text>
        ) : (
          <Text mb="md">
            Are you sure you want to delete the historical metrics record from{" "}
            {selectedMetric &&
              new Date(selectedMetric.timestamp).toLocaleDateString()}
            ?
          </Text>
        )}
        <Group justify="flex-end">
          <Button variant="outline" onClick={() => setDeleteModalOpen(false)}>
            Cancel
          </Button>
          <Button color="red" onClick={confirmDelete}>
            Delete
          </Button>
        </Group>
      </Modal>

      <Box mt="xs">
        <Text size="xs" c="dimmed">
          Volatility (Ann.) is the rolling standard deviation of daily portfolio
          returns, annualized by multiplying by √252. Daily return is computed
          from market value as (MVₜ − MVₜ₋₁) / MVₜ₋₁ and resets across large
          date gaps.
        </Text>
        <Text size="xs" c="dimmed" mt={4}>
          Why this matters: one of the simplest ways to estimate future
          volatility is to use a measure of recent standard deviation. This
          often works because volatility tends to persist—if the market has been
          crazy over the last few weeks, it will probably continue to be crazy
          for a few more weeks.
        </Text>
      </Box>
    </div>
  );
};

export default MetricsTable;
