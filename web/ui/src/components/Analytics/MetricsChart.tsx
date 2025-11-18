import React, { useEffect, useRef, useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  createChart,
  ColorType,
  UTCTimestamp,
  PriceScaleMode,
  LineSeries,
  AreaSeries,
} from "lightweight-charts";
import {
  Box,
  Title,
  Text,
  Paper,
  Select,
  Stack,
  Group,
  Button,
} from "@mantine/core";
import { getUrl } from "../../utils/url";
import { notifications } from "@mantine/notifications";
import { TimestampedMetrics, MetricsJob } from "./types";

interface ValueData {
  time: UTCTimestamp;
  value: number;
}

interface IRRData {
  time: UTCTimestamp;
  value: number;
}

const MetricsChart: React.FC = () => {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<any>(null); // To keep chart instance for resizing
  const [leftAxisSelection, setLeftAxisSelection] = useState<string>("PnL");
  const [rightAxisSelection, setRightAxisSelection] = useState<string>("IRR");
  const [selectedBookFilter, setSelectedBookFilter] = useState<string | null>(
    "None"
  );
  const [timelineFilter, setTimelineFilter] = useState<string>("All");

  // Fetch metrics jobs to populate book filter dropdown
  const fetchMetricsJobs = async (): Promise<MetricsJob[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/historical/metrics/jobs"));
      return await resp.json();
    } catch (error: any) {
      console.error("Error fetching metrics jobs:", error);
      return [];
    }
  };

  // Query to fetch metrics jobs
  const { data: metricsJobs = [] } = useQuery({
    queryKey: ["metricsJobs"],
    queryFn: fetchMetricsJobs,
  });

  // Create book filter options
  const bookFilterOptions = useMemo(() => {
    const options = [{ value: "None", label: "None" }];
    metricsJobs.forEach((job) => {
      if (!options.find((opt) => opt.value === job.BookFilter)) {
        options.push({ value: job.BookFilter, label: job.BookFilter });
      }
    });
    return options;
  }, [metricsJobs]);

  // Fetch all historical metrics, reusing the same query function from MetricsTable
  const fetchHistoricalMetrics = async (): Promise<TimestampedMetrics[]> => {
    try {
      const bookFilter =
        selectedBookFilter === "None" ? "" : selectedBookFilter || "";
      const url = bookFilter
        ? getUrl(
            `/api/v1/historical/metrics?book_filter=${encodeURIComponent(
              bookFilter
            )}`
          )
        : getUrl("/api/v1/historical/metrics");
      const resp = await fetch(url);
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

  // Get the current users primary locale
  const currentLocale = window.navigator.languages[0];
  // Create a custom formatter to show only $ sign
  const pxFormatter = (price: number): string => {
    // Format the number with proper thousand separators and decimals
    return (
      "$" +
      price.toLocaleString(currentLocale, {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      })
    );
  };

  // Query to fetch historical metrics
  const {
    data: historicalMetrics = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ["historicalMetrics", selectedBookFilter],
    queryFn: fetchHistoricalMetrics,
  });

  const filteredMetrics = useMemo(() => {
    if (!historicalMetrics || historicalMetrics.length === 0) return [];

    const sorted = [...historicalMetrics].sort(
      (a, b) =>
        new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );

    if (timelineFilter === "All") {
      return sorted;
    }

    const now = new Date();
    let startDate = new Date(sorted[0].timestamp);

    if (timelineFilter === "MTD") {
      startDate = new Date(now.getFullYear(), now.getMonth(), 1);
    } else if (timelineFilter === "YTD") {
      startDate = new Date(now.getFullYear(), 0, 1);
    } else if (timelineFilter === "3Y") {
      startDate = new Date(now);
      startDate.setFullYear(now.getFullYear() - 3);
    }

    return sorted.filter(
      (item) => new Date(item.timestamp).getTime() >= startDate.getTime()
    );
  }, [historicalMetrics, timelineFilter]);

  useEffect(() => {
    if (!chartContainerRef.current || filteredMetrics.length === 0) return;

    // Convert API data to chart series data format
    const marketValueData: ValueData[] = [];
    const pnlData: ValueData[] = [];
    const irrData: IRRData[] = [];

    // Process data by timestamp
    filteredMetrics.forEach((item) => {
      const timestamp = Math.floor(
        new Date(item.timestamp).getTime() / 1000
      ) as UTCTimestamp;
      marketValueData.push({
        time: timestamp,
        value: item.metrics.mv,
      });
      pnlData.push({
        time: timestamp,
        value:
          item.metrics.mv +
          item.metrics.totalDividends -
          item.metrics.pricePaid,
      });
      irrData.push({
        time: timestamp,
        value: item.metrics.irr * 100, // Convert IRR to percentage for better visualization
      });
    });

    // Create chart
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: "#ffffff" },
        textColor: "#333333",
        attributionLogo: false,
      },
      grid: {
        vertLines: { color: "#f0f0f0" },
        horzLines: { color: "#f0f0f0" },
      },
      width: chartContainerRef.current.clientWidth,
      height: 500,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
      },
    });
    chartRef.current = chart;

    // Add Market Value series
    const marketValueSeries = chart.addSeries(AreaSeries, {
      lineColor: "#2962FF",
      topColor: "rgba(41, 98, 255, 0.4)",
      bottomColor: "rgba(41, 98, 255, 0.0)",
      lineWidth: 2,
      title: "Market Value",
      priceScaleId: "left",
      priceFormat: {
        type: "custom",
        formatter: pxFormatter,
      },
      visible: leftAxisSelection === "MV",
    });
    marketValueSeries.setData(marketValueData);

    // Add PnL series
    const pnlSeries = chart.addSeries(AreaSeries, {
      lineColor: "#4CAF50", // Green color for PnL
      topColor: "rgba(76, 175, 80, 0.4)",
      bottomColor: "rgba(76, 175, 80, 0.0)",
      lineWidth: 2,
      title: "PnL",
      priceScaleId: "left",
      priceFormat: {
        type: "custom",
        formatter: pxFormatter,
      },
      visible: leftAxisSelection === "PnL",
    });
    pnlSeries.setData(pnlData);

    const rightAxisConfigMap = {
      IRR: {
        color: "#FF6B6B",
        title: "IRR (%)",
        priceFormat: {
          type: "percent" as const,
          precision: 2,
        },
        data: irrData,
      },
      MV: {
        color: "#FF9800",
        title: "Market Value (Right)",
        priceFormat: {
          type: "custom" as const,
          formatter: pxFormatter,
        },
        data: marketValueData,
      },
      PnL: {
        color: "#8E24AA",
        title: "P&L (Right)",
        priceFormat: {
          type: "custom" as const,
          formatter: pxFormatter,
        },
        data: pnlData,
      },
    } as const;

    const rightAxisConfig =
      rightAxisConfigMap[
        (rightAxisSelection || "IRR") as keyof typeof rightAxisConfigMap
      ] || rightAxisConfigMap.IRR;

    const rightSeries = chart.addSeries(LineSeries, {
      color: rightAxisConfig.color,
      lineWidth: 2,
      title: rightAxisConfig.title,
      priceScaleId: "right",
      priceFormat: rightAxisConfig.priceFormat,
    });
    rightSeries.setData(rightAxisConfig.data);

    // Configure separate price scales for both series
    chart.priceScale("left").applyOptions({
      borderVisible: true,
      borderColor: leftAxisSelection === "MV" ? "#2962FF" : "#4CAF50",
      mode: PriceScaleMode.Normal,
      visible: true,
      autoScale: true,
    });

    chart.priceScale("right").applyOptions({
      scaleMargins: {
        top: 0.1,
        bottom: 0.3,
      },
      borderVisible: true,
      borderColor:
        rightAxisSelection === "IRR"
          ? "#FF6B6B"
          : rightAxisSelection === "MV"
          ? "#FF9800"
          : "#8E24AA",
      mode: PriceScaleMode.Normal,
      visible: true,
      autoScale: true,
    });

    chart.timeScale().fitContent();

    // Use ResizeObserver for robust resizing
    const resizeObserver = new window.ResizeObserver(() => {
      if (chartContainerRef.current && chart) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
        });
        chart.timeScale().fitContent();
      }
    });
    resizeObserver.observe(chartContainerRef.current);

    // Clean up
    return () => {
      resizeObserver.disconnect();
      chart.remove();
    };
  }, [filteredMetrics, leftAxisSelection, rightAxisSelection]);

  // Handle loading states
  if (isLoading) return <div>Loading historical metrics chart...</div>;
  if (error) return <div>Error loading historical metrics chart</div>;

  if (historicalMetrics.length === 0 && !isLoading && !error) {
    return (
      <Paper p="xl" withBorder>
        <Select
          label="Book Filter"
          placeholder="Select book filter"
          data={bookFilterOptions}
          value={selectedBookFilter}
          onChange={setSelectedBookFilter}
          clearable={false}
          w={200}
        />
        <Text c="dimmed" ta="center">
          No historical metrics records found to display chart
        </Text>
      </Paper>
    );
  }

  return (
    <Box>
      <Title order={3} mb="md">
        Portfolio Performance Chart
      </Title>
      <Paper p="md" withBorder>
        <Stack>
          <Group gap="md" align="flex-end">
            <Select
              label="Left Axis Metric"
              value={leftAxisSelection}
              onChange={(value) => setLeftAxisSelection(value || "PnL")}
              data={[
                { value: "MV", label: "Market Value" },
                { value: "PnL", label: "P&L" },
              ]}
              w={200}
            />
            <Select
              label="Right Axis Metric"
              value={rightAxisSelection}
              onChange={(value) => setRightAxisSelection(value || "IRR")}
              data={[
                { value: "IRR", label: "IRR" },
                { value: "MV", label: "Market Value" },
                { value: "PnL", label: "P&L" },
              ]}
              w={200}
            />
            <Select
              label="Book Filter"
              placeholder="Select book filter"
              data={bookFilterOptions}
              value={selectedBookFilter}
              onChange={setSelectedBookFilter}
              clearable={false}
              w={200}
            />
          </Group>
          <Group gap="xs">
            {[
              { label: "MTD", value: "MTD" },
              { label: "YTD", value: "YTD" },
              { label: "3Y", value: "3Y" },
              { label: "All", value: "All" },
            ].map((option) => (
              <Button
                key={option.value}
                size="xs"
                variant={timelineFilter === option.value ? "filled" : "light"}
                onClick={() => setTimelineFilter(option.value)}
              >
                {option.label}
              </Button>
            ))}
          </Group>
          <Box
            ref={chartContainerRef}
            style={{ width: "100%", minHeight: 500, minWidth: 0 }}
          />
        </Stack>
        <Text size="xs" c="dimmed" mt="xs">
          {leftAxisSelection === "MV" ? "Market Value" : "P&L"} shown as{" "}
          {leftAxisSelection === "MV" ? "blue" : "green"} area (left scale).{" "}
          {rightAxisSelection === "IRR"
            ? "IRR shown as red line"
            : rightAxisSelection === "MV"
            ? "Market Value shown as orange line"
            : "P&L shown as purple line"}{" "}
          (right scale)
        </Text>
      </Paper>
    </Box>
  );
};

export default MetricsChart;
