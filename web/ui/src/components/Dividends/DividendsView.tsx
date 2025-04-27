import React, { useState } from "react";
import { Tabs } from "@mantine/core";
import DividendsTable from "./DividendsTable";
import DividendsAggregatedTable from "./DividendsAggregatedTable";

const DividendsView: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string | null>("ticker");

  return (
    <Tabs value={activeTab} onChange={setActiveTab}>
      <Tabs.List>
        <Tabs.Tab value="ticker">Dividends by Ticker</Tabs.Tab>
        <Tabs.Tab value="yearly">Yearly Summary</Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel value="ticker" pt="md">
        {activeTab === "ticker" && <DividendsTable />}
      </Tabs.Panel>

      <Tabs.Panel value="yearly" pt="md">
        {activeTab === "yearly" && <DividendsAggregatedTable />}
      </Tabs.Panel>
    </Tabs>
  );
};

export default DividendsView;
