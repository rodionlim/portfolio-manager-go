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
