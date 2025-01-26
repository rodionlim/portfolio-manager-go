import { createSlice, createAsyncThunk } from "@reduxjs/toolkit";
import { ReferenceData } from "../types";

interface ReferenceDataState {
  data: ReferenceData | null;
  status: "idle" | "loading" | "succeeded" | "failed";
  error: string | null;
}

const initialState: ReferenceDataState = {
  data: null,
  status: "idle",
  error: null,
};

export const fetchReferenceData = createAsyncThunk(
  "referenceData/fetchReferenceData",
  async () => {
    const response = await fetch("http://localhost:8080/api/v1/refdata");
    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    return response.json();
  }
);

const referenceDataSlice = createSlice({
  name: "referenceData",
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchReferenceData.pending, (state) => {
        state.status = "loading";
      })
      .addCase(fetchReferenceData.fulfilled, (state, action) => {
        state.status = "succeeded";
        state.data = action.payload;
      })
      .addCase(fetchReferenceData.rejected, (state, action) => {
        state.status = "failed";
        state.error = action.error.message || "Failed to fetch reference data";
      });
  },
});

export default referenceDataSlice.reducer;
