import React, { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Badge,
  Button,
  Card,
  Divider,
  Group,
  LoadingOverlay,
  NumberInput,
  Paper,
  SegmentedControl,
  Select,
  SimpleGrid,
  Stack,
  Text,
  Title,
  useMantineColorScheme,
} from "@mantine/core";
import { DateInput } from "@mantine/dates";
import { IconCalculator, IconInfoCircle } from "@tabler/icons-react";
import { notifications } from "@mantine/notifications";
import { useSelector } from "react-redux";

import { RootState } from "../../store";
import { getUrl } from "../../utils/url";

interface AssetPriceResponse {
  ticker: string;
  price: number;
  currency: string;
  timestamp: number;
}

interface OptionPricingResponse {
  ticker: string;
  style: string;
  pricingModel: string;
  optionType: string;
  spot: number;
  strike: number;
  expiry: string;
  timeToExpiryYears: number;
  rate: number;
  rateSource?: string;
  rateCurveDate?: string;
  dividendYield: number;
  premium?: number;
  volatility: number;
  volatilitySource: string;
  volatilityLookbackDays?: number;
  npv: number;
  delta: number;
  gamma: number;
  theta: number;
}

const historicalVolatilityLookbackOptions = [
  { value: "30", label: "30D" },
  { value: "60", label: "60D" },
  { value: "180", label: "180D" },
  { value: "360", label: "360D" },
];

const getNextWeekFriday = () => {
  const now = new Date();
  const nextWeekMonday = new Date(now);
  const daysUntilNextMonday = (8 - now.getDay()) % 7 || 7;
  nextWeekMonday.setDate(now.getDate() + daysUntilNextMonday);

  const nextWeekFriday = new Date(nextWeekMonday);
  nextWeekFriday.setDate(nextWeekMonday.getDate() + 4);
  nextWeekFriday.setHours(0, 0, 0, 0);
  return nextWeekFriday;
};

const roundToNearestFive = (value: number) => Math.round(value / 5) * 5;

const formatDate = (value: Date | null): string => {
  if (!value) return "";

  const year = value.getFullYear();
  const month = `${value.getMonth() + 1}`.padStart(2, "0");
  const day = `${value.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const formatNumber = (value: number, digits = 4) => value.toFixed(digits);

const valueColor = (value: number) => {
  if (value > 0) return "teal";
  if (value < 0) return "red";
  return "gray";
};

const ResultMetric = ({
  label,
  value,
  digits = 4,
  highlighted = false,
}: {
  label: string;
  value: number;
  digits?: number;
  highlighted?: boolean;
}) => {
  return (
    <Paper
      withBorder
      radius="md"
      p="sm"
      bg="var(--mantine-color-body)"
      style={
        highlighted
          ? {
              borderColor: "var(--mantine-color-teal-5)",
              boxShadow: "inset 0 0 0 1px rgba(18, 184, 134, 0.18)",
            }
          : undefined
      }
    >
      <Text size="xs" c="dimmed" tt="uppercase" fw={700}>
        {label}
      </Text>
      <Text size={highlighted ? "md" : "sm"} fw={700} c={valueColor(value)}>
        {formatNumber(value, digits)}
      </Text>
    </Paper>
  );
};

const OptionPricer: React.FC = () => {
  const { colorScheme } = useMantineColorScheme();
  const { data: referenceData, status: referenceDataStatus } = useSelector(
    (state: RootState) => state.referenceData,
  );

  const tickerOptions = useMemo(() => {
    if (!referenceData) return [];

    return Object.values(referenceData)
      .filter((item) => item.asset_class === "eq")
      .sort((left, right) => left.id.localeCompare(right.id))
      .map((item) => ({
        value: item.id,
        label: `${item.id} - ${item.name}`,
      }));
  }, [referenceData]);

  const [selectedTicker, setSelectedTicker] = useState<string | null>(null);
  const [optionType, setOptionType] = useState<string>("call");
  const [expiry, setExpiry] = useState<Date | null>(() => getNextWeekFriday());
  const [spot, setSpot] = useState<number | string>("");
  const [strike, setStrike] = useState<number | string>("");
  const [rate, setRate] = useState<number | string>("");
  const [dividendYield, setDividendYield] = useState<number | string>(0);
  const [premium, setPremium] = useState<number | string>("");
  const [volatility, setVolatility] = useState<number | string>("");
  const [
    historicalVolatilityLookbackDays,
    setHistoricalVolatilityLookbackDays,
  ] = useState<string>("180");
  const [spotLoading, setSpotLoading] = useState(false);
  const [pricingLoading, setPricingLoading] = useState(false);
  const [pricingResult, setPricingResult] =
    useState<OptionPricingResponse | null>(null);

  const selectedReference = useMemo(() => {
    if (!referenceData || !selectedTicker) return null;
    return referenceData[selectedTicker] ?? null;
  }, [referenceData, selectedTicker]);

  useEffect(() => {
    if (!selectedTicker) {
      setPricingResult(null);
      return;
    }

    let cancelled = false;

    const fetchSpotPrice = async () => {
      setSpotLoading(true);
      try {
        const response = await fetch(
          getUrl(`/api/v1/mdata/price/${encodeURIComponent(selectedTicker)}`),
        );
        if (!response.ok) {
          throw new Error(`Failed to fetch spot price for ${selectedTicker}`);
        }

        const data: AssetPriceResponse = await response.json();
        if (!cancelled) {
          setSpot(data.price);
          setStrike(roundToNearestFive(data.price));
        }
      } catch (error: any) {
        if (!cancelled) {
          notifications.show({
            title: "Spot price fetch failed",
            message: error.message,
            color: "red",
          });
        }
      } finally {
        if (!cancelled) {
          setSpotLoading(false);
        }
      }
    };

    fetchSpotPrice();

    return () => {
      cancelled = true;
    };
  }, [selectedTicker]);

  const cardBackground =
    colorScheme === "dark"
      ? "linear-gradient(145deg, rgba(13, 48, 61, 0.92), rgba(24, 27, 31, 0.98))"
      : "linear-gradient(145deg, rgba(247, 251, 255, 0.98), rgba(242, 247, 236, 0.98))";

  const resultsBackground =
    colorScheme === "dark"
      ? "rgba(15, 26, 33, 0.88)"
      : "rgba(255, 255, 255, 0.92)";

  const toNumber = (value: number | string): number => Number(value);
  const toOptionalNumber = (value: number | string): number | null => {
    if (value === "") return null;
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  };

  const handleCalculate = async () => {
    if (!selectedTicker) {
      notifications.show({
        title: "Missing ticker",
        message: "Select an equity ticker first.",
        color: "red",
      });
      return;
    }

    if (!expiry) {
      notifications.show({
        title: "Missing expiry",
        message: "Choose an expiry date.",
        color: "red",
      });
      return;
    }

    if (
      spot === "" ||
      Number(spot) <= 0 ||
      strike === "" ||
      Number(strike) <= 0
    ) {
      notifications.show({
        title: "Invalid inputs",
        message: "Spot and strike must be greater than zero.",
        color: "red",
      });
      return;
    }

    const volatilityValue = toOptionalNumber(volatility);
    const premiumValue = toOptionalNumber(premium);
    const rateValue = toOptionalNumber(rate);

    if (volatilityValue !== null && volatilityValue <= 0) {
      notifications.show({
        title: "Invalid volatility",
        message: "Volatility must be greater than zero when provided.",
        color: "red",
      });
      return;
    }

    if (premiumValue !== null && premiumValue <= 0) {
      notifications.show({
        title: "Invalid premium",
        message: "Premium must be greater than zero when provided.",
        color: "red",
      });
      return;
    }

    setPricingLoading(true);
    try {
      const payload: Record<string, unknown> = {
        ticker: selectedTicker,
        optionType,
        spot: toNumber(spot),
        strike: toNumber(strike),
        expiry: formatDate(expiry),
        dividendYield: toNumber(dividendYield),
        volatilityLookbackDays: Number(historicalVolatilityLookbackDays),
      };

      if (rateValue !== null) {
        payload.rate = rateValue;
      }

      if (premiumValue !== null) {
        payload.premium = premiumValue;
      }

      if (volatilityValue !== null) {
        payload.volatility = volatilityValue;
      }

      const response = await fetch(getUrl("/api/v1/options/price"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.message || "Failed to price option");
      }

      setPricingResult(data as OptionPricingResponse);
    } catch (error: any) {
      notifications.show({
        title: "Pricing failed",
        message: error.message,
        color: "red",
      });
    } finally {
      setPricingLoading(false);
    }
  };

  return (
    <Stack gap="lg">
      <Paper
        p="lg"
        radius="xl"
        withBorder
        style={{
          position: "relative",
          overflow: "hidden",
          background: cardBackground,
        }}
      >
        <LoadingOverlay
          visible={spotLoading || pricingLoading}
          overlayProps={{ radius: "lg", blur: 1 }}
        />
        <Stack gap="md">
          <Group justify="space-between" align="flex-start">
            <div>
              <Title order={2}>Option Pricer</Title>
              <Text c="dimmed" mt={4}>
                Black-Scholes approximation for American equity options with
                premium-driven implied volatility and historical fallback.
              </Text>
            </div>
            <Badge variant="light" size="lg">
              American Equity Options
            </Badge>
          </Group>

          <Alert
            icon={<IconInfoCircle size={16} />}
            color="blue"
            variant="light"
          >
            Rates, dividend yield, and volatility are decimals. Example: 5%
            should be entered as 0.05. Leave the rate blank to use the Fed H15
            Treasury curve, with a 3.75% fallback if the fetch fails. Leave
            volatility blank to solve it from premium, or leave both volatility
            and premium blank to use the selected historical window.
          </Alert>
        </Stack>
      </Paper>

      <Card
        withBorder
        radius="xl"
        shadow="sm"
        style={{
          position: "relative",
          background: cardBackground,
          overflow: "hidden",
        }}
      >
        <Stack gap="lg">
          <Group justify="space-between" align="center">
            <div>
              <Title order={4}>Inputs</Title>
              <Text c="dimmed" size="sm">
                Select a ticker, check the fetched spot, and price the option
                without leaving this panel.
              </Text>
            </div>
            {selectedReference ? (
              <Badge variant="outline">{selectedReference.ccy}</Badge>
            ) : null}
          </Group>

          <Text size="xs" c="dimmed">
            Default expiry is next week&apos;s Friday. Strike defaults to the
            fetched ATM level rounded to the nearest $5.
          </Text>

          <SimpleGrid
            cols={{ base: 1, sm: 2, md: 3 }}
            spacing={{ base: "sm", sm: "md" }}
            verticalSpacing={{ base: "sm", sm: "md" }}
          >
            <Select
              label="Equity ticker"
              placeholder={
                referenceDataStatus === "loading"
                  ? "Loading reference data..."
                  : "Select a ticker"
              }
              searchable
              data={tickerOptions}
              value={selectedTicker}
              onChange={setSelectedTicker}
              disabled={referenceDataStatus === "loading"}
              nothingFoundMessage="No equity tickers found"
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <Stack gap={6}>
              <Text size="sm" fw={600}>
                Option type
              </Text>
              <SegmentedControl
                fullWidth
                radius="md"
                value={optionType}
                onChange={(value) => setOptionType(value as "call" | "put")}
                data={[
                  { value: "call", label: "Call" },
                  { value: "put", label: "Put" },
                ]}
              />
            </Stack>

            <NumberInput
              label="Spot price"
              placeholder="Fetched from backend"
              value={spot}
              onChange={setSpot}
              min={0}
              decimalScale={4}
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <NumberInput
              label="Strike"
              value={strike}
              onChange={setStrike}
              min={0}
              decimalScale={4}
              step={5}
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <DateInput
              label="Expiry"
              value={expiry}
              onChange={setExpiry}
              minDate={new Date()}
              valueFormat="YYYY-MM-DD"
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <NumberInput
              label="Observed premium"
              value={premium}
              onChange={setPremium}
              min={0}
              decimalScale={4}
              placeholder="Optional"
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <NumberInput
              label="Annualized volatility"
              value={volatility}
              onChange={setVolatility}
              min={0}
              decimalScale={4}
              placeholder="Optional"
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <Select
              label="Historical vol window"
              data={historicalVolatilityLookbackOptions}
              value={historicalVolatilityLookbackDays}
              onChange={(value) =>
                setHistoricalVolatilityLookbackDays(value ?? "180")
              }
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <NumberInput
              label="Risk-free rate"
              value={rate}
              onChange={setRate}
              decimalScale={4}
              placeholder="Auto from Fed H15"
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />

            <NumberInput
              label="Dividend yield"
              value={dividendYield}
              onChange={setDividendYield}
              min={0}
              decimalScale={4}
              styles={{ label: { marginBottom: 6, fontWeight: 600 } }}
            />
          </SimpleGrid>

          <Group justify="space-between" align="center">
            {selectedReference ? (
              <Text size="sm" c="dimmed">
                {selectedReference.name}
              </Text>
            ) : (
              <Text size="sm" c="dimmed">
                Results will appear here after calculation.
              </Text>
            )}

            <Button
              leftSection={<IconCalculator size={16} />}
              onClick={handleCalculate}
              loading={pricingLoading}
              radius="md"
            >
              Calculate
            </Button>
          </Group>

          {pricingResult ? (
            <>
              <Divider />
              <Paper
                withBorder
                radius="lg"
                p="md"
                style={{ background: resultsBackground }}
              >
                <Stack gap="sm">
                  <Group justify="space-between" align="center">
                    <Group gap="xs">
                      <Badge variant="light" color="teal">
                        {pricingResult.pricingModel}
                      </Badge>
                      <Badge
                        variant="light"
                        color={
                          pricingResult.volatilitySource ===
                          "implied_from_premium"
                            ? "green"
                            : pricingResult.volatilitySource ===
                                "estimated_historical"
                              ? "orange"
                              : "blue"
                        }
                      >
                        {pricingResult.volatilitySource}
                      </Badge>
                    </Group>
                    <Text size="xs" c="dimmed">
                      Theta shown per day
                    </Text>
                  </Group>

                  <SimpleGrid cols={{ base: 2, md: 5 }} spacing="sm">
                    <ResultMetric
                      label="NPV"
                      value={pricingResult.npv}
                      highlighted
                    />
                    <ResultMetric label="Delta" value={pricingResult.delta} />
                    <ResultMetric
                      label="Gamma"
                      value={pricingResult.gamma}
                      digits={6}
                    />
                    <ResultMetric
                      label="Theta"
                      value={pricingResult.theta}
                      digits={6}
                    />
                    <ResultMetric
                      label="Volatility"
                      value={pricingResult.volatility}
                    />
                  </SimpleGrid>

                  <Group gap="md">
                    <Text size="xs" c="dimmed">
                      T: {formatNumber(pricingResult.timeToExpiryYears, 4)}y
                    </Text>
                    <Text size="xs" c="dimmed">
                      r: {formatNumber(pricingResult.rate, 4)}
                    </Text>
                    <Text size="xs" c="dimmed">
                      q: {formatNumber(pricingResult.dividendYield, 4)}
                    </Text>
                  </Group>

                  {pricingResult.rateSource ? (
                    <Text size="xs" c="dimmed">
                      Rate source: {pricingResult.rateSource}
                      {pricingResult.rateCurveDate
                        ? ` (${pricingResult.rateCurveDate})`
                        : ""}
                    </Text>
                  ) : null}

                  {pricingResult.volatilitySource === "estimated_historical" ? (
                    <Text size="xs" c="dimmed">
                      Historical volatility used from the last{" "}
                      {pricingResult.volatilityLookbackDays} calendar days.
                    </Text>
                  ) : null}

                  {pricingResult.volatilitySource === "implied_from_premium" ? (
                    <Text size="xs" c="dimmed">
                      Implied volatility was solved from the premium you
                      entered.
                    </Text>
                  ) : null}
                </Stack>
              </Paper>
            </>
          ) : null}
        </Stack>
      </Card>
    </Stack>
  );
};

export default OptionPricer;
