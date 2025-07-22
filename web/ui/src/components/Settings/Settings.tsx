import { useState, useEffect } from "react";
import {
  IconAlertCircle,
  IconUser,
  IconCalendarX,
  IconSettings,
  IconTrash,
  IconCurrencyDollar,
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
import { useSelector, useDispatch } from "react-redux";
import { RootState, AppDispatch } from "../../store";
import {
  fetchUserProfile,
  updateUserProfile,
  clearError,
} from "../../slices/userSlice";
import { getUrl } from "../../utils/url";

const Settings = () => {
  const dispatch = useDispatch<AppDispatch>();
  const {
    profile,
    loading: userLoading,
    error: userError,
  } = useSelector((state: RootState) => state.user);

  const [loading, setLoading] = useState(false);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [avatar, setAvatar] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [fxInferLoading, setFxInferLoading] = useState(false);

  // Load user profile data into form fields when component mounts or profile changes
  useEffect(() => {
    dispatch(fetchUserProfile());
  }, [dispatch]);

  useEffect(() => {
    setUsername(profile.username);
    setEmail(profile.email);
    setAvatar(profile.avatar);
  }, [profile]);

  // Clear error when component mounts
  useEffect(() => {
    dispatch(clearError());
  }, [dispatch]);

  const handleDeleteAllData = async () => {
    if (
      !window.confirm(
        "WARNING: This will permanently delete ALL portfolio positions and trades. This action cannot be undone. Are you sure?"
      )
    ) {
      return;
    }

    setDeleteLoading(true);
    try {
      const blotterResponse = await fetch(getUrl("api/v1/blotter/trade/all"), {
        method: "DELETE",
      });

      if (!blotterResponse.ok) {
        throw new Error("Failed to delete blotter trades");
      }

      const portfolioResponse = await fetch(
        getUrl("api/v1/portfolio/positions"),
        {
          method: "DELETE",
        }
      );

      if (!portfolioResponse.ok) {
        throw new Error("Failed to delete portfolio positions");
      }

      notifications.show({
        title: "Success",
        message: "All portfolio positions and trades have been deleted",
        color: "green",
      });
    } catch (error) {
      console.error("Error deleting data:", error);
      notifications.show({
        title: "Error",
        message: `Failed to delete all data: ${error}`,
        color: "red",
      });
    } finally {
      setDeleteLoading(false);
    }
  };

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

  const handleInferFxRates = async () => {
    setFxInferLoading(true);
    try {
      // This endpoint both infers FX rates and returns the CSV data
      const exportUrl = getUrl("api/v1/blotter/export-with-fx");
      const link = document.createElement("a");
      link.href = exportUrl;
      link.setAttribute("download", "trades-with-fx.csv");
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      notifications.show({
        title: "Success",
        message: "Successfully inferred FX rates and downloaded trades",
        color: "green",
      });
    } catch (error) {
      console.error("Error inferring FX rates:", error);
      notifications.show({
        title: "Error",
        message: `Failed to infer FX rates: ${error}`,
        color: "red",
      });
    } finally {
      setFxInferLoading(false);
    }
  };

  const handleSaveProfile = async () => {
    if (!username.trim() || !email.trim()) {
      notifications.show({
        title: "Validation Error",
        message: "Username and email are required",
        color: "red",
      });
      return;
    }

    try {
      await dispatch(
        updateUserProfile({
          username: username.trim(),
          email: email.trim(),
          avatar: avatar.trim(),
        })
      ).unwrap();

      notifications.show({
        title: "Profile Updated",
        message: "Your profile has been successfully updated",
        color: "green",
      });
    } catch (error) {
      notifications.show({
        title: "Error",
        message: `Failed to update profile: ${error}`,
        color: "red",
      });
    }
  };

  const handleUploadProfilePicture = () => {
    console.log("Upload profile picture functionality to be implemented");
    notifications.show({
      title: "Feature Not Implemented",
      message: "Profile picture upload functionality is not yet implemented",
      color: "yellow",
    });
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
          <Tabs.Tab
            value="blotter"
            leftSection={<IconCurrencyDollar size={16} />}
          >
            Blotter
          </Tabs.Tab>
          <Tabs.Tab value="general" leftSection={<IconSettings size={16} />}>
            General
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="profile">
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Box pos="relative">
              <LoadingOverlay
                visible={userLoading || loading}
                overlayProps={{ blur: 2 }}
              />
              <Stack>
                <Title order={4}>Profile Information</Title>
                <Divider />

                {userError && (
                  <Alert
                    icon={<IconAlertCircle size={16} />}
                    title="Error"
                    color="red"
                    mb="md"
                  >
                    {userError}
                  </Alert>
                )}

                <Group align="flex-start">
                  <Avatar size="xl" radius="xl" src={avatar || undefined}>
                    {!avatar && username.charAt(0).toUpperCase()}
                  </Avatar>
                  <Button
                    variant="outline"
                    onClick={handleUploadProfilePicture}
                  >
                    Upload Picture
                  </Button>
                </Group>

                <TextInput
                  label="Avatar URL"
                  placeholder="https://example.com/avatar.png"
                  value={avatar}
                  onChange={(e) => setAvatar(e.target.value)}
                />

                <TextInput
                  label="Username"
                  placeholder="Enter your username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                />

                <TextInput
                  label="Email"
                  placeholder="Enter your email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  type="email"
                />

                <Button mt="md" onClick={handleSaveProfile} loading={loading}>
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
                <Divider my="lg" />

                <Alert
                  icon={<IconAlertCircle size={16} />}
                  title="Delete All Data"
                  color="red"
                  mb="md"
                >
                  This will permanently delete ALL portfolio positions and
                  trades from the system. This action cannot be undone.
                </Alert>

                <Button
                  color="red"
                  onClick={handleDeleteAllData}
                  loading={deleteLoading}
                  leftSection={<IconTrash size={16} />}
                >
                  Delete All Data
                </Button>
              </Stack>
            </Box>
          </Card>
        </Tabs.Panel>

        <Tabs.Panel value="blotter">
          <Card shadow="sm" padding="lg" radius="md" withBorder>
            <Box pos="relative">
              <LoadingOverlay
                visible={fxInferLoading}
                overlayProps={{ blur: 2 }}
              />
              <Stack>
                <Title order={4}>Blotter Management</Title>
                <Divider />

                <Text c="dimmed" mb="md">
                  Manage your blotter settings and perform maintenance
                  operations.
                </Text>

                <Alert
                  icon={<IconAlertCircle size={16} />}
                  title="Infer FX Rates"
                  color="blue"
                  mb="md"
                >
                  This will infer FX rates for historical trades and download
                  blotter trades to a csv file.
                </Alert>
                <Button
                  color="blue"
                  onClick={handleInferFxRates}
                  loading={fxInferLoading}
                  leftSection={<IconCurrencyDollar size={16} />}
                >
                  Infer FX Rates
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
