import React, { useState, useEffect } from "react";
import { Container, Title, Text, Stack, Tabs } from "@mantine/core";
import { useLocation } from "react-router-dom";
import ReportsTable from "./ReportsTable";
import SGXMostTradedStocksView from "./SGXMostTradedStocksView";
import SGXTop10StocksView from "./SGXTop10StocksView";
import SGXSectorView from "./SGXSectorView";

const ReportsView: React.FC = () => {
  const location = useLocation();
  const [activeTab, setActiveTab] = useState<string | null>("reports");

  // Handle navigation state to switch tabs
  useEffect(() => {
    const navigationState = location.state as { activeTab?: string } | null;
    if (navigationState?.activeTab) {
      setActiveTab(navigationState.activeTab);
    }
  }, [location.state]);

  return (
    <Container size="xl">
      <Stack gap="md">
        <div>
          <Title order={2}>Analytics</Title>
          <Text c="dimmed" size="sm">
            Download and analyze SGX reports, visualize trading patterns with
            AI-powered insights
          </Text>
        </div>

        <Tabs value={activeTab} onChange={setActiveTab}>
          <Tabs.List>
            <Tabs.Tab value="reports">Download Reports</Tabs.Tab>
            <Tabs.Tab value="visualization">
              Most Traded Stocks (Weekly)
            </Tabs.Tab>
            <Tabs.Tab value="top10">Top 10 Stocks (Weekly)</Tabs.Tab>
            <Tabs.Tab value="sectors">Sector Funds Flow (Weekly)</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="reports" pt="md">
            {activeTab === "reports" && <ReportsTable />}
          </Tabs.Panel>

          <Tabs.Panel value="visualization" pt="md">
            {activeTab === "visualization" && <SGXMostTradedStocksView />}
          </Tabs.Panel>

          <Tabs.Panel value="top10" pt="md">
            {activeTab === "top10" && <SGXTop10StocksView />}
          </Tabs.Panel>

          <Tabs.Panel value="sectors" pt="md">
            {activeTab === "sectors" && <SGXSectorView />}
          </Tabs.Panel>
        </Tabs>
      </Stack>
    </Container>
  );
};

export default ReportsView;
