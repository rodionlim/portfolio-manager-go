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
import { useSelector } from "react-redux";
import { RootState } from "../../store";

interface Dividend {
  ExDate: string;
  Amount: number;
  AmountPerShare: number;
  Qty: number;
}

interface YearlyDividends {
  Year: number;
  Dividends: number;
  DividendsSSB: number;
  DividendsTBill: number;
  DividendsEquity: number;
}

const DividendsAggregatedTable: React.FC = () => {
  const refData = useSelector((state: RootState) => state.referenceData.data);
  console.log(refData);

  // Fetch all dividends for all tickers
  const fetchAllDividends = async (): Promise<Record<string, Dividend[]>> => {
    try {
      const resp = await fetch(getUrl("/api/v1/dividends"));
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching all dividends:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch all dividends: ${error.message}`,
      });
      return {};
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
  });

  // Aggregate dividends by year
  const aggregatedData = useMemo(() => {
    const yearlyData: Record<number, YearlyDividends> = {};

    if (!refData) return [];

    // Process all tickers and their dividends
    Object.entries(allDividends).forEach(([ticker, dividends]) => {
      dividends?.forEach((dividend) => {
        const date = new Date(dividend.ExDate);
        const year = date.getFullYear();

        // Initialize year record if it doesn't exist
        if (!yearlyData[year]) {
          yearlyData[year] = {
            Year: year,
            Dividends: 0,
            DividendsSSB: 0,
            DividendsTBill: 0,
            DividendsEquity: 0,
          };
        }

        // Add to total dividends
        yearlyData[year].Dividends += dividend.Amount;

        const tickerRef = refData[ticker];
        const isSgGovies =
          tickerRef?.ccy === "SGD" &&
          tickerRef?.asset_sub_class === "govies" &&
          tickerRef?.asset_class === "bond";
        let isSSB = false;
        let isSgTBill = false;

        if (isSgGovies) {
          if (ticker.startsWith("SB")) {
            isSSB = true;
          } else {
            // mas bill, hopefully assumption of isSgGovies stay true
            isSgTBill = true;
          }
        }

        // Categorize by asset type
        if (isSSB) {
          yearlyData[year].DividendsSSB += dividend.Amount;
        } else if (isSgTBill) {
          yearlyData[year].DividendsTBill += dividend.Amount;
        } else {
          yearlyData[year].DividendsEquity += dividend.Amount;
        }
      });
    });

    // Convert to array and sort by year descending
    return Object.values(yearlyData).sort((a, b) => b.Year - a.Year);
  }, [allDividends, refData]);

  // Calculate totals for all years
  const totals = useMemo(() => {
    return aggregatedData.reduce(
      (acc, curr) => {
        acc.Dividends += curr.Dividends;
        acc.DividendsSSB += curr.DividendsSSB;
        acc.DividendsTBill += curr.DividendsTBill;
        acc.DividendsEquity += curr.DividendsEquity;
        return acc;
      },
      {
        Year: 0,
        Dividends: 0,
        DividendsSSB: 0,
        DividendsTBill: 0,
        DividendsEquity: 0,
      }
    );
  }, [aggregatedData, refData]);

  // Define table columns
  const columns = useMemo<MRT_ColumnDef<YearlyDividends>[]>(
    () => [
      {
        accessorKey: "Year",
        header: "Year",
        Footer: () => <strong>Total</strong>,
      },
      {
        accessorKey: "Dividends",
        header: "Dividends",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toFixed(2)}`;
        },
        Footer: () => <strong>${totals.Dividends.toFixed(2)}</strong>,
      },
      {
        accessorKey: "DividendsSSB",
        header: "Dividends SSB",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toFixed(2)}`;
        },
        Footer: () => <strong>${totals.DividendsSSB.toFixed(2)}</strong>,
      },
      {
        accessorKey: "DividendsTBill",
        header: "Dividends T-Bill",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toFixed(2)}`;
        },
        Footer: () => <strong>${totals.DividendsTBill.toFixed(2)}</strong>,
      },
      {
        accessorKey: "DividendsEquity",
        header: "Dividends Equity",
        Cell: ({ cell }) => {
          return `$${cell.getValue<number>().toFixed(2)}`;
        },
        Footer: () => <strong>${totals.DividendsEquity.toFixed(2)}</strong>,
      },
    ],
    [totals]
  );

  // Configure the table
  const table = useMantineReactTable({
    columns,
    data: aggregatedData,
    state: { density: "xs" },
    enablePagination: false,
    enableColumnFilters: false,
    enableGlobalFilter: false,
    enableRowSelection: false,
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
          Yearly Dividend Summary
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
      {Object.keys(allDividends).length === 0 && !isLoading && !error && (
        <Text c="dimmed" ta="center" mt="md">
          No dividend records found
        </Text>
      )}
    </div>
  );
};

export default DividendsAggregatedTable;
