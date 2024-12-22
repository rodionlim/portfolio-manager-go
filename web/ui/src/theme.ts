import { createTheme, DEFAULT_THEME } from "@mantine/core";
import { themeToVars } from "@mantine/vanilla-extract";

export const theme = createTheme({
  fontFamily: "Roboto, sans-serif",
  headings: {
    fontFamily: `Roboto, ${DEFAULT_THEME.fontFamily}`,
  },
});
export const vars = themeToVars(theme);
