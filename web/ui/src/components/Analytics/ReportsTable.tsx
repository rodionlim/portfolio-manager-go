import React, { useState, useEffect } from "react";
import {
  Table,
  Button,
  Group,
  Modal,
  TextInput,
  NumberInput,
  Alert,
  Collapse,
  Text,
  Card,
  Stack,
  ActionIcon,
  Badge,
  Loader,
  Box,
  Checkbox,
} from "@mantine/core";
import {
  IconDownload,
  IconChevronDown,
  IconChevronRight,
  IconInfoCircle,
  IconFileText,
} from "@tabler/icons-react";
import { showNotification } from "@mantine/notifications";
import { getUrl } from "../../utils/url";
import { useMediaQuery } from "@mantine/hooks";

export interface ReportFile {
  path: string;
  name: string;
  hasAnalysis: boolean;
  analysis?: ReportAnalysis;
}

export interface ReportAnalysis {
  summary: string;
  keyInsights: string[];
  reportDate: number;
  reportTitle: string;
  reportType: string;
  filePath: string;
  analysisDate: number;
  metadata: Record<string, string>;
}

export const shortenReportName = (name: string): string => {
  // Shorten report names to fit in table cells, e.g. "SGX_Fund_Flow_Weekly_Tracker_Week_of_12_May_2025.xlsx"
  return name.replace(
    /(Fund_Flow_Weekly[A-Za-z_]+)\d{1,2}_[A-Za-z]{3}/,
    (match, p1) => match.replace(p1, "FF_")
  );
};

// Sort reports by date descending (newest first)
export const sortReportsByDate = (reportFiles: ReportFile[]): ReportFile[] => {
  return [...reportFiles].sort((a, b) => {
    const extractDate = (filename: string) => {
      // Extract date from filename like "SGX_Fund_Flow_Weekly_Tracker_Week_of_12_May_2025.xlsx"
      const match = filename.match(/Week_of_(\d+)_(\w+)_(\d+)/);
      if (match) {
        const [, day, month, year] = match;
        // Convert abbreviated month name to number
        const monthAbbreviations = [
          "Jan",
          "Feb",
          "Mar",
          "Apr",
          "May",
          "Jun",
          "Jul",
          "Aug",
          "Sep",
          "Oct",
          "Nov",
          "Dec",
        ];
        const monthIndex = monthAbbreviations.findIndex(
          (m) => m.toLowerCase() === month.toLowerCase().slice(0, 3)
        );
        if (monthIndex !== -1) {
          return new Date(parseInt(year), monthIndex, parseInt(day));
        }
      }
      return new Date(0); // fallback for unparseable dates
    };

    const dateA = extractDate(a.name);
    const dateB = extractDate(b.name);
    return dateB.getTime() - dateA.getTime(); // descending order
  });
};

const ReportsTable: React.FC = () => {
  const [reports, setReports] = useState<ReportFile[]>([]);
  const [loading, setLoading] = useState(false);
  const [downloadModalOpen, setDownloadModalOpen] = useState(false);
  const [downloadCount, setDownloadCount] = useState(1);
  const [reportType, setReportType] = useState<string>("fund flow");
  const [analyzeReports, setAnalyzeReports] = useState(false);
  const [forceReanalysis, setForceReanalysis] = useState(false);
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  const isMobile = useMediaQuery("(max-width: 768px)");

  // Fetch downloaded reports from disk
  const fetchReports = async () => {
    try {
      const response = await fetch(getUrl("/api/v1/analytics/list_files"));
      if (!response.ok) throw new Error("Failed to fetch reports");
      const reportPaths: string[] = await response.json();

      const reportFiles = reportPaths.map((path) => ({
        path,
        name: path.split("/").pop() || path,
        hasAnalysis: false,
      }));

      // Sort reports by date descending and set state
      const sortedReports = sortReportsByDate(reportFiles);
      setReports(sortedReports);
    } catch (error) {
      console.error("Error fetching reports:", error);
      showNotification({
        title: "Error",
        message: "Failed to fetch reports",
        color: "red",
      });
    }
  };

  // Fetch analysis results
  const fetchAnalyses = async () => {
    try {
      const response = await fetch(getUrl("/api/v1/analytics/list_analysis"));
      if (!response.ok) throw new Error("Failed to fetch analyses");
      const analysisData: ReportAnalysis[] = await response.json();
      return analysisData;
    } catch (error) {
      console.error("Error fetching analyses:", error);
      showNotification({
        title: "Error",
        message: "Failed to fetch analyses",
        color: "red",
      });
      return [];
    }
  };

  // Match reports with their analyses
  const matchReportsWithAnalyses = (
    reportFiles: ReportFile[],
    analysisData: ReportAnalysis[]
  ) => {
    const updatedReports = reportFiles.map((report) => {
      const analysis = analysisData.find((a) => {
        const analysisFileName = a.filePath.split("/").pop();
        return analysisFileName === report.name;
      });

      return {
        ...report,
        hasAnalysis: !!analysis,
        analysis,
      };
    });

    // Sort reports by date descending and set state
    const sortedReports = sortReportsByDate(updatedReports);
    setReports(sortedReports);
  };

  // Load data on component mount
  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      try {
        await fetchReports();
        const analysisData = await fetchAnalyses();

        // Re-fetch reports to ensure we have the latest data before matching
        const response = await fetch(getUrl("/api/v1/analytics/list_files"));
        if (response.ok) {
          const reportPaths: string[] = await response.json();
          const reportFiles = reportPaths.map((path) => ({
            path,
            name: path.split("/").pop() || path,
            hasAnalysis: false,
          }));

          matchReportsWithAnalyses(reportFiles, analysisData);
        }
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, []);

  // Download latest N reports
  const handleDownloadReports = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        n: downloadCount.toString(),
      });

      if (reportType) {
        params.append("type", reportType);
      }

      // First download the reports
      const downloadResponse = await fetch(
        getUrl(`/api/v1/analytics/download_latest_n?${params}`)
      );
      if (!downloadResponse.ok) throw new Error("Failed to download reports");

      const downloadedFiles: string[] = await downloadResponse.json();

      showNotification({
        title: "Success",
        message: `Downloaded ${downloadedFiles.length} reports`,
        color: "green",
      });

      // If analyze option is selected, also analyze the reports
      if (analyzeReports) {
        const analyzeParams = new URLSearchParams({
          n: downloadCount.toString(),
        });

        if (reportType) {
          analyzeParams.append("type", reportType);
        }

        if (forceReanalysis) {
          analyzeParams.append("force", "true");
        }

        const analyzeResponse = await fetch(
          getUrl(`/api/v1/analytics/analyze_latest_n?${analyzeParams}`)
        );

        if (analyzeResponse.ok) {
          const analyses: ReportAnalysis[] = await analyzeResponse.json();
          showNotification({
            title: "Analysis Complete",
            message: `Analyzed ${analyses.length} reports`,
            color: "blue",
          });
        } else {
          showNotification({
            title: "Analysis Warning",
            message: "Reports downloaded but analysis failed",
            color: "yellow",
          });
        }
      }

      setDownloadModalOpen(false);

      // Reset form state
      setAnalyzeReports(false);
      setForceReanalysis(false);

      // Refresh the reports list
      await fetchReports();
      const analysisData = await fetchAnalyses();

      const updatedResponse = await fetch(
        getUrl("/api/v1/analytics/list_files")
      );
      if (updatedResponse.ok) {
        const reportPaths: string[] = await updatedResponse.json();
        const reportFiles = reportPaths.map((path) => ({
          path,
          name: path.split("/").pop() || path,
          hasAnalysis: false,
        }));

        matchReportsWithAnalyses(reportFiles, analysisData);
      }
    } catch (error) {
      console.error("Error downloading reports:", error);
      showNotification({
        title: "Error",
        message: "Failed to download reports",
        color: "red",
      });
    } finally {
      setLoading(false);
    }
  };

  // Toggle row expansion
  const toggleRowExpansion = (reportName: string) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(reportName)) {
      newExpanded.delete(reportName);
    } else {
      newExpanded.add(reportName);
    }
    setExpandedRows(newExpanded);
  };

  // Format date
  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString();
  };

  // Render text with markdown-style formatting
  const renderFormattedText = (text: string) => {
    // Split text by ## headers and ** bold text, keeping the delimiters
    const parts = text.split(/(##[^#\n]+|\*\*[^*]+\*\*)/g);

    return parts.map((part, index) => {
      if (part.startsWith("##")) {
        // Header text - remove ## and render as heading
        return (
          <Text key={index} fw={700} size="md" mt="md" mb="sm" component="h4">
            {part.replace(/##\s*/, "")}
          </Text>
        );
      } else if (part.startsWith("**") && part.endsWith("**")) {
        // Bold text - remove ** and render as bold
        return (
          <Text key={index} size="sm" fw={600} component="span">
            {part.slice(2, -2)}
          </Text>
        );
      } else {
        // Regular text
        return <span key={index}>{part}</span>;
      }
    });
  };

  return (
    <Box>
      <Group justify="space-between" mb="md">
        <Text size="lg" fw={600}>
          Reports
        </Text>
        <Button
          leftSection={<IconDownload size={16} />}
          onClick={() => setDownloadModalOpen(true)}
          loading={loading}
        >
          Download Reports
        </Button>
      </Group>

      {loading && reports.length === 0 ? (
        <Group justify="center" py="xl">
          <Loader />
          <Text>Loading reports...</Text>
        </Group>
      ) : (
        <Table striped highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Report Name</Table.Th>
              <Table.Th>Analysis Status</Table.Th>
              <Table.Th>Analysis</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {reports.map((report) => (
              <React.Fragment key={report.name}>
                <Table.Tr>
                  <Table.Td>
                    <Group gap="xs">
                      <IconFileText size={16} />
                      <Text>
                        {isMobile
                          ? shortenReportName(report.name)
                          : report.name}
                      </Text>
                    </Group>
                  </Table.Td>
                  <Table.Td>
                    {report.hasAnalysis ? (
                      <Badge color="green" size="sm">
                        Analyzed
                      </Badge>
                    ) : (
                      <Badge color="gray" size="sm">
                        No Analysis
                      </Badge>
                    )}
                  </Table.Td>
                  <Table.Td>
                    {report.hasAnalysis && (
                      <ActionIcon
                        variant="subtle"
                        onClick={() => toggleRowExpansion(report.name)}
                      >
                        {expandedRows.has(report.name) ? (
                          <IconChevronDown size={16} />
                        ) : (
                          <IconChevronRight size={16} />
                        )}
                      </ActionIcon>
                    )}
                  </Table.Td>
                </Table.Tr>
                {report.hasAnalysis &&
                  expandedRows.has(report.name) &&
                  report.analysis && (
                    <Table.Tr>
                      <Table.Td colSpan={3}>
                        <Collapse in={expandedRows.has(report.name)}>
                          <Card withBorder p="md" mt="xs">
                            <Stack gap="sm">
                              <Group>
                                <Text fw={600}>Report Analysis</Text>
                                <Badge variant="light" color="blue">
                                  {report.analysis.reportType}
                                </Badge>
                              </Group>

                              <Group>
                                <Text size="sm" c="dimmed">
                                  Report Date:{" "}
                                  {formatDate(report.analysis.reportDate)}
                                </Text>
                                <Text size="sm" c="dimmed">
                                  Analysis Date:{" "}
                                  {formatDate(report.analysis.analysisDate)}
                                </Text>
                              </Group>

                              <div>
                                <Text fw={500} mb="xs">
                                  Summary:
                                </Text>
                                <div
                                  style={{
                                    fontSize: "14px",
                                    lineHeight: "1.5",
                                  }}
                                >
                                  {renderFormattedText(
                                    report.analysis.summary.replace(
                                      "**1. CONCISE SUMMARY:**",
                                      ""
                                    )
                                  )}
                                </div>
                              </div>

                              {report.analysis.keyInsights.length > 0 && (
                                <div>
                                  <Text fw={500} mb="xs">
                                    Key Insights:
                                  </Text>
                                  <Stack gap="xs">
                                    {report.analysis.keyInsights.map(
                                      (insight, index) => (
                                        <Alert
                                          key={index}
                                          icon={<IconInfoCircle size={16} />}
                                          variant="light"
                                          color="blue"
                                        >
                                          {renderFormattedText(insight)}
                                        </Alert>
                                      )
                                    )}
                                  </Stack>
                                </div>
                              )}
                            </Stack>
                          </Card>
                        </Collapse>
                      </Table.Td>
                    </Table.Tr>
                  )}
              </React.Fragment>
            ))}
            {reports.length === 0 && !loading && (
              <Table.Tr>
                <Table.Td colSpan={3} style={{ textAlign: "center" }}>
                  <Text c="dimmed" py="xl">
                    No reports found. Click "Download Reports" to get started.
                  </Text>
                </Table.Td>
              </Table.Tr>
            )}
          </Table.Tbody>
        </Table>
      )}

      {/* Download Modal */}
      <Modal
        opened={downloadModalOpen}
        onClose={() => {
          setDownloadModalOpen(false);
          setAnalyzeReports(false);
          setForceReanalysis(false);
        }}
        title="Download SGX Reports"
        size="md"
      >
        <Stack gap="md">
          <NumberInput
            label="Number of reports to download"
            description="How many of the latest reports to download"
            value={downloadCount}
            onChange={(value) => setDownloadCount(Number(value) || 1)}
            min={1}
            max={52}
            required
            allowNegative={false}
            allowDecimal={false}
            clampBehavior="none"
            placeholder="Enter number of reports"
          />

          <TextInput
            label="Report type filter"
            description="Filter by report type (e.g., 'fund flow', 'daily momentum')"
            placeholder="Leave empty to download all types"
            value={reportType}
            onChange={(event) => setReportType(event.currentTarget.value)}
          />

          <Checkbox
            label="Analyze reports after download"
            description="Automatically analyze downloaded reports with AI"
            checked={analyzeReports}
            onChange={(event) => setAnalyzeReports(event.currentTarget.checked)}
            mt="xs"
          />

          <Checkbox
            label="Force re-analysis"
            description="Force re-analysis with Gemini even if analysis already exists"
            checked={forceReanalysis}
            onChange={(event) =>
              setForceReanalysis(event.currentTarget.checked)
            }
            disabled={!analyzeReports}
          />

          <Group justify="flex-end">
            <Button
              variant="default"
              onClick={() => {
                setDownloadModalOpen(false);
                setAnalyzeReports(false);
                setForceReanalysis(false);
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleDownloadReports}
              loading={loading}
              leftSection={<IconDownload size={16} />}
            >
              Download
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
};

export default ReportsTable;
