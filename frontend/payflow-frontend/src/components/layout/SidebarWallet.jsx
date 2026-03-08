import { useWalletStore } from "../../store/walletStore";
import { formatIDR } from "../../lib/utils";

export default function SidebarWallet() {
  const { wallet } = useWalletStore();

  return (
    <div className="mx-4 mb-6 card px-4 py-3">
      <p className="text-slate-subtle text-xs mb-1">Saldo Kamu</p>
      <p className="balance-text text-jade text-lg">
        {wallet ? formatIDR(wallet.balance) : "—"}
      </p>
    </div>
  );
}
