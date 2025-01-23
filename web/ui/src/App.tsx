// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/App.tsx

import "@mantine/core/styles.css";
import "@mantine/dates/styles.css"; // mantine date picker styles
import "mantine-react-table/styles.css";

import {
  MantineProvider,
  AppShellNavbar,
  AppShell,
  AppShellMain,
} from "@mantine/core";
import { theme } from "./theme";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import NavbarNested from "./components/NavbarNested/NavbarNested";
import { ColorSchemeToggle } from "./components/ColorSchemeToggle/ColorSchemeToggle";
import Controller from "./Controller/controller";
import { useState } from "react";

const queryClient = new QueryClient();

export default function App() {
  const [currentTab, setCurrentTab] = useState("");

  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme}>
        <AppShell navbar={{ width: 300, breakpoint: "sm" }} padding="md">
          <AppShellNavbar>
            <NavbarNested setCurrentTab={setCurrentTab} />
          </AppShellNavbar>
          <AppShellMain>
            <Controller currentTab={currentTab} />
          </AppShellMain>
          <ColorSchemeToggle />
        </AppShell>
      </MantineProvider>
    </QueryClientProvider>
  );
}
