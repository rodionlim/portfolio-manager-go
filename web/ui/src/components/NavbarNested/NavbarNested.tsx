import {
  IconAdjustments,
  IconBook,
  IconCalendarStats,
  IconCoin,
  IconDeviceLaptop,
  IconChartLine,
} from "@tabler/icons-react";
import { Code, Group, ScrollArea } from "@mantine/core";
import React from "react";

import { LinksGroup } from "../NavbarLinksGroup/NavbarLinksGroup";
import { UserButton } from "../UserButton/UserButton";
import { Logo } from "./Logo";
import { VERSION } from "../../utils/version";

import classes from "./NavbarNested.module.css";

const items = [
  { label: "Positions", icon: IconCoin, link: "/positions" },
  {
    label: "Blotter",
    icon: IconDeviceLaptop,
    initiallyOpened: true,
    links: [
      { label: "Fetch trades", link: "/blotter" },
      { label: "Add trade", link: "/blotter/add_trade" },
    ],
  },
  {
    label: "Dividends",
    icon: IconCalendarStats,
    link: "/dividends",
  },
  {
    label: "Analytics",
    icon: IconChartLine,
    initiallyOpened: false,
    links: [
      { label: "Metrics", link: "/analytics/metrics" },
      { label: "Reports", link: "/analytics/reports" },
    ],
  },
  {
    label: "Reference Data",
    icon: IconBook,
    initiallyOpened: false,
    links: [
      { label: "Fetch data", link: "/refdata" },
      { label: "Add data", link: "/refdata/add_ref_data" },
    ],
  },
  { label: "Settings", icon: IconAdjustments, link: "/settings" },
];

interface NavbarNestedProps {
  onLinkClick?: () => void;
}

const NavbarNested: React.FC<NavbarNestedProps> = ({ onLinkClick }) => {
  const links = items.map((item) => (
    <LinksGroup {...item} key={item.label} onLinkClick={onLinkClick} />
  ));

  return (
    <nav className={classes.navbar}>
      <div className={classes.header}>
        <Group justify="space-between">
          <Logo style={{ width: 140 }} />
          <Code fw={700}>v{VERSION}</Code>
        </Group>
      </div>

      <ScrollArea className={classes.links}>
        <div className={classes.linksInner}>{links}</div>
      </ScrollArea>

      <div className={classes.footer}>
        <UserButton />
      </div>
    </nav>
  );
};

export default NavbarNested;
