import {
  Autocomplete,
  Button,
  Container,
  Group,
  NumberInput,
  Stack,
  Switch,
  TextInput,
  Title,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useForm } from "@mantine/form";
import { DatePickerInput } from "@mantine/dates";
import { useSelector } from "react-redux";
import { RootState } from "../../store";
import { refDataByAssetClass } from "../../utils/referenceData";

export default function BlotterForm() {
  const defaultTrader = localStorage.getItem("defaultTrader") || "TraderA";
  const defaultBroker = localStorage.getItem("defaultBroker") || "DBS";
  const defaultAccount = localStorage.getItem("defaultAccount") || "CDP";

  const refData = useSelector((state: RootState) => state.referenceData.data);

  const form = useForm({
    mode: "uncontrolled",
    initialValues: {
      date: new Date(),
      ticker: "",
      trader: defaultTrader,
      broker: defaultBroker,
      account: defaultAccount,
      qty: 0,
      price: 0,
      value: 0, // either value or price must be specified
      tradeType: false, // false for BUY, true for SELL
    },
    validate: {
      date: (value) => !value && "Date is required",
      ticker: (value) => value.length < 1 && "Ticker is required",
      account: (value) => value.length < 1 && "Account is required",
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
      date: values.date.toISOString().split(".")[0] + "Z",
    }),
  });

  function addTrade(
    values: Omit<typeof form.values, "date"> & { date: string }
  ) {
    return fetch("http://localhost:8080/api/v1/blotter/trade", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        tradeDate: values.date, // need to convert to 2024-12-09T00:00:00Z
        ticker: values.ticker,
        trader: values.trader,
        broker: values.broker,
        account: values.account,
        quantity: values.qty,
        price: values.price,
        side: values.tradeType ? "sell" : "buy",
      }),
    })
      .then((resp) => {
        if (!resp.ok) {
          return resp.json().then((error) => {
            throw new Error(error.message || "An error occurred");
          });
        }
        return resp.json();
      })
      .then((data) => {
        notifications.show({
          title: "Trade successfully added",
          message: `Trade [${data.TradeID}] was successfully added to the blotter`,
          autoClose: 10000,
        });
      })
      .catch((error) => {
        console.error(error);
        notifications.show({
          color: "red",
          title: "Error",
          message: `Unable to add trade to the blotter\n ${error}`,
        });
        throw new Error("An error occurred while submitting trade to blotter");
      });
  }

  const handleSubmit = (
    values: Omit<typeof form.values, "date"> & { date: string }
  ) => {
    localStorage.setItem("defaultTrader", values.trader);
    localStorage.setItem("defaultBroker", values.broker);
    addTrade(values); // TODO: add error handling
  };

  return (
    <Container size="sm">
      <Title order={2} mb="lg">
        Add Trade to Blotter
      </Title>
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
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
            placeholder="ticker to be added, e.g. es3.si"
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
          <NumberInput
            withAsterisk
            label="Quantity"
            placeholder="Quantity"
            thousandSeparator=","
            allowDecimal={false}
            key={form.key("qty")}
            {...form.getInputProps("qty")}
          />
          <NumberInput
            withAsterisk
            label="Price"
            placeholder="Price"
            allowDecimal={true}
            decimalScale={4}
            key={form.key("price")}
            {...form.getInputProps("price")}
          />
          <NumberInput
            withAsterisk
            label="Value"
            placeholder="Value"
            allowDecimal={true}
            decimalScale={4}
            key={form.key("value")}
            {...form.getInputProps("value")}
          />

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
        </Stack>
      </form>
    </Container>
  );
}
