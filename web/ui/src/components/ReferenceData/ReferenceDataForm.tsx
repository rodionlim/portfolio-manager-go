import {
  Button,
  Container,
  Group,
  NumberInput,
  SimpleGrid,
  Text,
  TextInput,
  Title,
  Autocomplete,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useForm } from "@mantine/form";
import { DatePickerInput } from "@mantine/dates";
import { useLocation } from "react-router-dom";
import { getUrl } from "../../utils/url";
import { useDispatch } from "react-redux";
import { AppDispatch } from "../../store";
import { fetchReferenceData } from "../../slices/referenceDataSlice";

const assetClassOptions = ["bond", "cash", "cmdty", "crypto", "eq", "fx"];

const assetSubClassOptions = [
  "bond",
  "cash",
  "etf",
  "future",
  "govies",
  "option",
  "reit",
  "spot",
  "stock",
];

const categoryOptions = [
  "agriculture",
  "consumercyclicals",
  "consumernoncyclicals",
  "crypto",
  "energy",
  "finance",
  "funeral",
  "healthcare",
  "industrials",
  "materials",
  "realestate",
  "reits",
  "telecommunications",
  "technology",
  "utilities",
];

const currencyOptions = ["SGD", "HKD", "GBP", "USD"];
const domicileOptions = ["SG", "US", "HK", "IE"];

export default function ReferenceDataForm() {
  const location = useLocation();
  const dispatch = useDispatch<AppDispatch>();

  const form = useForm({
    mode: "uncontrolled",
    initialValues: {
      id: location.state?.id || "",
      name: location.state?.name || "",
      underlying_ticker: location.state?.underlying_ticker || "",
      yahoo_ticker: location.state?.yahoo_ticker || "",
      google_ticker: location.state?.google_ticker || "",
      dividends_sg_ticker: location.state?.dividends_sg_ticker || "",
      nasdaq_ticker: location.state?.nasdaq_ticker || "",
      barchart_ticker: location.state?.barchart_ticker || "",
      asset_class: location.state?.asset_class || "",
      asset_sub_class: location.state?.asset_sub_class || "",
      category: location.state?.category || "",
      sub_category: location.state?.sub_category || "",
      ccy: location.state?.ccy || "",
      domicile: location.state?.domicile || "",
      coupon_rate: location.state?.coupon_rate || 0,
      maturity_date: location.state?.maturity_date || null,
      strike_price: location.state?.strike_price || 0,
      call_put: location.state?.call_put || "",
    },
    validate: {
      name: (value) => value.length < 1 && "Name is required",
      underlying_ticker: (value) =>
        value.length < 1 && "Underlying Ticker is required",
      asset_class: (value) => value.length < 1 && "Asset Class is required",
      asset_sub_class: (value) =>
        value.length < 1 && "Asset Subclass is required",
      ccy: (value) => value.length < 1 && "Currency is required",
    },
  });

  function upsertReferenceData(values: typeof form.values) {
    const action = !values.id ? "add" : "update";
    const actionPastTense = !values.id ? "added" : "updated";
    const body = {
      ...values,
      ID: values.id || values.underlying_ticker,
      maturity_date: values.maturity_date
        ? values.maturity_date.toISOString().split(".")[0] + "Z"
        : null,
    };

    return fetch(getUrl("/api/v1/refdata"), {
      method: values.id ? "PUT" : "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    })
      .then((resp) => {
        if (!resp.ok) {
          return resp.json().then((error) => {
            throw new Error(error.message || "An error occurred");
          });
        }
        return resp.json();
      })
      .then((_) => {
        notifications.show({
          title: "Reference Data successfully added",
          message: `Reference Data [${
            values.id || values.underlying_ticker
          }] was successfully ${actionPastTense}`,
          autoClose: 6000,
        });
        dispatch(fetchReferenceData());
      })
      .catch((error) => {
        console.error(error);
        notifications.show({
          color: "red",
          title: "Error",
          message: `Unable to ${action} reference data\n ${error}`,
        });
        throw new Error("An error occurred while submitting reference data");
      });
  }

  const handleSubmit = (values: typeof form.values) => {
    upsertReferenceData(values); // TODO: add error handling
  };

  return (
    <Container size="sm">
      <Title order={2}>
        {form.getValues().id ? "Update" : "Add"} Reference Data
      </Title>
      <Text size="sm" mb="sm">
        If reference data id has changed, please delete record and recreate
        instead
      </Text>
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <SimpleGrid cols={2}>
          {form.getValues().id && (
            <TextInput
              withAsterisk
              label="ID"
              placeholder="ID"
              disabled={true}
              key={form.key("id")}
              {...form.getInputProps("id")}
            />
          )}
          <TextInput
            withAsterisk
            label="Name"
            placeholder="Name"
            key={form.key("name")}
            {...form.getInputProps("name")}
          />
          <TextInput
            withAsterisk
            label="Underlying Ticker (Ref ID) "
            placeholder="Underlying Ticker"
            key={form.key("underlying_ticker")}
            {...form.getInputProps("underlying_ticker")}
          />
          <TextInput
            label="Yahoo Ticker"
            placeholder="Yahoo Ticker"
            key={form.key("yahoo_ticker")}
            {...form.getInputProps("yahoo_ticker")}
          />
          <TextInput
            label="Google Ticker"
            placeholder="Google Ticker"
            key={form.key("google_ticker")}
            {...form.getInputProps("google_ticker")}
          />
          <TextInput
            label="Dividends SG Ticker"
            placeholder="Dividends SG Ticker"
            key={form.key("dividends_sg_ticker")}
            {...form.getInputProps("dividends_sg_ticker")}
          />
          <TextInput
            label="Nasdaq Ticker"
            placeholder="Nasdaq Ticker"
            key={form.key("nasdaq_ticker")}
            {...form.getInputProps("nasdaq_ticker")}
          />
          <TextInput
            label="Barchart Ticker"
            placeholder="Barchart Ticker"
            key={form.key("barchart_ticker")}
            {...form.getInputProps("barchart_ticker")}
          />
          <Autocomplete
            withAsterisk
            label="Asset Class"
            placeholder="Asset Class"
            key={form.key("asset_class")}
            data={assetClassOptions}
            {...form.getInputProps("asset_class")}
          />
          <Autocomplete
            withAsterisk
            label="Asset Subclass"
            placeholder="Asset Subclass"
            key={form.key("asset_sub_class")}
            data={assetSubClassOptions}
            {...form.getInputProps("asset_sub_class")}
          />
          <Autocomplete
            label="Category"
            placeholder="Category"
            key={form.key("category")}
            data={categoryOptions}
            {...form.getInputProps("category")}
          />
          <TextInput
            label="Subcategory"
            placeholder="Subcategory"
            key={form.key("sub_category")}
            {...form.getInputProps("sub_category")}
          />
          <Autocomplete
            withAsterisk
            label="Currency"
            placeholder="Currency"
            key={form.key("ccy")}
            data={currencyOptions}
            {...form.getInputProps("ccy")}
          />
          <Autocomplete
            label="Domicile"
            placeholder="Domicile"
            key={form.key("domicile")}
            data={domicileOptions}
            {...form.getInputProps("domicile")}
          />
          <NumberInput
            label="Coupon Rate"
            placeholder="Coupon Rate"
            key={form.key("coupon_rate")}
            {...form.getInputProps("coupon_rate")}
          />
          <DatePickerInput
            label="Maturity Date"
            placeholder="Select the maturity date"
            key={form.key("maturity_date")}
            {...form.getInputProps("maturity_date")}
          />
          <NumberInput
            label="Strike Price"
            placeholder="Strike Price"
            key={form.key("strike_price")}
            {...form.getInputProps("strike_price")}
          />
          <TextInput
            label="Call/Put"
            placeholder="Call/Put"
            key={form.key("call_put")}
            {...form.getInputProps("call_put")}
          />
          <Group justify="flex-end">
            <Button type="submit">Submit</Button>
          </Group>
        </SimpleGrid>
      </form>
    </Container>
  );
}
