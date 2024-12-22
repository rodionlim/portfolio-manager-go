// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/App.tsx
import "@mantine/core/styles.css";
import {
  MantineProvider,
  AppShellNavbar,
  AppShell,
  AppShellMain,
} from "@mantine/core";
import { theme } from "./theme";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NavbarNested } from "./components/NavbarNested/NavbarNested";
import { ColorSchemeToggle } from "./components/ColorSchemeToggle/ColorSchemeToggle";
import BlotterTable from "./Blotter/BlotterTable";

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider theme={theme}>
        <AppShell navbar={{ width: 300, breakpoint: "sm" }} padding="md">
          <AppShellNavbar>
            <NavbarNested />
          </AppShellNavbar>
          <AppShellMain>PLACE HOLDER</AppShellMain>
          <ColorSchemeToggle />
        </AppShell>
      </MantineProvider>
    </QueryClientProvider>
  );
}
