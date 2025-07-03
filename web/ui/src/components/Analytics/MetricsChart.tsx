import React, { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  createChart,
  ColorType,
  UTCTimestamp,
  PriceScaleMode,
  LineSeries,
  AreaSeries,
} from "lightweight-charts";
import { Box, Title, Text, Paper, Select, Stack } from "@mantine/core";
import { getUrl } from "../../utils/url";
import { notifications } from "@mantine/notifications";
import { TimestampedMetrics } from "./types";

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

  // Fetch all historical metrics, reusing the same query function from MetricsTable
  const fetchHistoricalMetrics = async (): Promise<TimestampedMetrics[]> => {
    try {
      const resp = await fetch(getUrl("/api/v1/historical/metrics"));
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
    queryKey: ["historicalMetrics"],
    queryFn: fetchHistoricalMetrics,
  });

  useEffect(() => {
    if (!chartContainerRef.current || historicalMetrics.length === 0) return;

    // Convert API data to chart series data format
    const marketValueData: ValueData[] = [];
    const pnlData: ValueData[] = [];
    const irrData: IRRData[] = [];

    // Process and sort data by timestamp
    const sortedMetrics = [...historicalMetrics].sort(
      (a, b) =>
        new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );

    sortedMetrics.forEach((item) => {
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

    // Add IRR series with separate scale
    const irrSeries = chart.addSeries(LineSeries, {
      color: "#FF6B6B",
      lineWidth: 2,
      title: "IRR (%)",
      priceScaleId: "right",
      priceFormat: {
        type: "percent",
        precision: 2,
      },
      visible: true,
    });
    irrSeries.setData(irrData);

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
      borderColor: "#FF6B6B",
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
  }, [historicalMetrics, leftAxisSelection]);

  // Handle loading states
  if (isLoading) return <div>Loading historical metrics chart...</div>;
  if (error) return <div>Error loading historical metrics chart</div>;

  if (historicalMetrics.length === 0 && !isLoading && !error) {
    return (
      <Paper p="xl" withBorder>
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
          <Box mb="sm">
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
          </Box>
          <Box
            ref={chartContainerRef}
            style={{ width: "100%", minHeight: 500, minWidth: 0 }}
          />
        </Stack>
        <Text size="xs" c="dimmed" mt="xs">
          {leftAxisSelection === "MV" ? "Market Value" : "P&L"} shown as{" "}
          {leftAxisSelection === "MV" ? "blue" : "green"} area (left scale), IRR
          shown as red line (right scale)
        </Text>
      </Paper>
    </Box>
  );
};

export default MetricsChart;
