import React from "react";
import BlotterTable from "../Blotter/BlotterTable";

interface ControllerProps {
  navigationLinksGroup: string;
}

const Controller: React.FC<ControllerProps> = ({ navigationLinksGroup }) => {
  const renderComponent = () => {
    switch (navigationLinksGroup) {
      case "blotter":
        return <BlotterTable />;
      default:
        return <div>Select a valid action on the left navigation bar</div>;
    }
  };

  return <div>{renderComponent()}</div>;
};

export default Controller;
