import React, { useState } from "react";
import {
  Box,
  Collapse,
  Group,
  Loader,
  Paper,
  Table,
  Text,
  Title,
  UnstyledButton,
} from "@mantine/core";
import { IconChevronRight } from "@tabler/icons-react";
import type { ReferenceData } from "../../types";
import type { MonthlyPortfolioActivity } from "./monthlyActivity";

const numberStyle: React.CSSProperties = {
  fontVariantNumeric: "tabular-nums",
  fontFeatureSettings: '"tnum" 1',
};

const formatAmount = (value?: number) =>
  value === undefined
    ? "—"
    : value.toLocaleString(undefined, {
        minimumFractionDigits: 0,
        maximumFractionDigits: 0,
      });

const formatNumber = (value: number, maximumFractionDigits = 2) =>
  value.toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits,
  });

const formatDate = (value: string) =>
  new Date(`${value}T00:00:00`).toLocaleDateString();

interface MonthlyPortfolioActivityTableProps {
  rows: MonthlyPortfolioActivity[];
  referenceData: ReferenceData | null;
  isMetricsLoading: boolean;
  areTradesLoading: boolean;
  hasMetricsError: boolean;
  hasTradeError: boolean;
}

const MonthlyPortfolioActivityTable: React.FC<
  MonthlyPortfolioActivityTableProps
> = ({
  rows,
  referenceData,
  isMetricsLoading,
  areTradesLoading,
  hasMetricsError,
  hasTradeError,
}) => {
  const [expandedMonths, setExpandedMonths] = useState<Set<string>>(new Set());

  const toggleMonth = (monthKey: string) => {
    setExpandedMonths((current) => {
      const next = new Set(current);
      if (next.has(monthKey)) {
        next.delete(monthKey);
      } else {
        next.add(monthKey);
      }
      return next;
    });
  };

  return (
    <Paper withBorder p={{ base: "xs", sm: "md" }} radius="sm" mt="md">
      <Group justify="space-between" mb="sm">
        <Box>
          <Title order={4} size="h5">
            Past 12 Months
          </Title>
          <Text c="dimmed" size="xs">
            Monthly portfolio performance and aggregated trading activity
          </Text>
        </Box>
        {isMetricsLoading || areTradesLoading ? <Loader size="sm" /> : null}
      </Group>

      <Table.ScrollContainer minWidth={720}>
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
              <Table.Th>Month</Table.Th>
              <Table.Th style={{ textAlign: "right" }}>MTD P&amp;L</Table.Th>
              <Table.Th style={{ textAlign: "right" }}>Dividends</Table.Th>
              <Table.Th style={{ textAlign: "right" }}>Net CF</Table.Th>
              <Table.Th style={{ textAlign: "right" }}>
                Portfolio Market Value
              </Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {rows.map((row) => {
              const isExpanded = expandedMonths.has(row.monthKey);
              const monthNetPricePaid = row.netCashFlow;

              return (
                <React.Fragment key={row.monthKey}>
                  <Table.Tr>
                    <Table.Td>
                      <UnstyledButton
                        onClick={() => toggleMonth(row.monthKey)}
                        style={{ width: "100%" }}
                        aria-expanded={isExpanded}
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
                          <Text fw={600} size="sm">
                            {row.monthLabel}
                          </Text>
                        </Group>
                      </UnstyledButton>
                    </Table.Td>
                    <Table.Td
                      style={{
                        textAlign: "right",
                        color:
                          row.mtdPnl === undefined
                            ? "var(--mantine-color-dimmed)"
                            : row.mtdPnl < 0
                              ? "var(--mantine-color-red-6)"
                              : "var(--mantine-color-green-6)",
                        ...numberStyle,
                      }}
                    >
                      {formatAmount(row.mtdPnl)}
                    </Table.Td>
                    <Table.Td
                      style={{
                        textAlign: "right",
                        color:
                          row.dividends === undefined
                            ? "var(--mantine-color-dimmed)"
                            : "var(--mantine-color-green-6)",
                        ...numberStyle,
                      }}
                    >
                      {formatAmount(row.dividends)}
                    </Table.Td>
                    <Table.Td
                      style={{
                        textAlign: "right",
                        color:
                          row.netCashFlow < 0
                            ? "var(--mantine-color-red-6)"
                            : undefined,
                        ...numberStyle,
                      }}
                    >
                      {formatAmount(row.netCashFlow)}
                    </Table.Td>
                    <Table.Td style={{ textAlign: "right", ...numberStyle }}>
                      {formatAmount(row.marketValue)}
                    </Table.Td>
                  </Table.Tr>
                  <Table.Tr>
                    <Table.Td colSpan={5} p={0}>
                      <Collapse in={isExpanded}>
                        <Box p="xs">
                          {hasTradeError ? (
                            <Text c="red" size="sm">
                              Trade data is unavailable.
                            </Text>
                          ) : areTradesLoading ? (
                            <Text c="dimmed" size="sm">
                              Loading trades…
                            </Text>
                          ) : row.trades.length === 0 ? (
                            <Text c="dimmed" size="sm">
                              No trades recorded for this month.
                            </Text>
                          ) : (
                            <Table.ScrollContainer minWidth={560}>
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
                                    <Table.Th>Dt</Table.Th>
                                    <Table.Th style={{ textAlign: "right" }}>
                                      Bought
                                    </Table.Th>
                                    <Table.Th style={{ textAlign: "right" }}>
                                      Sold
                                    </Table.Th>
                                    <Table.Th style={{ textAlign: "right" }}>
                                      Net Quantity
                                    </Table.Th>
                                    <Table.Th style={{ textAlign: "right" }}>
                                      Avg Transaction Price
                                    </Table.Th>
                                    <Table.Th style={{ textAlign: "right" }}>
                                      Net Price Paid
                                    </Table.Th>
                                  </Table.Tr>
                                </Table.Thead>
                                <Table.Tbody>
                                  {row.trades.map((trade) => {
                                    const ref = referenceData?.[trade.ticker];
                                    return (
                                      <Table.Tr key={trade.ticker}>
                                        <Table.Td>{trade.ticker}</Table.Td>
                                        <Table.Td>
                                          {ref?.name || trade.ticker}
                                        </Table.Td>
                                        <Table.Td>
                                          {formatDate(trade.earliestTradeDate)}
                                        </Table.Td>
                                        <Table.Td
                                          style={{
                                            textAlign: "right",
                                            ...numberStyle,
                                          }}
                                        >
                                          {formatNumber(
                                            trade.quantityBought,
                                            4,
                                          )}
                                        </Table.Td>
                                        <Table.Td
                                          style={{
                                            textAlign: "right",
                                            ...numberStyle,
                                          }}
                                        >
                                          {formatNumber(
                                            trade.quantitySold,
                                            4,
                                          )}
                                        </Table.Td>
                                        <Table.Td
                                          style={{
                                            textAlign: "right",
                                            ...numberStyle,
                                          }}
                                        >
                                          {formatNumber(trade.netQuantity, 4)}
                                        </Table.Td>
                                        <Table.Td
                                          style={{
                                            textAlign: "right",
                                            ...numberStyle,
                                          }}
                                        >
                                          {formatNumber(trade.averagePrice, 4)}
                                        </Table.Td>
                                        <Table.Td
                                          style={{
                                            textAlign: "right",
                                            ...numberStyle,
                                            color:
                                              trade.pricePaid < 0
                                                ? "var(--mantine-color-red-6)"
                                                : undefined,
                                          }}
                                        >
                                          {formatAmount(trade.pricePaid)}
                                        </Table.Td>
                                      </Table.Tr>
                                    );
                                  })}
                                </Table.Tbody>
                                <Table.Tfoot>
                                  <Table.Tr>
                                    <Table.Th colSpan={7}>Month total</Table.Th>
                                    <Table.Th
                                      style={{
                                        textAlign: "right",
                                        color:
                                          monthNetPricePaid < 0
                                            ? "var(--mantine-color-red-6)"
                                            : undefined,
                                      }}
                                    >
                                      {formatAmount(monthNetPricePaid)}
                                    </Table.Th>
                                  </Table.Tr>
                                </Table.Tfoot>
                              </Table>
                            </Table.ScrollContainer>
                          )}
                        </Box>
                      </Collapse>
                    </Table.Td>
                  </Table.Tr>
                </React.Fragment>
              );
            })}
          </Table.Tbody>
        </Table>
      </Table.ScrollContainer>
      {hasMetricsError ? (
        <Text c="red" size="xs" mt="xs">
          Historical portfolio metrics are unavailable.
        </Text>
      ) : null}
      <Text c="dimmed" size="xs" mt="xs">
        Monthly P&amp;L and dividends require a historical snapshot at or before
        the start of the month. Net price paid includes trade FX, with sell
        proceeds offsetting purchases.
      </Text>
    </Paper>
  );
};

export default MonthlyPortfolioActivityTable;
