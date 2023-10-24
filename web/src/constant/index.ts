export const enum SupportLocales {
  En = "en",
  ZH_CN = "zh-CN",
  ZH_TW = "zh-TW",
}

// localStorage store keys
export const enum LocalStorageKeys {
  TOKEN_INFO = "keep_share_credentials",
  REMEMBER_TOKEN = 'keep_share_remember_token',
}

// shared links table column keys, consistent with the database table name
export const enum SharedLinkTableKey {
  ORIGINAL_LINKS = "Original Links",
  KEEP_SHARING_LINK = "Keep Sharing Link",
  TITLE = "Title",
  HOST_SHARED_LINK = "Host Shared Link",
  CREATED_AT = "Created at",
  VISITOR = "Visitor",
  STORED = "Stored",
  SIZE = "Size",
  DAYS_NOT_VISIT = "Days not visit",
  STATE = "State",
  ACTION = "Action",

  CREATED_BY = "Created by",
}

// support table columns list
export const supportTableColumns: SharedLinkTableKey[] = [
  SharedLinkTableKey.ORIGINAL_LINKS,
  SharedLinkTableKey.KEEP_SHARING_LINK,
  SharedLinkTableKey.TITLE,
  SharedLinkTableKey.HOST_SHARED_LINK,
  SharedLinkTableKey.CREATED_AT,
  SharedLinkTableKey.VISITOR,
  SharedLinkTableKey.STORED,
  SharedLinkTableKey.SIZE,
  SharedLinkTableKey.DAYS_NOT_VISIT,
  SharedLinkTableKey.STATE,
  SharedLinkTableKey.ACTION,
];

// CreatedBy use by filter shared links list
export const enum CreatedBy {
  AUTO_SHARE = "Auto Share",
  LINT_TO_SHARE = "Link to Share",
}

// the link format types
export const enum LinkFormatType {
  TEXT = "Text",
  HTML = "Html",
  BB_CODE = "BBCode",
}

// reset password steps
export const enum ResetPasswordSteps {
  SEND_VERIFICATION_CODE = "SEND_VERIFICATION_CODE",
  ENTER_VERIFICATION_CODE = "ENTER_VERIFICATION_CODE",
  ENTER_NEW_PASSWORD = "ENTER_NEW_PASSWORD",
  RESET_PASSWORD_RESULT = "RESET_PASSWORD_RESULT",
}
// reset password component params
export interface StepComponentParams {
  setStep: (step: ResetPasswordSteps) => void;
}