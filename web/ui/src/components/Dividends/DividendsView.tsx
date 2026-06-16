import React, { useState, useEffect } from "react";
import { Tabs } from "@mantine/core";
import DividendsTable from "./DividendsTable";
import DividendsAggregatedTable from "./DividendsAggregatedTable";
import DividendsMonthlySummaryTable from "./DividendsMonthlySummaryTable";
import DividendsAll from "./DividendsAll";
import { useLocation } from "react-router-dom";

// Interface for location state containing ticker information
interface LocationState {
  ticker?: string;
  activeTab?: string;
  dividendMonth?: string;
}

const DividendsView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string | null>("ticker");
  const location = useLocation();
  const state = location.state as LocationState;
  const selectedTicker = state?.ticker || null;
  const selectedDividendMonth = state?.dividendMonth || null;

  // If coming to this page with a ticker, make sure we're on ticker tab
  useEffect(() => {
    if (selectedTicker) {
      setActiveTab("ticker");
    } else if (state?.activeTab) {
      setActiveTab(state.activeTab);
    }
  }, [selectedTicker, state?.activeTab]);

  return (
    <Tabs value={activeTab} onChange={setActiveTab}>
      <Tabs.List>
        <Tabs.Tab value="ticker">Dividends by Ticker</Tabs.Tab>
        <Tabs.Tab value="monthly">Monthly Summary</Tabs.Tab>
        <Tabs.Tab value="yearly">Yearly Summary</Tabs.Tab>
        <Tabs.Tab value="all">All Dividends</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="ticker" pt="md">
        {activeTab === "ticker" && (
          <DividendsTable initialTicker={selectedTicker} />
        )}
      </Tabs.Panel>

      <Tabs.Panel value="monthly" pt="md">
        {activeTab === "monthly" && <DividendsMonthlySummaryTable />}
      </Tabs.Panel>

      <Tabs.Panel value="yearly" pt="md">
        {activeTab === "yearly" && <DividendsAggregatedTable />}
      </Tabs.Panel>

      <Tabs.Panel value="all" pt="md">
        {activeTab === "all" && (
          <DividendsAll initialMonth={selectedDividendMonth} />
        )}
      </Tabs.Panel>
    </Tabs>
  );
};

export default DividendsView;
