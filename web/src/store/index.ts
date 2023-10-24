import { create } from "zustand";
import { AccountState, createAccountStore } from "./account";
import { GlobalState, createGlobalStore } from "./global";
import { LinkState, createLinkStore } from "./link";
import { PikPakState, createPikPakStore } from "./pikpak";

const useStore = create<GlobalState & AccountState & LinkState & PikPakState>((...set) => ({
  ...createAccountStore(...set),
  ...createGlobalStore(...set),
  ...createLinkStore(...set),
  ...createPikPakStore(...set),
}));

export default useStore;
