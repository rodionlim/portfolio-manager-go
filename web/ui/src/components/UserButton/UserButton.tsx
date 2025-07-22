import { IconChevronRight } from "@tabler/icons-react";
import { Avatar, Group, Text, UnstyledButton } from "@mantine/core";
import { useSelector, useDispatch } from "react-redux";
import { useEffect } from "react";
import { RootState, AppDispatch } from "../../store";
import { fetchUserProfile } from "../../slices/userSlice";
import classes from "./UserButton.module.css";

export function UserButton() {
  const dispatch = useDispatch<AppDispatch>();
  const { profile, loading } = useSelector((state: RootState) => state.user);

  useEffect(() => {
    dispatch(fetchUserProfile());
  }, [dispatch]);

  // Show loading state if needed
  if (loading) {
    return (
      <UnstyledButton className={classes.user}>
        <Group>
          <Avatar radius="xl" />
          <div style={{ flex: 1 }}>
            <Text size="sm" fw={500}>
              Loading...
            </Text>
          </div>
          <IconChevronRight size={14} stroke={1.5} />
        </Group>
      </UnstyledButton>
    );
  }

  return (
    <UnstyledButton className={classes.user}>
      <Group>
        <Avatar 
          src={profile.avatar || undefined} 
          radius="xl"
        >
          {!profile.avatar && profile.username.charAt(0).toUpperCase()}
        </Avatar>

        <div style={{ flex: 1 }}>
          <Text size="sm" fw={500}>
            {profile.username}
          </Text>

          <Text c="dimmed" size="xs">
            {profile.email}
          </Text>
        </div>

        <IconChevronRight size={14} stroke={1.5} />
      </Group>
    </UnstyledButton>
  );
}
