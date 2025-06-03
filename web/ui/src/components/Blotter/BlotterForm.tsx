import {
  Autocomplete,
  Button,
  Container,
  Group,
  NumberInput,
  SimpleGrid,
  Switch,
  TextInput,
  Title,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useForm } from "@mantine/form";
import { DatePickerInput } from "@mantine/dates";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import { IsSGGovies, refDataByAssetClass } from "../../utils/referenceData";
import { useLocation } from "react-router-dom";
import { getUrl } from "../../utils/url";

export default function BlotterForm() {
  const location = useLocation();

  const defaultTrader = localStorage.getItem("defaultTrader") || "TraderA";
  const defaultBroker = localStorage.getItem("defaultBroker") || "DBS";
  const defaultAccount = localStorage.getItem("defaultAccount") || "CDP";

  const refData = useSelector((state: RootState) => state.referenceData.data);

  const form = useForm({
    mode: "uncontrolled",
    initialValues: {
      tradeId: location.state?.tradeId || "",
      date:
        location.state?.date ||
        (() => {
          // Create date at midnight
          const date = new Date();
          // Add 9 hours to get 9:00 AM local time
          date.setHours(9, 0, 0, 0);
          return date;
        })(),
      ticker: location.state?.ticker || "",
      trader: location.state?.trader || defaultTrader,
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
    },
    validate: {
      date: (value) => !value && "Date is required",
      ticker: (value) => value.length < 1 && "Ticker is required",
      account: (value) => value.length < 1 && "Account is required",
      status: (value) =>
        !["open", "autoclosed", "closed"].includes(value) &&
        "Status is required, and must be either open, autoclosed, or closed",
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
      date: values.date.toLocaleDateString("sv-SE") + "T00:00:00Z",
      price: values.price > 0 ? values.price : values.value / values.qty,
    }),
  });

  async function upsertTrade(
    values: Omit<typeof form.values, "date"> & { date: string }
  ) {
    const tradeTypeAction = !values.tradeId ? "add" : "update";
    const tradeTypeActionPastTense = !values.tradeId ? "added" : "updated";
    const baseCcy = "SGD"; // TODO: make this dynamic based on user settings or location

    if (values.fx === 0) {
      // Check if it is SG Govies, if so, set FX to 1
      if (IsSGGovies(values.ticker) && baseCcy === "SGD") {
        values.fx = 1; // SG Govies are always in SGD
      } else if (refData && refData[values.ticker]?.ccy) {
        const quoteCcy = refData[values.ticker].ccy;
        if (quoteCcy === baseCcy) {
          values.fx = 1;
        } else {
          const dt = values.date.replaceAll("-", "").slice(0, 8); // YYYYMMDD
          // fetch price as of historical date
          const url = getUrl(
            `api/v1/mdata/price/historical/${quoteCcy}-SGD?start=${dt}&end=${dt}`
          );
          const resp = await fetch(url);
          if (!resp.ok) {
            notifications.show({
              color: "red",
              title: "Error",
              message: `Unable to fetch FX rate for ${values.ticker}`,
            });
            throw new Error("Unable to fetch FX rate");
          }
          const vals = await resp.json();
          if (Array.isArray(vals) && vals.length === 0) {
            notifications.show({
              color: "red",
              title: "Error",
              message: `No FX rate found for ${values.ticker} on ${dt}`,
            });
            throw new Error("No FX rate found");
          }
          const price = vals[0]["Price"];
          values.fx = price;
        }
      } else {
        notifications.show({
          color: "red",
          title: "Error",
          message: `Unable to infer FX rate for ${values.ticker}`,
        });
        throw new Error("Unable to infer FX rate");
      }
    }

    const body = {
      id: values.tradeId,
      tradeDate: values.date, // need to convert to 2024-12-09T00:00:00Z
      ticker: values.ticker,
      trader: values.trader,
      broker: values.broker,
      account: values.account,
      status: values.status,
      originalTradeId: values.originalTradeId,
      quantity: values.qty,
      price: values.price,
      fx: values.fx, // Add FX rate to request body (0 means infer from backend)
      side: values.tradeType ? "sell" : "buy",
      seqNum: values.seqNum,
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

      notifications.show({
        title: "Trade successfully added",
        message: `Trade [${data.TradeID}] was successfully ${tradeTypeActionPastTense} in the blotter`,
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

  const handleSubmit = (
    values: Omit<typeof form.values, "date"> & { date: string }
  ) => {
    localStorage.setItem("defaultTrader", values.trader);
    localStorage.setItem("defaultBroker", values.broker);
    upsertTrade(values); // TODO: add error handling
  };

  return (
    <Container size="md">
      <Title order={2} mb="lg">
        {form.getValues().tradeId ? "Update" : "Add"} Trade to Blotter
      </Title>
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <SimpleGrid cols={2}>
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
            {...form.getInputProps("date")}
          />
          <Autocomplete
            withAsterisk
            label="Ticker"
            placeholder="ticker to be added, e.g. es3.si, sbjun25"
            data={refDataByAssetClass(refData)}
            key={form.key("ticker")}
            {...form.getInputProps("ticker")}
          />
          <TextInput
            withAsterisk
            label="Trader"
            placeholder="trader to be added, e.g. Trader A"
            key={form.key("trader")}
            {...form.getInputProps("trader")}
          />
          <TextInput
            withAsterisk
            label="Broker"
            placeholder="broker to be added, e.g. DBS"
            key={form.key("broker")}
            {...form.getInputProps("broker")}
          />
          <TextInput
            withAsterisk
            label="Account"
            placeholder="account to be added, e.g. CDP"
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

          <div />
          <Group justify="flex-end">
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
