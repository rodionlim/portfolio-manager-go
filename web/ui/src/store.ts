import { configureStore } from "@reduxjs/toolkit";
import referenceDataReducer from "./slices/referenceDataSlice";
import userReducer from "./slices/userSlice";

const store = configureStore({
  reducer: {
    referenceData: referenceDataReducer,
    user: userReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export default store;
