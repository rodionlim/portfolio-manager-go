import { useState } from "react";
import { IconChevronRight } from "@tabler/icons-react";
import { Box, Collapse, Group, ThemeIcon, UnstyledButton } from "@mantine/core";
import classes from "./NavbarLinksGroup.module.css";
import { Link, useNavigate } from "react-router-dom";

interface LinksGroupProps {
  icon: React.FC<any>;
  label: string;
  initiallyOpened?: boolean;
  links?: { label: string; link: string }[];
  link?: string;
  onLinkClick?: () => void;
}

export function LinksGroup({
  icon: Icon,
  label,
  initiallyOpened,
  links,
  link,
  onLinkClick,
}: LinksGroupProps) {
  const hasLinks = Array.isArray(links);
  const [opened, setOpened] = useState(initiallyOpened || false);
  const items = (hasLinks ? links : []).map((link) => (
    <Link
      className={classes.link}
      to={link.link}
      key={link.label}
      onClick={onLinkClick}
    >
      {link.label}
    </Link>
  ));
  const navigate = useNavigate();

  return (
    <>
      <UnstyledButton
        onClick={() => {
          if (hasLinks) {
            setOpened((o) => !o);
          } else {
            navigate(link!);
            // Add a small delay to ensure navigation completes before navbar closes
            setTimeout(() => {
              onLinkClick?.();
            }, 50);
          }
        }}
        className={classes.control}
      >
        <Group justify="space-between" gap={0}>
          <Box style={{ display: "flex", alignItems: "center" }}>
            <ThemeIcon variant="light" size={30}>
              <Icon size={18} />
            </ThemeIcon>
            <Box ml="md">{label}</Box>
          </Box>
          {hasLinks && (
            <IconChevronRight
              className={classes.chevron}
              stroke={1.5}
              size={16}
              style={{ transform: opened ? "rotate(-90deg)" : "none" }}
            />
          )}
        </Group>
      </UnstyledButton>
      {hasLinks ? <Collapse in={opened}>{items}</Collapse> : null}
    </>
  );
}
