import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import { SupportLocales } from "@/constant";
import en from "@/locale/en.json";
import zhCN from '@/locale/zh-CN.json';
import zhTW from '@/locale/zh-TW.json';

// support language content config
const resourceConfig = [
  {
    key: SupportLocales.En,
    value: en,
  },
  {
    key: SupportLocales.ZH_CN,
    value: zhCN,
  },
  {
    key: SupportLocales.ZH_TW,
    value: zhTW,
  }
];

const resources = Object.fromEntries(
  resourceConfig.map(({ key, value }) => {
    return [key, { translation: value }];
  }),
);

const defaultLocale = SupportLocales.En;

i18n.use(initReactI18next).init({
  resources,
  lng: defaultLocale,
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
