import { axiosWrapper } from ".";

/* PikPak account management */

// get pikpak host info
export interface PikPakHostInfo {
  master: {
    user_id: string;
    keepshare_user_id: string;
    email: string;
    password: string;
    created_at: string;
    updated_at: string;
  };
  revenue: number;
  workers: {
    premium: {
      count: number;
      used: number;
      limit: number;
    };
    free: {
      count: number;
      used: number;
      limit: number;
    };
  };
}
export const getPikPakHostInfo = () => {
  return axiosWrapper<PikPakHostInfo>({
    url: "/api/host/info",
    method: "GET",
  });
};

export interface GetPikPakAccountStatisticsParams {
  host: string;
  stored_count_lt: number[];
  not_stored_days_gt: number[];
}
export type GetPikPakAccountStatisticsResponse = Record<
  "stored_count_lt" | "not_stored_days_gt",
  Array<{
    number: number;
    total_count: number;
    total_size: number;
  }>
>;
// get account storage statistics
export const getPikPakAccountStatistics = (
  params: GetPikPakAccountStatisticsParams,
) => {
  return axiosWrapper<GetPikPakAccountStatisticsResponse>({
    url: "/api/storage/statistics",
    method: "POST",
    data: params,
  });
};

export interface ClearPikPakAccountStorageParams {
  host: string;
  stored_count_lt?: number;
  not_stored_days_gt?: number;
  only_for_premium?: boolean;
}
// clear account usage storage
export const clearPikPakAccountStorage = (
  params: ClearPikPakAccountStorageParams,
) => {
  return axiosWrapper({
    url: "/api/storage/release",
    method: "POST",
    data: params,
  });
};
