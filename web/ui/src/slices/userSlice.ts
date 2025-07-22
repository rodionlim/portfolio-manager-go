import { createSlice, createAsyncThunk, PayloadAction } from "@reduxjs/toolkit";
import { getUrl } from "../utils/url";

// User profile interface
export interface UserProfile {
  username: string;
  email: string;
  avatar: string;
}

// Initial state interface
interface UserState {
  profile: UserProfile;
  loading: boolean;
  error: string | null;
}

// Initial state with default values
const initialState: UserState = {
  profile: {
    username: "User",
    email: "user@example.com",
    avatar: "",
  },
  loading: false,
  error: null,
};

// Async thunk to fetch user profile
export const fetchUserProfile = createAsyncThunk(
  "user/fetchProfile",
  async (_, { rejectWithValue }) => {
    try {
      const response = await fetch(getUrl("api/v1/user/profile"));
      if (!response.ok) {
        throw new Error("Failed to fetch user profile");
      }
      const data = await response.json();
      return data as UserProfile;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : "Unknown error");
    }
  }
);

// Async thunk to update user profile
export const updateUserProfile = createAsyncThunk(
  "user/updateProfile",
  async (profileData: UserProfile, { rejectWithValue }) => {
    try {
      const response = await fetch(getUrl("api/v1/user/profile"), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(profileData),
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to update user profile");
      }
      
      const data = await response.json();
      return data as UserProfile;
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : "Unknown error");
    }
  }
);

// User slice
const userSlice = createSlice({
  name: "user",
  initialState,
  reducers: {
    // Clear error
    clearError: (state) => {
      state.error = null;
    },
    // Reset user state (for testing or logout scenarios)
    resetUserState: (state) => {
      state.profile = initialState.profile;
      state.loading = false;
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch user profile
      .addCase(fetchUserProfile.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchUserProfile.fulfilled, (state, action: PayloadAction<UserProfile>) => {
        state.loading = false;
        state.profile = action.payload;
        state.error = null;
      })
      .addCase(fetchUserProfile.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      // Update user profile
      .addCase(updateUserProfile.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateUserProfile.fulfilled, (state, action: PayloadAction<UserProfile>) => {
        state.loading = false;
        state.profile = action.payload;
        state.error = null;
      })
      .addCase(updateUserProfile.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { clearError, resetUserState } = userSlice.actions;
export default userSlice.reducer;