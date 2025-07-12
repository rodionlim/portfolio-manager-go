import { createTheme, DEFAULT_THEME } from "@mantine/core";

export const theme = createTheme({
  fontFamily: "Roboto, sans-serif",
  headings: {
    fontFamily: `Roboto, ${DEFAULT_THEME.fontFamily}`,
  },
});
