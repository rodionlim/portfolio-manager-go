import {
  Button,
  Container,
  Group,
  NumberInput,
  Stack,
  Switch,
  TextInput,
  Title,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { DatePickerInput } from "@mantine/dates";
import { useState } from "react";

export default function BlotterForm() {
  const form = useForm({
    mode: "uncontrolled",
    initialValues: {
      date: null,
      ticker: "",
      trader: "traderA",
      broker: "dbs",
      qty: 0,
      price: 0,
      tradeType: false, // false for BUY, true for SELL
    },
    validate: {
      ticker: (value) => value.length < 1 && "Ticker is required",
      qty: (value) => value <= 0 && "Quantity must be greater than 0",
      price: (value) => value <= 0 && "Price must be greater than 0",
    },
  });

  const [tradeDt, setTradeDt] = useState<Date | null>(null);

  return (
    <Container size="sm">
      <Title order={2} mb="lg">
        Add Trade to Blotter
      </Title>
      <form onSubmit={form.onSubmit((values) => console.log(values))}>
        <Stack gap="md">
          <DatePickerInput
            withAsterisk
            label="Trade Date"
            placeholder="Select the trade date"
            value={tradeDt}
            onChange={setTradeDt}
          />
          <TextInput
            withAsterisk
            label="Ticker"
            placeholder="ticker to be added, e.g. es3.si"
            key={form.key("ticker")}
            {...form.getInputProps("ticker")}
          />
          <TextInput
            withAsterisk
            label="Trader"
            placeholder="trader to be added, e.g. trader A"
            key={form.key("trader")}
            {...form.getInputProps("trader")}
          />
          <TextInput
            withAsterisk
            label="Broker"
            placeholder="broker to be added, e.g. dbs"
            key={form.key("broker")}
            {...form.getInputProps("broker")}
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
