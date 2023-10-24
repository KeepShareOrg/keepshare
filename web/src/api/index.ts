import useStore from "@/store";
import { getTokenInfo, removeTokenInfo } from "@/util";
import axios from "axios";
import type { AxiosRequestConfig, AxiosInstance } from "axios";
import { refreshToken } from "./account";
import { RoutePaths } from "@/router";

// keepShare base api address
export const BASE_API = import.meta.env.DEV ? "http://localhost:8080/" : window.location.origin;

export const fetcher = axios.create({
  baseURL: BASE_API,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request intercept and automatically populate the Authorization http header field
fetcher.interceptors.request.use(
  (config) => {
    try {
      const accessToken = useStore.getState().accessToken;
      if (accessToken) {
        config.headers.Authorization = `Bearer ${accessToken}`;
      }
    } catch (err) {
      console.error("request interceptor error: ", err);
    }
    return config;
  },
  (error) => Promise.reject(error),
);

let isRefreshing = false;
let refreshSubscribers: ((token: string) => void)[] = [];
// Response interception, with some unified logic processing when the results return
fetcher.interceptors.response.use(
  (response) => {
    try {
      // Check that the response content contains the access_token field
      if (response.data.access_token) {
        useStore.getState().signIn({
          accessToken: response.data.access_token,
          refreshToken: response.data?.refresh_token,
        });
      }
    } catch (err) {
      console.warn("response interceptor error: ", err);
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;
    // If status code 401 is returned, it automatically redirect to the login page
    if (
      error.response &&
      error.response.status === 401 &&
      !originalRequest._retry
    ) {
      if (isRefreshing) {
        // If the token is already being refreshed, wait for the refresh to complete and try again
        return new Promise<string>((resolve) => {
          refreshSubscribers.push(resolve);
        }).then((newToken) => {
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return fetcher(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const tokenInfo = getTokenInfo();
        if (!tokenInfo.refreshToken) {
          throw new Error("refresh token not found!");
        }
        // refresh token
        const { data } = await refreshToken(tokenInfo.refreshToken);

        const newToken = data?.access_token || "";
        if (!newToken) {
          throw new Error("refresh token failed!");
        }
        // update token, and notify waiting request
        fetcher.defaults.headers.common["Authorization"] = `Bearer ${newToken}`;
        isRefreshing = false;
        refreshSubscribers.forEach((subscriber) => subscriber(newToken));
        refreshSubscribers = [];

        // Retry the original request
        return fetcher(originalRequest);
      } catch (refreshError) {
        console.warn(refreshError);
        // Failed to refresh the token, clear the token and jump to the login page
        removeTokenInfo();
        location.href = RoutePaths.Login;
      }
    }

    return Promise.reject(error);
  },
);

// transfer response data to { data: ..., error: ... }
export type AxiosWrapperResponse<T> = Promise<{
  data: T | null;
  error:
  | ({ error?: string; message?: string } & Record<string, unknown>)
  | null;
}>;
// custom axios wrapper by difference fetcher
export const getAxiosWrapper = (client: AxiosInstance) => {
  return <T>(options: AxiosRequestConfig): AxiosWrapperResponse<T> =>
    new Promise((resolve) => {
      client<T>(options)
        .then((response) => resolve({ data: response.data, error: null }))
        // eslint-disable-next-line
        .catch((error: any) =>
          resolve({ data: null, error: error.response?.data }),
        );
    });
}

export const axiosWrapper = ((client: AxiosInstance) => getAxiosWrapper(client))(fetcher);
