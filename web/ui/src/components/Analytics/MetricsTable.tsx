import React, { useMemo, useState } from "react";
import { Text, Box, Button, Group, Modal } from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import { IconDownload, IconUpload, IconTrash } from "@tabler/icons-react";
import { TimestampedMetrics } from "./types";

interface DeleteMetricsResponse {
  deleted: number;
  failed: number;
  failures: string[];
}

const MetricsTable: React.FC = () => {
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedMetric, setSelectedMetric] =
    useState<TimestampedMetrics | null>(null);
  const [selectedMetrics, setSelectedMetrics] = useState<TimestampedMetrics[]>(
    []
  );
  const [isBatchDelete, setIsBatchDelete] = useState(false);

  // Fetch all historical metrics
  const fetchHistoricalMetrics = async (): Promise<TimestampedMetrics[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/historical/metrics"));
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
      const response = await fetch(getUrl("/api/v1/historical/metrics/export"));

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
      const response = await fetch(
        getUrl(`/api/v1/historical/metrics/delete`),
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ timestamps }),
        }
      );

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
    queryKey: ["historicalMetrics"],
    queryFn: fetchHistoricalMetrics,
  });

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
    ],
    []
  );

  // Configure the table
  const table = useMantineReactTable({
    columns,
    data: historicalMetrics,
    initialState: {
      sorting: [{ id: "timestamp", desc: true }], // Sort by date descending
    },
    state: {
      density: "xs",
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
    </div>
  );
};

export default MetricsTable;
