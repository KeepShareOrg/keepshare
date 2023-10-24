import { PikPakHostInfo } from "@/api/pikpak";
import { StateCreator } from "zustand";

export interface PikPakState {
	pikPakInfo: Partial<PikPakHostInfo>;

	setPikPakInfo: (pikPakInfo: PikPakHostInfo) => void;
}

export const createPikPakStore: StateCreator<PikPakState> = (set) => ({
	pikPakInfo: {},

	setPikPakInfo: (pikPakInfo) => set({ pikPakInfo }),
});
