import React, { useEffect, useMemo, useState } from "react";
import {
  ActionIcon,
  Autocomplete,
  Badge,
  Button,
  Card,
  Group,
  NumberInput,
  Select,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconPlus, IconTrash } from "@tabler/icons-react";
import { getUrl } from "../../utils/url";
import type { MetricsJob } from "./types";

interface BenchmarkTickerInput {
  id: string;
  ticker: string;
  weight: number;
}

interface BenchmarkCashFlow {
  date: string;
  cash: number;
  ticker: string;
  description: string;
}

interface BenchmarkResponse {
  portfolio_metrics: {
    irr: number;
    pricePaid: number;
    mv: number;
    totalDividends: number;
  };
  benchmark_metrics: {
    irr: number;
    pricePaid: number;
    mv: number;
    fees: number;
  };
  portfolio_irr: number;
  benchmark_irr: number;
  irr_difference: number;
  winner: string;
  benchmark_cash_flows: BenchmarkCashFlow[];
}

const MetricsBenchmark: React.FC = () => {
  const [bookFilter, setBookFilter] = useState("");
  const [mode, setMode] = useState("match_trades");
  const [notional, setNotional] = useState<number | "">(100000);
  const [costPct, setCostPct] = useState<number | "">(0.12);
  const [costAbsolute, setCostAbsolute] = useState<number | "">(10.9);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<BenchmarkResponse | null>(null);
  const [bookOptions, setBookOptions] = useState<string[]>([]);

  const [tickers, setTickers] = useState<BenchmarkTickerInput[]>([
    { id: "1", ticker: "ES3.SI", weight: 1 },
  ]);

  const totalWeight = useMemo(() => {
    return tickers.reduce((sum, t) => sum + (Number(t.weight) || 0), 0);
  }, [tickers]);

  useEffect(() => {
    const fetchBooks = async () => {
      try {
        const resp = await fetch(getUrl("/api/v1/historical/metrics/jobs"));
        if (!resp.ok) return;
        const jobs: MetricsJob[] = await resp.json();
        const books = Array.from(
          new Set(jobs.map((j) => j.BookFilter).filter((b) => b && b.trim()))
        ).sort((a, b) => a.localeCompare(b));
        setBookOptions(books);
      } catch (e) {
        // non-blocking
      }
    };

    fetchBooks();
  }, []);

  const handleAddTicker = () => {
    setTickers((prev) => [
      ...prev,
      { id: `${Date.now()}`, ticker: "", weight: 0 },
    ]);
  };

  const handleRemoveTicker = (id: string) => {
    setTickers((prev) => prev.filter((t) => t.id !== id));
  };

  const handleTickerChange = (
    id: string,
    key: "ticker" | "weight",
    value: string | number
  ) => {
    setTickers((prev) =>
      prev.map((t) => (t.id === id ? { ...t, [key]: value } : t))
    );
  };

  const formatPct = (value: number | null | undefined) => {
    if (value === null || value === undefined) return "-";
    return `${(value * 100).toFixed(2)}%`;
  };

  const formatCurrency = (value: number | null | undefined) => {
    if (value === null || value === undefined) return "-";
    return value.toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
  };

  const handleSubmit = async () => {
    const cleanedTickers = tickers
      .filter((t) => t.ticker.trim() !== "" && Number(t.weight) > 0)
      .map((t) => ({
        ticker: t.ticker.trim().toUpperCase(),
        weight: t.weight,
      }));

    if (cleanedTickers.length === 0) {
      notifications.show({
        title: "Missing tickers",
        message: "Please add at least one benchmark ticker.",
        color: "red",
      });
      return;
    }

    if (mode === "buy_at_start" && (!notional || Number(notional) <= 0)) {
      notifications.show({
        title: "Invalid notional",
        message: "Please provide a notional amount for buy_at_start mode.",
        color: "red",
      });
      return;
    }

    setLoading(true);
    try {
      const payload = {
        book_filter: bookFilter.trim(),
        benchmark_cost: {
          pct: (Number(costPct) || 0) / 100,
          absolute: Number(costAbsolute) || 0,
        },
        mode,
        notional: mode === "buy_at_start" ? Number(notional) || 0 : undefined,
        benchmark_tickers: cleanedTickers,
      };

      const resp = await fetch(getUrl("/api/v1/metrics/benchmark"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const json = await resp.json();
      if (!resp.ok) {
        throw new Error(json?.error || "Failed to run benchmark");
      }

      setResult(json);
    } catch (e: any) {
      notifications.show({
        title: "Benchmark failed",
        message: e?.message || "Failed to run benchmark",
        color: "red",
      });
    } finally {
      setLoading(false);
    }
  };

  const excessIrr = result?.irr_difference ?? 0;
  const excessColor = excessIrr > 0 ? "green" : excessIrr < 0 ? "red" : "gray";

  const sortedCashFlows = useMemo(() => {
    if (!result?.benchmark_cash_flows) return [];
    return [...result.benchmark_cash_flows].sort(
      (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
    );
  }, [result]);

  const formatDate = (value: string) => {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString();
  };

  return (
    <Stack gap="lg">
      <Card withBorder shadow="sm">
        <Stack gap="sm">
          <Title order={4}>Benchmark Configuration</Title>
          <Group align="flex-end" wrap="wrap">
            <Autocomplete
              label="Book filter"
              placeholder="optional"
              value={bookFilter}
              onChange={setBookFilter}
              data={bookOptions}
              style={{ width: 200 }}
            />
            <Select
              label="Mode"
              value={mode}
              onChange={(v) => setMode(v || "buy_at_start")}
              data={[
                { value: "buy_at_start", label: "buy_at_start" },
                { value: "match_trades", label: "match_trades" },
              ]}
              style={{ width: 180 }}
            />
            <NumberInput
              label="Notional"
              value={notional}
              onChange={(v) => setNotional(typeof v === "number" ? v : "")}
              disabled={mode !== "buy_at_start"}
              min={0}
              style={{ width: 160 }}
            />
            <NumberInput
              label="Brokerage pct (%)"
              value={costPct}
              onChange={(v) => setCostPct(typeof v === "number" ? v : "")}
              min={0}
              step={0.01}
              decimalScale={4}
              style={{ width: 160 }}
            />
            <NumberInput
              label="Brokerage min ($)"
              value={costAbsolute}
              onChange={(v) => setCostAbsolute(typeof v === "number" ? v : "")}
              min={0}
              style={{ width: 160 }}
            />
            <Button onClick={handleSubmit} loading={loading}>
              Run Benchmark
            </Button>
          </Group>

          <Text size="xs" c="dimmed">
            Excess return is computed as portfolio IRR minus benchmark IRR.
          </Text>
        </Stack>
      </Card>

      <Card withBorder shadow="sm">
        <Stack gap="sm">
          <Group justify="space-between" align="center">
            <Title order={4}>Benchmark Portfolio</Title>
            <Button
              leftSection={<IconPlus size={16} />}
              variant="light"
              onClick={handleAddTicker}
            >
              Add Ticker
            </Button>
          </Group>

          <Table withColumnBorders>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Ticker</Table.Th>
                <Table.Th>Weight</Table.Th>
                <Table.Th>Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {tickers.map((t) => (
                <Table.Tr key={t.id}>
                  <Table.Td>
                    <TextInput
                      placeholder="ES3.SI"
                      value={t.ticker}
                      onChange={(e) =>
                        handleTickerChange(
                          t.id,
                          "ticker",
                          e.currentTarget.value
                        )
                      }
                    />
                  </Table.Td>
                  <Table.Td>
                    <NumberInput
                      value={t.weight}
                      onChange={(v) =>
                        handleTickerChange(t.id, "weight", v || 0)
                      }
                      min={0}
                      step={0.1}
                    />
                  </Table.Td>
                  <Table.Td>
                    <ActionIcon
                      color="red"
                      variant="light"
                      onClick={() => handleRemoveTicker(t.id)}
                    >
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Table.Td>
                </Table.Tr>
              ))}
              {tickers.length === 0 && (
                <Table.Tr>
                  <Table.Td colSpan={3} style={{ textAlign: "center" }}>
                    No benchmark tickers configured
                  </Table.Td>
                </Table.Tr>
              )}
            </Table.Tbody>
          </Table>
          <Text size="xs" c="dimmed">
            Current weight sum: {totalWeight.toFixed(2)} (normalized
            server-side)
          </Text>
        </Stack>
      </Card>

      {result && (
        <Card withBorder shadow="sm">
          <Stack gap="md">
            <Group justify="space-between" align="center">
              <Title order={4}>Benchmark Results</Title>
              <Badge color={excessColor} variant="light">
                {result.winner.toUpperCase()}
              </Badge>
            </Group>

            <Group grow>
              <Card withBorder radius="md" p="md">
                <Text size="sm" c="dimmed">
                  Portfolio IRR
                </Text>
                <Text fw={700} size="lg">
                  {formatPct(result.portfolio_irr)}
                </Text>
              </Card>
              <Card withBorder radius="md" p="md">
                <Text size="sm" c="dimmed">
                  Benchmark IRR
                </Text>
                <Text fw={700} size="lg">
                  {formatPct(result.benchmark_irr)}
                </Text>
              </Card>
              <Card withBorder radius="md" p="md">
                <Text size="sm" c="dimmed">
                  Excess IRR
                </Text>
                <Text fw={700} size="lg" c={excessColor}>
                  {formatPct(excessIrr)}
                </Text>
              </Card>
            </Group>

            <Card withBorder radius="md" p="md">
              <Group justify="space-between" align="center">
                <Text size="sm" c="dimmed">
                  Benchmark Fees
                </Text>
                <Text fw={600}>
                  {formatCurrency(result.benchmark_metrics.fees)}
                </Text>
              </Group>
              <Group justify="space-between" align="center" mt="xs">
                <Text size="sm" c="dimmed">
                  Benchmark MV
                </Text>
                <Text fw={600}>
                  {formatCurrency(result.benchmark_metrics.mv)}
                </Text>
              </Group>
            </Card>

            <Card withBorder radius="md" p="md">
              <Title order={6} mb="xs">
                Excess Return Summary
              </Title>
              <Group justify="space-between" align="center">
                <Text size="sm" c="dimmed">
                  Portfolio vs Benchmark
                </Text>
                <Badge color={excessColor} variant="filled">
                  {formatPct(excessIrr)}
                </Badge>
              </Group>
              <Text size="xs" c="dimmed" mt="xs">
                Positive means the portfolio outperformed the benchmark.
              </Text>
            </Card>

            <Card withBorder radius="md" p="md">
              <Group justify="space-between" align="center" mb="xs">
                <Title order={6}>Benchmark Cash Flows</Title>
                <Text size="xs" c="dimmed">
                  {sortedCashFlows.length} entries
                </Text>
              </Group>
              <Table.ScrollContainer minWidth={700}>
                <Table striped highlightOnHover withColumnBorders>
                  <Table.Thead>
                    <Table.Tr>
                      <Table.Th>Date</Table.Th>
                      <Table.Th>Ticker</Table.Th>
                      <Table.Th>Description</Table.Th>
                      <Table.Th style={{ textAlign: "right" }}>Cash</Table.Th>
                    </Table.Tr>
                  </Table.Thead>
                  <Table.Tbody>
                    {sortedCashFlows.map((cf, idx) => (
                      <Table.Tr key={`${cf.date}-${cf.ticker}-${idx}`}>
                        <Table.Td>{formatDate(cf.date)}</Table.Td>
                        <Table.Td>{cf.ticker}</Table.Td>
                        <Table.Td>{cf.description}</Table.Td>
                        <Table.Td style={{ textAlign: "right" }}>
                          <Text c={cf.cash >= 0 ? "green" : "red"}>
                            {formatCurrency(cf.cash)}
                          </Text>
                        </Table.Td>
                      </Table.Tr>
                    ))}
                    {sortedCashFlows.length === 0 && (
                      <Table.Tr>
                        <Table.Td colSpan={4} style={{ textAlign: "center" }}>
                          No cash flows returned
                        </Table.Td>
                      </Table.Tr>
                    )}
                  </Table.Tbody>
                </Table>
              </Table.ScrollContainer>
            </Card>
          </Stack>
        </Card>
      )}
    </Stack>
  );
};

export default MetricsBenchmark;
