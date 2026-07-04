import React, { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Alert,
  Box,
  Collapse,
  ActionIcon,
  Group,
  Loader,
  Paper,
  SimpleGrid,
  Table,
  Text,
  Title,
  Tooltip,
  UnstyledButton,
} from "@mantine/core";
import {
  IconArrowsSort,
  IconChevronRight,
  IconPencil,
  IconSortAscending,
  IconSortDescending,
} from "@tabler/icons-react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { RootState } from "../../store";
import { ReferenceDataItem } from "../../types";
import { getUrl } from "../../utils/url";
import { IsSGGovies } from "../../utils/referenceData";

interface Position {
  Ticker: string;
  Name: string;
  Book: string;
  Ccy: string;
  AssetClass: string;
  AssetSubClass: string;
  Mv: number;
  PnL: number;
  FxRate: number;
  Qty: number;
  Px: number;
}

interface SummaryDetail {
  ticker: string;
  name: string;
  ccy: string;
  assetClass: string;
  assetSubClass: string;
  marketValue: number;
  pnl: number;
  dailyPnl: number;
  referenceData?: ReferenceDataItem;
}

interface SummaryRow {
  label: string;
  displayLabel?: string;
  marketValue: number;
  pnl: number;
  dailyPnl: number;
  currency: string;
  details?: SummaryDetail[];
}

interface CachedPrice {
  ticker: string;
  price: number;
  timestamp: string;
}

interface CachedPricesResponse {
  metrics?: {
    timestamp: string;
    metrics: {
      irr: number;
      pricePaid: number;
      mv: number;
      totalDividends: number;
    };
  };
  prices: CachedPrice[];
  pricesPrev2?: CachedPrice[];
  missing?: string[];
}

type SummarySortKey = "marketValue" | "pnl" | "dailyPnl";
type SummarySortDirection = "asc" | "desc";

const formatMoney = (value: number, currency: string) =>
  `${currency} ${value.toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })}`;

const formatAmount = (value: number) =>
  value.toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  });

const formatPercentage = (value: number, total: number) => {
  if (!total) {
    return "0%";
  }
  return `${((value / total) * 100).toFixed(0)}%`;
};

const numberStyle: React.CSSProperties = {
  fontVariantNumeric: "tabular-nums",
  fontFeatureSettings: '"tnum" 1',
};

const inlineMetricStyle: React.CSSProperties = {
  ...numberStyle,
  display: "inline-flex",
  alignItems: "baseline",
  justifyContent: "flex-end",
  gap: 6,
  whiteSpace: "nowrap",
};

const AmountWithShare: React.FC<{
  value: number;
  shareValue?: number;
  total: number;
}> = ({ value, shareValue = value, total }) => (
  <span style={inlineMetricStyle}>
    <span>{formatAmount(value)}</span>
    <Text c="dimmed" component="span" size="xs">
      {formatPercentage(Math.abs(shareValue), total)}
    </Text>
  </span>
);

const categoryLabels: Record<string, string> = {
  consumercyclicals: "Consumer Cyclicals",
  consumernoncyclicals: "Consumer Non-Cyclicals",
  realestate: "Real Estate",
  telecommunications: "Telecommunications",
};

const formatLabel = (value: string) => {
  if (value === "Rest") {
    return value;
  }
  return (
    categoryLabels[value] ||
    value
      .replace(/[_-]+/g, " ")
      .replace(/\b\w/g, (char) => char.toUpperCase())
  );
};

const summarizeDetails = (details: SummaryDetail[]) => {
  const totals = new Map<string, SummaryDetail>();

  details.forEach((detail) => {
    const existing =
      totals.get(detail.ticker) ||
      ({
        ticker: detail.ticker,
        name: detail.name,
        ccy: detail.ccy,
        assetClass: detail.assetClass,
        assetSubClass: detail.assetSubClass,
        marketValue: 0,
        pnl: 0,
        dailyPnl: 0,
        referenceData: detail.referenceData,
      } satisfies SummaryDetail);
    existing.marketValue += detail.marketValue;
    existing.pnl += detail.pnl;
    existing.dailyPnl += detail.dailyPnl;
    existing.referenceData = existing.referenceData || detail.referenceData;
    totals.set(detail.ticker, existing);
  });

  return Array.from(totals.values()).sort(
    (left, right) =>
      right.marketValue - left.marketValue ||
      left.ticker.localeCompare(right.ticker),
  );
};

const SummaryTable: React.FC<{
  title: string;
  rows: SummaryRow[];
  showDailyPnl?: boolean;
  defaultSort?: {
    key: SummarySortKey;
    direction: SummarySortDirection;
  };
  expandable?: boolean;
  onEditReferenceData?: (detail: SummaryDetail) => void;
  onViewPositionsByBook?: (book: string) => void;
}> = ({
  title,
  rows,
  showDailyPnl = false,
  defaultSort = null,
  expandable = false,
  onEditReferenceData,
  onViewPositionsByBook,
}) => {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [sort, setSort] = useState<{
    key: SummarySortKey;
    direction: SummarySortDirection;
  } | null>(defaultSort);
  const totalMarketValue = rows.reduce((sum, row) => sum + row.marketValue, 0);
  const totalPnl = rows.reduce((sum, row) => sum + Math.abs(row.pnl), 0);
  const totalDailyPnl = rows.reduce(
    (sum, row) => sum + Math.abs(row.dailyPnl),
    0,
  );
  const detailColSpan = showDailyPnl ? 4 : 3;
  const sortedRows = useMemo(() => {
    if (!sort) {
      return rows;
    }

    return [...rows].sort((left, right) => {
      const value =
        sort.direction === "desc"
          ? right[sort.key] - left[sort.key]
          : left[sort.key] - right[sort.key];
      return (
        value ||
        (left.displayLabel || left.label).localeCompare(
          right.displayLabel || right.label,
        )
      );
    });
  }, [rows, sort]);

  const toggleRow = (label: string) => {
    setExpandedRows((current) => {
      const next = new Set(current);
      if (next.has(label)) {
        next.delete(label);
      } else {
        next.add(label);
      }
      return next;
    });
  };

  const toggleSort = (key: SummarySortKey) => {
    setSort((current) => {
      if (!current || current.key !== key) {
        return { key, direction: "desc" };
      }
      return {
        key,
        direction: current.direction === "desc" ? "asc" : "desc",
      };
    });
  };

  const SortHeader: React.FC<{ label: string; sortKey: SummarySortKey }> = ({
    label,
    sortKey,
  }) => {
    const isActive = sort?.key === sortKey;
    const Icon = !isActive
      ? IconArrowsSort
      : sort.direction === "desc"
        ? IconSortDescending
        : IconSortAscending;

    return (
      <UnstyledButton
        onClick={() => toggleSort(sortKey)}
        style={{ width: "100%" }}
      >
        <Group gap={4} justify="flex-end" wrap="nowrap">
          <Text fw={600} size="sm">
            {label}
          </Text>
          <Icon size={14} />
        </Group>
      </UnstyledButton>
    );
  };

  return (
    <Paper
      withBorder
      p={{ base: "xs", sm: "md" }}
      radius="sm"
      style={{ minWidth: 0 }}
    >
      <Title order={4} mb="sm" size="h5">
        {title}
      </Title>
      <Table.ScrollContainer minWidth={showDailyPnl ? 560 : 420}>
        <Table
          striped
          highlightOnHover
          withTableBorder
          fz="sm"
          horizontalSpacing="xs"
          verticalSpacing="xs"
        >
        <Table.Thead>
          <Table.Tr>
            <Table.Th />
            <Table.Th style={{ textAlign: "right" }}>
              <SortHeader label="Value" sortKey="marketValue" />
            </Table.Th>
            <Table.Th style={{ textAlign: "right" }}>
              <SortHeader label="P&L" sortKey="pnl" />
            </Table.Th>
            {showDailyPnl ? (
              <Table.Th style={{ textAlign: "right" }}>
                <SortHeader label="Day" sortKey="dailyPnl" />
              </Table.Th>
            ) : null}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sortedRows.map((row) => {
            const isExpanded = expandedRows.has(row.label);
            const details = summarizeDetails(row.details || []);
            const canExpand = expandable && details.length > 0;
            const canViewPositionsByBook = Boolean(onViewPositionsByBook);

            return (
              <React.Fragment key={row.label}>
                <Table.Tr>
                  <Table.Td>
                    {canExpand ? (
                      <UnstyledButton
                        onClick={() => toggleRow(row.label)}
                        style={{ width: "100%" }}
                      >
                        <Group gap="xs" wrap="nowrap">
                          <IconChevronRight
                            size={14}
                            style={{
                              flexShrink: 0,
                              transform: isExpanded
                                ? "rotate(90deg)"
                                : "none",
                            }}
                          />
                          <Text fw={500} size="sm">
                            {row.displayLabel || row.label}
                          </Text>
                        </Group>
                      </UnstyledButton>
                    ) : canViewPositionsByBook ? (
                      <UnstyledButton
                        onClick={() => onViewPositionsByBook?.(row.label)}
                        style={{ width: "100%" }}
                      >
                        <Group gap={4} wrap="nowrap">
                          <Text fw={500} size="sm" c="blue">
                            {row.displayLabel || row.label}
                          </Text>
                          <IconChevronRight size={14} style={{ flexShrink: 0 }} />
                        </Group>
                      </UnstyledButton>
                    ) : (
                      <Text fw={500} size="sm">
                        {row.displayLabel || row.label}
                      </Text>
                    )}
                  </Table.Td>
                  <Table.Td style={{ textAlign: "right" }}>
                    <AmountWithShare
                      value={row.marketValue}
                      total={totalMarketValue}
                    />
                  </Table.Td>
                  <Table.Td
                    style={{
                      textAlign: "right",
                      color:
                        row.pnl < 0
                          ? "var(--mantine-color-red-6)"
                          : "var(--mantine-color-green-6)",
                    }}
                  >
                    <AmountWithShare value={row.pnl} total={totalPnl} />
                  </Table.Td>
                  {showDailyPnl ? (
                    <Table.Td
                      style={{
                        textAlign: "right",
                        color:
                          row.dailyPnl < 0
                            ? "var(--mantine-color-red-6)"
                            : "var(--mantine-color-green-6)",
                      }}
                    >
                      <AmountWithShare
                        value={row.dailyPnl}
                        total={totalDailyPnl}
                      />
                    </Table.Td>
                  ) : null}
                </Table.Tr>
                {canExpand ? (
                  <Table.Tr>
                    <Table.Td colSpan={detailColSpan} p={0}>
                      <Collapse in={isExpanded}>
                        <Box p="xs">
                          <Table.ScrollContainer
                            minWidth={showDailyPnl ? 600 : 480}
                          >
                            <Table
                              withColumnBorders
                              fz="sm"
                              horizontalSpacing="xs"
                              verticalSpacing="xs"
                            >
                            <Table.Thead>
                              <Table.Tr>
                                <Table.Th>Ticker</Table.Th>
                                <Table.Th>Name</Table.Th>
                                <Table.Th style={{ width: 36 }} />
                                <Table.Th style={{ textAlign: "right" }}>
                                  Value
                                </Table.Th>
                                <Table.Th style={{ textAlign: "right" }}>
                                  P&amp;L
                                </Table.Th>
                                {showDailyPnl ? (
                                  <Table.Th style={{ textAlign: "right" }}>
                                    Day
                                  </Table.Th>
                                ) : null}
                              </Table.Tr>
                            </Table.Thead>
                            <Table.Tbody>
                              {details.map((detail) => (
                                <Table.Tr key={detail.ticker}>
                                  <Table.Td>{detail.ticker}</Table.Td>
                                  <Table.Td>{detail.name}</Table.Td>
                                  <Table.Td>
                                    {onEditReferenceData ? (
                                      <Tooltip label="Update reference data">
                                        <ActionIcon
                                          variant="subtle"
                                          size="sm"
                                          aria-label={`Update reference data for ${detail.ticker}`}
                                          onClick={() =>
                                            onEditReferenceData(detail)
                                          }
                                        >
                                          <IconPencil size={14} />
                                        </ActionIcon>
                                      </Tooltip>
                                    ) : null}
                                  </Table.Td>
                                  <Table.Td
                                    style={{
                                      textAlign: "right",
                                      ...numberStyle,
                                    }}
                                  >
                                    {formatAmount(detail.marketValue)}
                                  </Table.Td>
                                  <Table.Td
                                    style={{
                                      textAlign: "right",
                                      ...numberStyle,
                                      color:
                                        detail.pnl < 0
                                          ? "var(--mantine-color-red-6)"
                                          : "var(--mantine-color-green-6)",
                                    }}
                                  >
                                    {formatAmount(detail.pnl)}
                                  </Table.Td>
                                  {showDailyPnl ? (
                                    <Table.Td
                                      style={{
                                        textAlign: "right",
                                        ...numberStyle,
                                        color:
                                          detail.dailyPnl < 0
                                            ? "var(--mantine-color-red-6)"
                                            : "var(--mantine-color-green-6)",
                                      }}
                                    >
                                      {formatAmount(detail.dailyPnl)}
                                    </Table.Td>
                                  ) : null}
                                </Table.Tr>
                              ))}
                            </Table.Tbody>
                            </Table>
                          </Table.ScrollContainer>
                        </Box>
                      </Collapse>
                    </Table.Td>
                  </Table.Tr>
                ) : null}
              </React.Fragment>
            );
          })}
        </Table.Tbody>
        </Table>
      </Table.ScrollContainer>
    </Paper>
  );
};

const SummaryView: React.FC = () => {
  const navigate = useNavigate();
  const refData = useSelector((state: RootState) => state.referenceData.data);
  const {
    data: positions = [],
    isLoading,
    error,
  } = useQuery<Position[], Error>({
    queryKey: ["positions"],
    queryFn: async () => {
      const resp = await fetch(getUrl("/api/v1/portfolio/positions"));
      if (!resp.ok) {
        const errorPayload = await resp
          .json()
          .catch(() => ({ message: "Failed to get positions" }));
        throw new Error(errorPayload?.message || "Failed to get positions");
      }
      return resp.json();
    },
    retry: false,
    refetchOnWindowFocus: false,
  });

  const uniqueTickers = useMemo(() => {
    const tickers = new Set<string>();
    positions.forEach((position) => {
      if (position.Ticker && position.Qty !== 0) {
        tickers.add(position.Ticker);
      }
    });
    return Array.from(tickers);
  }, [positions]);

  const { data: cachedPricesData } = useQuery<CachedPricesResponse | null>({
    queryKey: ["cachedDailyPrices", uniqueTickers],
    queryFn: async () => {
      if (uniqueTickers.length === 0) return null;
      const resp = await fetch(getUrl("/api/v1/historical/prices/cached"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ tickers: uniqueTickers }),
      });
      if (!resp.ok) {
        return null;
      }
      return resp.json();
    },
    retry: false,
    refetchOnWindowFocus: false,
    enabled: uniqueTickers.length > 0,
  });

  const hasCachedMetrics = Boolean(cachedPricesData?.metrics);

  const cachedPriceMap = useMemo(() => {
    const map = new Map<string, CachedPrice>();
    if (!hasCachedMetrics || !cachedPricesData?.prices) {
      return map;
    }
    cachedPricesData.prices.forEach((price) => {
      map.set(price.ticker, price);
    });
    return map;
  }, [cachedPricesData, hasCachedMetrics]);

  const overallPnl = useMemo(
    () =>
      positions.reduce(
        (sum, position) => sum + position.PnL * (position.FxRate || 1),
        0,
      ),
    [positions],
  );
  const overallMarketValue = useMemo(
    () =>
      positions.reduce(
        (sum, position) => sum + position.Mv * (position.FxRate || 1),
        0,
      ),
    [positions],
  );

  const { byBook, byCurrency, byCategory, byAssetSubClass } = useMemo(() => {
    const bookTotals = new Map<string, SummaryRow>();
    const currencyTotals = new Map<string, SummaryRow>();
    const categoryTotals = new Map<string, SummaryRow>();
    const assetSubClassTotals = new Map<string, SummaryRow>();

    const addSGDTotal = (
      totals: Map<string, SummaryRow>,
      label: string,
      marketValue: number,
      pnl: number,
      dailyPnl: number,
    ) => {
      const row =
        totals.get(label) ||
        ({
          label,
          marketValue: 0,
          pnl: 0,
          dailyPnl: 0,
          currency: "SGD",
          details: [],
        } satisfies SummaryRow);
      row.marketValue += marketValue;
      row.pnl += pnl;
      row.dailyPnl += dailyPnl;
      totals.set(label, row);
    };

    positions.forEach((position) => {
      const fxRate = position.FxRate || 1;
      const book = position.Book || "Unknown";
      const currency = position.Ccy || "SGD";
      const marketValue = position.Mv * fxRate;
      const pnl = position.PnL * fxRate;
      const previousPrice = cachedPriceMap.get(position.Ticker)?.price;
      const dailyPnl =
        hasCachedMetrics &&
        previousPrice !== undefined &&
        previousPrice > 0 &&
        position.Px > 0 &&
        position.Qty !== 0
          ? (position.Px - previousPrice) * position.Qty * fxRate
          : 0;
      const ref =
        refData?.[position.Ticker] ||
        Object.values(refData || {}).find(
          (item) =>
            item.id === position.Ticker ||
            item.underlying_ticker === position.Ticker ||
            item.yahoo_ticker === position.Ticker,
        );
      const assetSubClass =
        position.AssetSubClass || ref?.asset_sub_class || "Unknown";
      const isGovies =
        assetSubClass === "govies" || IsSGGovies(position.Ticker);
      const category = isGovies ? "Govies" : ref?.category || "Rest";
      const detail = {
        ticker: position.Ticker,
        name: position.Name || ref?.name || position.Ticker,
        ccy: position.Ccy,
        assetClass: position.AssetClass || ref?.asset_class || "",
        assetSubClass,
        marketValue,
        pnl,
        dailyPnl,
        referenceData: ref,
      } satisfies SummaryDetail;

      addSGDTotal(bookTotals, book, marketValue, pnl, dailyPnl);
      addSGDTotal(currencyTotals, currency, marketValue, pnl, dailyPnl);
      addSGDTotal(categoryTotals, category, marketValue, pnl, dailyPnl);
      categoryTotals.get(category)?.details?.push(detail);
      addSGDTotal(
        assetSubClassTotals,
        assetSubClass,
        marketValue,
        pnl,
        dailyPnl,
      );
    });

    const sortedRows = (totals: Map<string, SummaryRow>) =>
      Array.from(totals.values())
        .map((row) => ({ ...row, displayLabel: formatLabel(row.label) }))
        .sort((a, b) => a.displayLabel.localeCompare(b.displayLabel));

    return {
      byBook: sortedRows(bookTotals),
      byCurrency: sortedRows(currencyTotals),
      byCategory: sortedRows(categoryTotals),
      byAssetSubClass: sortedRows(assetSubClassTotals),
    };
  }, [positions, refData, cachedPriceMap, hasCachedMetrics]);

  const handleEditReferenceData = (detail: SummaryDetail) => {
    const ref = detail.referenceData;

    navigate("/refdata/update_ref_data", {
      state: {
        id: ref?.id || detail.ticker,
        name: ref?.name || detail.name,
        underlying_ticker: ref?.underlying_ticker || detail.ticker,
        yahoo_ticker: ref?.yahoo_ticker || "",
        google_ticker: ref?.google_ticker || "",
        dividends_sg_ticker: ref?.dividends_sg_ticker || "",
        nasdaq_ticker: ref?.nasdaq_ticker || "",
        barchart_ticker: ref?.barchart_ticker || "",
        asset_class: ref?.asset_class || detail.assetClass,
        asset_sub_class: ref?.asset_sub_class || detail.assetSubClass,
        category: ref?.category || "",
        sub_category: ref?.sub_category || "",
        ccy: ref?.ccy || detail.ccy,
        domicile: ref?.domicile || "",
        coupon_rate: ref?.coupon_rate || 0,
        maturity_date: ref?.maturity_date || "",
        strike_price: ref?.strike_price || 0,
        call_put: ref?.call_put || "",
      },
    });
  };

  const handleViewPositionsByBook = (book: string) => {
    navigate("/positions", {
      state: { book },
    });
  };

  if (isLoading) {
    return (
      <Group justify="center" py="xl">
        <Loader size="sm" />
        <Text c="dimmed">Loading summary</Text>
      </Group>
    );
  }

  if (error) {
    return (
      <Alert color="red" title="Error loading summary">
        {error.message}
      </Alert>
    );
  }

  if (!positions.length) {
    return (
      <Box py="md">
        <Text c="dimmed">No positions found.</Text>
      </Box>
    );
  }

  return (
    <Box style={{ minWidth: 0 }}>
      <Title order={2} mb="md">
        Summary
      </Title>
      <Paper withBorder p="md" radius="sm" mb="md">
        <SimpleGrid cols={{ base: 1, xs: 2 }} spacing="md">
          <Box>
            <Text c="dimmed" size="sm" fw={600}>
              Market Value
            </Text>
            <Text fw={700} size="xl" style={{ lineHeight: 1.2, ...numberStyle }}>
              {formatMoney(overallMarketValue, "SGD")}
            </Text>
          </Box>
          <Box>
            <Text c="dimmed" size="sm" fw={600}>
              Overall P&amp;L
            </Text>
            <Text
              fw={700}
              size="xl"
              c={overallPnl < 0 ? "red" : "green"}
              style={{ lineHeight: 1.2, ...numberStyle }}
            >
              {formatMoney(overallPnl, "SGD")}
            </Text>
          </Box>
        </SimpleGrid>
      </Paper>
      <SimpleGrid cols={{ base: 1, lg: 2 }} spacing="md" style={{ minWidth: 0 }}>
        <SummaryTable
          title="By Book"
          rows={byBook}
          showDailyPnl={hasCachedMetrics}
          onViewPositionsByBook={handleViewPositionsByBook}
        />
        <SummaryTable
          title="By Currency"
          rows={byCurrency}
          showDailyPnl={hasCachedMetrics}
        />
        <SummaryTable
          title="By Category"
          rows={byCategory}
          showDailyPnl={hasCachedMetrics}
          defaultSort={{ key: "marketValue", direction: "desc" }}
          expandable
          onEditReferenceData={handleEditReferenceData}
        />
        <SummaryTable
          title="By Asset Subclass"
          rows={byAssetSubClass}
          showDailyPnl={hasCachedMetrics}
        />
      </SimpleGrid>
    </Box>
  );
};

export default SummaryView;
