import { useState } from "react";
import {
  Modal,
  Text,
  Select,
  TextInput,
  Button,
  Group,
  Stack,
  Alert,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconInfoCircle } from "@tabler/icons-react";
import { Trade } from "../../types/blotter";
import { getUrl } from "../../utils/url";

interface BlotterBulkUpdateModalProps {
  opened: boolean;
  onClose: () => void;
  selectedTrades: Trade[];
  onSuccess: () => void;
}

type BulkUpdateField = "ticker" | "book";

const BlotterBulkUpdateModal: React.FC<BlotterBulkUpdateModalProps> = ({
  opened,
  onClose,
  selectedTrades,
  onSuccess,
}) => {
  const [selectedField, setSelectedField] = useState<BulkUpdateField | null>(
    null
  );
  const [newValue, setNewValue] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  const handleBulkUpdate = async () => {
    if (!selectedField || !newValue.trim()) {
      notifications.show({
        color: "red",
        title: "Error",
        message: "Please select a field and provide a value",
      });
      return;
    }

    setIsUpdating(true);

    try {
      const updatePromises = selectedTrades.map(async (trade) => {
        const body = {
          id: trade.TradeID,
          tradeDate: trade.TradeDate,
          ticker:
            selectedField === "ticker" ? newValue.toUpperCase() : trade.Ticker,
          book: selectedField === "book" ? newValue : trade.Book,
          broker: trade.Broker,
          account: trade.Account,
          status: "open", // Default status for bulk updates
          originalTradeId: "", // Not available in Trade interface
          quantity: trade.Quantity,
          price: trade.Price,
          fx: trade.Fx,
          side: trade.Side,
          seqNum: trade.SeqNum,
        };

        const resp = await fetch(getUrl("api/v1/blotter/trade"), {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(body),
        });

        if (!resp.ok) {
          const error = await resp.json();
          throw new Error(
            error.message || `Failed to update trade ${trade.TradeID}`
          );
        }

        return await resp.json();
      });

      await Promise.all(updatePromises);

      notifications.show({
        title: "Bulk Update Successful",
        message: `Successfully updated ${selectedTrades.length} trades`,
        autoClose: 6000,
      });

      onSuccess();
      onClose();
      resetForm();
    } catch (error) {
      console.error(error);
      notifications.show({
        color: "red",
        title: "Bulk Update Failed",
        message: `Failed to update trades: ${error}`,
      });
    } finally {
      setIsUpdating(false);
    }
  };

  const resetForm = () => {
    setSelectedField(null);
    setNewValue("");
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  return (
    <Modal
      opened={opened}
      onClose={handleClose}
      title="Bulk Update Trades"
      size="md"
    >
      <Stack gap="md">
        <Alert
          icon={<IconInfoCircle size="1rem" />}
          color="blue"
          title="Important Information"
        >
          <Text size="sm">
            Bulk updates are only supported for specific fields due to their
            complexity. For large-scale updates, it is recommended to export
            your trades, wipe all data in settings, make amendments in the CSV,
            and re-upload.
          </Text>
        </Alert>

        <Text size="sm" c="dimmed">
          You are about to update {selectedTrades.length} selected trades.
        </Text>

        <Select
          label="Field to Update"
          placeholder="Select field to update"
          data={[
            { value: "ticker", label: "Ticker" },
            { value: "book", label: "Book" },
          ]}
          value={selectedField}
          onChange={(value) => setSelectedField(value as BulkUpdateField)}
          withAsterisk
        />

        <TextInput
          label="New Value"
          placeholder="Enter the new value"
          value={newValue}
          onChange={(event) => setNewValue(event.currentTarget.value)}
          withAsterisk
        />

        <Group justify="flex-end" mt="md">
          <Button variant="outline" onClick={handleClose} disabled={isUpdating}>
            Cancel
          </Button>
          <Button
            onClick={handleBulkUpdate}
            loading={isUpdating}
            disabled={!selectedField || !newValue.trim()}
          >
            Update {selectedTrades.length} Trades
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};

export default BlotterBulkUpdateModal;
