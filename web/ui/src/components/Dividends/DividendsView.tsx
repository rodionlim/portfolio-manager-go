import React, { useState, useEffect } from "react";
import { Tabs } from "@mantine/core";
import DividendsTable from "./DividendsTable";
import DividendsAggregatedTable from "./DividendsAggregatedTable";
import { useLocation } from "react-router-dom";

// Interface for location state containing ticker information
interface LocationState {
  ticker?: string;
}

const DividendsView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string | null>("ticker");
  const location = useLocation();
  const state = location.state as LocationState;
  const selectedTicker = state?.ticker || null;

  // If coming to this page with a ticker, make sure we're on ticker tab
  useEffect(() => {
    if (selectedTicker) {
      setActiveTab("ticker");
    }
  }, [selectedTicker]);

  return (
    <Tabs value={activeTab} onChange={setActiveTab}>
      <Tabs.List>
        <Tabs.Tab value="ticker">Dividends by Ticker</Tabs.Tab>
        <Tabs.Tab value="yearly">Yearly Summary</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="ticker" pt="md">
        {activeTab === "ticker" && (
          <DividendsTable initialTicker={selectedTicker} />
        )}
      </Tabs.Panel>

      <Tabs.Panel value="yearly" pt="md">
        {activeTab === "yearly" && <DividendsAggregatedTable />}
      </Tabs.Panel>
    </Tabs>
  );
};

export default DividendsView;
