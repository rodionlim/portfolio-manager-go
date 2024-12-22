import { IconChevronRight } from "@tabler/icons-react";
import { Avatar, Group, Text, UnstyledButton } from "@mantine/core";
import classes from "./UserButton.module.css";

export function UserButton() {
  return (
    <UnstyledButton className={classes.user}>
      <Group>
        <Avatar src="https://github.com/rodionlim.png" radius="xl" />

        <div style={{ flex: 1 }}>
          <Text size="sm" fw={500}>
            Rodion Lim
          </Text>

          <Text c="dimmed" size="xs">
            rodion.lim@hotmail.com
          </Text>
        </div>

        <IconChevronRight size={14} stroke={1.5} />
      </Group>
    </UnstyledButton>
  );
}
