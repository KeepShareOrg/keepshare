import { SupportLocales } from "@/constant";
import { StateCreator } from "zustand";

type SupportThemeMode = "dark" | "light";
export interface GlobalState {
  locale: SupportLocales;
  setLocale: (locale: SupportLocales) => void;
  asideCollapsed: boolean;
  setAsideCollapsed: (v: boolean) => void;
  themeMode: SupportThemeMode;
  setThemeMode: (themeMode: SupportThemeMode) => void;
  isMobile: boolean;
  setIsMobile: (isMobile: boolean) => void;
  showMenuDrawer: boolean;
  setShowMenuDrawer: (showMenuDrawer: boolean) => void;
}

export const createGlobalStore: StateCreator<GlobalState> = (set) => ({
  locale: SupportLocales.En,
  setLocale: (locale) => set(() => ({ locale })),
  asideCollapsed: false,
  setAsideCollapsed: (v: boolean) => set(() => ({ asideCollapsed: v })),
  themeMode: "light",
  setThemeMode: (themeMode) => {
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', themeMode === 'light' ? '#FFFFFF' : '#141414');
    return set(() => ({ themeMode }));
  },
  isMobile: false,
  setIsMobile: (isMobile) => set(() => ({ isMobile })),
  showMenuDrawer: false,
  setShowMenuDrawer: (showMenuDrawer) => set(() => ({ showMenuDrawer })),
});
