// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/App.tsx

import "@mantine/core/styles.css";
import "@mantine/dates/styles.css"; // mantine date picker styles
import "mantine-react-table/styles.css";

import { useEffect, useState } from "react";
import { Provider, useDispatch, useSelector } from "react-redux";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  MantineProvider,
  AppShell,
  AppShellNavbar,
  AppShellMain,
} from "@mantine/core";

import { theme } from "./theme";
import store, { RootState, AppDispatch } from "./store";
import { fetchReferenceData } from "./slices/referenceDataSlice";
import NavbarNested from "./components/NavbarNested/NavbarNested";
import { ColorSchemeToggle } from "./components/ColorSchemeToggle/ColorSchemeToggle";
import Controller from "./Controller/Controller";

const queryClient = new QueryClient();

export default function AppWrapper() {
  return (
    <Provider store={store}>
      <App />
    </Provider>
  );
}

function App() {
  const [currentTab, setCurrentTab] = useState("");
  const dispatch = useDispatch<AppDispatch>();
  const referenceData = useSelector(
    (state: RootState) => state.referenceData.data
  );

  useEffect(() => {
    if (!referenceData) {
      dispatch(fetchReferenceData());
    }
  }, [dispatch, referenceData]);

  return (
    <Provider store={store}>
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
    </Provider>
  );
}
