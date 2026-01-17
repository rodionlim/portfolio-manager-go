import React, { useState } from "react";
import { Tabs } from "@mantine/core";
import MetricsTable from "./MetricsTable";
import MetricsChart from "./MetricsChart";
import MetricsCashFlow from "./MetricsCashFlow";
import CustomJobsTable from "./CustomJobsTable";
import MetricsBenchmark from "./MetricsBenchmark";
import type { VolatilityMethod } from "./volatility";

const MetricsView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string | null>("table");

  const [volatilityMethod, setVolatilityMethod] =
    useState<VolatilityMethod>("sma");
  const [volatilityWindow, setVolatilityWindow] = useState<number>(25);

  return (
    <Tabs defaultValue="table" onChange={setActiveTab}>
      <Tabs.List>
        <Tabs.Tab value="table">Historical Metrics Table</Tabs.Tab>
        <Tabs.Tab value="chart">Metrics Chart</Tabs.Tab>
        <Tabs.Tab value="cashflow">Current Metrics & Cash Flows</Tabs.Tab>
        <Tabs.Tab value="benchmark">Benchmark Simulation</Tabs.Tab>
        <Tabs.Tab value="jobs">Custom Jobs</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="table" pt="md">
        {activeTab === "table" && (
          <MetricsTable
            volatilityMethod={volatilityMethod}
            setVolatilityMethod={(method) => {
              setVolatilityMethod(method);
              setVolatilityWindow(method === "ewma" ? 36 : 25);
            }}
            volatilityWindow={volatilityWindow}
            setVolatilityWindow={setVolatilityWindow}
          />
        )}
      </Tabs.Panel>

      <Tabs.Panel value="chart" pt="md">
        {activeTab === "chart" && (
          <MetricsChart
            volatilityMethod={volatilityMethod}
            setVolatilityMethod={(method) => {
              setVolatilityMethod(method);
              setVolatilityWindow(method === "ewma" ? 36 : 25);
            }}
            volatilityWindow={volatilityWindow}
            setVolatilityWindow={setVolatilityWindow}
          />
        )}
      </Tabs.Panel>

      <Tabs.Panel value="cashflow" pt="md">
        {activeTab === "cashflow" && <MetricsCashFlow />}
      </Tabs.Panel>

      <Tabs.Panel value="jobs" pt="md">
        {activeTab === "jobs" && <CustomJobsTable />}
      </Tabs.Panel>

      <Tabs.Panel value="benchmark" pt="md">
        {activeTab === "benchmark" && <MetricsBenchmark />}
      </Tabs.Panel>
    </Tabs>
  );
};

export default MetricsView;
