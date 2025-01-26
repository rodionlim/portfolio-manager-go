import React from "react";
import { Routes, Route } from "react-router-dom";

import BlotterTable from "../Blotter/BlotterTable";
import BlotterForm from "../Blotter/BlotterForm";

const Controller: React.FC = () => {
  return (
    <Routes>
      <Route path="/blotter/add_trade" element={<BlotterForm />} />
      <Route path="/blotter" element={<BlotterTable />} />
      <Route
        path="/*"
        element={<div>Select a valid action on the left navigation bar</div>}
      />
    </Routes>
  );
};

export default Controller;
