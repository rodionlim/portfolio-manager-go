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
import { Trade } from "../../types/blotter";

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
  Purchases: number;
  Sales: number;
  Net: number;
  CumulativeNet: number;
  DividendYield: number;
}

const DividendsAggregatedTable: React.FC = () => {
  const refData = useSelector((state: RootState) => state.referenceData.data);

  // Fetch all dividends and trades for all tickers
  const fetchAllDividendsAndTrades = async (): Promise<{
    dividends: Record<string, Dividend[]>;
    trades: Trade[];
  }> => {
    try {
      const [dividendsResp, tradesResp] = await Promise.all([
        fetch(getUrl("/api/v1/dividends")),
        fetch(getUrl("/api/v1/blotter/trade")),
      ]);

      const dividends = await dividendsResp.json();
      const trades = await tradesResp.json();

      return { dividends, trades };
    } catch (error: any) {
      console.error("Error fetching all dividends and trades:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch data: ${error.message}`,
      });
      return { dividends: {}, trades: [] };
    }
  };

  // Query to fetch all dividends and trades
  const {
    data: allDividendsAndTrades = { dividends: {}, trades: [] },
    isLoading,
    error,
  } = useQuery({
    queryKey: ["allDividendsAndTrades"],
    queryFn: fetchAllDividendsAndTrades,
  });

  // Aggregate dividends by year
  const aggregatedData = useMemo(() => {
    const yearlyData: Record<number, YearlyDividends> = {};

    if (!refData) return [];

    // Process all tickers and their dividends
    Object.entries(allDividendsAndTrades.dividends).forEach(
      ([ticker, dividends]) => {
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
              Purchases: 0,
              Sales: 0,
              Net: 0,
              CumulativeNet: 0,
              DividendYield: 0,
            };
          }

          // TODO: handle non-sgd dividends

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
      }
    );

    // Process all trades for purchases and sales by year
    allDividendsAndTrades.trades?.forEach((trade) => {
      const date = new Date(trade.TradeDate);
      const year = date.getFullYear();

      // TODO: handle non sgd trades
      const tradeValue = trade.Quantity * trade.Price * trade.Fx;

      // Initialize year record if it doesn't exist
      if (!yearlyData[year]) {
        yearlyData[year] = {
          Year: year,
          Dividends: 0,
          DividendsSSB: 0,
          DividendsTBill: 0,
          DividendsEquity: 0,
          Purchases: 0,
          Sales: 0,
          Net: 0,
          CumulativeNet: 0,
          DividendYield: 0,
        };
      }

      // Update purchases or sales based on trade side
      if (trade.Side.toLowerCase() === "buy") {
        yearlyData[year].Purchases += tradeValue;
      } else if (trade.Side.toLowerCase() === "sell") {
        yearlyData[year].Sales += tradeValue;
      }
    });

    // Calculate Net and CumulativeNet for each year
    let cumulativeNet = 0;
    const sortedYears = Object.keys(yearlyData)
      .map(Number)
      .sort((a, b) => a - b);

    sortedYears.forEach((year) => {
      const yearData = yearlyData[year];
      yearData.Net = yearData.Purchases - yearData.Sales;
      cumulativeNet += yearData.Net;
      yearData.CumulativeNet = cumulativeNet;

      // Calculate Dividend Yield (avoid division by zero)
      if (yearData.CumulativeNet > 0) {
        yearData.DividendYield =
          (yearData.Dividends / yearData.CumulativeNet) * 100;
      } else {
        yearData.DividendYield = 0;
      }
    });

    // Convert to array and sort by year descending
    return Object.values(yearlyData).sort((a, b) => b.Year - a.Year);
  }, [allDividendsAndTrades, refData]);

  // Calculate totals for all years
  const totals = useMemo(() => {
    return aggregatedData.reduce(
      (acc, curr) => {
        acc.Dividends += curr.Dividends;
        acc.DividendsSSB += curr.DividendsSSB;
        acc.DividendsTBill += curr.DividendsTBill;
        acc.DividendsEquity += curr.DividendsEquity;
        acc.Purchases += curr.Purchases;
        acc.Sales += curr.Sales;
        acc.Net += curr.Net;
        // CumulativeNet is not summed as it's already the final value
        // Use the most recent year's CumulativeNet as the total
        if (aggregatedData.length > 0) {
          acc.CumulativeNet = aggregatedData[0].CumulativeNet;
        }
        // Calculate overall dividend yield
        if (acc.CumulativeNet > 0) {
          acc.DividendYield = (acc.Dividends / acc.CumulativeNet) * 100;
        }
        return acc;
      },
      {
        Year: 0,
        Dividends: 0,
        DividendsSSB: 0,
        DividendsTBill: 0,
        DividendsEquity: 0,
        Purchases: 0,
        Sales: 0,
        Net: 0,
        CumulativeNet: 0,
        DividendYield: 0,
      }
    );
  }, [aggregatedData]);

  // Helper function to format numbers with thousands separators
  const formatNumber = (num: number): string => {
    return num.toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  };

  // Define table columns
  const columns = useMemo<MRT_ColumnDef<YearlyDividends>[]>(
    () => [
      {
        accessorKey: "Year",
        header: "Year",
        Footer: () => <strong>Total</strong>,
      },
      {
        accessorKey: "DividendYield",
        header: "Dividend Yield (%)",
        Cell: ({ cell }) => {
          return `${formatNumber(cell.getValue<number>())}%`;
        },
        Footer: () => <strong>${formatNumber(totals.DividendYield)}%</strong>,
      },
      {
        accessorKey: "Dividends",
        header: "Dividends",
        Cell: ({ cell }) => {
          return (
            <span style={{ color: "#2E8B57" }}>
              ${formatNumber(cell.getValue<number>())}
            </span>
          );
        },
        Footer: () => (
          <strong style={{ color: "#2E8B57" }}>
            ${formatNumber(totals.Dividends)}
          </strong>
        ),
      },
      {
        accessorKey: "DividendsSSB",
        header: "Dividends SSB",
        Cell: ({ cell }) => {
          return `$${formatNumber(cell.getValue<number>())}`;
        },
        Footer: () => <strong>${formatNumber(totals.DividendsSSB)}</strong>,
      },
      {
        accessorKey: "DividendsTBill",
        header: "Dividends MAS Bills",
        Cell: ({ cell }) => {
          return `$${formatNumber(cell.getValue<number>())}`;
        },
        Footer: () => <strong>${formatNumber(totals.DividendsTBill)}</strong>,
      },
      {
        accessorKey: "DividendsEquity",
        header: "Dividends Equity",
        Cell: ({ cell }) => {
          return `$${formatNumber(cell.getValue<number>())}`;
        },
        Footer: () => <strong>${formatNumber(totals.DividendsEquity)}</strong>,
      },
      {
        accessorKey: "Purchases",
        header: "Purchases",
        Cell: ({ cell }) => {
          return (
            <span style={{ color: "#2E8B57" }}>
              ${formatNumber(cell.getValue<number>())}
            </span>
          );
        },
        Footer: () => (
          <strong style={{ color: "#2E8B57" }}>
            ${formatNumber(totals.Purchases)}
          </strong>
        ),
      },
      {
        accessorKey: "Sales",
        header: "Sales",
        Cell: ({ cell }) => {
          return (
            <span style={{ color: "#DC143C" }}>
              ${formatNumber(cell.getValue<number>())}
            </span>
          );
        },
        Footer: () => (
          <strong style={{ color: "#DC143C" }}>
            ${formatNumber(totals.Sales)}
          </strong>
        ),
      },
      {
        accessorKey: "Net",
        header: "Net",
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          const color =
            value < 0 ? "#DC143C" : value > 0 ? "#2E8B57" : "inherit";
          return <span style={{ color }}>${formatNumber(value)}</span>;
        },
        Footer: () => {
          const color =
            totals.Net < 0 ? "#DC143C" : totals.Net > 0 ? "#2E8B57" : "inherit";
          return <strong style={{ color }}>${formatNumber(totals.Net)}</strong>;
        },
      },
      {
        accessorKey: "CumulativeNet",
        header: "Cumulative Net",
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          const color =
            value < 0 ? "#DC143C" : value > 0 ? "#2E8B57" : "inherit";
          return <span style={{ color }}>${formatNumber(value)}</span>;
        },
        Footer: () => {
          const color =
            totals.CumulativeNet < 0
              ? "#DC143C"
              : totals.CumulativeNet > 0
              ? "#2E8B57"
              : "inherit";
          return (
            <strong style={{ color }}>
              ${formatNumber(totals.CumulativeNet)}
            </strong>
          );
        },
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
      {Object.keys(allDividendsAndTrades.dividends).length === 0 &&
        !isLoading &&
        !error && (
          <Text c="dimmed" ta="center" mt="md">
            No dividend records found
          </Text>
        )}
    </div>
  );
};

export default DividendsAggregatedTable;
