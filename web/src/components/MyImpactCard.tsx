import { Link } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../hooks/useApi";
import { getTgUser } from "../lib/telegram";
import { compactSom } from "../lib/format";
import Avatar from "./ui/Avatar";

// Foydalanuvchining shaxsiy hissasi (Telegram ID bo'yicha backend'dan).
function tier(total: number): string {
  if (total >= 5_000_000) return "Platinum yordamchi";
  if (total >= 2_000_000) return "Oltin yordamchi";
  if (total >= 500_000) return "Kumush yordamchi";
  if (total > 0) return "Bronza yordamchi";
  return "Yangi yordamchi";
}

export default function MyImpactCard() {
  const tg = getTgUser();
  const tgName = [tg.first_name, tg.last_name].filter(Boolean).join(" ");
  const { data } = useApi(() => api.myDonations(), []);

  const total = data?.total_paid_som ?? 0;
  const people = data?.paid_people ?? 0;
  const count = data?.count ?? 0;

  return (
    <div className="card overflow-hidden">
      <div className="bg-gradient-to-br from-ink to-[#1E293B] p-5 text-white">
        <div className="mb-3 flex items-center gap-2.5">
          <Avatar name={tgName} src={tg.photo_url} size={36} />
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold">{tg.first_name || "Mehmon"}</p>
            <p className="text-xs opacity-70">★ {tier(total)}</p>
          </div>
        </div>
        <p className="text-3xl font-extrabold">
          {compactSom(total)} <span className="text-base font-bold">so'm</span>
        </p>
        <p className="mt-1 text-sm opacity-80">jami yordam berdingiz</p>
      </div>
      <div className="grid grid-cols-2 divide-x divide-line">
        <div className="p-4 text-center">
          <p className="text-2xl font-extrabold">{people}</p>
          <p className="text-xs text-ink-3">kishiga yordam</p>
        </div>
        <div className="p-4 text-center">
          <p className="text-2xl font-extrabold">{count}</p>
          <p className="text-xs text-ink-3">marta</p>
        </div>
      </div>
      <div className="p-4 pt-0">
        <Link to="/mening-yordamlarim" className="btn-ghost w-full">
          Yordamlarim tarixi
        </Link>
      </div>
    </div>
  );
}
