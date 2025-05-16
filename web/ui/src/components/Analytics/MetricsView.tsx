import React from "react";
import { Tabs } from "@mantine/core";
import MetricsTable from "./MetricsTable";

const MetricsView: React.FC = () => {
  return (
    <Tabs defaultValue="table">
      <Tabs.List>
        <Tabs.Tab value="table">Historical Metrics Table</Tabs.Tab>
        <Tabs.Tab value="chart">Metrics Chart</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="table" pt="md">
        <MetricsTable />
      </Tabs.Panel>

      <Tabs.Panel value="chart" pt="md">
        <h2>Work In Progress</h2>
        <p>Charts visualization coming soon.</p>
      </Tabs.Panel>
    </Tabs>
  );
};

export default MetricsView;
