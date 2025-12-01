import React, { useState, useEffect, useRef } from "react";
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
  Button,
  MultiSelect,
} from "@mantine/core";
import { IconInfoCircle, IconBriefcase } from "@tabler/icons-react";
import { getUrl } from "../../utils/url";
import { useLocation } from "react-router-dom";

interface Position {
  Ticker: string;
  Name: string;
  AssetClass: string;
}

interface ReferenceData {
  id: string;
  underlying_ticker: string;
  category: string;
  category_sgx: string;
}

interface Top10Stock {
  stockName: string;
  stockCode: string;
  netBuySellSGDM: number;
  isNetBuy: boolean;
  investorType: string; // "institutional" or "retail"
}

interface Top10WeeklyReport {
  reportDate: string;
  reportTitle: string;
  filePath: string;
  weekEndingDate: string;
  institutionalNetSellTotalSGDM: number;
  institutionalNetSellPreviousSGDM: number;
  retailNetBuyTotalSGDM: number;
  retailNetBuyPreviousSGDM: number;
  top10Stocks: Top10Stock[];
  extractedAt: number;
}

const SGXTop10StocksView: React.FC = () => {
  const { colorScheme } = useMantineColorScheme();
  const location = useLocation();
  const [investorType, setInvestorType] = useState<string>("institutional");
  const [sortMethod, setSortMethod] = useState<string>("balanced");
  const [topCount, setTopCount] = useState<string>("10");
  const [reports, setReports] = useState<Top10WeeklyReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSectors, setSelectedSectors] = useState<string[]>([]);
  const [portfolioTickers, setPortfolioTickers] = useState<string[]>([]);
  const [showPortfolioOnly, setShowPortfolioOnly] = useState(false);
  const [portfolioLoading, setPortfolioLoading] = useState(true);
  const [stockSectorMap, setStockSectorMap] = useState<Map<string, string>>(
    new Map()
  );
  const [availableSectors, setAvailableSectors] = useState<string[]>([]);
  const [sectorLoading, setSectorLoading] = useState(true);

  // Track if navigation state was already applied to prevent re-applying on tab switches
  const navigationAppliedRef = useRef(false);

  // Get navigation state (sector filter from SGXSectorView)
  const navigationState = location.state as { selectedSector?: string } | null;

  useEffect(() => {
    // Only set initial sector filter once when first coming from navigation
    if (navigationState?.selectedSector && !navigationAppliedRef.current) {
      setSelectedSectors([navigationState.selectedSector]);
      navigationAppliedRef.current = true;
    }
  }, [navigationState]);

  useEffect(() => {
    fetchTop10Stocks();
    fetchPortfolioPositions();
    fetchSectorData();
  }, []);

  // Fetch portfolio positions to get tickers for filtering (using lite endpoint for performance)
  const fetchPortfolioPositions = async () => {
    setPortfolioLoading(true);
    try {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions/lite"));
      if (!resp.ok) {
        console.error("Error fetching positions");
        return;
      }
      const positions: Position[] = await resp.json();
      // Filter out bonds (SSB, TBill) and extract unique SGX tickers
      const sgxTickers = positions
        .filter((p) => {
          // Exclude Singapore Savings Bonds (SSB) and T-Bills
          const ticker = p.Ticker;
          const isSsb = ticker.startsWith("SB") && ticker.length === 7;
          const isTbill =
            ticker.length === 8 &&
            /^[A-Za-z]{2}/.test(ticker) &&
            /[A-Za-z]$/.test(ticker);
          return !isSsb && !isTbill && p.AssetClass !== "Bond";
        })
        .map((p) => p.Ticker);
      setPortfolioTickers([...new Set(sgxTickers)]);
    } catch (err) {
      console.error("Failed to fetch portfolio positions:", err);
    } finally {
      setPortfolioLoading(false);
    }
  };

  // Fetch sector data from reference data API to map stock codes to sectors/categories
  const fetchSectorData = async () => {
    setSectorLoading(true);
    try {
      const response = await fetch(getUrl(`/api/v1/refdata`));
      if (!response.ok) {
        console.error("Failed to fetch reference data");
        return;
      }
      // API returns an object/map format: { "ticker": {...}, "ticker2": {...} }
      const data: Record<string, ReferenceData> = await response.json();
      if (data && Object.keys(data).length > 0) {
        const sectorMap = new Map<string, string>();
        const sectors = new Set<string>();
        Object.values(data).forEach((ref: ReferenceData) => {
          // Use category_sgx for SGX sector mapping, with fallback to category
          const sectorName =
            (ref.category_sgx && ref.category_sgx.trim()) ||
            (ref.category && ref.category.trim());
          if (sectorName) {
            // Also map without .SI suffix for matching
            if (ref.id.endsWith(".SI")) {
              sectorMap.set(ref.id.replace(".SI", ""), sectorName);
            }
            sectors.add(sectorName);
          }
        });
        setStockSectorMap(sectorMap);
        setAvailableSectors(Array.from(sectors).sort());
      }
    } catch (err) {
      console.error("Failed to fetch sector data:", err);
    } finally {
      setSectorLoading(false);
    }
  };

  const fetchTop10Stocks = async () => {
    try {
      setLoading(true);
      setError(null);

      // Limit to latest 52 weeks of reports
      const response = await fetch(
        getUrl(`/api/v1/analytics/top10_stocks?n=52`)
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

    const maxValue = 150; // Normalize to a reasonable range for Top 10 data
    const normalizedValue = Math.min(Math.abs(value) / maxValue, 1);

    if (value > 0) {
      // Green shades for positive values (net buy)
      const opacity = 0.1 + normalizedValue * 0.9;
      return `rgba(34, 197, 94, ${opacity})`; // green-500 with varying opacity
    } else {
      // Red shades for negative values (net sell)
      const opacity = 0.1 + normalizedValue * 0.9;
      return `rgba(239, 68, 68, ${opacity})`; // red-500 with varying opacity
    }
  };

  // Get text color for readability
  const getTextColor = (value: number): string => {
    if (value === 0) return colorScheme === "dark" ? "#d1d5db" : "#374151"; // gray-300 for dark, gray-700 for light

    const maxValue = 150;
    const normalizedValue = Math.min(Math.abs(value) / maxValue, 1);

    // Use appropriate text colors based on theme and background intensity
    if (normalizedValue > 0.6) {
      // For darker backgrounds, use white text
      return "#ffffff";
    } else {
      // For lighter backgrounds, use theme-appropriate text
      return colorScheme === "dark" ? "#f3f4f6" : "#374151"; // gray-100 for dark theme, gray-700 for light
    }
  };

  // Get unique sectors - uses category data from reference data API
  const getUniqueSectors = (): string[] => {
    return availableSectors;
  };

  // Get sector for a stock code using the sector map
  const getStockSector = (stockCode: string): string | undefined => {
    return stockSectorMap.get(stockCode);
  };

  // Get stocks for the selected investor type and calculate cumulative values
  const getProcessedStocks = () => {
    if (reports.length === 0) return [];

    // Create a map to track cumulative values for each stock
    const stockMap = new Map<
      string,
      {
        stockName: string;
        stockCode: string;
        cumulativeValue: number;
        weeklyValues: Map<string, number>;
        sector?: string;
      }
    >();

    // Process each report
    reports.forEach((report) => {
      const reportDate = report.reportDate;

      // Filter stocks by investor type
      const filteredStocks = report.top10Stocks.filter(
        (stock) => stock.investorType === investorType
      );

      filteredStocks.forEach((stock) => {
        if (!stockMap.has(stock.stockCode)) {
          stockMap.set(stock.stockCode, {
            stockName: stock.stockName,
            stockCode: stock.stockCode,
            cumulativeValue: 0,
            weeklyValues: new Map(),
            sector: stockSectorMap.get(stock.stockCode),
          });
        }

        const stockData = stockMap.get(stock.stockCode)!;
        stockData.cumulativeValue += stock.netBuySellSGDM;
        stockData.weeklyValues.set(reportDate, stock.netBuySellSGDM);
      });
    });

    // Convert to array and sort by cumulative value (descending)
    let allStocks = Array.from(stockMap.values()).sort(
      (a, b) => b.cumulativeValue - a.cumulativeValue
    );

    // Apply sector filtering using the sector map from most traded stocks
    if (selectedSectors.length > 0) {
      allStocks = allStocks.filter((stock) => {
        const sector = getStockSector(stock.stockCode);
        return sector && selectedSectors.includes(sector);
      });
    }

    // Apply portfolio filtering
    if (showPortfolioOnly && portfolioTickers.length > 0) {
      allStocks = allStocks.filter((stock) =>
        portfolioTickers.some(
          (ticker) =>
            ticker === stock.stockCode ||
            ticker.replace(".SI", "") === stock.stockCode ||
            stock.stockCode.replace(".SI", "") === ticker.replace(".SI", "")
        )
      );
    }

    const count = parseInt(topCount);

    // Apply sort method filtering
    switch (sortMethod) {
      case "positive":
        // Only positive cumulative values
        const positiveStocks = allStocks.filter(
          (stock) => stock.cumulativeValue > 0
        );
        return positiveStocks.slice(0, count);

      case "negative":
        // Only negative cumulative values, sorted by most negative first
        const negativeStocks = allStocks.filter(
          (stock) => stock.cumulativeValue < 0
        );
        return negativeStocks.slice(0, count).reverse();

      case "balanced":
      default:
        // Split the count: half positive (net buyers), half negative (net sellers)
        const halfCount = Math.floor(count / 2);
        const positives = allStocks.filter(
          (stock) => stock.cumulativeValue > 0
        );
        const negatives = allStocks.filter(
          (stock) => stock.cumulativeValue < 0
        );

        const topPositives = positives.slice(0, halfCount);
        const topNegatives = negatives.reverse().slice(0, halfCount);

        return [...topPositives, ...topNegatives];
    }
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

  // Get value for specific stock and period
  const getStockValue = (stockCode: string, periodIndex: number): number => {
    const periods = getPeriods();
    if (periodIndex >= periods.length) return 0;

    const targetDate = periods[periodIndex].date;
    const report = reports.find((r) => r.reportDate === targetDate);
    if (!report) return 0;

    const stock = report.top10Stocks.find(
      (s) => s.stockCode === stockCode && s.investorType === investorType
    );
    return stock ? stock.netBuySellSGDM : 0;
  };

  const processedStocks = getProcessedStocks();
  const periods = getPeriods();

  if (loading) {
    return (
      <Stack align="center" py="xl">
        <Loader size="lg" />
        <Text>Loading Top 10 stocks data...</Text>
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
        No Top 10 stocks data available. Please ensure SGX Fund Flow reports
        have been downloaded and processed.
      </Alert>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end">
        <div>
          <Text size="lg" fw={600}>
            Weekly Top 10 Stocks Heat Map
          </Text>
          <Text size="sm" c="dimmed">
            {investorType === "institutional" ? "Institutional" : "Retail"} Net
            Buy/Sell (SGD Million) - Latest {reports.length} reports
            {selectedSectors.length > 0 && (
              <Text component="span" c="blue">
                {" "}
                • Filtered by {selectedSectors.length} sector
                {selectedSectors.length !== 1 ? "s" : ""}
              </Text>
            )}
            {showPortfolioOnly && (
              <Text component="span" c="teal">
                {" "}
                • Portfolio stocks only
              </Text>
            )}
          </Text>
        </div>

        <Group gap="md" align="flex-end">
          <Button
            variant={showPortfolioOnly ? "filled" : "light"}
            color="teal"
            leftSection={<IconBriefcase size={16} />}
            onClick={() => setShowPortfolioOnly(!showPortfolioOnly)}
            disabled={portfolioLoading || portfolioTickers.length === 0}
            loading={portfolioLoading}
            title={
              portfolioLoading
                ? "Loading portfolio positions..."
                : portfolioTickers.length === 0
                ? "No portfolio positions found"
                : showPortfolioOnly
                ? "Click to show all stocks"
                : "Click to show only portfolio stocks"
            }
          >
            {showPortfolioOnly ? "Show All" : "My Portfolio"}
          </Button>

          <Select
            label="Investor Type"
            value={investorType}
            onChange={(value) => setInvestorType(value || "institutional")}
            data={[
              { value: "institutional", label: "Institutional" },
              { value: "retail", label: "Retail" },
            ]}
            w={140}
          />

          <Select
            label="Sort Method"
            value={sortMethod}
            onChange={(value) => setSortMethod(value || "balanced")}
            data={[
              { value: "balanced", label: "Balanced" },
              { value: "positive", label: "Only Positive" },
              { value: "negative", label: "Only Negative" },
            ]}
            w={140}
          />

          <Select
            label="Show top stocks"
            value={topCount}
            onChange={(value) => setTopCount(value || "10")}
            data={[
              { value: "10", label: "Top 10" },
              { value: "20", label: "Top 20" },
              { value: "30", label: "Top 30" },
            ]}
            w={140}
          />

          {/* Show category filter on same row when nothing selected */}
          {selectedSectors.length === 0 && (
            <MultiSelect
              label="Filter by Category"
              placeholder="All categories"
              value={selectedSectors}
              onChange={setSelectedSectors}
              data={getUniqueSectors()}
              clearable
              w={200}
              disabled={sectorLoading}
            />
          )}
        </Group>
      </Group>

      {/* Sector filter on separate row only when sectors are selected */}
      {selectedSectors.length > 0 && (
        <Group gap="md" align="flex-end">
          <MultiSelect
            label="Filter by Category"
            placeholder={`${selectedSectors.length} selected`}
            value={selectedSectors}
            onChange={setSelectedSectors}
            data={getUniqueSectors()}
            clearable
            w={500}
          />
          <Button
            variant="light"
            color="blue"
            size="sm"
            onClick={() => setSelectedSectors([])}
          >
            Clear Filter
          </Button>
        </Group>
      )}

      <ScrollArea>
        <Box style={{ minWidth: Math.max(800, processedStocks.length * 120) }}>
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
                {processedStocks.map((stock) => {
                  const cumulativeValue = stock.cumulativeValue;

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
                            {stock.stockName} ({stock.stockCode})
                          </Text>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Investor Type:
                            </Text>
                            <Text size="xs" fw={500}>
                              {investorType === "institutional"
                                ? "Institutional"
                                : "Retail"}
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Cumulative Net Buy/Sell:
                            </Text>
                            <Text
                              size="xs"
                              fw={600}
                              c={cumulativeValue >= 0 ? "green" : "red"}
                            >
                              {cumulativeValue >= 0 ? "+" : ""}
                              {cumulativeValue.toFixed(1)}M SGD
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Weeks Appeared:
                            </Text>
                            <Text size="xs" fw={500}>
                              {stock.weeklyValues.size} / {reports.length}
                            </Text>
                          </Group>
                          <Group justify="space-between">
                            <Text size="xs" c="dimmed">
                              Sector:
                            </Text>
                            <Text size="xs" fw={500}>
                              {stock.sector}
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
                      backgroundColor:
                        colorScheme === "dark" ? "#1a1b1e" : "white",
                      color: colorScheme === "dark" ? "#f8f9fa" : "#000000",
                      zIndex: 1,
                      fontWeight: 500,
                    }}
                  >
                    <div>
                      <Text
                        size="xs"
                        fw={600}
                        c={colorScheme === "dark" ? "gray.4" : "dimmed"}
                      >
                        {new Date(period.date).toLocaleDateString()}
                      </Text>
                    </div>
                  </Table.Td>
                  {processedStocks.map((stock) => {
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
                        {value !== 0 ? value.toFixed(1) : ""}
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
          Legend:{" "}
          {sortMethod === "balanced"
            ? "Balanced view - top net buyers (left) and top net sellers (right)"
            : sortMethod === "positive"
            ? "Only positive cumulative net buyers (highest left, lowest right)"
            : "Only negative cumulative net sellers (most negative left, least negative right)"}
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
          No Activity
        </Badge>
      </Group>

      {/* Summary Statistics */}
      {reports.length > 0 && (
        <Group gap="md" align="center">
          <Text size="xs" c="dimmed">
            Latest Week Summary:
          </Text>
          <Badge variant="light" color="blue">
            Institutional Net:{" "}
            {investorType === "institutional"
              ? `${
                  reports[0]?.institutionalNetSellTotalSGDM > 0 ? "+" : ""
                }${reports[0]?.institutionalNetSellTotalSGDM?.toFixed(1)}M SGD`
              : `${
                  reports[0]?.retailNetBuyTotalSGDM > 0 ? "+" : ""
                }${reports[0]?.retailNetBuyTotalSGDM?.toFixed(1)}M SGD`}
          </Badge>
          <Badge variant="light" color="gray">
            Previous Week:{" "}
            {investorType === "institutional"
              ? `${
                  reports[0]?.institutionalNetSellPreviousSGDM > 0 ? "+" : ""
                }${reports[0]?.institutionalNetSellPreviousSGDM?.toFixed(
                  1
                )}M SGD`
              : `${
                  reports[0]?.retailNetBuyPreviousSGDM > 0 ? "+" : ""
                }${reports[0]?.retailNetBuyPreviousSGDM?.toFixed(1)}M SGD`}
          </Badge>
        </Group>
      )}
    </Stack>
  );
};

export default SGXTop10StocksView;
