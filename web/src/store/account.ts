import type { GetUserInfoResponse } from "@/api/account";
import { TokenInfo, removeTokenInfo, setTokenInfo } from "@/util";
import { StateCreator } from "zustand";

interface ResetInfo {
  email: string;
  verificationCode: string;
  verificationToken: string;
}
export interface AccountState {
  isLogin: boolean;
  changeLoginStatus: (isLogin: boolean) => void;
  accessToken: string;
  setAccessToken: (accessToken: string) => void;
  refreshToken: string;
  setRefreshToken: (refreshToken: string) => void;
  userInfo: Partial<GetUserInfoResponse>;
  setUserInfo: (userInfo: Partial<GetUserInfoResponse>) => void;
  keepLogin: boolean;
  setKeepLogin: (keepLogin: boolean) => void;
  resetInfo: ResetInfo;
  setResetInfo: (values: Partial<ResetInfo>) => void;

  signIn: (info: TokenInfo) => void;
  signOut: () => void;
}

export const createAccountStore: StateCreator<AccountState> = (set) => ({
  isLogin: false,
  changeLoginStatus: (isLogin) => set({ isLogin }),
  accessToken: "",
  setAccessToken: (accessToken) => set({ accessToken }),
  refreshToken: "",
  setRefreshToken: (refreshToken) => set({ refreshToken }),
  userInfo: {},
  setUserInfo: (userInfo) => set({ userInfo }),
  keepLogin: true,
  setKeepLogin: (keepLogin) => set({ keepLogin }),
  resetInfo: {
    email: '',
    verificationCode: '',
    verificationToken: '',
  },
  setResetInfo: (values) => {
    set((state) => ({
      resetInfo: {
        ...state.resetInfo,
        ...values,
      },
    }))
  },

  signOut: () => {
    set({ isLogin: false, accessToken: "", refreshToken: "", userInfo: {} });
    removeTokenInfo();
  },
  signIn: ({ accessToken, refreshToken }) => {
    set({ isLogin: true, accessToken, refreshToken });
    // Persistently store token information
    setTokenInfo({ accessToken, refreshToken });
  },
});
