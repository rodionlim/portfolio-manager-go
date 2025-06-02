import { ReferenceData } from "../types";

export function refDataByAssetClass(
  refData: ReferenceData | null
): { group: string; items: string[] }[] {
  if (!refData) return [];
  const grp = Object.values(refData).reduce((acc, item) => {
    if (!acc[item.asset_class]) {
      acc[item.asset_class] = [];
    }
    acc[item.asset_class].push(item.id);
    return acc;
  }, {} as Record<string, string[]>);

  return Object.keys(grp).map((k) => {
    return { group: k, items: grp[k] };
  });
}

export function IsSSB(ticker: string): boolean {
  return ticker.startsWith("SB") && ticker.length === 7;
}

export function IsSgTBill(ticker: string): boolean {
  return (
    (ticker.startsWith("BS") || ticker.startsWith("BY")) && ticker.length === 8
  );
}

export function IsSGGovies(ticker: string): boolean {
  return IsSSB(ticker) || IsSgTBill(ticker);
}
