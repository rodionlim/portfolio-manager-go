export function getUrl(apiPath: string = "api/v1/refdata"): string {
  // In development mode, use a fixed backend URL
  if (process.env.NODE_ENV === "development") {
    return `http://localhost:8080/${apiPath}`;
  }

  // In production, use the current browser's URL
  return `${window.location.protocol}//${window.location.host}/${apiPath}`;
}
