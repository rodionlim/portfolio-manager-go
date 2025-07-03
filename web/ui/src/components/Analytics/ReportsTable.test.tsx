import { describe, it, expect } from "vitest";
import { sortReportsByDate, ReportFile } from "./ReportsTable";

describe("sortReportsByDate", () => {
  it("should sort reports by date descending (newest first)", () => {
    const input: ReportFile[] = [
      {
        path: "",
        name: "SGX_Fund_Flow_Weekly_Tracker_Week_of_23_June_2025.xlsx",
        hasAnalysis: false,
      },
      {
        path: "",
        name: "SGX_Fund_Flow_Weekly_Tracker_Week_of_16_Jun_2025.xlsx",
        hasAnalysis: false,
      },
      {
        path: "",
        name: "SGX_Fund_Flow_Weekly_Tracker_Week_of_9_Jun_2025.xlsx",
        hasAnalysis: false,
      },
      {
        path: "",
        name: "SGX_Fund_Flow_Weekly_Tracker_Week_of_2_Jun_2025.xlsx",
        hasAnalysis: false,
      },
      {
        path: "",
        name: "SGX_Fund_Flow_Weekly_Tracker_Week_of_31_Mar_2025.xlsx",
        hasAnalysis: false,
      },
      {
        path: "",
        name: "SGX_Fund_Flow_Weekly_Tracker_Week_of_24_Mar_2025.xlsx",
        hasAnalysis: false,
      },
    ];

    const sorted = sortReportsByDate(input);
    const expectedOrder = [
      "SGX_Fund_Flow_Weekly_Tracker_Week_of_23_June_2025.xlsx",
      "SGX_Fund_Flow_Weekly_Tracker_Week_of_16_Jun_2025.xlsx",
      "SGX_Fund_Flow_Weekly_Tracker_Week_of_9_Jun_2025.xlsx",
      "SGX_Fund_Flow_Weekly_Tracker_Week_of_2_Jun_2025.xlsx",
      "SGX_Fund_Flow_Weekly_Tracker_Week_of_31_Mar_2025.xlsx",
      "SGX_Fund_Flow_Weekly_Tracker_Week_of_24_Mar_2025.xlsx",
    ];
    expect(sorted.map((r: ReportFile) => r.name)).toEqual(expectedOrder);
  });
});
