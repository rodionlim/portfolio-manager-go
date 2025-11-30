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
  useMantineColorScheme,
} from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import { getUrl } from "../../utils/url";
import { useNavigate } from "react-router-dom";

interface SectorFlow {
  sectorName: string;
  netBuySellSGDM: number;
}

interface SectorFundsFlowReport {
  reportDate: string;
  reportTitle: string;
  filePath: string;
  weekEndingDate: string;
  overallNetBuySell: number;
  sectorFlows: SectorFlow[];
  extractedAt: number;
}

const SGXSectorView: React.FC = () => {
  const { colorScheme } = useMantineColorScheme();
  const navigate = useNavigate();
  const [topCount, setTopCount] = useState<string>("12");
  const [sortMode, setSortMode] = useState<string>("absolute");
  const [reports, setReports] = useState<SectorFundsFlowReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handleSectorClick = (sectorName: string) => {
    // Navigate to Top 10 stocks view with sector filter
    navigate("/analytics/reports", {
      state: {
        selectedSector: sectorName,
        activeTab: "top10", // Switch to Top 10 Stocks tab
      },
    });
  };

  useEffect(() => {
    fetchSectorFundsFlow();
  }, []);

  const fetchSectorFundsFlow = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch 52 weeks of data for cumulative analysis
      const response = await fetch(
        getUrl(`/api/v1/analytics/sector_funds_flow?n=52`)
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

    const maxValue = 200; // Normalize to a reasonable range for sector flows
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

    const maxValue = 200;
    const normalizedValue = Math.min(Math.abs(value) / maxValue, 1);

    // Use white text for darker backgrounds
    return normalizedValue > 0.6 ? "#ffffff" : "#374151";
  };

  // Process and sort sectors
  const getTopSectors = () => {
    if (reports.length === 0) return [];

    // Sort reports by date to get the actual latest and earliest
    const sortedReports = [...reports].sort(
      (a, b) =>
        new Date(b.reportDate).getTime() - new Date(a.reportDate).getTime()
    );

    // Get the latest report
    const latestReport = sortedReports[0];

    let sortedSectors: (SectorFlow & {
      cumulativeValue?: number;
      percentageChange?: number;
      momentum?: number;
    })[] = [];

    // Calculate cumulative values for each sector
    const cumulativeData = new Map<string, number>();
    sortedReports
      .slice()
      .reverse()
      .forEach((report) => {
        report.sectorFlows.forEach((flow) => {
          const current = cumulativeData.get(flow.sectorName) || 0;
          cumulativeData.set(flow.sectorName, current + flow.netBuySellSGDM);
        });
      });

    switch (sortMode) {
      case "absolute":
        // Sort by cumulative institutional net buy/sell
        sortedSectors = latestReport.sectorFlows
          .map((sector) => ({
            ...sector,
            cumulativeValue: cumulativeData.get(sector.sectorName) || 0,
          }))
          .sort((a, b) => (b.cumulativeValue || 0) - (a.cumulativeValue || 0));
        break;

      case "relative":
        // Sort by latest week's net buy/sell
        sortedSectors = [...latestReport.sectorFlows].sort(
          (a, b) => b.netBuySellSGDM - a.netBuySellSGDM
        );
        break;

      case "percentage":
        // Sort by percentage change from earliest to latest cumulative
        sortedSectors = latestReport.sectorFlows
          .map((sector) => {
            const latestCumulative = cumulativeData.get(sector.sectorName) || 0;

            // Calculate cumulative value up to earliest report
            let earliestCumulative = 0;
            const reportsUpToEarliest = sortedReports.slice(1).reverse(); // Exclude latest, then reverse to get chronological order
            reportsUpToEarliest.forEach((report) => {
              const flow = report.sectorFlows.find(
                (f) => f.sectorName === sector.sectorName
              );
              if (flow) {
                earliestCumulative += flow.netBuySellSGDM;
              }
            });

            let percentageChange = 0;
            if (Math.abs(earliestCumulative) > 1) {
              percentageChange =
                ((latestCumulative - earliestCumulative) /
                  Math.abs(earliestCumulative)) *
                100;
            } else if (Math.abs(latestCumulative) > 1) {
              percentageChange = latestCumulative > 0 ? 1000 : -1000;
            }

            return {
              ...sector,
              cumulativeValue: latestCumulative,
              percentageChange,
            };
          })
          .sort(
            (a, b) => (b.percentageChange || 0) - (a.percentageChange || 0)
          );
        break;

      default:
        sortedSectors = latestReport.sectorFlows.map((sector) => ({
          ...sector,
          cumulativeValue: cumulativeData.get(sector.sectorName) || 0,
        }));
    }

    const count = parseInt(topCount);
    return sortedSectors.slice(0, count);
  };

  // Create periods for rows (using report dates)
  const getPeriods = () => {
    return reports
      .map((report) => ({
        date: report.reportDate,
        weekEnding: report.weekEndingDate,
        title: report.reportTitle,
      }))
      .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
  };

  // Get cumulative value for specific sector up to a specific period
  const getSectorCumulativeValue = (
    sectorName: string,
    periodIndex: number
  ): number => {
    const periods = getPeriods();
    if (periodIndex >= periods.length) return 0;

    // Get all reports from the earliest up to the target period
    const targetDate = periods[periodIndex].date;
    const sortedReports = [...reports].sort(
      (a, b) =>
        new Date(a.reportDate).getTime() - new Date(b.reportDate).getTime()
    );

    let cumulative = 0;
    for (const report of sortedReports) {
      if (new Date(report.reportDate) <= new Date(targetDate)) {
        const flow = report.sectorFlows.find(
          (f) => f.sectorName === sectorName
        );
        if (flow) {
          cumulative += flow.netBuySellSGDM;
        }
      }
    }

    return cumulative;
  };

  // Get weekly value for specific sector and period
  const getSectorWeeklyValue = (
    sectorName: string,
    periodIndex: number
  ): number => {
    const periods = getPeriods();
    if (periodIndex >= periods.length) return 0;

    const targetDate = periods[periodIndex].date;
    const report = reports.find((r) => r.reportDate === targetDate);
    if (!report) return 0;

    const flow = report.sectorFlows.find((f) => f.sectorName === sectorName);
    return flow ? flow.netBuySellSGDM : 0;
  };

  // Helper function to get percentage change for a sector
  /**
   * Calculates the percentage change in cumulative net buy/sell value (SGD millions)
   * for a given sector over the available reports.
   *
   * The calculation compares the cumulative value from the first half of the reports
   * (earliest period) to the cumulative value from all reports (latest period).
   *
   * - If the earliest cumulative value is significant (absolute value > 1),
   *   returns the percentage change between latest and earliest cumulative values.
   * - If the earliest cumulative value is negligible but the latest is significant,
   *   returns 1000 or -1000 to indicate a large relative change.
   * - Returns 0 if both cumulative values are negligible or if there are no reports.
   *
   * @param sectorName - The name of the sector to calculate the percentage change for.
   * @returns The percentage change in cumulative net buy/sell value for the sector.
   */
  const getSectorPercentageChange = (sectorName: string): number => {
    if (reports.length === 0) return 0;

    const sortedReports = [...reports].sort(
      (a, b) =>
        new Date(a.reportDate).getTime() - new Date(b.reportDate).getTime()
    );

    // Calculate cumulative values
    let earliestCumulative = 0;
    let latestCumulative = 0;

    // Calculate earliest cumulative (first half of reports)
    const halfIndex = Math.floor(sortedReports.length / 2);
    for (let i = 0; i < halfIndex; i++) {
      const flow = sortedReports[i].sectorFlows.find(
        (f) => f.sectorName === sectorName
      );
      if (flow) earliestCumulative += flow.netBuySellSGDM;
    }

    // Calculate latest cumulative (all reports)
    for (const report of sortedReports) {
      const flow = report.sectorFlows.find((f) => f.sectorName === sectorName);
      if (flow) latestCumulative += flow.netBuySellSGDM;
    }

    if (Math.abs(earliestCumulative) > 1) {
      return (
        ((latestCumulative - earliestCumulative) /
          Math.abs(earliestCumulative)) *
        100
      );
    } else if (Math.abs(latestCumulative) > 1) {
      return latestCumulative > 0 ? 1000 : -1000;
    }

    return 0;
  };

  const topSectors = getTopSectors();
  const periods = getPeriods();

  if (loading) {
    return (
      <Stack align="center" py="xl">
        <Loader size="lg" />
        <Text>Loading sector funds flow data...</Text>
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
        No sector funds flow data available. Please ensure SGX Fund Flow reports
        have been downloaded and processed.
      </Alert>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end">
        <div>
          <Text size="lg" fw={600}>
            SGX Sector Funds Flow Heat Map
          </Text>
          <Text size="sm" c="dimmed">
            Institutional Net Buy/Sell by Sector (SGD Million) - Latest{" "}
            {reports.length} reports
          </Text>
        </div>

        <Group gap="md" align="flex-end">
          <Select
            label="Sort Method"
            value={sortMode}
            onChange={(value) => setSortMode(value || "absolute")}
            data={[
              { value: "absolute", label: "Cumulative" },
              { value: "relative", label: "Latest Week" },
              { value: "percentage", label: "% Change" },
            ]}
            w={140}
          />

          <Select
            label="Show sectors"
            value={topCount}
            onChange={(value) => setTopCount(value || "12")}
            data={[
              { value: "6", label: "Top 6" },
              { value: "8", label: "Top 8" },
              { value: "10", label: "Top 10" },
              { value: "12", label: "All 12" },
            ]}
            w={140}
          />
        </Group>
      </Group>

      <Alert
        icon={<IconInfoCircle size="1rem" />}
        title="Interactive Analysis"
        color="blue"
        variant="light"
        mb="md"
      >
        Click on any sector header to drill down and view individual stocks
        within that sector.
      </Alert>

      <ScrollArea>
        <Box style={{ minWidth: Math.max(800, topSectors.length * 140) }}>
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
                    backgroundColor:
                      colorScheme === "dark" ? "#1a1b1e" : "white",
                    color: colorScheme === "dark" ? "#f8f9fa" : "#000000",
                    zIndex: 1,
                  }}
                >
                  Report Period
                </Table.Th>
                {topSectors.map((sector) => {
                  const percentageChange = getSectorPercentageChange(
                    sector.sectorName
                  );
                  const cumulativeValue = sector.cumulativeValue || 0;
                  const latestWeekValue = sector.netBuySellSGDM;

                  return (
                    <HoverCard key={sector.sectorName} width={300} shadow="md">
                      <HoverCard.Target>
                        <Table.Th
                          style={{
                            minWidth: 120,
                            textAlign: "center",
                            writingMode: "vertical-rl",
                            textOrientation: "mixed",
                            padding: "8px 4px",
                            cursor: "pointer",
                            transition: "background-color 0.2s ease",
                          }}
                          onClick={() => handleSectorClick(sector.sectorName)}
                          onMouseEnter={(e) => {
                            if (colorScheme === "light") {
                              e.currentTarget.style.backgroundColor = "#f8f9fa";
                            } else {
                              e.currentTarget.style.backgroundColor = "#1a1b1e";
                            }
                          }}
                          onMouseLeave={(e) => {
                            if (colorScheme === "light") {
                              e.currentTarget.style.backgroundColor = "white";
                            }
                          }}
                        >
                          <div style={{ transform: "rotate(180deg)" }}>
                            <Text size="xs" fw={600} c="blue.6">
                              {sector.sectorName}
                            </Text>
                            <Text size="xs" c="dimmed">
                              ({percentageChange.toFixed(1)}%)
                            </Text>
                          </div>
                        </Table.Th>
                      </HoverCard.Target>
                      <HoverCard.Dropdown>
                        <Stack gap="xs">
                          <Text fw={600} size="sm">
                            {sector.sectorName}
                          </Text>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Cumulative Value:
                            </Text>
                            <Text size="xs" fw={500}>
                              {cumulativeValue.toFixed(1)}M SGD
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Latest Week:
                            </Text>
                            <Text size="xs" fw={500}>
                              {latestWeekValue.toFixed(1)}M SGD
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Cumulative Change:
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
                      zIndex: 1,
                      fontWeight: 500,
                      backgroundColor:
                        colorScheme === "dark" ? "#1a1b1e" : "white",
                      color: colorScheme === "dark" ? "#f8f9fa" : "#000000",
                    }}
                  >
                    <div>
                      <Text size="xs" fw={600} c="dimmed">
                        {period.weekEnding}
                      </Text>
                    </div>
                  </Table.Td>
                  {topSectors.map((sector) => {
                    // Show cumulative or weekly value based on sort mode
                    let value: number;
                    if (sortMode === "relative") {
                      value = getSectorWeeklyValue(
                        sector.sectorName,
                        periodIndex
                      );
                    } else {
                      value = getSectorCumulativeValue(
                        sector.sectorName,
                        periodIndex
                      );
                    }

                    const backgroundColor = getValueColor(value);
                    const textColor = getTextColor(value);

                    return (
                      <Table.Td
                        key={sector.sectorName}
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
            "Legend: Sorted by cumulative institutional net buy/sell (highest cumulative inflows are sorted to the left)"}
          {sortMode === "relative" &&
            "Legend: Sorted by latest week's institutional net buy/sell (highest weekly inflows are sorted to the left)"}
          {sortMode === "percentage" &&
            "Legend: Sorted by percentage change in cumulative flows (highest growth are sorted to the left)"}
        </Text>
      </Group>

      <Group gap="xs" align="center">
        <Text size="xs" c="dimmed">
          Values: {sortMode === "relative" ? "Weekly" : "Cumulative"} | Colors:
        </Text>
        <Badge
          color="green"
          variant="filled"
          size="xs"
          style={{ backgroundColor: "rgba(34, 197, 94, 0.8)" }}
        >
          Positive (Net Inflows)
        </Badge>
        <Badge
          color="red"
          variant="filled"
          size="xs"
          style={{ backgroundColor: "rgba(239, 68, 68, 0.8)" }}
        >
          Negative (Net Outflows)
        </Badge>
        <Badge color="gray" variant="filled" size="xs">
          Neutral (0)
        </Badge>
      </Group>
    </Stack>
  );
};

export default SGXSectorView;
