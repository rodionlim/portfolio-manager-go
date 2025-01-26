import React from "react";
import BlotterTable from "../Blotter/BlotterTable";
import BlotterForm from "../Blotter/BlotterForm";

interface ControllerProps {
  currentTab: string;
}

const Controller: React.FC<ControllerProps> = ({ currentTab }) => {
  const renderComponent = () => {
    switch (currentTab) {
      case "/blotter":
        return <BlotterTable />;
      case "/blotter/add_trade":
        return <BlotterForm />;
      default:
        return <div>Select a valid action on the left navigation bar</div>;
    }
  };

  return <div>{renderComponent()}</div>;
};

export default Controller;
