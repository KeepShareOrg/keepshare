import { SharedLinkInfo } from "@/api/link";
import { SharedLinkTableKey, supportTableColumns } from "@/constant";
import { StateCreator } from "zustand";

export interface LinkState {
  isLoading: boolean;
  tableData: SharedLinkInfo[];
  totalSharedNum: number;
  visibleTableColumns: SharedLinkTableKey[];
  selectedSharedLinkKeys: React.Key[];
  totalSharedLinks: SharedLinkInfo[];

  setTableData: (tableData: SharedLinkInfo[]) => void;
  setIsLoading: (isLoading: boolean) => void;
  setTotalSharedNum: (totalSharedNum: number) => void;
  setVisibleTableColumns: (visibleTableColumns: SharedLinkTableKey[]) => void;
  setSelectedSharedLinkKeys: (selectedSharedLinkKeys: React.Key[]) => void;
  setTotalSharedLinks: (totalSharedLinks: SharedLinkInfo[]) => void;
}

export const createLinkStore: StateCreator<LinkState> = (set) => ({
  isLoading: false,
  tableData: [],
  totalSharedNum: 0,
  visibleTableColumns: supportTableColumns.filter(v => ![SharedLinkTableKey.ORIGINAL_LINKS].includes(v)),
  selectedSharedLinkKeys: [],
  totalSharedLinks: [],

  setIsLoading: (isLoading) => set({ isLoading }),
  setTableData: (tableData) => set({ tableData }),
  setTotalSharedNum: (totalSharedNum) => set({ totalSharedNum }),
  setVisibleTableColumns: (visibleTableColumns) => set({ visibleTableColumns }),
  setSelectedSharedLinkKeys: (selectedSharedLinkKeys) =>
    set({ selectedSharedLinkKeys }),
  setTotalSharedLinks: (totalSharedLinks) => set({ totalSharedLinks }),
});
