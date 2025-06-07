// create a table of the YTDInstitutionNetBuySellSGDM as the values #file:interfaces.go and stock name as the columns, sort the columns by showing those with largest YTDInstitutionNetBuySellSGDM on the left, showing 100 columns can be abit overwhelming, for a start, give me a user dropdown where they can select either top 10, 20, 50, 100, and default it to 30.
// if the table supports colour coding, then use a darker green for more positive values and a light green for less positive values, vary the greeness by the value, and vice versa for negative values, use a darker red for more negative values and a light red for less negative values, vary the redness by the value, if the value is 0, then use a grey colour.

import React, { useState, useEffect } from "react";
import {
  Stack,
  Select,
  Text,
  Table,
  ScrollArea,
  Alert,
  Loader,
  Group,
  Badge,
  Box,
  HoverCard,
} from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import { getUrl } from "../../utils/url";

interface MostTradedStock {
  stockName: string;
  stockCode: string;
  ytdAvgDailyTurnoverSGDM: number;
  ytdInstitutionNetBuySellSGDM: number;
  past5SessionsInstitutionNetSGDM: number;
  sector: string;
  institutionNetBuySellChange?: number;
}

interface MostTradedStocksReport {
  reportDate: string;
  reportTitle: string;
  filePath: string;
  stocks: MostTradedStock[];
  extractedAt: number;
}

const MostTradedStocksView: React.FC = () => {
  const [topCount, setTopCount] = useState<string>("10");
  const [sortMode, setSortMode] = useState<string>("absolute");
  const [reports, setReports] = useState<MostTradedStocksReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchMostTradedStocks();
  }, []);

  const fetchMostTradedStocks = async () => {
    try {
      setLoading(true);
      setError(null);

      // TODO: values are YTD values, to evaluate if 52 weeks is appropriate on the next roll over
      const response = await fetch(
        getUrl(`/api/v1/analytics/most_traded_stocks?n=52`)
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch data: ${response.statusText}`);
      }

      const data = await response.json();
      setReports(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  };

  // Get color based on value
  const getValueColor = (value: number): string => {
    if (value === 0) return "#9ca3af"; // gray-400

    const maxValue = 100; // Normalize to a reasonable range
    const normalizedValue = Math.min(Math.abs(value) / maxValue, 1);

    if (value > 0) {
      // Green shades for positive values
      const opacity = 0.1 + normalizedValue * 0.9;
      return `rgba(34, 197, 94, ${opacity})`; // green-500 with varying opacity
    } else {
      // Red shades for negative values
      const opacity = 0.1 + normalizedValue * 0.9;
      return `rgba(239, 68, 68, ${opacity})`; // red-500 with varying opacity
    }
  };

  // Get text color for readability
  const getTextColor = (value: number): string => {
    if (value === 0) return "#374151"; // gray-700

    const maxValue = 100;
    const normalizedValue = Math.min(Math.abs(value) / maxValue, 1);

    // Use white text for darker backgrounds
    return normalizedValue > 0.6 ? "#ffffff" : "#374151";
  };

  // Process and sort stocks
  const getTopStocks = () => {
    if (reports.length === 0) return [];

    // Sort reports by date to get the actual latest and earliest
    const sortedReports = [...reports].sort(
      (a, b) =>
        new Date(b.reportDate).getTime() - new Date(a.reportDate).getTime()
    );

    // Get the latest and earliest reports by date
    const latestReport = sortedReports[0];
    const earliestReport = sortedReports[sortedReports.length - 1];

    let sortedStocks: MostTradedStock[] = [];

    switch (sortMode) {
      case "absolute":
        // Sort by absolute YTD values (existing logic)
        sortedStocks = [...latestReport.stocks].sort(
          (a, b) =>
            b.ytdInstitutionNetBuySellSGDM - a.ytdInstitutionNetBuySellSGDM
        );
        break;

      case "percentage":
        // Sort by percentage change from earliest to latest report
        sortedStocks = [...latestReport.stocks]
          .map((stock) => {
            const earliestStock = earliestReport.stocks.find(
              (s) => s.stockCode === stock.stockCode
            );
            const earliestValue =
              earliestStock?.ytdInstitutionNetBuySellSGDM || 0;
            const latestValue = stock.ytdInstitutionNetBuySellSGDM;

            // Calculate percentage change, handling edge cases
            let percentageChange = 0;
            if (Math.abs(earliestValue) > 0.1) {
              // Avoid division by very small numbers
              percentageChange =
                ((latestValue - earliestValue) / earliestValue) * 100;
            } else if (Math.abs(latestValue) > 0.1) {
              // If earliest is ~0 but latest is significant, treat as large positive change
              percentageChange = latestValue > 0 ? 1000 : -1000;
            }

            return { ...stock, percentageChange };
          })
          .sort(
            (a, b) => (b.percentageChange || 0) - (a.percentageChange || 0)
          );
        break;

      case "momentum":
        // Sort by momentum score (combination of absolute change and consistency)
        sortedStocks = [...latestReport.stocks]
          .map((stock) => {
            const earliestStock = earliestReport.stocks.find(
              (s) => s.stockCode === stock.stockCode
            );
            const earliestValue =
              earliestStock?.ytdInstitutionNetBuySellSGDM || 0;
            const latestValue = stock.ytdInstitutionNetBuySellSGDM;

            // Calculate absolute change
            const absoluteChange = latestValue - earliestValue;

            // Calculate consistency (how many periods show positive trend)
            let positivePeriodsCount = 0;
            let previousValue = earliestValue;

            reports
              .slice()
              .reverse()
              .forEach((report) => {
                const reportStock = report.stocks.find(
                  (s) => s.stockCode === stock.stockCode
                );
                if (reportStock) {
                  if (
                    reportStock.ytdInstitutionNetBuySellSGDM > previousValue
                  ) {
                    positivePeriodsCount++;
                  }
                  previousValue = reportStock.ytdInstitutionNetBuySellSGDM;
                }
              });

            const consistencyScore =
              positivePeriodsCount / (reports.length - 1);

            // Momentum score: weighted combination of absolute change and consistency
            const momentumScore =
              absoluteChange * 0.7 +
              consistencyScore * Math.abs(absoluteChange) * 0.3;

            return {
              ...stock,
              momentumScore,
              absoluteChange,
              consistencyScore,
            };
          })
          .sort((a, b) => (b.momentumScore || 0) - (a.momentumScore || 0));
        break;

      default:
        sortedStocks = latestReport.stocks;
    }

    const count = parseInt(topCount);
    const halfCount = Math.floor(count / 2);

    // Take top buyers and top sellers
    const topHalf = sortedStocks.slice(0, halfCount);
    const bottomHalf = sortedStocks.slice(-halfCount);

    return [...topHalf, ...bottomHalf];
  };

  // Create periods for rows (using report dates)
  const getPeriods = () => {
    return reports
      .map((report) => ({
        date: report.reportDate,
        title: report.reportTitle,
      }))
      .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()); // Sort by date descending (latest first)
  };

  // Helper function to get percentage change for a stock
  const getStockPercentageChange = (stockCode: string): number => {
    if (reports.length === 0) return 0;

    // Sort reports by date to get the actual latest and earliest
    const sortedReports = [...reports].sort(
      (a, b) =>
        new Date(b.reportDate).getTime() - new Date(a.reportDate).getTime()
    );

    const latestReport = sortedReports[0];
    const earliestReport = sortedReports[sortedReports.length - 1];

    const latestStock = latestReport.stocks.find(
      (s) => s.stockCode === stockCode
    );
    const earliestStock = earliestReport.stocks.find(
      (s) => s.stockCode === stockCode
    );

    if (!latestStock || !earliestStock) return 0;

    const earliestValue = earliestStock.ytdInstitutionNetBuySellSGDM;
    const latestValue = latestStock.ytdInstitutionNetBuySellSGDM;

    if (Math.abs(earliestValue) > 0.1) {
      return ((latestValue - earliestValue) / earliestValue) * 100;
    } else if (Math.abs(latestValue) > 0.1) {
      return latestValue > 0 ? 1000 : -1000;
    }

    return 0;
  };

  // Get value for specific stock and period
  const getStockValue = (stockCode: string, periodIndex: number): number => {
    const periods = getPeriods();
    if (periodIndex >= periods.length) return 0;

    const targetDate = periods[periodIndex].date;
    const report = reports.find((r) => r.reportDate === targetDate);
    if (!report) return 0;

    const stock = report.stocks.find((s) => s.stockCode === stockCode);
    return stock ? stock.ytdInstitutionNetBuySellSGDM : 0;
  };

  const topStocks = getTopStocks();
  const periods = getPeriods();

  if (loading) {
    return (
      <Stack align="center" py="xl">
        <Loader size="lg" />
        <Text>Loading most traded stocks data...</Text>
      </Stack>
    );
  }

  if (error) {
    return (
      <Alert icon={<IconInfoCircle size="1rem" />} title="Error" color="red">
        {error}
      </Alert>
    );
  }

  if (reports.length === 0) {
    return (
      <Alert
        icon={<IconInfoCircle size="1rem" />}
        title="No Data"
        color="yellow"
      >
        No most traded stocks data available. Please ensure SGX Fund Flow
        reports have been downloaded and processed.
      </Alert>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end">
        <div>
          <Text size="lg" fw={600}>
            Most Traded Stocks Heat Map
          </Text>
          <Text size="sm" c="dimmed">
            YTD Institution Net Buy/Sell (SGD Million) - Latest {reports.length}{" "}
            reports
          </Text>
        </div>

        <Group gap="md" align="flex-end">
          <Select
            label="Sort Method"
            value={sortMode}
            onChange={(value) => setSortMode(value || "absolute")}
            data={[
              { value: "absolute", label: "Absolute" },
              { value: "percentage", label: "% Change" },
              { value: "momentum", label: "Momentum" },
            ]}
            w={140}
          />

          <Select
            label="Show top stocks"
            value={topCount}
            onChange={(value) => setTopCount(value || "30")}
            data={[
              { value: "10", label: "Top 10" },
              { value: "20", label: "Top 20" },
              { value: "30", label: "Top 30" },
              { value: "50", label: "Top 50" },
              { value: "100", label: "Top 100" },
            ]}
            w={140}
          />
        </Group>
      </Group>

      <ScrollArea>
        <Box style={{ minWidth: Math.max(800, topStocks.length * 120) }}>
          <Table
            striped
            highlightOnHover
            withTableBorder
            withColumnBorders
            style={{ fontSize: "12px" }}
          >
            <Table.Thead>
              <Table.Tr>
                <Table.Th
                  style={{
                    minWidth: 120,
                    position: "sticky",
                    left: 0,
                    backgroundColor: "white",
                    zIndex: 1,
                  }}
                >
                  Report Period
                </Table.Th>
                {topStocks.map((stock) => {
                  const percentageChange = getStockPercentageChange(
                    stock.stockCode
                  );

                  // Sort reports by date to get the actual earliest value
                  const sortedReports = [...reports].sort(
                    (a, b) =>
                      new Date(b.reportDate).getTime() -
                      new Date(a.reportDate).getTime()
                  );

                  const earliestValue =
                    sortedReports.length > 0
                      ? sortedReports[sortedReports.length - 1].stocks.find(
                          (s) => s.stockCode === stock.stockCode
                        )?.ytdInstitutionNetBuySellSGDM || 0
                      : 0;
                  const latestValue = stock.ytdInstitutionNetBuySellSGDM;

                  return (
                    <HoverCard key={stock.stockCode} width={280} shadow="md">
                      <HoverCard.Target>
                        <Table.Th
                          style={{
                            minWidth: 100,
                            textAlign: "center",
                            writingMode: "vertical-rl",
                            textOrientation: "mixed",
                            padding: "8px 4px",
                            cursor: "help",
                          }}
                        >
                          <div style={{ transform: "rotate(180deg)" }}>
                            <Text size="xs" fw={600}>
                              {stock.stockName}
                            </Text>
                            <Text size="xs" c="dimmed">
                              ({stock.stockCode})
                            </Text>
                          </div>
                        </Table.Th>
                      </HoverCard.Target>
                      <HoverCard.Dropdown>
                        <Stack gap="xs">
                          <Text fw={600} size="sm">
                            {stock.stockName} ({stock.stockCode}) -{" "}
                            {stock.sector}
                          </Text>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Earliest Value:
                            </Text>
                            <Text size="xs" fw={500}>
                              {earliestValue.toFixed(1)}M SGD
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Latest Value:
                            </Text>
                            <Text size="xs" fw={500}>
                              {latestValue.toFixed(1)}M SGD
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Change:
                            </Text>
                            <Text
                              size="xs"
                              fw={600}
                              c={percentageChange >= 0 ? "green" : "red"}
                            >
                              {percentageChange >= 0 ? "+" : ""}
                              {percentageChange.toFixed(1)}%
                            </Text>
                          </Group>
                        </Stack>
                      </HoverCard.Dropdown>
                    </HoverCard>
                  );
                })}
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {periods.map((period, periodIndex) => (
                <Table.Tr key={`${period.date}-${periodIndex}`}>
                  <Table.Td
                    style={{
                      position: "sticky",
                      left: 0,
                      backgroundColor: "white",
                      zIndex: 1,
                      fontWeight: 500,
                    }}
                  >
                    <div>
                      <Text size="xs" fw={600} c="dimmed">
                        {new Date(period.date).toLocaleDateString()}
                      </Text>
                    </div>
                  </Table.Td>
                  {topStocks.map((stock) => {
                    const value = getStockValue(stock.stockCode, periodIndex);
                    const backgroundColor = getValueColor(value);
                    const textColor = getTextColor(value);

                    return (
                      <Table.Td
                        key={stock.stockCode}
                        style={{
                          backgroundColor,
                          color: textColor,
                          textAlign: "center",
                          fontWeight: 500,
                          padding: "8px 4px",
                        }}
                      >
                        {value.toFixed(1)}
                      </Table.Td>
                    );
                  })}
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Box>
      </ScrollArea>

      <Group gap="xs" align="center">
        <Text size="xs" c="dimmed">
          {sortMode === "absolute" &&
            "Legend: Sorted by absolute YTD institutional net buy/sell amounts (top buyers left, top sellers right)"}
          {sortMode === "percentage" &&
            "Legend: Sorted by percentage change from earliest to latest report (highest gains left, biggest losses right)"}
          {sortMode === "momentum" &&
            "Legend: Sorted by momentum score - combines absolute change + consistency (strongest momentum left, weakest right)"}
        </Text>
      </Group>

      <Group gap="xs" align="center">
        <Text size="xs" c="dimmed">
          Colors:
        </Text>
        <Badge
          color="green"
          variant="filled"
          size="xs"
          style={{ backgroundColor: "rgba(34, 197, 94, 0.8)" }}
        >
          Positive (Net Buy)
        </Badge>
        <Badge
          color="red"
          variant="filled"
          size="xs"
          style={{ backgroundColor: "rgba(239, 68, 68, 0.8)" }}
        >
          Negative (Net Sell)
        </Badge>
        <Badge color="gray" variant="filled" size="xs">
          Neutral (0)
        </Badge>
      </Group>
    </Stack>
  );
};

export default MostTradedStocksView;
