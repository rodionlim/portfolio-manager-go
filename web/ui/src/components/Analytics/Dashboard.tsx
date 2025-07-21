import React, { useState, useEffect, useRef } from "react";
import {
  Card,
  Text,
  Group,
  Stack,
  Badge,
  Loader,
  Alert,
  Container,
  Grid,
  Title,
  Box,
  Progress,
} from "@mantine/core";
import {
  IconAlertCircle,
  IconTrendingUp,
  IconTrendingDown,
} from "@tabler/icons-react";
import {
  createChart,
  ColorType,
  UTCTimestamp,
  IChartApi,
  LineSeries,
  AreaSeries,
} from "lightweight-charts";
import { getUrl } from "../../utils/url";

interface InterestRateData {
  date: string;
  rate: number;
  tenor: string;
  country: string;
  rate_type: string;
}

interface ChartData {
  time: UTCTimestamp;
  value: number;
}

const Dashboard: React.FC = () => {
  const [interestRates, setInterestRates] = useState<InterestRateData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentRate, setCurrentRate] = useState<number | null>(null);
  const [rateRange, setRateRange] = useState<{
    min: number;
    max: number;
  } | null>(null);
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);

  // Calculate 3 years ago from current date
  const getThreeYearsAgoPoints = () => {
    const now = new Date();
    const threeYearsAgo = new Date(
      now.getFullYear() - 3,
      now.getMonth(),
      now.getDate()
    );
    const daysDiff = Math.floor(
      (now.getTime() - threeYearsAgo.getTime()) / (1000 * 60 * 60 * 24)
    );
    return Math.min(daysDiff, 1095); // Cap at 3 years worth of daily data
  };

  useEffect(() => {
    const fetchInterestRates = async () => {
      try {
        setLoading(true);
        const points = getThreeYearsAgoPoints();
        const response = await fetch(
          getUrl(`/api/v1/mdata/interest-rates/SG?points=${points}`)
        );

        if (!response.ok) {
          throw new Error(
            `Failed to fetch interest rates: ${response.statusText}`
          );
        }

        const data: InterestRateData[] = await response.json();
        setInterestRates(data);

        if (data.length > 0) {
          // Sort by date to ensure proper ordering
          const sortedData = data.sort(
            (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
          );

          // Get current rate (most recent)
          const latest = sortedData[sortedData.length - 1];
          setCurrentRate(latest.rate);

          // Calculate overall range
          const rates = sortedData.map((d) => d.rate);
          const min = Math.min(...rates);
          const max = Math.max(...rates);
          setRateRange({ min, max });
        }
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch interest rates"
        );
      } finally {
        setLoading(false);
      }
    };

    fetchInterestRates();
  }, []);

  useEffect(() => {
    if (!chartContainerRef.current || interestRates.length === 0) return;

    // Convert data for lightweight-charts
    const chartData: ChartData[] = interestRates
      .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
      .map((item) => ({
        time: (new Date(item.date).getTime() / 1000) as UTCTimestamp,
        value: item.rate,
      }));

    // Create chart
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: "transparent" },
        textColor: "#333333",
      },
      grid: {
        vertLines: { color: "rgba(197, 203, 206, 0.3)" },
        horzLines: { color: "rgba(197, 203, 206, 0.3)" },
      },
      width: chartContainerRef.current.clientWidth,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        tickMarkFormatter: (time: UTCTimestamp) => {
          const date = new Date(time * 1000);
          return date.toLocaleDateString("en-US", {
            month: "short",
            year: "2-digit",
          });
        },
        visible: true,
        borderVisible: true,
        borderColor: "#333333",
      },
      rightPriceScale: {
        visible: true,
        borderColor: "#333333",
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
    });

    chartRef.current = chart;

    // Add line series
    const lineSeries = chart.addSeries(LineSeries, {
      color: "#2962FF",
      lineWidth: 2,
      priceScaleId: "right",
    });

    lineSeries.setData(chartData);

    // Add area series for range visualization
    const areaSeries = chart.addSeries(AreaSeries, {
      topColor: "rgba(41, 98, 255, 0.2)",
      bottomColor: "rgba(41, 98, 255, 0.05)",
      lineColor: "rgba(41, 98, 255, 0.5)",
      lineWidth: 1,
      priceScaleId: "right",
      lastValueVisible: false,
    });

    areaSeries.setData(chartData);

    // Add max, min and current price lines with labels
    if (rateRange) {
      // Max value line
      lineSeries.createPriceLine({
        price: rateRange.max,
        color: "#ff6b6b",
        lineWidth: 2,
        lineStyle: 2, // Dashed line
        axisLabelVisible: true,
        title: `Max: ${rateRange.max.toFixed(2)}%`,
      });

      // Min value line
      lineSeries.createPriceLine({
        price: rateRange.min,
        color: "#51cf66",
        lineWidth: 2,
        lineStyle: 2, // Dashed line
        axisLabelVisible: true,
        title: `Min: ${rateRange.min.toFixed(2)}%`,
      });

      // Current value line
      if (currentRate) {
        lineSeries.createPriceLine({
          price: currentRate,
          color: "#ffd43b",
          lineWidth: 2,
          lineStyle: 0, // Solid line
          axisLabelVisible: false, // Hide axis label
          title: `Current: ${currentRate.toFixed(2)}%`,
        });
      }
    }

    // Fit content
    chart.timeScale().fitContent();

    // Cleanup function
    return () => {
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
    };
  }, [interestRates, rateRange, currentRate]);

  const getTrendIcon = () => {
    if (!interestRates.length || interestRates.length < 2) return null;

    const sortedRates = [...interestRates].sort(
      (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
    );

    // Get the most recent data point and one year ago
    const latest = sortedRates[sortedRates.length - 1];
    const oneYearAgo = new Date();
    oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

    // Find the rate closest to one year ago
    const oneYearAgoData = sortedRates.find((rate) => {
      const rateDate = new Date(rate.date);
      return (
        Math.abs(rateDate.getTime() - oneYearAgo.getTime()) <
        30 * 24 * 60 * 60 * 1000
      ); // Within 30 days
    });

    if (!oneYearAgoData) {
      // Fallback to previous day comparison if no one-year data
      const previous = sortedRates[sortedRates.length - 2];
      if (latest.rate > previous.rate) {
        return <IconTrendingUp size={16} color="red" />;
      } else if (latest.rate < previous.rate) {
        return <IconTrendingDown size={16} color="green" />;
      }
      return null;
    }

    if (latest.rate > oneYearAgoData.rate) {
      return <IconTrendingUp size={16} color="red" />;
    } else if (latest.rate < oneYearAgoData.rate) {
      return <IconTrendingDown size={16} color="green" />;
    }
    return null;
  };

  const getCurrentRatePosition = () => {
    if (!currentRate || !rateRange) return 0;
    return (
      ((currentRate - rateRange.min) / (rateRange.max - rateRange.min)) * 100
    );
  };

  if (loading) {
    return (
      <Container size="lg" p="md">
        <Stack align="center" gap="md">
          <Loader size="lg" />
          <Text>Loading market dashboard...</Text>
        </Stack>
      </Container>
    );
  }

  if (error) {
    return (
      <Container size="lg" p="md">
        <Alert icon={<IconAlertCircle size="1rem" />} title="Error" color="red">
          {error}
        </Alert>
      </Container>
    );
  }

  return (
    <Container size="lg" p="md">
      <Stack gap="lg">
        <Title order={2}>Market Dashboard</Title>

        {/* Key Metrics Cards */}
        <Grid>
          <Grid.Col span={{ base: 12, sm: 6, md: 4 }}>
            <Card shadow="sm" padding="lg" radius="md" withBorder>
              <Group justify="space-between" mb="xs">
                <Text fw={500}>Current SG Rate</Text>
                {getTrendIcon()}
              </Group>
              <Text size="xl" fw={700} c="blue">
                {currentRate ? `${currentRate.toFixed(2)}%` : "N/A"}
              </Text>
              <Text size="xs" c="dimmed">
                SORA Overnight Rate
              </Text>
            </Card>
          </Grid.Col>

          <Grid.Col span={{ base: 12, sm: 6, md: 4 }}>
            <Card shadow="sm" padding="lg" radius="md" withBorder>
              <Text fw={500} mb="xs">
                3-Year Range
              </Text>
              <Text size="xl" fw={700} c="orange">
                {rateRange
                  ? `${rateRange.min.toFixed(2)}% - ${rateRange.max.toFixed(
                      2
                    )}%`
                  : "N/A"}
              </Text>
              <Text size="xs" c="dimmed">
                Min - Max Range
              </Text>
            </Card>
          </Grid.Col>

          <Grid.Col span={{ base: 12, sm: 12, md: 4 }}>
            <Card shadow="sm" padding="lg" radius="md" withBorder>
              <Text fw={500} mb="xs">
                Data Points
              </Text>
              <Text size="xl" fw={700} c="green">
                {interestRates.length}
              </Text>
              <Text size="xs" c="dimmed">
                Past 3 Years
              </Text>
            </Card>
          </Grid.Col>
        </Grid>

        {/* Current Rate Position Indicator */}
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Group justify="space-between" mb="md">
            <Text fw={500} size="lg">
              Rate Position in 3-Year Range
            </Text>
            <Badge color="yellow" variant="light">
              Current: {currentRate ? `${currentRate.toFixed(2)}%` : "N/A"}
            </Badge>
          </Group>

          <Stack gap="xs">
            <Progress
              value={getCurrentRatePosition()}
              size="lg"
              color="orange"
              bg="gray.1"
              radius="md"
            />
            <Group justify="space-between">
              <Text size="xs" c="dimmed">
                Min: {rateRange ? `${rateRange.min.toFixed(2)}%` : "N/A"}
              </Text>
              <Text size="xs" c="dimmed">
                Max: {rateRange ? `${rateRange.max.toFixed(2)}%` : "N/A"}
              </Text>
            </Group>
          </Stack>
        </Card>

        {/* Interest Rate Trend Chart */}
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Group justify="space-between" mb="md">
            <Text fw={500} size="lg">
              SG Interest Rate Trends (3 Years)
            </Text>
            <Badge color="blue" variant="light">
              SORA
            </Badge>
          </Group>

          <Box h={{ base: 400, sm: 350 }} style={{ position: "relative" }}>
            <div
              ref={chartContainerRef}
              style={{
                width: "100%",
                height: "100%",
                position: "absolute",
                top: 0,
                left: 0,
              }}
            />
          </Box>

          <Group justify="space-between" mt="md">
            <Group gap="lg">
              <Group gap="xs">
                <Box w={12} h={12} style={{ backgroundColor: "#2962FF" }} />
                <Text size="xs" c="dimmed">
                  Interest Rate Trend
                </Text>
              </Group>
              <Group gap="xs">
                <Box
                  w={12}
                  h={2}
                  style={{ backgroundColor: "#ff6b6b", borderStyle: "dashed" }}
                />
                <Text size="xs" c="dimmed">
                  Max
                </Text>
              </Group>
              <Group gap="xs">
                <Box
                  w={12}
                  h={2}
                  style={{ backgroundColor: "#51cf66", borderStyle: "dashed" }}
                />
                <Text size="xs" c="dimmed">
                  Min
                </Text>
              </Group>
              <Group gap="xs">
                <Box w={12} h={2} style={{ backgroundColor: "#ffd43b" }} />
                <Text size="xs" c="dimmed">
                  Current
                </Text>
              </Group>
            </Group>
            <Text size="xs" c="dimmed">
              Past {interestRates.length} data points
            </Text>
          </Group>
        </Card>
      </Stack>
    </Container>
  );
};

export default Dashboard;
