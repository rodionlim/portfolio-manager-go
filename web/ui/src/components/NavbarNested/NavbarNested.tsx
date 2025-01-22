import {
  IconAdjustments,
  IconBook,
  IconCalendarStats,
  IconCoin,
  IconDeviceLaptop,
} from "@tabler/icons-react";
import { Code, Group, ScrollArea } from "@mantine/core";
import { LinksGroup } from "../NavbarLinksGroup/NavbarLinksGroup";
import { UserButton } from "../UserButton/UserButton";
import { Logo } from "./Logo";
import classes from "./NavbarNested.module.css";

const items = [
  { label: "Positions", icon: IconCoin },
  {
    label: "Blotter",
    icon: IconDeviceLaptop,
    initiallyOpened: true,
    links: [
      { label: "Fetch trades", link: "/" },
      { label: "Add trade", link: "/add_trade" },
      { label: "Update trade", link: "/" },
    ],
  },
  {
    label: "Dividends",
    icon: IconCalendarStats,
  },
  {
    label: "Reference Data",
    icon: IconBook,
    initiallyOpened: false,
    links: [
      { label: "Fetch", link: "/" },
      { label: "Add / Update", link: "/" },
    ],
  },
  { label: "Settings", icon: IconAdjustments },
];

export function NavbarNested() {
  const links = items.map((item) => <LinksGroup {...item} key={item.label} />);

  return (
    <nav className={classes.navbar}>
      <div className={classes.header}>
        <Group justify="space-between">
          <Logo style={{ width: 140 }} />
          <Code fw={700}>v1.0.0</Code>
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
}
