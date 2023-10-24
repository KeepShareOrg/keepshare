import useSWR from "swr";
import { axiosWrapper, fetcher, getAxiosWrapper } from ".";
import type { AxiosResponse } from "axios";
import axios from "axios";

export type SharedLinkStatus =
  | "PENDING"
  | "UNKNOWN"
  | "CREATED"
  | "OK"
  | "DELETED"
  | "NOT_FOUND"
  | "SENSITIVE"
  | "BLOCKED";
export interface SharedLinkInfo {
  key?: string;
  auto_id: number;
  user_id: string;
  state: SharedLinkStatus;
  host: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  size: string | number;
  visitor: number;
  stored: number;
  revenue: number;
  title: string;
  days_not_visit: number;
  resource_link_hash: string;
  shared_link_hash: string;
  original_link: string;
  host_shared_link: string;
  last_visited_at: string;
}
export interface QuerySharedLinksParams {
  // KeepShare search dsl string: eg. size=123456;status="ok";
  search?: string;
  limit?: number;
  filter?: string;
  pageToken?: string;
  pageIndex?: number;
}
export interface QuerySharedLinksResponse {
  list: SharedLinkInfo[];
  next_page_token: string;
  page_size: number;
  total: number;
}
// query shared links
export const querySharedLinks = ({
  search,
  limit,
  filter,
  pageIndex,
  pageToken,
}: QuerySharedLinksParams = {}) => {
  const ps = new URLSearchParams();
  search && ps.set("search", search);
  filter && ps.set("filter", filter);
  limit && ps.set("limit", String(limit));
  pageIndex && ps.set("page_index", String(pageIndex));
  pageToken && ps.set("page_token", pageToken);

  // eslint-disable-next-line
  return useSWR<AxiosResponse<QuerySharedLinksResponse, any>>(
    `/api/shared_links?${ps.toString()}`,
    fetcher,
    {
      revalidateOnFocus: true,
      revalidateOnMount: true,
    },
  );
};

interface CreateShareLinkResponse {
  message: string;
  links: SharedLinkInfo[];
}
// create share link by resource link(current just support magnet)
export const createShareLink = (links: string[]) => {
  return axiosWrapper<CreateShareLinkResponse>({
    method: "POST",
    url: "/api/shared_links",
    data: { links },
  });
};

// get shared link info by auto id
export const getSharedLinkInfo = (id: string) => {
  return axiosWrapper<SharedLinkInfo>({
    method: "GET",
    url: `/api/shared_link?id=${id}`,
  });
};

// link to share submission result
interface SubmissionResultResponse {
  list: SharedLinkInfo[];
}
export const getLinkToShareSubmissionResult = (links: string[]) => {
  const url = `/api/query_shared_links`;
  return axiosWrapper<SubmissionResultResponse>({
    url,
    method: "POST",
    data: {
      links,
    },
  });
};

// delete shared link
export interface DeleteSharedLinkParams {
  host: string;
  links: string[];
}
export const deleteSharedLink = ({ host, links }: DeleteSharedLinkParams) => {
  const url = `/api/shared_links`;
  return axiosWrapper({
    url,
    method: "DELETE",
    data: { host, links },
  });
};

/* shared links management */
// list black list shared links
interface GetBlackListParams {
  pageIndex: number;
  limit: number;
}
export const getBlacklist = ({ pageIndex, limit }: GetBlackListParams) => {
  const ps = new URLSearchParams();
  pageIndex && ps.set("page_index", `${pageIndex}`);
  ps.set("limit", `${limit || 10}`);

  // eslint-disable-next-line
  return useSWR<AxiosResponse<QuerySharedLinksResponse>>(
    `/api/blacklist?${ps.toString()}`,
    fetcher,
    {
      revalidateOnFocus: true,
      revalidateOnMount: true,
    },
  );
};

// add to shared links to blacklist
export const addToBlacklist = (links: string[]) => {
  return axiosWrapper({
    url: "/api/blacklist",
    method: "POST",
    data: { links },
  });
};

// remove shared links from blacklist
export const removeFromBlacklist = (links: string[], isDeleteFile = false) => {
  return axiosWrapper({
    url: "/api/blacklist",
    method: "DELETE",
    data: { links, delete: isDeleteFile },
  });
};

/* 
  get resource link from "WhatsLinks"
  https://whatslink.info/
*/
export interface GetLinkInfoFromWhatsLinkResponse {
  type: string; // The content type for the link
  file_type: string; // The type of the content corresponding to the link, Possible values: unknown, folder, video, text, image, audio, archive, font, document
  name: string; // The name of the content corresponding to the link
  size: number; // The total size of the content corresponding to the link
  count: number; // The number of included files corresponding to the link
  screenshots: {
    time: number; // Position of the screenshot within the content
    screenshot: string; // The URL of the screenshot image
  }[]; // List of content screenshots corresponding to the link
}
export const getLinkInfoFromWhatsLink = (link: string) => {
  const client = axios.create({ baseURL: `https://whatslink.info/api/v1` });
  return getAxiosWrapper(client)<GetLinkInfoFromWhatsLinkResponse>({
    url: `/link?url=${window.encodeURIComponent(link)}`,
    method: 'GET',
  });
}