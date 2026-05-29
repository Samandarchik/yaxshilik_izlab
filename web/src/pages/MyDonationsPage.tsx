import { Link } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../hooks/useApi";
import { getTgUser, isTelegram } from "../lib/telegram";
import { formatSom, compactSom, timeAgo } from "../lib/format";
import type { MyDonationItem } from "../lib/types";
import Avatar from "../components/ui/Avatar";
import Spinner from "../components/ui/Spinner";

export default function MyDonationsPage() {
  const tg = getTgUser();
  const tgName = [tg.first_name, tg.last_name].filter(Boolean).join(" ");
  const { data, loading, error } = useApi(() => api.myDonations(), []);

  return (
    <div className="mx-auto max-w-3xl px-4 py-10 sm:px-6">
      {/* Foydalanuvchi (Telegram) */}
      <div className="card flex items-center gap-4 p-5">
        <Avatar name={tgName} src={tg.photo_url} size={56} />
        <div className="min-w-0 flex-1">
          <h1 className="truncate text-xl font-bold">{tgName || "Mehmon"}</h1>
          <p className="text-sm text-ink-3">
            {tg.username ? `@${tg.username}` : "Telegram foydalanuvchisi"}
          </p>
        </div>
        {!isTelegram && (
          <span className="chip bg-amber-50 text-amber-600">Test rejimi</span>
        )}
      </div>

      <h2 className="mb-4 mt-8 text-lg font-bold">
        Mening <span className="font-display italic text-accent">yordamlarim</span>
      </h2>

      {/* Statistika */}
      {data && (
        <div className="grid grid-cols-3 gap-4">
          <StatBox value={compactSom(data.total_paid_som)} unit="so'm" label="Jami yordam" />
          <StatBox value={`${data.paid_people}`} label="Kishiga" />
          <StatBox value={`${data.count}`} label="Marta" />
        </div>
      )}

      {/* Ro'yxat */}
      <div className="mt-8">
        {loading && <Spinner label="Yuklanmoqda..." />}
        {error && (
          <p className="py-10 text-center text-ink-2">
            Ma'lumotni yuklab bo'lmadi. Iltimos, qayta urinib ko'ring.
          </p>
        )}

        {data && data.items.length === 0 && (
          <div className="card flex flex-col items-center gap-4 py-16 text-center">
            <p className="text-4xl">🤍</p>
            <p className="font-semibold">Hali yordam bermagansiz</p>
            <p className="max-w-xs text-sm text-ink-2">
              Birinchi yordamingizni bering — bu yerda tarix paydo bo'ladi.
            </p>
            <Link to="/#shoshilinch" className="btn-primary">
              Loyihalarni ko'rish
            </Link>
          </div>
        )}

        {data && data.items.length > 0 && (
          <ul className="space-y-3">
            {data.items.map((d) => (
              <DonationRow key={d.id} d={d} />
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

function StatBox({ value, unit, label }: { value: string; unit?: string; label: string }) {
  return (
    <div className="card p-4 text-center">
      <p className="text-2xl font-extrabold">
        {value} {unit && <span className="text-sm font-bold text-ink-3">{unit}</span>}
      </p>
      <p className="mt-0.5 text-xs text-ink-3">{label}</p>
    </div>
  );
}

// Holatni o'zbekcha va rangli ko'rsatish
function statusBadge(status: string): { text: string; cls: string } {
  switch (status) {
    case "paid":
      return { text: "Yordam berildi ✓", cls: "bg-emerald-50 text-emerald-600" };
    case "cancelled":
      return { text: "Bekor qilingan", cls: "bg-red-50 text-red-500" };
    case "prepared":
    case "pending":
    default:
      return { text: "Kutilmoqda", cls: "bg-amber-50 text-amber-600" };
  }
}

function DonationRow({ d }: { d: MyDonationItem }) {
  const badge = statusBadge(d.status);
  const when = d.paid_at || d.created_at;
  return (
    <li className="card flex items-center gap-4 p-4">
      <Link
        to={`/loyiha/${d.person_id}`}
        className="flex h-11 w-11 items-center justify-center rounded-xl bg-accent-soft font-bold text-accent"
      >
        {(d.person_name || "?").charAt(0)}
      </Link>
      <div className="min-w-0 flex-1">
        <Link to={`/loyiha/${d.person_id}`} className="truncate font-semibold hover:text-accent">
          {d.person_name || "Noma'lum"}
        </Link>
        <p className="text-xs text-ink-3">
          {d.provider === "click" ? "Click" : "Payme"} · {timeAgo(when)}
        </p>
      </div>
      <div className="text-right">
        <p className="font-bold">{formatSom(d.amount_som)} so'm</p>
        <span className={`chip mt-0.5 px-2 py-0.5 text-[11px] ${badge.cls}`}>{badge.text}</span>
      </div>
    </li>
  );
}
