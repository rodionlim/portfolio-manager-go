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
}

interface SummaryDetail {
  ticker: string;
  name: string;
  ccy: string;
  assetClass: string;
  assetSubClass: string;
  marketValue: number;
  pnl: number;
  referenceData?: ReferenceDataItem;
}

interface SummaryRow {
  label: string;
  displayLabel?: string;
  marketValue: number;
  pnl: number;
  currency: string;
  details?: SummaryDetail[];
}

type SummarySortKey = "marketValue" | "pnl";
type SummarySortDirection = "asc" | "desc";

const formatMoney = (value: number, currency: string) =>
  `${currency} ${value.toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  })}`;

const formatPercentage = (value: number, total: number) => {
  if (!total) {
    return "0%";
  }
  return `${((value / total) * 100).toFixed(0)}%`;
};

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
        referenceData: detail.referenceData,
      } satisfies SummaryDetail);
    existing.marketValue += detail.marketValue;
    existing.pnl += detail.pnl;
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
  expandable?: boolean;
  onEditReferenceData?: (detail: SummaryDetail) => void;
}> = ({
  title,
  rows,
  expandable = false,
  onEditReferenceData,
}) => {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [sort, setSort] = useState<{
    key: SummarySortKey;
    direction: SummarySortDirection;
  } | null>(null);
  const totalMarketValue = rows.reduce((sum, row) => sum + row.marketValue, 0);
  const totalPnl = rows.reduce((sum, row) => sum + Math.abs(row.pnl), 0);
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
    <Paper withBorder p="md" radius="sm">
      <Title order={4} mb="sm">
        {title}
      </Title>
      <Table striped highlightOnHover withTableBorder>
        <Table.Thead>
          <Table.Tr>
            <Table.Th />
            <Table.Th style={{ textAlign: "right" }}>
              <SortHeader label="Market Value" sortKey="marketValue" />
            </Table.Th>
            <Table.Th style={{ textAlign: "right" }}>
              <SortHeader label="P&L" sortKey="pnl" />
            </Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sortedRows.map((row) => {
            const isExpanded = expandedRows.has(row.label);
            const details = summarizeDetails(row.details || []);
            const canExpand = expandable && details.length > 0;

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
                              transform: isExpanded
                                ? "rotate(90deg)"
                                : "none",
                            }}
                          />
                          <Text fw={500}>{row.displayLabel || row.label}</Text>
                        </Group>
                      </UnstyledButton>
                    ) : (
                      <Text fw={500}>{row.displayLabel || row.label}</Text>
                    )}
                  </Table.Td>
                  <Table.Td style={{ textAlign: "right" }}>
                    {formatMoney(row.marketValue, row.currency)}{" "}
                    <Text c="dimmed" component="span" size="sm">
                      ({formatPercentage(row.marketValue, totalMarketValue)})
                    </Text>
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
                    {formatMoney(row.pnl, row.currency)}{" "}
                    <Text c="dimmed" component="span" size="sm">
                      ({formatPercentage(Math.abs(row.pnl), totalPnl)})
                    </Text>
                  </Table.Td>
                </Table.Tr>
                {canExpand ? (
                  <Table.Tr>
                    <Table.Td colSpan={3} p={0}>
                      <Collapse in={isExpanded}>
                        <Box p="xs">
                          <Table withColumnBorders>
                            <Table.Thead>
                              <Table.Tr>
                                <Table.Th>Ticker</Table.Th>
                                <Table.Th>Name</Table.Th>
                                <Table.Th style={{ width: 44 }} />
                                <Table.Th style={{ textAlign: "right" }}>
                                  Market Value
                                </Table.Th>
                                <Table.Th style={{ textAlign: "right" }}>
                                  P&amp;L
                                </Table.Th>
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
                                  <Table.Td style={{ textAlign: "right" }}>
                                    {formatMoney(detail.marketValue, "SGD")}
                                  </Table.Td>
                                  <Table.Td
                                    style={{
                                      textAlign: "right",
                                      color:
                                        detail.pnl < 0
                                          ? "var(--mantine-color-red-6)"
                                          : "var(--mantine-color-green-6)",
                                    }}
                                  >
                                    {formatMoney(detail.pnl, "SGD")}
                                  </Table.Td>
                                </Table.Tr>
                              ))}
                            </Table.Tbody>
                          </Table>
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
    ) => {
      const row =
        totals.get(label) ||
        ({
          label,
          marketValue: 0,
          pnl: 0,
          currency: "SGD",
          details: [],
        } satisfies SummaryRow);
      row.marketValue += marketValue;
      row.pnl += pnl;
      totals.set(label, row);
    };

    positions.forEach((position) => {
      const fxRate = position.FxRate || 1;
      const book = position.Book || "Unknown";
      const currency = position.Ccy || "SGD";
      const marketValue = position.Mv * fxRate;
      const pnl = position.PnL * fxRate;
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
        referenceData: ref,
      } satisfies SummaryDetail;

      addSGDTotal(bookTotals, book, marketValue, pnl);
      addSGDTotal(currencyTotals, currency, marketValue, pnl);
      addSGDTotal(categoryTotals, category, marketValue, pnl);
      categoryTotals.get(category)?.details?.push(detail);
      addSGDTotal(assetSubClassTotals, assetSubClass, marketValue, pnl);
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
  }, [positions, refData]);

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
    <Box>
      <Title order={2} mb="md">
        Summary
      </Title>
      <SimpleGrid cols={{ base: 1, md: 2 }} spacing="md">
        <SummaryTable title="By Book" rows={byBook} />
        <SummaryTable title="By Currency" rows={byCurrency} />
        <SummaryTable
          title="By Category"
          rows={byCategory}
          expandable
          onEditReferenceData={handleEditReferenceData}
        />
        <SummaryTable title="By Asset Subclass" rows={byAssetSubClass} />
      </SimpleGrid>
    </Box>
  );
};

export default SummaryView;
