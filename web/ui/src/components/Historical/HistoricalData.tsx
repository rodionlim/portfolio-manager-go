import React, { useEffect, useState, useMemo } from "react";
import {
  Title,
  Card,
  Collapse,
  Table,
  Button,
  Group,
  Select,
  Switch,
  ActionIcon,
  Tooltip,
  Modal,
  Pagination,
  LoadingOverlay,
  NumberInput,
  Autocomplete,
  Text,
  Stack,
} from "@mantine/core";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import {
  IconTrash,
  IconRefresh,
  IconPlus,
  IconEdit,
  IconChevronDown,
  IconChevronUp,
} from "@tabler/icons-react";
import { notifications } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import { DateInput } from "@mantine/dates";

import HistoricalCorrelation from "./HistoricalCorrelation";

interface AssetConfig {
  ticker: string;
  source: string;
  enabled: boolean;
  last_sync?: number;
  lookback_years?: number;
}

interface HistoricalRecord {
  ticker: string;
  price: number;
  adj_close: number;
  currency: string;
  timestamp: number;
}

const HistoricalData: React.FC = () => {
  const [configs, setConfigs] = useState<AssetConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [configTableCollapsed, setConfigTableCollapsed] = useState(false);

  const refData = useSelector((state: RootState) => state.referenceData.data);
  const tickerOptions = useMemo(() => {
    if (!refData) return [];
    const tickers = Object.values(refData)
      .map((item) => item.yahoo_ticker)
      .filter((t) => t && t.trim() !== "");
    // Unique and sorted
    return Array.from(new Set(tickers)).sort();
  }, [refData]);

  // Add new config state
  const [newTicker, setNewTicker] = useState("");
  const [newSource, setNewSource] = useState("yahoo");
  const [newLookback, setNewLookback] = useState(5);
  const [resyncingAll, setResyncingAll] = useState(false);

  const matchedName = useMemo(() => {
    if (!refData || !newTicker) return null;
    const match = Object.values(refData).find(
      (item) => item.yahoo_ticker === newTicker || item.id === newTicker
    );
    return match ? match.name : null;
  }, [refData, newTicker]);

  const tickerNameMap = useMemo(() => {
    const map = new Map<string, string>();
    if (!refData) return map;
    for (const item of Object.values(refData)) {
      if (item?.yahoo_ticker && item?.name)
        map.set(item.yahoo_ticker, item.name);
      if (item?.id && item?.name) map.set(item.id, item.name);
    }
    return map;
  }, [refData]);

  const sortedConfigs = useMemo(() => {
    return [...configs].sort((a, b) => a.ticker.localeCompare(b.ticker));
  }, [configs]);

  const truncate = (value: string, maxChars: number) => {
    if (!value) return value;
    if (value.length <= maxChars) return value;
    return value.slice(0, maxChars) + "...";
  };

  // Modal State
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedTicker, setSelectedTicker] = useState<string | null>(null);

  // Edit Modal State
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editConfig, setEditConfig] = useState<AssetConfig | null>(null);

  const [historyData, setHistoryData] = useState<HistoricalRecord[]>([]);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [historyPage, setHistoryPage] = useState(1);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [fromDate, setFromDate] = useState<Date | null>(null);
  const [toDate, setToDate] = useState<Date | null>(null);

  useEffect(() => {
    fetchConfigs();
  }, []);

  const fetchConfigs = async () => {
    setLoading(true);
    try {
      const response = await fetch(getUrl("/api/v1/historical/config"));
      if (!response.ok) throw new Error("Failed to fetch configs");
      const data = await response.json();
      setConfigs(data || []);
    } catch (error: any) {
      notifications.show({
        title: "Error",
        message: error.message,
        color: "red",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleRowClick = (ticker: string) => {
    setSelectedTicker(ticker);
    setHistoryPage(1);
    setModalOpen(true);
    setFromDate(null);
    setToDate(null);
    fetchHistoryWrapper(ticker, 1, null, null);
  };

  const fetchHistoryWrapper = (
    ticker: string,
    page: number,
    from: Date | null,
    to: Date | null
  ) => {
    setHistoryLoading(true);
    const p = new URLSearchParams();
    p.append("page", page.toString());
    p.append("limit", "100");
    if (from) p.append("from", (from.getTime() / 1000).toString());
    if (to) p.append("to", (to.getTime() / 1000).toString());

    fetch(getUrl(`/api/v1/historical/data/${ticker}?${p.toString()}`))
      .then((res) => res.json())
      .then((data) => {
        setHistoryData(data.data || []);
        setHistoryTotal(data.total || 0);
      })
      .catch((err) =>
        notifications.show({
          title: "Error",
          message: err.message,
          color: "red",
        })
      )
      .finally(() => setHistoryLoading(false));
  };

  const handlePageChange = (p: number) => {
    setHistoryPage(p);
    if (selectedTicker) {
      fetchHistoryWrapper(selectedTicker, p, fromDate, toDate);
    }
  };

  const handleDateSearch = () => {
    if (selectedTicker) {
      setHistoryPage(1);
      fetchHistoryWrapper(selectedTicker, 1, fromDate, toDate);
    }
  };

  const handleAdd = async () => {
    if (!newTicker) return;
    try {
      const config: AssetConfig = {
        ticker: newTicker,
        source: newSource,
        enabled: true,
        lookback_years: newLookback,
      };

      const response = await fetch(getUrl("/api/v1/historical/config"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(config),
      });

      if (!response.ok) throw new Error("Failed to add config");

      notifications.show({ message: "Added " + newTicker, color: "green" });
      setNewTicker("");
      fetchConfigs();
    } catch (error: any) {
      notifications.show({
        title: "Error",
        message: error.message,
        color: "red",
      });
    }
  };

  const handleDelete = async (ticker: string) => {
    try {
      const response = await fetch(
        getUrl(`/api/v1/historical/config/${ticker}`),
        {
          method: "DELETE",
        }
      );
      if (!response.ok) throw new Error("Failed to delete config");

      notifications.show({ message: "Deleted " + ticker, color: "green" });
      fetchConfigs();
    } catch (error: any) {
      notifications.show({
        title: "Error",
        message: error.message,
        color: "red",
      });
    }
  };

  const handleToggle = async (config: AssetConfig) => {
    try {
      const newConfig = { ...config, enabled: !config.enabled };
      const response = await fetch(getUrl("/api/v1/historical/config"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newConfig),
      });
      if (!response.ok) throw new Error("Failed to update config");
      fetchConfigs();
    } catch (error: any) {
      notifications.show({
        title: "Error",
        message: error.message,
        color: "red",
      });
    }
  };

  const handleEditClick = (e: React.MouseEvent, config: AssetConfig) => {
    e.stopPropagation();
    setEditConfig({ ...config });
    setEditModalOpen(true);
  };

  const handleSaveEdit = async () => {
    if (!editConfig) return;
    try {
      const response = await fetch(getUrl("/api/v1/historical/config"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(editConfig),
      });

      if (!response.ok) throw new Error("Failed to update config");

      notifications.show({
        message: "Updated " + editConfig.ticker,
        color: "green",
      });
      setEditModalOpen(false);
      setEditConfig(null);
      fetchConfigs();
    } catch (error: any) {
      notifications.show({
        title: "Error",
        message: error.message,
        color: "red",
      });
    }
  };

  const handleSyncAll = async () => {
    if (
      !window.confirm(
        "Are you sure you want to resync ALL enabled assets? This may take time."
      )
    )
      return;

    setResyncingAll(true);
    const enabledConfigs = configs.filter((c) => c.enabled);
    let successCount = 0;
    let failCount = 0;

    for (const [index, config] of enabledConfigs.entries()) {
      try {
        const message = `Syncing ${config.ticker} (${index + 1}/${
          enabledConfigs.length
        })...`;

        if (index === 0) {
          notifications.show({
            id: "sync-all-progress",
            loading: true,
            message,
            autoClose: false,
            withCloseButton: false,
          });
        } else {
          notifications.update({
            id: "sync-all-progress",
            loading: true,
            message,
            autoClose: false,
            withCloseButton: false,
          });
        }

        // Call existing API
        const response = await fetch(getUrl("/api/v1/historical/sync"), {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ ticker: config.ticker }),
        });

        // Always consume response body
        await response.json();

        if (!response.ok) throw new Error("Failed");
        successCount++;
      } catch (e) {
        failCount++;
        console.error(`Failed to sync ${config.ticker}`, e);
      }

      // Delay 3s if not last
      if (index < enabledConfigs.length - 1) {
        await new Promise((r) => setTimeout(r, 3000));
      }
    }

    setResyncingAll(false);
    notifications.update({
      id: "sync-all-progress",
      message: `Completed! Success: ${successCount}, Failed: ${failCount}`,
      color: "blue",
      loading: false,
      autoClose: 5000,
      withCloseButton: true,
    });
    fetchConfigs();
  };

  const handleSync = async (e: React.MouseEvent, ticker: string) => {
    e.stopPropagation(); // Prevent row click
    try {
      notifications.show({
        message: `Syncing ${ticker}...`,
        loading: true,
        id: `sync-${ticker}`,
      });
      const response = await fetch(getUrl("/api/v1/historical/sync"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ticker }),
      });

      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "Failed to sync");

      notifications.show({
        id: `sync-${ticker}`,
        message: data.message || `Synced ${ticker}`,
        color: "green",
        autoClose: 5000,
      });
      fetchConfigs();
    } catch (error: any) {
      notifications.show({
        id: `sync-${ticker}`,
        title: "Error",
        message: error.message,
        color: "red",
      });
    }
  };

  const getSyncTooltip = (config: AssetConfig) => {
    const lastSync = config.last_sync
      ? new Date(config.last_sync * 1000)
      : null;
    const today = new Date();

    if (!lastSync) {
      return `Sync / Backfill data (Last ${config.lookback_years || 5} Years)`;
    }

    // Calculate next sync range
    return `Sync data from ${lastSync.toLocaleDateString()} to ${today.toLocaleDateString()}`;
  };

  return (
    <div style={{ padding: "20px" }}>
      <Title order={2} mb="lg">
        Historical Market Data
      </Title>

      <Card withBorder shadow="sm" mb="lg">
        <Group align="flex-start">
          <div style={{ flex: 1, minWidth: "250px" }}>
            <Stack gap={4}>
              <Autocomplete
                label="Ticker"
                placeholder="e.g. AAPL"
                value={newTicker}
                onChange={(val) => setNewTicker(val.toUpperCase())}
                data={tickerOptions}
              />
              {matchedName && (
                <Text c="dimmed" size="xs">
                  {matchedName}
                </Text>
              )}
            </Stack>
          </div>
          <Select
            label="Source"
            data={[
              { value: "yahoo", label: "Yahoo Finance" },
              { value: "google", label: "Google Finance" },
            ]}
            value={newSource}
            onChange={(val) => setNewSource(val || "yahoo")}
            style={{ width: 200 }}
          />
          <NumberInput
            label="Years"
            value={newLookback}
            onChange={(val) => setNewLookback(Number(val) || 5)}
            min={1}
            max={30}
            style={{ width: 100 }}
          />
          <Group gap="xs" mt={24}>
            <Button leftSection={<IconPlus size={16} />} onClick={handleAdd}>
              Add
            </Button>
            <Button
              color="orange"
              variant="light"
              leftSection={<IconRefresh size={16} />}
              onClick={handleSyncAll}
              loading={resyncingAll}
            >
              Sync All
            </Button>
          </Group>
        </Group>
      </Card>

      <Card withBorder shadow="sm">
        <Group justify="space-between" mb="sm" wrap="wrap">
          <Title order={4}>Configurations</Title>
          <ActionIcon
            variant="subtle"
            onClick={() => setConfigTableCollapsed((v) => !v)}
            aria-label={
              configTableCollapsed
                ? "Expand configurations"
                : "Collapse configurations"
            }
          >
            {configTableCollapsed ? (
              <IconChevronDown size={18} />
            ) : (
              <IconChevronUp size={18} />
            )}
          </ActionIcon>
        </Group>

        <Collapse in={!configTableCollapsed}>
          <Table.ScrollContainer minWidth={800}>
            <Table highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Ticker</Table.Th>
                  <Table.Th>Source</Table.Th>
                  <Table.Th>Lookback (Y)</Table.Th>
                  <Table.Th>Status</Table.Th>
                  <Table.Th>Last Sync</Table.Th>
                  <Table.Th>Actions</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {sortedConfigs.map((c) => (
                  <Table.Tr
                    key={c.ticker}
                    onClick={() => handleRowClick(c.ticker)}
                    style={{ cursor: "pointer" }}
                  >
                    <Table.Td>
                      <Group gap={6} wrap="nowrap">
                        <Text fw={600}>{c.ticker}</Text>
                        {tickerNameMap.get(c.ticker) && (
                          <Tooltip
                            label={tickerNameMap.get(c.ticker) || ""}
                            withArrow
                          >
                            <Text
                              size="xs"
                              c="dimmed"
                              style={{ whiteSpace: "nowrap" }}
                            >
                              {truncate(tickerNameMap.get(c.ticker) || "", 10)}
                            </Text>
                          </Tooltip>
                        )}
                      </Group>
                    </Table.Td>
                    <Table.Td>{c.source}</Table.Td>
                    <Table.Td>{c.lookback_years || 5}</Table.Td>
                    <Table.Td>
                      <Switch
                        checked={c.enabled}
                        onChange={() => handleToggle(c)}
                        onClick={(e) => e.stopPropagation()}
                        label={c.enabled ? "Active" : "Disabled"}
                      />
                    </Table.Td>
                    <Table.Td>
                      {c.last_sync
                        ? new Date(c.last_sync * 1000).toLocaleString()
                        : "Never"}
                    </Table.Td>
                    <Table.Td>
                      <Group gap="xs">
                        <Tooltip label={getSyncTooltip(c)} withArrow>
                          <ActionIcon
                            variant="light"
                            color="blue"
                            onClick={(e) => handleSync(e, c.ticker)}
                            disabled={!c.enabled}
                          >
                            <IconRefresh size={16} />
                          </ActionIcon>
                        </Tooltip>

                        <Tooltip label="Edit configuration" withArrow>
                          <ActionIcon
                            variant="light"
                            color="orange"
                            onClick={(e) => handleEditClick(e, c)}
                          >
                            <IconEdit size={16} />
                          </ActionIcon>
                        </Tooltip>

                        <Tooltip
                          label="Delete configuration and all historical data"
                          withArrow
                          color="red"
                        >
                          <ActionIcon
                            variant="light"
                            color="red"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDelete(c.ticker);
                            }}
                          >
                            <IconTrash size={16} />
                          </ActionIcon>
                        </Tooltip>
                      </Group>
                    </Table.Td>
                  </Table.Tr>
                ))}
                {sortedConfigs.length === 0 && !loading && (
                  <Table.Tr>
                    <Table.Td colSpan={6} style={{ textAlign: "center" }}>
                      No historical data configurations found
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>
        </Collapse>
      </Card>

      <Modal
        opened={modalOpen}
        onClose={() => setModalOpen(false)}
        title={<Title order={4}>Historical Data: {selectedTicker}</Title>}
        size="lg"
      >
        <Group mb="md">
          <DateInput
            value={fromDate}
            onChange={setFromDate}
            label="From Date"
            placeholder="Start Date"
            clearable
          />
          <DateInput
            value={toDate}
            onChange={setToDate}
            label="To Date"
            placeholder="End Date"
            clearable
          />
          <Button onClick={handleDateSearch} mt={24}>
            Search
          </Button>
        </Group>

        <div style={{ position: "relative", minHeight: 200 }}>
          <LoadingOverlay visible={historyLoading} />
          <Table.ScrollContainer minWidth={500}>
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Date</Table.Th>
                  <Table.Th>Close</Table.Th>
                  <Table.Th>Adj Close</Table.Th>
                  <Table.Th>Currency</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {historyData.map((d, i) => (
                  <Table.Tr key={i}>
                    <Table.Td>
                      {new Date(d.timestamp * 1000).toLocaleDateString()}
                    </Table.Td>
                    <Table.Td>{d.price.toFixed(4)}</Table.Td>
                    <Table.Td>
                      {d.adj_close ? d.adj_close.toFixed(4) : "-"}
                    </Table.Td>
                    <Table.Td>{d.currency}</Table.Td>
                  </Table.Tr>
                ))}
                {historyData.length === 0 && (
                  <Table.Tr>
                    <Table.Td colSpan={4} align="center">
                      No data found
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>
        </div>

        <Group justify="center" mt="md">
          <Pagination
            total={Math.ceil(historyTotal / 100)}
            value={historyPage}
            onChange={handlePageChange}
          />
        </Group>
      </Modal>

      <Modal
        opened={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        title={
          <Title order={4}>Edit Configuration: {editConfig?.ticker}</Title>
        }
      >
        {editConfig && (
          <div
            style={{ display: "flex", flexDirection: "column", gap: "1rem" }}
          >
            <Select
              label="Source"
              data={[
                { value: "yahoo", label: "Yahoo Finance" },
                { value: "google", label: "Google Finance" },
              ]}
              value={editConfig.source}
              onChange={(val) =>
                setEditConfig({ ...editConfig, source: val || "yahoo" })
              }
            />
            <NumberInput
              label="Lookback Years"
              value={editConfig.lookback_years || 5}
              onChange={(val) =>
                setEditConfig({
                  ...editConfig,
                  lookback_years: Number(val) || 5,
                })
              }
              min={1}
              max={30}
            />
            <Group justify="flex-end" mt="md">
              <Button variant="default" onClick={() => setEditModalOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleSaveEdit}>Save</Button>
            </Group>
          </div>
        )}
      </Modal>

      <HistoricalCorrelation />
    </div>
  );
};

export default HistoricalData;
