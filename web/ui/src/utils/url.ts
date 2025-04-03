export function getUrl(apiPath: string = "api/v1/refdata"): string {
  // remove all leading slashes at the start
  while (apiPath.length > 0 && apiPath[0] === "/") {
    apiPath = apiPath.substring(1);
  }

  // In development mode, use a fixed backend URL
  if (process.env.NODE_ENV === "development") {
    return `http://localhost:8080/${apiPath}`;
  }

  // In production, use the current browser's URL
  return `${window.location.protocol}//${window.location.host}/${apiPath}`;
}
