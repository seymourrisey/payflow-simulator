import { create } from "zustand";
import { walletApi } from "../lib/api";

export const useWalletStore = create((set) => ({
  wallet: null,
  loading: false,
  error: null,

  fetchWallet: async () => {
    set({ loading: true, error: null });
    try {
      const res = await walletApi.getWallet();
      set({ wallet: res.data.data, loading: false });
    } catch (err) {
      set({ error: err.message, loading: false });
    }
  },

  // Update balance optimistically setelah transaksi berhasil
  updateBalance: (newBalance) =>
    set((state) => ({
      wallet: state.wallet ? { ...state.wallet, balance: newBalance } : null,
    })),
}));
