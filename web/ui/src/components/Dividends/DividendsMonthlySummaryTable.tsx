import React, { useEffect, useMemo, useState } from "react";
import { Box, Button, Select, Text } from "@mantine/core";
import { IconColumns, IconTable } from "@tabler/icons-react";
import {
  MantineReactTable,
  MRT_Cell,
  MRT_ColumnDef,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { RootState } from "../../store";
import { Trade } from "../../types/blotter";
import { getUrl } from "../../utils/url";
import {
  aggregateMonthlyDividends,
  Dividend,
  FxRates,
  getYearsFromMonthlyDividends,
  MonthlyDividends,
  pivotMonthlyDividendsByYear,
  PivotedMonthlyDividends,
} from "./dividendsSummary";

type MonthlySummaryRow = MonthlyDividends | PivotedMonthlyDividends;

const DividendsMonthlySummaryTable: React.FC = () => {
  const refData = useSelector((state: RootState) => state.referenceData.data);
  const navigate = useNavigate();
  const currentYear = new Date().getFullYear();
  const [firstYear, setFirstYear] = useState<string>(String(currentYear));
  const [secondYear, setSecondYear] = useState<string>(
    String(currentYear - 1)
  );
  const [showAllColumns, setShowAllColumns] = useState(false);
  const [isPivoted, setIsPivoted] = useState(false);

  const fetchAllDividendsAndTradesAndFx = async (): Promise<{
    dividends: Record<string, Dividend[]>;
    trades: Trade[];
    fx: FxRates;
  }> => {
    try {
      const [dividendsResp, tradesResp, fxResp] = await Promise.all([
        fetch(getUrl("/api/v1/dividends")),
        fetch(getUrl("/api/v1/blotter/trade")),
        fetch(getUrl("/api/v1/blotter/fx")),
      ]);

      if (!dividendsResp.ok) {
        const errorData = await dividendsResp.json();
        throw new Error(errorData?.message || "Failed to fetch dividends");
      }

      const trades = await tradesResp.json();
      const fx = await fxResp.json();
      const dividends = await dividendsResp.json();

      return { dividends, trades, fx };
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : "Unknown fetch error";
      console.error("Error fetching monthly dividend summary data:", error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to fetch data: ${message}`,
        autoClose: 15000,
      });
      return { dividends: {}, trades: [], fx: {} };
    }
  };

  const {
    data: allDividendsAndTradesAndFx = { dividends: {}, trades: [], fx: {} },
    isLoading,
    error,
  } = useQuery({
    queryKey: ["monthlyDividendsAndTradesAndFx"],
    queryFn: fetchAllDividendsAndTradesAndFx,
  });

  const yearOptions = useMemo(() => {
    const years = new Set<number>([currentYear, currentYear - 1]);

    Object.values(allDividendsAndTradesAndFx.dividends).forEach(
      (tickerDividends) => {
        tickerDividends?.forEach((dividend) => {
          const year = Number(dividend.ExDate.slice(0, 4));
          if (Number.isInteger(year)) {
            years.add(year);
          }
        });
      }
    );

    allDividendsAndTradesAndFx.trades?.forEach((trade) => {
      const year = Number(trade.TradeDate.slice(0, 4));
      if (Number.isInteger(year)) {
        years.add(year);
      }
    });

    return [...years]
      .sort((a, b) => b - a)
      .map((year) => ({ value: String(year), label: String(year) }));
  }, [allDividendsAndTradesAndFx, currentYear]);

  useEffect(() => {
    if (firstYear === secondYear) {
      const fallbackYear = yearOptions.find(
        (option) => option.value !== firstYear
      );
      if (fallbackYear) {
        setSecondYear(fallbackYear.value);
      }
    }
  }, [firstYear, secondYear, yearOptions]);

  const monthlyData = useMemo(() => {
    if (!refData) return [];
    const selectedYears = [Number(firstYear), Number(secondYear)].filter(
      (year) => Number.isInteger(year)
    );

    return aggregateMonthlyDividends({
      dividends: allDividendsAndTradesAndFx.dividends,
      trades: allDividendsAndTradesAndFx.trades,
      fx: allDividendsAndTradesAndFx.fx,
      refData,
      years: selectedYears,
    });
  }, [allDividendsAndTradesAndFx, firstYear, refData, secondYear]);

  const totals = useMemo(() => {
    return monthlyData.reduce(
      (acc, curr) => {
        acc.Dividends += curr.Dividends;
        acc.DividendsSSB += curr.DividendsSSB;
        acc.DividendsTBill += curr.DividendsTBill;
        acc.DividendsEquity += curr.DividendsEquity;
        acc.Purchases += curr.Purchases;
        acc.Sales += curr.Sales;
        acc.Net += curr.Net;
        acc.PurchasesExclGov += curr.PurchasesExclGov;
        acc.SalesExclGov += curr.SalesExclGov;
        acc.NetExclGov += curr.NetExclGov;

        if (monthlyData.length > 0) {
          acc.CumulativeNet = monthlyData[0].CumulativeNet;
          acc.CumulativeNetExclGov = monthlyData[0].CumulativeNetExclGov;
        }

        acc.DividendYield =
          acc.CumulativeNet > 0
            ? (acc.Dividends / acc.CumulativeNet) * 100
            : 0;
        acc.DividendYieldExclGov =
          acc.CumulativeNetExclGov > 0
            ? (acc.DividendsEquity / acc.CumulativeNetExclGov) * 100
            : 0;

        return acc;
      },
      {
        Month: "",
        MonthLabel: "",
        Dividends: 0,
        DividendsSSB: 0,
        DividendsTBill: 0,
        DividendsEquity: 0,
        Purchases: 0,
        Sales: 0,
        Net: 0,
        CumulativeNet: 0,
        DividendYield: 0,
        PurchasesExclGov: 0,
        SalesExclGov: 0,
        NetExclGov: 0,
        CumulativeNetExclGov: 0,
        DividendYieldExclGov: 0,
      }
    );
  }, [monthlyData]);

  const pivotedData = useMemo(
    () => pivotMonthlyDividendsByYear(monthlyData),
    [monthlyData]
  );

  const pivotedYears = useMemo(
    () => getYearsFromMonthlyDividends(monthlyData),
    [monthlyData]
  );

  const formatNumber = (num: number): string => {
    return num.toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  };

  const currencyCell = (value: number, color?: string) => {
    return <span style={{ color }}>${formatNumber(value)}</span>;
  };

  const netColor = (value: number) => {
    if (value < 0) return "#DC143C";
    if (value > 0) return "#2E8B57";
    return "inherit";
  };

  const drillDownToMonth = (monthKey: string) => {
    navigate("/dividends", {
      state: {
        activeTab: "all",
        dividendMonth: monthKey,
      },
    });
  };

  const standardColumns = useMemo<MRT_ColumnDef<MonthlySummaryRow>[]>(
    () => [
      {
        accessorKey: "MonthLabel",
        header: "Month",
        Cell: ({ row }) => {
          const month = (row.original as MonthlyDividends).Month;
          return (
            <Button
              onClick={() => drillDownToMonth(month)}
              variant="subtle"
              size="xs"
              px={4}
            >
              {(row.original as MonthlyDividends).MonthLabel}
            </Button>
          );
        },
        Footer: () => <strong>Total</strong>,
      },
      {
        accessorKey: "DividendYield",
        header: "Dividend Yield (%)",
        Cell: ({ cell }) => `${formatNumber(cell.getValue<number>())}%`,
      },
      {
        accessorKey: "DividendYieldExclGov",
        header: "Dividend Yield Ex. Gov (%)",
        Cell: ({ cell }) => `${formatNumber(cell.getValue<number>())}%`,
      },
      {
        accessorKey: "Dividends",
        header: "Dividends",
        Cell: ({ cell }) =>
          currencyCell(cell.getValue<number>(), "#2E8B57"),
        Footer: () => (
          <strong style={{ color: "#2E8B57" }}>
            ${formatNumber(totals.Dividends)}
          </strong>
        ),
      },
      {
        accessorKey: "DividendsSSB",
        header: "Dividends SSB",
        Cell: ({ cell }) => currencyCell(cell.getValue<number>()),
        Footer: () => <strong>${formatNumber(totals.DividendsSSB)}</strong>,
      },
      {
        accessorKey: "DividendsTBill",
        header: "Dividends MAS Bills",
        Cell: ({ cell }) => currencyCell(cell.getValue<number>()),
        Footer: () => <strong>${formatNumber(totals.DividendsTBill)}</strong>,
      },
      {
        accessorKey: "DividendsEquity",
        header: "Dividends Equity",
        Cell: ({ cell }) => currencyCell(cell.getValue<number>()),
        Footer: () => <strong>${formatNumber(totals.DividendsEquity)}</strong>,
      },
      {
        accessorKey: "Purchases",
        header: "Purchases",
        Cell: ({ cell }) =>
          currencyCell(cell.getValue<number>(), "#2E8B57"),
        Footer: () => (
          <strong style={{ color: "#2E8B57" }}>
            ${formatNumber(totals.Purchases)}
          </strong>
        ),
      },
      {
        accessorKey: "Sales",
        header: "Sales",
        Cell: ({ cell }) =>
          currencyCell(cell.getValue<number>(), "#DC143C"),
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
          return currencyCell(value, netColor(value));
        },
        Footer: () => (
          <strong style={{ color: netColor(totals.Net) }}>
            ${formatNumber(totals.Net)}
          </strong>
        ),
      },
      {
        accessorKey: "CumulativeNet",
        header: "Cumulative Net",
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return currencyCell(value, netColor(value));
        },
        Footer: () => (
          <strong style={{ color: netColor(totals.CumulativeNet) }}>
            ${formatNumber(totals.CumulativeNet)}
          </strong>
        ),
      },
      {
        accessorKey: "PurchasesExclGov",
        header: "PurchasesExclGov",
        Cell: ({ cell }) =>
          currencyCell(cell.getValue<number>(), "#2E8B57"),
        Footer: () => (
          <strong style={{ color: "#2E8B57" }}>
            ${formatNumber(totals.PurchasesExclGov)}
          </strong>
        ),
      },
      {
        accessorKey: "SalesExclGov",
        header: "SalesExclGov",
        Cell: ({ cell }) =>
          currencyCell(cell.getValue<number>(), "#DC143C"),
        Footer: () => (
          <strong style={{ color: "#DC143C" }}>
            ${formatNumber(totals.SalesExclGov)}
          </strong>
        ),
      },
      {
        accessorKey: "NetExclGov",
        header: "NetExclGov",
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return currencyCell(value, netColor(value));
        },
        Footer: () => (
          <strong style={{ color: netColor(totals.NetExclGov) }}>
            ${formatNumber(totals.NetExclGov)}
          </strong>
        ),
      },
      {
        accessorKey: "CumulativeNetExclGov",
        header: "Cumulative Net ExclGov",
        Cell: ({ cell }) => {
          const value = cell.getValue<number>();
          return currencyCell(value, netColor(value));
        },
        Footer: () => (
          <strong style={{ color: netColor(totals.CumulativeNetExclGov) }}>
            ${formatNumber(totals.CumulativeNetExclGov)}
          </strong>
        ),
      },
    ],
    [totals]
  );

  const pivotedColumns = useMemo<MRT_ColumnDef<MonthlySummaryRow>[]>(
    () => {
      const yearColumns: MRT_ColumnDef<MonthlySummaryRow>[] =
        pivotedYears.flatMap((year): MRT_ColumnDef<MonthlySummaryRow>[] => [
          {
            accessorKey: `${year}DividendYield`,
            header: `${year} Dividend Yield (%)`,
            Cell: ({
              cell,
            }: {
              cell: MRT_Cell<MonthlySummaryRow, number | undefined>;
            }) => {
              const value = cell.getValue();
              if (value === undefined) return "";

              const monthIndex = (cell.row.original as PivotedMonthlyDividends)
                .MonthIndex;
              return (
                <Button
                  onClick={() =>
                    drillDownToMonth(
                      `${year}-${String(monthIndex).padStart(2, "0")}`
                    )
                  }
                  variant="subtle"
                  size="xs"
                  px={4}
                >
                  {formatNumber(value)}%
                </Button>
              );
            },
          },
          {
            accessorKey: `${year}Dividends`,
            header: `${year} Dividends`,
            Cell: ({
              cell,
            }: {
              cell: MRT_Cell<MonthlySummaryRow, number | undefined>;
            }) => {
              const value = cell.getValue();
              if (value === undefined) return "";

              const monthIndex = (cell.row.original as PivotedMonthlyDividends)
                .MonthIndex;
              return (
                <Button
                  onClick={() =>
                    drillDownToMonth(
                      `${year}-${String(monthIndex).padStart(2, "0")}`
                    )
                  }
                  variant="subtle"
                  size="xs"
                  px={4}
                  c="green"
                >
                  ${formatNumber(value)}
                </Button>
              );
            },
          },
        ]);

      return [
        {
        accessorKey: "Month",
        header: "Month",
        },
        ...yearColumns,
      ];
    },
    [pivotedYears]
  );

  const columnVisibility = useMemo(
    () => ({
      DividendYieldExclGov: showAllColumns,
      Purchases: showAllColumns,
      Sales: showAllColumns,
      Net: showAllColumns,
      CumulativeNet: showAllColumns,
      PurchasesExclGov: showAllColumns,
      SalesExclGov: showAllColumns,
      NetExclGov: showAllColumns,
      CumulativeNetExclGov: showAllColumns,
    }),
    [showAllColumns]
  );

  const table = useMantineReactTable({
    columns: isPivoted ? pivotedColumns : standardColumns,
    data: isPivoted ? pivotedData : monthlyData,
    state: {
      columnVisibility: isPivoted ? {} : columnVisibility,
      density: "xs",
      isLoading,
      showLoadingOverlay: isLoading,
    },
    enablePagination: false,
    enableColumnFilters: false,
    enableGlobalFilter: false,
    enableRowSelection: false,
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: () => (
      <Box
        style={{
          overflowX: "auto",
          width: "100%",
        }}
      >
        <Box
          style={{
            display: "flex",
            gap: "12px",
            padding: "4px",
            alignItems: "center",
            flexWrap: "nowrap",
            minWidth: "max-content",
          }}
        >
          <Text fw={700} size="lg" style={{ whiteSpace: "nowrap" }}>
            Monthly Dividend Summary
          </Text>
          <Text size="sm" c="dimmed" style={{ whiteSpace: "nowrap" }}>
            Compare two years
          </Text>
          <Select
            aria-label="First comparison year"
            data={yearOptions}
            value={firstYear}
            onChange={(value) => {
              if (value) {
                setFirstYear(value);
              }
            }}
            allowDeselect={false}
            size="xs"
            w={110}
          />
          <Select
            aria-label="Second comparison year"
            data={yearOptions}
            value={secondYear}
            onChange={(value) => {
              if (value) {
                setSecondYear(value);
              }
            }}
            allowDeselect={false}
            size="xs"
            w={110}
          />
          <Button
            leftSection={<IconColumns size={14} />}
            onClick={() => setShowAllColumns((value) => !value)}
            variant="light"
            size="xs"
            disabled={isPivoted}
            style={{ whiteSpace: "nowrap" }}
          >
            {showAllColumns ? "Hide columns" : "Show all columns"}
          </Button>
          <Button
            leftSection={<IconTable size={14} />}
            onClick={() => setIsPivoted((value) => !value)}
            variant={isPivoted ? "filled" : "light"}
            size="xs"
            style={{ whiteSpace: "nowrap" }}
          >
            {isPivoted ? "Table" : "Pivot"}
          </Button>
        </Box>
      </Box>
    ),
  });

  if (isLoading) return <div>Fetching dividend data...</div>;
  if (error) return <div>Error fetching dividend data</div>;

  return (
    <div>
      <MantineReactTable table={table} />
      {Object.keys(allDividendsAndTradesAndFx.dividends).length === 0 &&
        !isLoading &&
        !error && (
          <Text c="dimmed" ta="center" mt="md">
            No dividend records found
          </Text>
        )}

      <Text
        size="xs"
        c="dimmed"
        ta="left"
        mt="sm"
        style={{ fontStyle: "italic" }}
      >
        Note: Non-SGD dividends are revalued at current FX rates instead of
        historical rates, matching the yearly summary behavior.
      </Text>
    </div>
  );
};

export default DividendsMonthlySummaryTable;
