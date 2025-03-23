import { useState } from "react";
import {
  IconAlertCircle,
  IconUser,
  IconCalendarX,
  IconSettings,
} from "@tabler/icons-react";
import { notifications } from "@mantine/notifications";
import {
  Box,
  Button,
  Card,
  Container,
  Divider,
  Group,
  Stack,
  Text,
  Title,
  Avatar,
  TextInput,
  Tabs,
  Alert,
  LoadingOverlay,
} from "@mantine/core";
import { getUrl } from "../../utils/url";

const Settings = () => {
  const [loading, setLoading] = useState(false);
  const [username, setUsername] = useState("User");

  const handleCloseExpiries = async () => {
    setLoading(true);
    try {
      const response = await fetch(getUrl("api/v1/portfolio/cleanup"), {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error("Failed to close expired positions");
      }

      const closedTrades = await response.json();

      notifications.show({
        title: "Success",
        message: `Closed ${closedTrades?.length || 0} expired position(s)`,
        color: "green",
      });
    } catch (error) {
      console.error("Error closing expired positions:", error);
      notifications.show({
        title: "Error",
        message: "Failed to close expired positions",
        color: "red",
      });
    } finally {
      setLoading(false);
    }
  };

  // Placeholder for future implementation
  const handleSaveProfile = () => {
    notifications.show({
      title: "Profile Updated",
      message: "Your profile has been successfully updated",
      color: "blue",
    });
  };

  // Placeholder for future implementation
  const handleUploadProfilePicture = () => {
    // This would be implemented when the feature is ready
    console.log("Upload profile picture functionality to be implemented");
  };

  return (
    <Container size="md" py="xl">
      <Title order={2} mb="lg">
        Settings
      </Title>

      <Tabs defaultValue="profile">
        <Tabs.List mb="md">
          <Tabs.Tab value="profile" leftSection={<IconUser size={16} />}>
            Profile
          </Tabs.Tab>
          <Tabs.Tab value="portfolio" leftSection={<IconCalendarX size={16} />}>
            Portfolio
          </Tabs.Tab>
          <Tabs.Tab value="general" leftSection={<IconSettings size={16} />}>
            General
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="profile">
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Box pos="relative">
              <LoadingOverlay visible={loading} overlayProps={{ blur: 2 }} />
              <Stack>
                <Title order={4}>Profile Information</Title>
                <Divider />

                {/* Add warning alert here */}
                <Alert
                  icon={<IconAlertCircle size={16} />}
                  title="Feature in Development"
                  color="yellow"
                  mb="md"
                >
                  Profile picture upload and username changes are not currently
                  functional. These features will be implemented in a future
                  update.
                </Alert>

                <Group align="flex-start">
                  <Avatar size="xl" radius="xl" color="blue">
                    {username.charAt(0).toUpperCase()}
                  </Avatar>
                  <Button
                    variant="outline"
                    onClick={handleUploadProfilePicture}
                  >
                    Upload Picture
                  </Button>
                </Group>

                <TextInput
                  label="Username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />

                <Button mt="md" onClick={handleSaveProfile}>
                  Save Profile
                </Button>
              </Stack>
            </Box>
          </Card>
        </Tabs.Panel>

        <Tabs.Panel value="portfolio">
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Box pos="relative">
              <LoadingOverlay visible={loading} overlayProps={{ blur: 2 }} />
              <Stack>
                <Title order={4}>Portfolio Management</Title>
                <Divider />

                <Text c="dimmed" mb="md">
                  Manage your portfolio settings and perform maintenance
                  operations.
                </Text>

                <Alert
                  icon={<IconAlertCircle size={16} />}
                  title="Auto-close Expired Positions"
                  color="blue"
                  mb="md"
                >
                  This will automatically close all positions that have expired
                  without a corresponding closure trade.
                </Alert>

                <Button
                  color="blue"
                  onClick={handleCloseExpiries}
                  loading={loading}
                  leftSection={<IconCalendarX size={16} />}
                >
                  Close Expired Positions
                </Button>
              </Stack>
            </Box>
          </Card>
        </Tabs.Panel>

        <Tabs.Panel value="general">
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Stack>
              <Title order={4}>General Settings</Title>
              <Divider />
              <Text color="dimmed">
                Additional settings will appear here in future updates.
              </Text>
            </Stack>
          </Card>
        </Tabs.Panel>
      </Tabs>
    </Container>
  );
};

export default Settings;
