/**
 * Formats a date as YYYYMMDD string
 * @param date - The date to format (defaults to current date)
 * @returns Date string in YYYYMMDD format
 */
export const formatDateAsYYYYMMDD = (date: Date = new Date()): string => {
  const year = date.getFullYear().toString();
  const month = (date.getMonth() + 1).toString().padStart(2, "0");
  const day = date.getDate().toString().padStart(2, "0");

  return `${year}${month}${day}`;
};

/**
 * Gets current date formatted as YYYYMMDD string
 * @returns Current date string in YYYYMMDD format
 */
export const getCurrentDateAsYYYYMMDD = (): string => {
  return formatDateAsYYYYMMDD();
};
