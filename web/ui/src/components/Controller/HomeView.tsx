import React from "react";
import { Button, Group, Stack, Text } from "@mantine/core";
import { IconChartLine, IconLayoutDashboard } from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";

const HomeView: React.FC = () => {
  const navigate = useNavigate();

  return (
    <Stack gap="md" align="flex-start">
      <Text>Select a valid action on the left navigation bar</Text>
      <Group gap="sm">
        <Button
          leftSection={<IconLayoutDashboard size={16} />}
          variant="light"
          onClick={() => navigate("/summary")}
        >
          Portfolio Summary
        </Button>
        <Button
          leftSection={<IconChartLine size={16} />}
          variant="light"
          onClick={() => navigate("/analytics/metrics")}
        >
          Historical Performance
        </Button>
      </Group>
    </Stack>
  );
};

export default HomeView;
