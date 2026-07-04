import React from "react";
import { Button, SimpleGrid, Stack, Text } from "@mantine/core";
import {
  IconChartLine,
  IconLayoutDashboard,
  IconListDetails,
  IconTableRow,
} from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";

const HomeView: React.FC = () => {
  const navigate = useNavigate();
  const quickLinks = [
    {
      label: "Portfolio Summary",
      icon: <IconLayoutDashboard size={16} />,
      path: "/summary",
    },
    {
      label: "Historical Performance",
      icon: <IconChartLine size={16} />,
      path: "/analytics/metrics",
    },
    {
      label: "Positions",
      icon: <IconTableRow size={16} />,
      path: "/positions",
    },
    {
      label: "Blotter",
      icon: <IconListDetails size={16} />,
      path: "/blotter",
    },
  ];

  return (
    <Stack gap="md" align="flex-start">
      <Text>Select a valid action on the left navigation bar</Text>
      <SimpleGrid cols={{ base: 1, xs: 2 }} spacing="sm" w="100%" maw={520}>
        {quickLinks.map((link) => (
          <Button
            key={link.path}
            leftSection={link.icon}
            variant="light"
            onClick={() => navigate(link.path)}
            justify="flex-start"
            fullWidth
          >
            {link.label}
          </Button>
        ))}
      </SimpleGrid>
    </Stack>
  );
};

export default HomeView;
