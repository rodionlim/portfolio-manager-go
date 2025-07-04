import React, { useState } from "react";
import { Tabs } from "@mantine/core";
import MetricsTable from "./MetricsTable";
import MetricsChart from "./MetricsChart";
import CustomJobsTable from "./CustomJobsTable";

const MetricsView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string | null>("table");

  return (
    <Tabs defaultValue="table" onChange={setActiveTab}>
      <Tabs.List>
        <Tabs.Tab value="table">Historical Metrics Table</Tabs.Tab>
        <Tabs.Tab value="chart">Metrics Chart</Tabs.Tab>
        <Tabs.Tab value="jobs">Custom Jobs</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="table" pt="md">
        {activeTab === "table" && <MetricsTable />}
      </Tabs.Panel>

      <Tabs.Panel value="chart" pt="md">
        {activeTab === "chart" && <MetricsChart />}
      </Tabs.Panel>

      <Tabs.Panel value="jobs" pt="md">
        {activeTab === "jobs" && <CustomJobsTable />}
      </Tabs.Panel>
    </Tabs>
  );
};

export default MetricsView;
