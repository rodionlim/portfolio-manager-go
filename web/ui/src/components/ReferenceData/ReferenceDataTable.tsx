// filepath: /Users/rodionlim/workspace/portfolio-manager-go/web/ui/src/components/BlotterTable.tsx
import React, { useMemo } from "react";
import { Box, Button } from "@mantine/core";
import {
  MantineReactTable,
  MRT_ColumnDef,
  MRT_TableInstance,
  useMantineReactTable,
} from "mantine-react-table";
import { useQuery } from "@tanstack/react-query";
import { notifications } from "@mantine/notifications";
import { useNavigate } from "react-router-dom";

interface RefData {
  id: string;
  name: string;
  underlying_ticker: string;
  yahoo_ticker: string;
  google_ticker: string;
  dividends_sg_ticker: string;
  asset_class: string;
  asset_sub_class: string;
  category: string;
  sub_category: string;
  ccy: string;
  domicile: string;
  coupon_rate: number;
  maturity_date: Date;
  strike_price: number;
  call_put: string;
}

const fetchData = async (): Promise<RefData[]> => {
  return fetch("http://localhost:8080/api/v1/refdata")
    .then((resp) => resp.json())
    .then(
      (data) => {
        return Object.values(data);
      },
      (error) => {
        console.error("error", error);
        throw new Error(
          `An error occurred while fetching reference data ${error.message}`
        );
      }
    );
};

const deleteRefData = async (
  refData: string[]
): Promise<{ message: string }> => {
  return fetch("http://localhost:8080/api/v1/refdata", {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(refData),
  })
    .then((resp) => resp.json())
    .then(
      (data) => {
        return data;
      },
      (error) => {
        console.error("error", error);
        throw new Error("An error occurred while deleting reference data");
      }
    );
};

const ReferenceDataTable: React.FC = () => {
  const navigate = useNavigate();

  const {
    data: refData = [],
    isLoading,
    error,
    refetch,
  } = useQuery({ queryKey: ["refData"], queryFn: fetchData });

  const columns = useMemo<MRT_ColumnDef<RefData>[]>(
    () => [
      { accessorKey: "id", header: "Ticker" },
      { accessorKey: "name", header: "Name" },
      { accessorKey: "yahoo_ticker", header: "Yahoo Ticker" },
      { accessorKey: "google_ticker", header: "Google Ticker" },
      { accessorKey: "dividends_sg_ticker", header: "Dividends SG Ticker" },
      { accessorKey: "asset_class", header: "Asset Class" },
      { accessorKey: "asset_sub_class", header: "Asset Subclass" },
      { accessorKey: "category", header: "Category" },
      { accessorKey: "sub_category", header: "Subcategory" },
      { accessorKey: "ccy", header: "Currency" },
      { accessorKey: "domicile", header: "Domicile" },
      //   { accessorKey: "CouponRate", header: "Coupon Rate" },
      { accessorKey: "maturity_date", header: "Maturity" },
      //   { accessorKey: "StrikePrice", header: "Strike Price" },
      //   { accessorKey: "CallPut", header: "Call/Put" },
    ],
    []
  );

  const table = useMantineReactTable({
    columns,
    data: refData,
    initialState: { showGlobalFilter: true, showColumnFilters: true },
    state: { density: "xs" },
    enableRowSelection: true,
    positionToolbarAlertBanner: "bottom",
    renderTopToolbarCustomActions: ({ table }) => (
      <Box style={{ display: "flex", gap: "16px", padding: "4px" }}>
        <Button
          color="teal"
          onClick={handleAddRefData()}
          disabled={table.getIsSomeRowsSelected()}
          variant="filled"
        >
          Add Reference Data
        </Button>
        <Button
          color="red"
          disabled={!table.getIsSomeRowsSelected()}
          onClick={handleDeleteRefData(table)}
          variant="filled"
        >
          Delete Selected Reference Data
        </Button>
        <Button
          color="blue"
          disabled={!(table.getSelectedRowModel().rows.length === 1)}
          onClick={handleUpdateRefData(table)}
          variant="filled"
        >
          Update Reference Data
        </Button>
      </Box>
    ),
  });

  // handle add reference data allows routing to the add reference data page
  const handleAddRefData = (): (() => void) => {
    return () => {
      navigate("/refdata/add_ref_data");
    };
  };

  const handleDeleteRefData = (
    table: MRT_TableInstance<RefData>
  ): (() => void) => {
    return () => {
      const refDataRows = table
        .getSelectedRowModel()
        .rows.map((data) => data.original.id);

      deleteRefData(refDataRows)
        .then(
          (resp: { message: string }) => {
            notifications.show({
              title: "Reference data successfully deleted",
              message: `${resp.message}`,
              autoClose: 10000,
            });
          },
          (error) => {
            notifications.show({
              color: "red",
              title: "Error",
              message: `Unable to delete reference data from the store\n ${error}`,
            });
          }
        )
        .finally(() => {
          refetch();
        });
    };
  };

  // handle update ref data allows routing to the update ref data page
  const handleUpdateRefData = (
    table: MRT_TableInstance<RefData>
  ): (() => void) => {
    return () => {
      // first check if there is any selections
      const selection = table
        .getSelectedRowModel()
        .rows.map((data) => data.original)[0];
      navigate("/refdata/update_ref_data", {
        state: {
          id: selection.id,
          name: selection.name,
          underlying_ticker: selection.underlying_ticker,
          yahoo_ticker: selection.yahoo_ticker,
          google_ticker: selection.google_ticker,
          dividends_sg_ticker: selection.dividends_sg_ticker,
          asset_class: selection.asset_class,
          asset_sub_class: selection.asset_sub_class,
          category: selection.category,
          sub_category: selection.sub_category,
          ccy: selection.ccy,
          domicile: selection.domicile,
          coupon_rate: selection.coupon_rate,
          maturity_date: selection.maturity_date,
          strike_price: selection.strike_price,
          call_put: selection.call_put,
        },
      });
    };
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading reference data</div>;

  return (
    <div>
      <MantineReactTable table={table} />
    </div>
  );
};

export default ReferenceDataTable;
