import React from "react";
import { Routes, Route } from "react-router-dom";

import BlotterTable from "../Blotter/BlotterTable";
import BlotterForm from "../Blotter/BlotterForm";
import ReferenceDataTable from "../ReferenceData/ReferenceDataTable";
import ReferenceDataForm from "../ReferenceData/ReferenceDataForm";
import PositionTable from "../Position/PositionTable";
import DividendsView from "../Dividends/DividendsView";
import Dashboard from "../Analytics/Dashboard";
import MetricsView from "../Analytics/MetricsView";
import ReportsView from "../Analytics/ReportsView";
import Settings from "../Settings/Settings";
import HistoricalData from "../Historical/HistoricalData";

const Controller: React.FC = () => {
  return (
    <Routes>
      <Route path="/blotter/add_trade" element={<BlotterForm />} />
      <Route path="/blotter/update_trade" element={<BlotterForm />} />
      <Route path="/blotter" element={<BlotterTable />} />
      <Route path="/dividends" element={<DividendsView />} />
      <Route path="/refdata/add_ref_data" element={<ReferenceDataForm />} />
      <Route path="/refdata/update_ref_data" element={<ReferenceDataForm />} />
      <Route path="/refdata" element={<ReferenceDataTable />} />
      <Route path="/positions" element={<PositionTable />} />
      <Route path="/analytics/dashboard" element={<Dashboard />} />
      <Route path="/analytics/metrics" element={<MetricsView />} />
      <Route path="/analytics/reports" element={<ReportsView />} />
      <Route path="/historical" element={<HistoricalData />} />
      <Route path="/settings" element={<Settings />} />
      <Route
        path="/*"
        element={<div>Select a valid action on the left navigation bar</div>}
      />
    </Routes>
  );
};

export default Controller;
