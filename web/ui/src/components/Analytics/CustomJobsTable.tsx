import React, { useMemo, useState } from "react";
import {
  Text,
  Box,
  Button,
  Group,
  Modal,
  TextInput,
  Select,
  Stack,
  ActionIcon,
  Loader,
} from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import { IconPlus, IconX } from "@tabler/icons-react";

interface MetricsJob {
  BookFilter: string;
  CronExpr: string;
  TaskId: string;
}

interface Trade {
  Book: string;
  // other trade properties
}

interface CreateJobRequest {
  bookFilter: string;
  cronExpr: string;
}

const CustomJobsTable: React.FC = () => {
  const [addJobModalOpen, setAddJobModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedJob, setSelectedJob] = useState<MetricsJob | null>(null);
  const [newJobBookFilter, setNewJobBookFilter] = useState("");
  const [newJobCronExpr, setNewJobCronExpr] = useState("");

  // Fetch custom metrics jobs
  const fetchCustomJobs = async (): Promise<MetricsJob[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/historical/metrics/jobs"));
      if (!resp.ok) {
        throw new Error(`Failed to fetch jobs: ${resp.status}`);
      }
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching custom jobs:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch custom jobs: ${error.message}`,
      });
      return [];
    }
  };

  // Fetch trades to get book options
  const fetchTrades = async (): Promise<Trade[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/blotter/trade"));
      if (!resp.ok) {
        throw new Error(`Failed to fetch trades: ${resp.status}`);
      }
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching trades:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch trades: ${error.message}`,
      });
      return [];
    }
  };

  // Query to fetch custom jobs
  const {
    data: customJobs = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ["customJobs"],
    queryFn: fetchCustomJobs,
  });

  // Query to fetch trades (only when modal is open)
  const { data: trades = [], isLoading: tradesLoading } = useQuery({
    queryKey: ["trades"],
    queryFn: fetchTrades,
    enabled: addJobModalOpen, // Only fetch when modal is open
  });

  // Extract unique books from trades
  const bookOptions = useMemo(() => {
    return Array.from(new Set(trades.map((trade) => trade.Book))).map(
      (book) => ({ value: book, label: book })
    );
  }, [trades]);

  // Create new metrics job
  const createJob = async (request: CreateJobRequest) => {
    try {
      const response = await fetch(getUrl("/api/v1/historical/metrics/jobs"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `Failed to create job: ${response.status} - ${errorText}`
        );
      }

      notifications.show({
        title: "Success",
        message: "Custom metrics job created successfully",
        color: "green",
      });

      // Refetch the data after creating
      refetch();
    } catch (error: any) {
      console.error("Error creating job:", error);
      notifications.show({
        color: "red",
        title: "Create Failed",
        message: error.message,
      });
    }
  };

  // Delete metrics job
  const deleteJob = async (bookFilter: string) => {
    try {
      const response = await fetch(
        getUrl(
          `/api/v1/historical/metrics/jobs/${encodeURIComponent(bookFilter)}`
        ),
        {
          method: "DELETE",
        }
      );

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `Failed to delete job: ${response.status} - ${errorText}`
        );
      }

      notifications.show({
        title: "Success",
        message: "Custom metrics job deleted successfully",
        color: "green",
      });

      // Refetch the data after deleting
      refetch();
    } catch (error: any) {
      console.error("Error deleting job:", error);
      notifications.show({
        color: "red",
        title: "Delete Failed",
        message: error.message,
      });
    }
  };

  const handleCreateJob = () => {
    if (!newJobBookFilter) {
      notifications.show({
        color: "red",
        title: "Validation Error",
        message: "Book filter is required",
      });
      return;
    }

    createJob({
      bookFilter: newJobBookFilter,
      cronExpr: newJobCronExpr,
    });

    // Reset form and close modal
    setNewJobBookFilter("");
    setNewJobCronExpr("");
    setAddJobModalOpen(false);
  };

  const handleDeleteJob = () => {
    if (selectedJob) {
      deleteJob(selectedJob.BookFilter);
    }
    setDeleteModalOpen(false);
    setSelectedJob(null);
  };

  const openDeleteModal = (job: MetricsJob) => {
    setSelectedJob(job);
    setDeleteModalOpen(true);
  };

  // Define table columns
  const columns = useMemo<MRT_ColumnDef<MetricsJob>[]>(
    () => [
      {
        accessorKey: "BookFilter",
        header: "Book Filter",
        Cell: ({ cell }) => <Text fw={500}>{cell.getValue<string>()}</Text>,
      },
      {
        accessorKey: "CronExpr",
        header: "Cron Schedule",
        Cell: ({ cell }) => (
          <Text ff="monospace" size="sm">
            {cell.getValue<string>()}
          </Text>
        ),
      },
      {
        id: "actions",
        header: "Actions",
        Cell: ({ row }) => (
          <ActionIcon
            color="red"
            variant="subtle"
            onClick={() => openDeleteModal(row.original)}
          >
            <IconX size={16} />
          </ActionIcon>
        ),
      },
    ],
    []
  );

  const table = useMantineReactTable({
    columns,
    data: customJobs,
    enableRowSelection: false,
    enableColumnActions: false,
    enableColumnFilters: false,
    enablePagination: false,
    enableSorting: false,
    mantineTableHeadCellProps: {
      style: {
        backgroundColor: "#f8f9fa",
      },
    },
  });

  if (error) {
    return (
      <Box>
        <Text c="red">Error loading custom jobs: {error.message}</Text>
      </Box>
    );
  }

  return (
    <Box>
      <Group justify="space-between" mb="md">
        <Text size="lg" fw={600}>
          Custom Metrics Jobs
        </Text>
        <Button
          leftSection={<IconPlus size={16} />}
          onClick={() => setAddJobModalOpen(true)}
        >
          Add Job
        </Button>
      </Group>

      {isLoading ? (
        <Group justify="center" py="xl">
          <Loader />
          <Text>Loading custom jobs...</Text>
        </Group>
      ) : (
        <MantineReactTable table={table} />
      )}

      {/* Add Job Modal */}
      <Modal
        opened={addJobModalOpen}
        onClose={() => {
          setAddJobModalOpen(false);
          setNewJobBookFilter("");
          setNewJobCronExpr("");
        }}
        title="Add Custom Metrics Job"
        size="md"
      >
        <Stack gap="md">
          <Select
            label="Book Filter"
            placeholder="Select a book"
            data={bookOptions}
            value={newJobBookFilter}
            onChange={(value) => setNewJobBookFilter(value || "")}
            required
            disabled={tradesLoading}
            searchable
          />

          <TextInput
            label="Cron Schedule"
            placeholder="e.g., 0 9 * * 1-5"
            description="Leave empty to use the default schedule for collecting overall portfolio metrics"
            value={newJobCronExpr}
            onChange={(event) => setNewJobCronExpr(event.currentTarget.value)}
          />

          <Group justify="flex-end">
            <Button
              variant="default"
              onClick={() => {
                setAddJobModalOpen(false);
                setNewJobBookFilter("");
                setNewJobCronExpr("");
              }}
            >
              Cancel
            </Button>
            <Button onClick={handleCreateJob} disabled={!newJobBookFilter}>
              Create Job
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        opened={deleteModalOpen}
        onClose={() => {
          setDeleteModalOpen(false);
          setSelectedJob(null);
        }}
        title="Delete Custom Job"
        size="sm"
      >
        <Stack gap="md">
          <Text>
            Are you sure you want to delete the custom metrics job for book "
            {selectedJob?.BookFilter}"? This action cannot be undone.
          </Text>

          <Group justify="flex-end">
            <Button
              variant="default"
              onClick={() => {
                setDeleteModalOpen(false);
                setSelectedJob(null);
              }}
            >
              Cancel
            </Button>
            <Button color="red" onClick={handleDeleteJob}>
              Delete
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
};

export default CustomJobsTable;
