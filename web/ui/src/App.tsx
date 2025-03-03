import "@mantine/core/styles.css";
import "@mantine/dates/styles.css"; // mantine date picker styles
import "@mantine/notifications/styles.css";
import "mantine-react-table/styles.css";

import { useEffect, useState } from "react";
import { BrowserRouter as Router } from "react-router-dom";
import { Provider, useDispatch } from "react-redux";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  MantineProvider,
  AppShell,
  AppShellNavbar,
  AppShellMain,
  AppShellHeader,
  Burger,
  Group,
  Box,
} from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";

import { theme } from "./theme";
import store, { AppDispatch } from "./store";
import { fetchReferenceData } from "./slices/referenceDataSlice";
import NavbarNested from "./components/NavbarNested/NavbarNested";
import { ColorSchemeToggle } from "./components/ColorSchemeToggle/ColorSchemeToggle";
import Controller from "./components/Controller/Controller";
import { Notifications } from "@mantine/notifications";

const queryClient = new QueryClient();

export default function AppWrapper() {
  return (
    <Provider store={store}>
      <Router>
        <App />
      </Router>
    </Provider>
  );
}

function App() {
  const dispatch = useDispatch<AppDispatch>();
  const [opened, setOpened] = useState(true);
  const isMobile = useMediaQuery("(max-width: 768px)");

  // Automatically close navbar on mobile devices
  useEffect(() => {
    if (isMobile) {
      setOpened(false);
    } else {
      setOpened(true);
    }
  }, [isMobile]);

  useEffect(() => {
    dispatch(fetchReferenceData());
  }, [dispatch]);

  return (
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <MantineProvider theme={theme}>
          <Notifications />
          <AppShell
            header={isMobile ? { height: 60 } : undefined}
            navbar={{
              width: 300,
              breakpoint: "sm",
              collapsed: { mobile: !opened, desktop: false },
            }}
            padding="lg"
          >
            <AppShellHeader>
              <Group h="100%" px="md">
                <Burger
                  opened={opened}
                  onClick={() => setOpened(!opened)}
                  hiddenFrom="sm"
                  size="sm"
                />
                <Box style={{ flexGrow: 1 }} />
                <ColorSchemeToggle />
              </Group>
            </AppShellHeader>
            <AppShellNavbar>
              <NavbarNested />
            </AppShellNavbar>
            <AppShellMain>
              <Controller />
            </AppShellMain>
          </AppShell>
        </MantineProvider>
      </QueryClientProvider>
    </Provider>
  );
}
