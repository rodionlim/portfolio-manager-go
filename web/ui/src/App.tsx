// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/App.tsx

import "@mantine/core/styles.css";
import "@mantine/dates/styles.css"; // mantine date picker styles
import "@mantine/notifications/styles.css";
import "mantine-react-table/styles.css";

import { useEffect } from "react";
import { BrowserRouter as Router } from "react-router-dom";
import { Provider, useDispatch } from "react-redux";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  MantineProvider,
  AppShell,
  AppShellNavbar,
  AppShellMain,
} from "@mantine/core";

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

  useEffect(() => {
    dispatch(fetchReferenceData());
  }, [dispatch]);

  return (
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <MantineProvider theme={theme}>
          <Notifications />
          <AppShell navbar={{ width: 300, breakpoint: "sm" }} padding="md">
            <AppShellNavbar>
              <NavbarNested />
            </AppShellNavbar>
            <AppShellMain>
              <Controller />
            </AppShellMain>
            <ColorSchemeToggle />
          </AppShell>
        </MantineProvider>
      </QueryClientProvider>
    </Provider>
  );
}
