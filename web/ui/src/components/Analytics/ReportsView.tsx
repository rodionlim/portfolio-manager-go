import React from "react";
import { Container, Title, Text, Stack } from "@mantine/core";
import ReportsTable from "./ReportsTable";

const ReportsView: React.FC = () => {
  return (
    <Container size="xl">
      <Stack gap="md">
        <div>
          <Title order={2}>SGX Reports</Title>
          <Text c="dimmed" size="sm">
            Download, view, and analyze SGX reports with AI-powered insights
          </Text>
        </div>

        <ReportsTable />
      </Stack>
    </Container>
  );
};

export default ReportsView;
