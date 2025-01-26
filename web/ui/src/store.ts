import { configureStore } from "@reduxjs/toolkit";
import referenceDataReducer from "./slices/referenceDataSlice";

const store = configureStore({
  reducer: {
    referenceData: referenceDataReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export default store;
