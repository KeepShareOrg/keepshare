import * as XLSX from "xlsx";
import { match } from "ts-pattern";
import dayJs from "dayjs";
import sha256 from "crypto-js/sha256";
import hex from "crypto-js/enc-hex";
import { LinkFormatType, LocalStorageKeys, SupportLocales } from "@/constant";
import { SharedLinkInfo } from "@/api/link";

// validate email format
export const isValidateEmail = (email: string) => {
  return /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/.test(
    email,
  );
};

// calc string sha256 hash, return 64 character
export const calcPasswordHash = (password: string): string => {
  return sha256(password).toString(hex);
};

export interface TokenInfo {
  accessToken?: string;
  refreshToken?: string;
  updatedTime?: number;
}
// get token info from local storage
export const getTokenInfo = (): TokenInfo => {
  try {
    const isRemember = getRememberToken();
    let storage = window.localStorage;
    if (!isRemember) {
      storage = window.sessionStorage;
      window.localStorage.removeItem(LocalStorageKeys.TOKEN_INFO);
    }
    const info = storage.getItem(
      LocalStorageKeys.TOKEN_INFO,
    ) as string;
    return JSON.parse(info) || {};
  } catch {
    return {};
  }
};

// set token info into local storage
export const setTokenInfo = (info: TokenInfo) => {
  try {
    info.updatedTime = Date.now();
    const isRemember = getRememberToken();
    const storage = isRemember ? localStorage : sessionStorage;
    storage.setItem(LocalStorageKeys.TOKEN_INFO, JSON.stringify(info));
  } catch {
    console.error("[keep share: ] set token info error!");
  }
};

// remove token info from local storage
export const removeTokenInfo = () => {
  const isRemember = getRememberToken();
  const storage = isRemember ? localStorage : sessionStorage;
  storage.removeItem(LocalStorageKeys.TOKEN_INFO);
};

export const setRememberToken = (remember: string) => {
  try {
    window.localStorage.setItem(LocalStorageKeys.REMEMBER_TOKEN, remember);
  } catch (err) {
    console.error('[keep share: ] set remember info error!');
  }
}

export const getRememberToken = (): boolean => {
  return /^true$/i.test(window.localStorage.getItem(LocalStorageKeys.REMEMBER_TOKEN) || '');
}

// transfer bytes unit to human friendly unit
export const formatBytes = (bytes: number, decimals: number = 2): string => {
  if (bytes === 0) return "0 Bytes";

  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
};

// shim sharedLinks table data
export const shimSharedLinksTableData = (
  data: SharedLinkInfo,
): SharedLinkInfo => {
  return {
    ...data,
    key: String(data.auto_id),
    created_at: dayJs(data.created_at).format("YYYY-MM-DD hh:mm"),
    updated_at: dayJs(data.updated_at).format("YYYY-MM-DD hh:mm"),
    size: typeof data.size === "number" ? formatBytes(data.size) : data.size,
  };
};

// How many bytes are equal to per GB
export const G_BYTES = 1 * 1024 * 1024 * 1024;

// format link with support type
export const formatLinkWithType = (url: string, formatType: LinkFormatType) => {
  return match(formatType)
    .with(LinkFormatType.TEXT, () => url)
    .with(LinkFormatType.HTML, () => `<a href="${url}">${url}</a>`)
    .with(LinkFormatType.BB_CODE, () => `[${url}]`)
    .exhaustive();
};

// parse magnets from string content
export const parseLinks = (content: string) => {
  return content.split(/[\n\r\s]/).filter((v) => {
    const content = v.trim();
    return content.split("\n");
  });
};

// copy text util function
export const copyToClipboard = (text: string) => {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  document.body.removeChild(textarea);
};

// export data to excel
export const exportExcel = async (data: unknown[][], filename: string) => {
  console.log("i wanna export excel data!");
  // create a excel workbook
  const wb = XLSX.utils.book_new();
  const ws = XLSX.utils.aoa_to_sheet(data);
  // add data to the workbook
  XLSX.utils.book_append_sheet(wb, ws, "Sheet1");
  // export excel
  XLSX.writeFile(wb, filename);
};

// parse user browser language
export const getSupportLanguage = (): SupportLocales => {
  const userLanguage = navigator.language;

  if (userLanguage.includes('zh')) {
    if (userLanguage.toLowerCase() === 'zh-tw' || userLanguage.toLowerCase() === 'zh-hk') {
      return SupportLocales.ZH_TW;
    } else if (userLanguage.toLowerCase() === 'zh-cn' || userLanguage.toLowerCase() === 'zh-sg') {
      return SupportLocales.ZH_CN;
    }
  }
  return SupportLocales.En;
}