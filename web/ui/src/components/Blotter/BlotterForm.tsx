import {
  Autocomplete,
  Box,
  Button,
  Container,
  Divider,
  Group,
  NumberInput,
  Select,
  SimpleGrid,
  Switch,
  Text,
  TextInput,
  Title,
  FileInput,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useForm } from "@mantine/form";
import { DatePickerInput } from "@mantine/dates";
import { useQuery } from "@tanstack/react-query";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import { Trade } from "../../types/blotter";
import { IsSGGovies, refDataByAssetClass } from "../../utils/referenceData";
import { useLocation } from "react-router-dom";
import { getUrl } from "../../utils/url";
import { useEffect, useMemo, useState } from "react";
import { IconPaperclip } from "@tabler/icons-react";

const instrumentTypeOptions = [
  { value: "outright", label: "Outright" },
  { value: "option", label: "Option" },
  { value: "future", label: "Future" },
];

const callPutOptions = [
  { value: "call", label: "Call" },
  { value: "put", label: "Put" },
];

const formatOptionStrike = (value: number) => Math.round(value).toString();

const buildOptionTickerPreview = (
  underlyingTicker: string,
  expiryDate: string,
  strikePrice: number,
  callPut: string,
) => {
  const normalizedUnderlying = underlyingTicker.trim().toUpperCase();
  const normalizedExpiry = expiryDate.trim().replaceAll("-", "");
  const normalizedCallPut = callPut.trim().toLowerCase();

  if (
    !normalizedUnderlying ||
    !normalizedExpiry ||
    strikePrice <= 0 ||
    !["call", "put"].includes(normalizedCallPut)
  ) {
    return "";
  }

  const callPutCode = normalizedCallPut === "put" ? "p" : "c";
  return `${normalizedUnderlying}_${normalizedExpiry}_${formatOptionStrike(strikePrice)}_${callPutCode}`.toUpperCase();
};

const formatDateForHistoricalQuery = (value: Date | null) => {
  if (!value) return "";
  const year = value.getFullYear();
  const month = `${value.getMonth() + 1}`.padStart(2, "0");
  const day = `${value.getDate()}`.padStart(2, "0");
  return `${year}${month}${day}`;
};

const fetchTrades = async (): Promise<Trade[]> => {
  const resp = await fetch(getUrl("/api/v1/blotter/trade"));
  if (!resp.ok) {
    throw new Error("Failed to fetch blotter trades");
  }

  return resp.json();
};

const buildAutocompleteOptions = (
  values: Array<string | null | undefined>,
): string[] =>
  Array.from(
    new Set(
      values
        .map((value) => value?.trim())
        .filter((value): value is string => Boolean(value)),
    ),
  ).sort((left, right) => left.localeCompare(right));

export default function BlotterForm() {
  const location = useLocation();
  const [confirmationFile, setConfirmationFile] = useState<File | null>(null);
  const [tradeDateValue, setTradeDateValue] = useState<Date | null>(
    location.state?.date instanceof Date
      ? location.state.date
      : location.state?.date
        ? new Date(location.state.date)
        : (() => {
            const date = new Date();
            date.setHours(9, 0, 0, 0);
            return date;
          })(),
  );
  const [instrumentTypeValue, setInstrumentTypeValue] = useState(
    location.state?.instrumentType || "outright",
  );
  const [underlyingTickerValue, setUnderlyingTickerValue] = useState(
    location.state?.underlyingTicker || "",
  );
  const [expiryDateValue, setExpiryDateValue] = useState(
    location.state?.expiryDate || "",
  );
  const [strikePriceValue, setStrikePriceValue] = useState<number | string>(
    location.state?.strikePrice || 0,
  );
  const [callPutValue, setCallPutValue] = useState(
    location.state?.callPut || "",
  );

  const defaultBook = localStorage.getItem("defaultBook") || "Main";
  const defaultBroker = localStorage.getItem("defaultBroker") || "DBS";
  const defaultAccount = localStorage.getItem("defaultAccount") || "CDP";

  const { data: trades = [] } = useQuery({
    queryKey: ["trades"],
    queryFn: fetchTrades,
    retry: false,
  });

  const refData = useSelector((state: RootState) => state.referenceData.data);
  const tickerOptions = useMemo(() => refDataByAssetClass(refData), [refData]);
  const bookOptions = useMemo(
    () =>
      buildAutocompleteOptions([
        defaultBook,
        location.state?.book,
        ...trades.map((trade) => trade.Book),
      ]),
    [defaultBook, location.state?.book, trades],
  );
  const brokerOptions = useMemo(
    () =>
      buildAutocompleteOptions([
        defaultBroker,
        location.state?.broker,
        ...trades.map((trade) => trade.Broker),
      ]),
    [defaultBroker, location.state?.broker, trades],
  );
  const accountOptions = useMemo(
    () =>
      buildAutocompleteOptions([
        defaultAccount,
        location.state?.account,
        ...trades.map((trade) => trade.Account),
      ]),
    [defaultAccount, location.state?.account, trades],
  );

  const normalizeString = (value: unknown) =>
    typeof value === "string"
      ? value.trim()
      : value
        ? String(value).trim()
        : "";

  const normalizeNumber = (value: unknown) => {
    const num = typeof value === "number" ? value : Number(value);
    return Number.isNaN(num) ? 0 : num;
  };

  const normalizeDate = (value: unknown) => {
    if (!value) return "";
    const date = value instanceof Date ? value : new Date(String(value));
    if (Number.isNaN(date.getTime())) return "";
    return date.toLocaleDateString("sv-SE") + "T00:00:00Z";
  };

  const normalizeTradeType = (value: unknown) => {
    if (typeof value === "boolean") return value;
    if (typeof value === "string") {
      const lower = value.toLowerCase();
      if (lower === "sell") return true;
      if (lower === "buy") return false;
    }
    return Boolean(value);
  };

  const buildComparableTrade = (values: {
    date?: unknown;
    ticker?: unknown;
    book?: unknown;
    broker?: unknown;
    account?: unknown;
    status?: unknown;
    originalTradeId?: unknown;
    qty?: unknown;
    price?: unknown;
    value?: unknown;
    fx?: unknown;
    tradeType?: unknown;
    seqNum?: unknown;
    instrumentType?: unknown;
    underlyingTicker?: unknown;
    underlyingSpotRef?: unknown;
    expiryDate?: unknown;
    strikePrice?: unknown;
    callPut?: unknown;
  }) => {
    const qty = normalizeNumber(values.qty);
    const priceRaw = normalizeNumber(values.price);
    const valueRaw = normalizeNumber(values.value);
    const price =
      priceRaw > 0
        ? priceRaw
        : valueRaw > 0 && qty > 0
          ? valueRaw / qty
          : priceRaw;

    return {
      date: normalizeDate(values.date),
      ticker: normalizeString(values.ticker).toUpperCase(),
      book: normalizeString(values.book),
      broker: normalizeString(values.broker),
      account: normalizeString(values.account),
      status: normalizeString(values.status).toLowerCase(),
      originalTradeId: normalizeString(values.originalTradeId),
      qty,
      price,
      fx: normalizeNumber(values.fx),
      tradeType: normalizeTradeType(values.tradeType),
      seqNum: normalizeNumber(values.seqNum),
      instrumentType: normalizeString(values.instrumentType).toLowerCase(),
      underlyingTicker: normalizeString(values.underlyingTicker).toUpperCase(),
      underlyingSpotRef: normalizeNumber(values.underlyingSpotRef),
      expiryDate: normalizeString(values.expiryDate),
      strikePrice: normalizeNumber(values.strikePrice),
      callPut: normalizeString(values.callPut).toLowerCase(),
    };
  };

  const areNumbersClose = (a: number, b: number, epsilon = 1e-6) =>
    Math.abs(a - b) <= epsilon;

  const areTradesEqual = (
    a: ReturnType<typeof buildComparableTrade>,
    b: ReturnType<typeof buildComparableTrade>,
  ) =>
    a.date === b.date &&
    a.ticker === b.ticker &&
    a.book === b.book &&
    a.broker === b.broker &&
    a.account === b.account &&
    a.status === b.status &&
    a.originalTradeId === b.originalTradeId &&
    areNumbersClose(a.qty, b.qty) &&
    areNumbersClose(a.price, b.price) &&
    areNumbersClose(a.fx, b.fx) &&
    a.tradeType === b.tradeType &&
    areNumbersClose(a.seqNum, b.seqNum) &&
    a.instrumentType === b.instrumentType &&
    a.underlyingTicker === b.underlyingTicker &&
    areNumbersClose(a.underlyingSpotRef, b.underlyingSpotRef) &&
    a.expiryDate === b.expiryDate &&
    areNumbersClose(a.strikePrice, b.strikePrice) &&
    a.callPut === b.callPut;

  const form = useForm({
    mode: "uncontrolled",
    initialValues: {
      tradeId: location.state?.tradeId || "",
      date: tradeDateValue,
      ticker: location.state?.ticker || "",
      book: location.state?.book || defaultBook,
      broker: location.state?.broker || defaultBroker,
      account: location.state?.account || defaultAccount,
      status: location.state?.status || "open",
      originalTradeId: location.state?.originalTradeId || "",
      qty: location.state?.qty || 0,
      price: location.state?.price || 0,
      value: 0, // either value or price must be specified
      fx: location.state?.fx || 0, // optional: 0 means the rate will be inferred
      tradeType: location.state?.tradeType || false, // false for BUY, true for SELL
      seqNum: location.state?.seqNum || 0,
      instrumentType: location.state?.instrumentType || "outright",
      underlyingTicker: location.state?.underlyingTicker || "",
      underlyingSpotRef: location.state?.underlyingSpotRef || 0,
      expiryDate: location.state?.expiryDate || "",
      strikePrice: location.state?.strikePrice || 0,
      callPut: location.state?.callPut || "",
    },
    validate: {
      date: (value) => !value && "Date is required",
      ticker: (value) => value.length < 1 && "Ticker is required",
      account: (value) => value.length < 1 && "Account is required",
      status: (value) =>
        !["open", "autoclosed", "closed"].includes(value) &&
        "Status is required, and must be either open, autoclosed, or closed",
      instrumentType: (value) =>
        !["outright", "option", "future"].includes(
          String(value).toLowerCase(),
        ) && "Instrument type must be either outright, option, or future",
      underlyingTicker: (value, values) =>
        values.instrumentType === "option" && !String(value).trim()
          ? "Underlying ticker is required for option trades"
          : null,
      expiryDate: (value, values) =>
        values.instrumentType === "option" && !String(value).trim()
          ? "Expiry date is required for option trades"
          : null,
      strikePrice: (value, values) =>
        values.instrumentType === "option" && Number(value) <= 0
          ? "Strike price must be greater than 0 for option trades"
          : values.instrumentType === "option" &&
              !Number.isInteger(Number(value))
            ? "Strike price must be a whole number for option trades"
            : null,
      callPut: (value, values) =>
        values.instrumentType === "option" &&
        !["call", "put"].includes(String(value).toLowerCase())
          ? "Call/Put is required for option trades"
          : null,
      qty: (value) => value <= 0 && "Quantity must be greater than 0",
      price: (value, values) => {
        if (value <= 0 && values.value <= 0) {
          return "Either Price or Value must be specified";
        }
        if (value > 0 && values.value > 0) {
          return "Either Price or Value must be specified, not both";
        }
        return null;
      },
      value: (value, values) => {
        if (value <= 0 && values.price <= 0) {
          return "Either Price or Value must be specified";
        }
        if (value > 0 && values.price > 0) {
          return "Either Price or Value must be specified, not both";
        }
        return null;
      },
    },
    transformValues: (values) => ({
      ...values,
      ticker: values.ticker.toUpperCase(),
      date:
        values.date instanceof Date
          ? values.date.toLocaleDateString("sv-SE") + "T00:00:00Z"
          : "",
      price: values.price > 0 ? values.price : values.value / values.qty,
      underlyingTicker: values.underlyingTicker.toUpperCase(),
    }),
  });

  const isOptionTrade = instrumentTypeValue === "option";
  const optionTickerPreview = useMemo(
    () =>
      buildOptionTickerPreview(
        underlyingTickerValue,
        expiryDateValue,
        Number(strikePriceValue) || 0,
        callPutValue,
      ),
    [underlyingTickerValue, expiryDateValue, strikePriceValue, callPutValue],
  );

  useEffect(() => {
    if (!isOptionTrade) {
      return;
    }

    if (
      optionTickerPreview &&
      form.getValues().ticker !== optionTickerPreview
    ) {
      form.setFieldValue("ticker", optionTickerPreview);
    }
  }, [form, isOptionTrade, optionTickerPreview]);

  useEffect(() => {
    if (!isOptionTrade || !underlyingTickerValue || !tradeDateValue) {
      return;
    }

    if (Number(form.getValues().underlyingSpotRef) > 0) {
      return;
    }

    let cancelled = false;
    const tradeDateKey = formatDateForHistoricalQuery(tradeDateValue);

    const populateSpotReference = async () => {
      const normalizedUnderlying = underlyingTickerValue.trim().toUpperCase();
      const historicalUrl = getUrl(
        `/api/v1/mdata/price/historical/${encodeURIComponent(normalizedUnderlying)}?start=${tradeDateKey}&end=${tradeDateKey}`,
      );

      try {
        const historicalResp = await fetch(historicalUrl);
        if (!historicalResp.ok) {
          throw new Error("Historical spot lookup failed");
        }

        const historicalData = await historicalResp.json();
        const historicalPrice = Array.isArray(historicalData)
          ? historicalData[historicalData.length - 1]?.Price ||
            historicalData[historicalData.length - 1]?.price
          : undefined;

        if (!cancelled && historicalPrice && Number(historicalPrice) > 0) {
          form.setFieldValue("underlyingSpotRef", Number(historicalPrice));
          return;
        }

        throw new Error("Historical spot not available");
      } catch {
        const currentResp = await fetch(
          getUrl(
            `/api/v1/mdata/price/${encodeURIComponent(normalizedUnderlying)}`,
          ),
        );
        if (!currentResp.ok) {
          return;
        }

        const currentData = await currentResp.json();
        const currentPrice = currentData.Price || currentData.price;
        if (!cancelled && currentPrice && Number(currentPrice) > 0) {
          form.setFieldValue("underlyingSpotRef", Number(currentPrice));
        }
      }
    };

    populateSpotReference();

    return () => {
      cancelled = true;
    };
  }, [form, isOptionTrade, tradeDateValue, underlyingTickerValue]);

  async function upsertTrade(
    values: Omit<typeof form.values, "date"> & { date: string },
  ) {
    const tradeTypeAction = !values.tradeId ? "add" : "update";
    const tradeTypeActionPastTense = !values.tradeId ? "added" : "updated";
    const baseCcy = "SGD"; // TODO: make this dynamic based on user settings or location
    let usedCurrentRate = false;
    let resolvedFx = normalizeNumber(values.fx);

    const fxLookupTicker =
      values.instrumentType === "option"
        ? values.underlyingTicker
        : values.ticker;

    if (resolvedFx === 0) {
      // Check if it is SG Govies, if so, set FX to 1
      if (IsSGGovies(fxLookupTicker) && baseCcy === "SGD") {
        resolvedFx = 1; // SG Govies are always in SGD
      } else if (refData && refData[fxLookupTicker]?.ccy) {
        const quoteCcy = refData[fxLookupTicker].ccy;
        if (quoteCcy === baseCcy) {
          resolvedFx = 1;
        } else {
          const dt = values.date.replaceAll("-", "").slice(0, 8); // YYYYMMDD
          // fetch price as of historical date
          const historicalUrl = getUrl(
            `api/v1/mdata/price/historical/${quoteCcy}-SGD?start=${dt}&end=${dt}`,
          );

          try {
            const resp = await fetch(historicalUrl);
            if (!resp.ok) {
              throw new Error("Unable to fetch historical FX rate");
            }
            const vals = await resp.json();
            if (Array.isArray(vals) && vals.length === 0) {
              throw new Error("No historical FX rate found");
            }
            const price = Number(vals[0]["Price"] || vals[0]["price"]);
            if (!Number.isFinite(price) || price <= 0) {
              throw new Error("Invalid historical FX rate");
            }
            resolvedFx = price;
          } catch (error) {
            // Fallback to current FX rate
            console.warn(
              `Historical FX rate not available, using current rate: ${error}`,
            );
            const currentUrl = getUrl(`api/v1/mdata/price/${quoteCcy}-SGD`);
            const currentResp = await fetch(currentUrl);
            if (!currentResp.ok) {
              notifications.show({
                color: "red",
                title: "Error",
                message: `Unable to fetch FX rate for ${fxLookupTicker}`,
              });
              throw new Error("Unable to fetch FX rate");
            }
            const currentData = await currentResp.json();
            const currentPrice = Number(currentData.Price || currentData.price);
            if (!Number.isFinite(currentPrice) || currentPrice <= 0) {
              notifications.show({
                color: "red",
                title: "Error",
                message: `Unable to infer FX rate for ${fxLookupTicker}`,
              });
              throw new Error("Unable to infer FX rate");
            }
            resolvedFx = currentPrice;
            usedCurrentRate = true;
          }
        }
      } else {
        notifications.show({
          color: "red",
          title: "Error",
          message: `Unable to infer FX rate for ${fxLookupTicker}`,
        });
        throw new Error("Unable to infer FX rate");
      }
    }

    form.setFieldValue("fx", resolvedFx);

    const body = {
      id: values.tradeId,
      tradeDate: values.date, // need to convert to 2024-12-09T00:00:00Z
      ticker: values.ticker,
      book: values.book,
      broker: values.broker,
      account: values.account,
      status: values.status,
      origTradeID: values.originalTradeId,
      quantity: values.qty,
      price: values.price,
      fx: resolvedFx, // Add FX rate to request body (0 means infer from backend)
      side: values.tradeType ? "sell" : "buy",
      seqNum: values.seqNum,
      instrumentType: values.instrumentType,
      underlyingTicker: values.underlyingTicker,
      underlyingSpotRef: values.underlyingSpotRef,
      expiryDate: values.expiryDate,
      strikePrice: values.strikePrice,
      callPut: values.callPut,
    };

    try {
      const resp = await fetch(getUrl("api/v1/blotter/trade"), {
        method: values.tradeId ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      });

      if (!resp.ok) {
        const error = await resp.json();
        throw new Error(error.message || "An error occurred");
      }

      const data = await resp.json();
      const fxParsed = parseFloat(resolvedFx.toFixed(4));

      notifications.show({
        title: "Trade successfully added",
        message: usedCurrentRate
          ? `Trade [${data.TradeID}] was successfully ${tradeTypeActionPastTense} in the blotter using current FX rate [${fxParsed}] (historical rate unavailable)`
          : `Trade [${data.TradeID}] was successfully ${tradeTypeActionPastTense} in the blotter using FX rate [${fxParsed}]`,
        autoClose: 6000,
      });

      return data;
    } catch (error) {
      console.error(error);
      notifications.show({
        color: "red",
        title: "Error",
        message: `Unable to ${tradeTypeAction} trade to the blotter\n ${error}`,
      });
      throw new Error("An error occurred while submitting trade to blotter");
    }
  }

  const initialTradeSnapshot = useMemo(() => {
    if (!location.state?.tradeId) return null;
    return buildComparableTrade({
      date: location.state?.date,
      ticker: location.state?.ticker,
      book: location.state?.book || defaultBook,
      broker: location.state?.broker || defaultBroker,
      account: location.state?.account || defaultAccount,
      status: location.state?.status || "open",
      originalTradeId:
        location.state?.origTradeID || location.state?.originalTradeId || "",
      qty: location.state?.qty || 0,
      price: location.state?.price || 0,
      value: location.state?.value || 0,
      fx: location.state?.fx || 0,
      tradeType: location.state?.tradeType || false,
      seqNum: location.state?.seqNum || 0,
      instrumentType: location.state?.instrumentType || "outright",
      underlyingTicker: location.state?.underlyingTicker || "",
      underlyingSpotRef: location.state?.underlyingSpotRef || 0,
      expiryDate: location.state?.expiryDate || "",
      strikePrice: location.state?.strikePrice || 0,
      callPut: location.state?.callPut || "",
    });
  }, [location.state, defaultBook, defaultBroker, defaultAccount]);

  const handleSubmit = async (
    values: Omit<typeof form.values, "date"> & { date: string },
  ) => {
    localStorage.setItem("defaultBook", values.book);
    localStorage.setItem("defaultBroker", values.broker);
    localStorage.setItem("defaultAccount", values.account);

    try {
      const isUpdate = Boolean(values.tradeId);
      const currentTradeSnapshot = buildComparableTrade(values);

      if (isUpdate && initialTradeSnapshot) {
        const hasChanges = !areTradesEqual(
          currentTradeSnapshot,
          initialTradeSnapshot,
        );

        if (!hasChanges) {
          if (confirmationFile && values.tradeId) {
            await uploadConfirmation(values.tradeId, confirmationFile);
          }
          return;
        }
      }

      const trade = await upsertTrade(values);

      // Upload confirmation if provided
      if (confirmationFile && trade.TradeID) {
        await uploadConfirmation(trade.TradeID, confirmationFile);
      }
    } catch (error) {
      console.error("Error submitting trade:", error);
    }
  };

  const uploadConfirmation = async (tradeId: string, file: File) => {
    const formData = new FormData();
    formData.append("file", file);

    try {
      const resp = await fetch(
        getUrl(`/api/v1/blotter/confirmation/${tradeId}`),
        {
          method: "POST",
          body: formData,
        },
      );

      if (!resp.ok) {
        throw new Error("Failed to upload confirmation");
      }

      notifications.show({
        title: "Confirmation uploaded",
        message: "Trade confirmation was successfully uploaded",
        autoClose: 5000,
      });

      setConfirmationFile(null);
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      notifications.show({
        color: "red",
        title: "Confirmation upload failed",
        message: `Failed to upload confirmation: ${errorMessage}`,
      });
    }
  };

  return (
    <Container size="md">
      <Title order={2} mb="sm">
        {form.getValues().tradeId ? "Update" : "Add"} Trade to Blotter
      </Title>
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="sm" verticalSpacing="sm">
          {form.getValues().tradeId && (
            <TextInput
              withAsterisk
              label="Trade ID"
              placeholder="trade id to be updated"
              disabled={true}
              key={form.key("tradeId")}
              {...form.getInputProps("tradeId")}
            />
          )}
          <DatePickerInput
            withAsterisk
            clearable
            label="Trade Date"
            placeholder="Select the trade date"
            key={form.key("date")}
            value={tradeDateValue}
            onChange={(value) => {
              setTradeDateValue(value);
              form.setFieldValue("date", value);
            }}
            error={form.errors.date}
          />
          <Select
            withAsterisk
            label="Instrument Type"
            data={instrumentTypeOptions}
            key={form.key("instrumentType")}
            value={instrumentTypeValue}
            onChange={(value) => {
              const normalizedValue = value || "outright";
              setInstrumentTypeValue(normalizedValue);
              form.setFieldValue("instrumentType", normalizedValue);
              if (normalizedValue !== "option") {
                setExpiryDateValue("");
                setStrikePriceValue(0);
                setCallPutValue("");
                form.setFieldValue("underlyingSpotRef", 0);
                form.setFieldValue("expiryDate", "");
                form.setFieldValue("strikePrice", 0);
                form.setFieldValue("callPut", "");
              }
            }}
            error={form.errors.instrumentType}
          />
          {isOptionTrade ? (
            <Autocomplete
              withAsterisk
              label="Underlying Ticker"
              placeholder="underlying ticker, e.g. AAPL"
              description="Used to derive the option preview and spot reference"
              data={tickerOptions}
              value={underlyingTickerValue}
              onChange={(value) => {
                const normalizedValue = value.toUpperCase();
                setUnderlyingTickerValue(normalizedValue);
                form.setFieldValue("underlyingTicker", normalizedValue);
              }}
              error={form.errors.underlyingTicker}
            />
          ) : (
            <Autocomplete
              withAsterisk
              label="Ticker"
              placeholder="ticker to be added, e.g. es3.si, sbjun25"
              data={tickerOptions}
              key={form.key("ticker")}
              {...form.getInputProps("ticker")}
            />
          )}
          {!isOptionTrade && (
            <Autocomplete
              label="Underlying Ticker"
              placeholder="Defaults to the ticker when left empty"
              description="Optional grouping tag stored on this trade"
              data={tickerOptions}
              value={underlyingTickerValue}
              onChange={(value) => {
                const normalizedValue = value.toUpperCase();
                setUnderlyingTickerValue(normalizedValue);
                form.setFieldValue("underlyingTicker", normalizedValue);
              }}
              error={form.errors.underlyingTicker}
            />
          )}
          {isOptionTrade && (
            <TextInput
              label="Option Ticker Preview"
              value={optionTickerPreview || form.getValues().ticker}
              readOnly
              description="Auto-generated from the option fields"
            />
          )}
          <Autocomplete
            withAsterisk
            label="Book"
            placeholder="book to be added, e.g. Main Book"
            data={bookOptions}
            key={form.key("book")}
            {...form.getInputProps("book")}
          />
          <Autocomplete
            withAsterisk
            label="Broker"
            placeholder="broker to be added, e.g. DBS"
            data={brokerOptions}
            key={form.key("broker")}
            {...form.getInputProps("broker")}
          />
          <Autocomplete
            withAsterisk
            label="Account"
            placeholder="account to be added, e.g. CDP"
            data={accountOptions}
            key={form.key("account")}
            {...form.getInputProps("account")}
          />
          <Autocomplete
            withAsterisk
            label="Status"
            placeholder="status to be added, e.g. open"
            data={["open", "autoclosed", "closed"]}
            key={form.key("status")}
            {...form.getInputProps("status")}
          />
          <TextInput
            label="Original Trade ID"
            placeholder="original trade id"
            key={form.key("originalTradeId")}
            {...form.getInputProps("originalTradeId")}
          />
          <NumberInput
            withAsterisk
            label="Quantity"
            placeholder="Quantity"
            thousandSeparator=","
            allowDecimal={true}
            key={form.key("qty")}
            {...form.getInputProps("qty")}
          />
          <NumberInput
            label="Price"
            placeholder="Price"
            allowDecimal={true}
            decimalScale={4}
            key={form.key("price")}
            {...form.getInputProps("price")}
          />
          <NumberInput
            label="Value (Only specify either Price or Value)"
            placeholder="Value"
            allowDecimal={true}
            decimalScale={4}
            key={form.key("value")}
            {...form.getInputProps("value")}
          />
          <NumberInput
            label="FX Rate (Set 0 to infer)"
            placeholder="FX Rate"
            allowDecimal={true}
            decimalScale={4}
            key={form.key("fx")}
            {...form.getInputProps("fx")}
          />
          {isOptionTrade && (
            <>
              <TextInput
                withAsterisk
                label="Expiry Date"
                placeholder="YYYY-MM-DD"
                value={expiryDateValue}
                onChange={(event) => {
                  const value = event.currentTarget.value;
                  setExpiryDateValue(value);
                  form.setFieldValue("expiryDate", value);
                }}
                error={form.errors.expiryDate}
              />
              <NumberInput
                withAsterisk
                label="Strike Price"
                placeholder="Strike"
                allowDecimal={false}
                decimalScale={0}
                value={strikePriceValue}
                onChange={(value) => {
                  setStrikePriceValue(value || 0);
                  form.setFieldValue("strikePrice", Number(value) || 0);
                }}
                error={form.errors.strikePrice}
              />
              <Select
                withAsterisk
                label="Call / Put"
                data={callPutOptions}
                value={callPutValue}
                onChange={(value) => {
                  const normalizedValue = value || "";
                  setCallPutValue(normalizedValue);
                  form.setFieldValue("callPut", normalizedValue);
                }}
                error={form.errors.callPut}
              />
              <NumberInput
                label="Underlying Spot Reference"
                placeholder="Auto-filled from historical spot if left empty"
                allowDecimal={true}
                decimalScale={4}
                key={form.key("underlyingSpotRef")}
                {...form.getInputProps("underlyingSpotRef")}
              />
            </>
          )}

          <FileInput
            label="Trade Confirmation (Optional)"
            placeholder="Upload confirmation file"
            accept="application/pdf,image/png,image/jpeg"
            value={confirmationFile}
            onChange={setConfirmationFile}
            leftSection={<IconPaperclip size={16} />}
            clearable
          />

          {isOptionTrade && (
            <Box style={{ gridColumn: "1 / -1" }}>
              <Divider my={4} label="Option Booking" labelPosition="center" />
              <Text size="xs" c="dimmed" mt={4}>
                Spot reference is persisted on the trade. Open options keep zero
                live MV and unrealized PnL until the closing trade is booked.
              </Text>
            </Box>
          )}

          <Group justify="flex-end" style={{ gridColumn: "1 / -1" }} mt={4}>
            <Switch
              size="xl"
              onLabel="SELL"
              offLabel="BUY"
              key={form.key("tradeType")}
              {...form.getInputProps("tradeType")}
            />

            <Button type="submit">Submit</Button>
          </Group>
        </SimpleGrid>
      </form>
    </Container>
  );
}
