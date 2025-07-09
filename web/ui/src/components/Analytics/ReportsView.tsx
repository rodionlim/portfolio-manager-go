import React, { useState } from "react";
import { Container, Title, Text, Stack, Tabs } from "@mantine/core";
import ReportsTable from "./ReportsTable";
import MostTradedStocksView from "./MostTradedStocksView";
import SGXSectorView from "./SGXSectorView";

const ReportsView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string | null>("reports");

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
            <Tabs.Tab value="sectors">Sector Funds Flow (Weekly)</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="reports" pt="md">
            {activeTab === "reports" && <ReportsTable />}
          </Tabs.Panel>

          <Tabs.Panel value="visualization" pt="md">
            {activeTab === "visualization" && <MostTradedStocksView />}
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
