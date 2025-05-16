import React, { useMemo, useState, useEffect } from "react";
import { Box, FileInput, Select, Text } from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import { IconUpload } from "@tabler/icons-react";

interface Position {
  Ticker: string;
}

interface Dividend {
  ExDate: string;
  Amount: number;
  AmountPerShare: number;
  Qty: number;
}

interface DividendsTableProps {
  initialTicker?: string | null;
}

const DividendsTable: React.FC<DividendsTableProps> = ({
  initialTicker = null,
}) => {
  const [selectedTicker, setSelectedTicker] = useState<string | null>(
    initialTicker
  );

  // Effect to update selectedTicker when initialTicker changes (e.g., from navigation)
  useEffect(() => {
    if (initialTicker) {
      setSelectedTicker(initialTicker);
    }
  }, [initialTicker]);

  // Fetch all positions to populate the dropdown of tickers
  const fetchPositions = async (): Promise<Position[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions"));
      const uniqueTickers: Position[] = await resp.json();
      return uniqueTickers;
    } catch (error: any) {
      console.error("Error fetching positions:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch positions: ${error.message}`,
      });
      return [];
    }
  };

  // Fetch dividends for a specific ticker
  const fetchDividends = async (ticker: string | null): Promise<Dividend[]> => {
    if (!ticker) return [];

    try {
      const resp = await fetch(getUrl(`/api/v1/dividends/${ticker}`));
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching dividends:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch dividends for ${ticker}: ${error.message}`,
      });
      return [];
    }
  };

  const uploadDividendsCSV = async (
    file: File
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
        }
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
      },
      (error) => {
        notifications.show({
          color: "red",
          title: "Error",
          message: `Unable to upload custom dividends to the market data service\n ${error}`,
        });
      }
    );
  };

  // Query to fetch positions for dropdown
  const {
    data: positions = [],
    isLoading: isLoadingPositions,
    error: positionsError,
  } = useQuery({
    queryKey: ["positions"],
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
    return dividends.reduce((sum, div) => sum + div.Amount, 0);
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
        accessorKey: "AmountPerShare",
        header: "Amount Per Share",
        Cell: ({ cell }) => {
          return cell.getValue<number>().toFixed(4);
        },
      },
      {
        accessorKey: "Amount",
        header: "Total Amount (After withholding tax)",
        Cell: ({ cell }) => {
          return cell.getValue<number>().toFixed(2);
        },
      },
    ],
    []
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
          data={positions.map((position) => ({
            label: position.Ticker,
            value: position.Ticker,
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

  // Handle loading states
  if (isLoadingPositions) return <div>Loading positions...</div>;
  if (positionsError) return <div>Error loading positions</div>;

  // Render the component
  return (
    <div>
      <MantineReactTable table={table} />
      {dividendsError && (
        <Text c="red" ta="center" mt="md">
          Error loading dividends
        </Text>
      )}
      {selectedTicker &&
        dividends.length === 0 &&
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
