import { V1ReportSpec } from "@rilldata/web-common/runtime-client";

export function getDashboardNameFromReport(
  reportSpec: V1ReportSpec | undefined,
): string | null {
  if (!reportSpec?.queryArgsJson) return null;
  const queryArgsJson = JSON.parse(reportSpec.queryArgsJson);
  return (
    queryArgsJson?.metrics_view_name ??
    queryArgsJson?.metricsViewName ??
    queryArgsJson?.metrics_view ??
    queryArgsJson?.metricsView ??
    null
  );
}
