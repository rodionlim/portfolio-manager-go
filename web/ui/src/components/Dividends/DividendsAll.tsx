import React, { useMemo } from "react";
import { Box, Text } from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";

import classes from "../../styles.module.css";

interface Dividend {
  ExDate: string;
  Amount: number;
  AmountPerShare: number;
  Qty: number;
}

interface DividendWithTicker extends Dividend {
  Ticker: string;
}

const DividendsAll: React.FC = () => {
  // Fetch all dividends for all tickers
  const fetchAllDividends = async (): Promise<Record<string, Dividend[]>> => {
    try {
      const resp = await fetch(getUrl("/api/v1/dividends"));
      if (!resp.ok) {
        const errorData = await resp.json();
        throw new Error(errorData?.message || "Failed to fetch dividends");
      }
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching all dividends:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch dividends: ${error.message}`,
      });
      throw error;
    }
  };

  // Query to fetch all dividends
  const {
    data: allDividends = {},
    isLoading,
    error,
  } = useQuery({
    queryKey: ["allDividends"],
    queryFn: fetchAllDividends,
    refetchOnWindowFocus: false,
    retry: false,
  });

  // Flatten and sort dividends by ExDate in descending order
  const flattenedDividends = useMemo(() => {
    const dividends: DividendWithTicker[] = [];

    Object.entries(allDividends).forEach(([ticker, tickerDividends]) => {
      tickerDividends?.forEach((dividend) => {
        dividends.push({
          ...dividend,
          Ticker: ticker,
        });
      });
    });

    // Sort by ExDate in descending order (most recent first)
    return dividends.sort((a, b) => {
      return new Date(b.ExDate).getTime() - new Date(a.ExDate).getTime();
    });
  }, [allDividends]);

  // Helper function to format numbers with thousands separators
  const formatNumber = (num: number): string => {
    return num.toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  };

  // Helper function to format date
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  // Define table columns
  const columns = useMemo<MRT_ColumnDef<DividendWithTicker>[]>(
    () => [
      {
        accessorKey: "ExDate",
        header: "Ex-Date",
        Cell: ({ cell }) => formatDate(cell.getValue<string>()),
        sortingFn: (rowA, rowB, columnId) => {
          const dateA = new Date(rowA.getValue(columnId));
          const dateB = new Date(rowB.getValue(columnId));

          // Handle invalid dates by putting them at the end
          const timeA = isNaN(dateA.getTime()) ? 0 : dateA.getTime();
          const timeB = isNaN(dateB.getTime()) ? 0 : dateB.getTime();

          return timeA - timeB;
        },
      },
      {
        accessorKey: "Ticker",
        header: "Ticker",
        Cell: ({ cell }) => (
          <Text className={classes["default-font-size"]} fw={500} c="blue">
            {cell.getValue<string>()}
          </Text>
        ),
      },
      {
        accessorKey: "Qty",
        header: "Quantity",
        Cell: ({ cell }) => formatNumber(cell.getValue<number>()),
      },
      {
        accessorKey: "AmountPerShare",
        header: "Amount Per Share",
        Cell: ({ cell }) => `$${formatNumber(cell.getValue<number>())}`,
      },
      {
        accessorKey: "Amount",
        header: "Total Amount",
        Cell: ({ cell }) => (
          <Text className={classes["default-font-size"]} fw={500} c="green">
            ${formatNumber(cell.getValue<number>())}
          </Text>
        ),
      },
    ],
    []
  );

  // Configure the table
  const table = useMantineReactTable({
    columns,
    data: flattenedDividends,
    state: {
      density: "xs",
      isLoading: isLoading,
      showLoadingOverlay: isLoading,
    },
    enablePagination: true,
    enableColumnFilters: true,
    enableGlobalFilter: true,
    enableRowSelection: false,
    enableSorting: true,
    initialState: {
      pagination: {
        pageSize: 20,
        pageIndex: 0,
      },
    },
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
        <Text fw={700} size="lg">
          All Dividends Received
        </Text>
        <Text size="sm" c="dimmed">
          ({flattenedDividends.length} records)
        </Text>
      </Box>
    ),
  });

  // Handle loading states
  if (isLoading) return <div>Fetching dividend data...</div>;
  if (error) return <div>Error fetching dividend data</div>;

  // Render the component
  return (
    <div>
      <MantineReactTable table={table} />
      {flattenedDividends.length === 0 && !isLoading && !error && (
        <Text c="dimmed" ta="center" mt="md">
          No dividend records found
        </Text>
      )}
    </div>
  );
};

export default DividendsAll;
