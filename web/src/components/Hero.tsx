import { Link } from "react-router-dom";
import type { Person, PublicStats } from "../lib/types";
import { compactSom, formatSom, percent } from "../lib/format";
import ProgressBar from "./ui/ProgressBar";

export default function Hero({
  featured,
  stats,
}: {
  featured: Person | null;
  stats: PublicStats | null;
}) {
  return (
    <section className="mx-auto grid max-w-6xl gap-6 px-4 pb-4 pt-10 sm:px-6 lg:grid-cols-[1.5fr_1fr]">
      {/* Chap: sarlavha + tanlangan bemor */}
      <div className="flex flex-col">
        <span className="chip w-fit bg-accent-soft text-accent">
          ❤️ Yaxshilik Izlab · ishonchli xayriya platformasi
        </span>
        <h1 className="mt-4 text-4xl font-bold leading-[1.1] tracking-tight sm:text-5xl">
          Yaxshilik —{" "}
          <span className="font-display italic text-accent">bir bosishda</span>
          <br />
          muhtojlarga bevosita yordam
        </h1>
        <p className="mt-4 max-w-lg text-lg text-ink-2">
          Tasdiqlangan hikoyani tanlang, Click yoki Payme orqali yordam bering —
          har bir so'm to'g'ridan-to'g'ri insonga yetadi.
        </p>
        <p className="mt-2 font-display text-lg italic text-ink-3">
          «Bizni kutar, biz kutgan kunlar!»
        </p>

        <div className="mt-6 flex flex-wrap gap-3">
          <a href="#shoshilinch" className="btn-primary">
            Loyihalarni ko'rish
          </a>
          <a href="#qanday" className="btn-ghost">
            Qanday ishlaydi?
          </a>
        </div>

        {/* Tanlangan bemor kartasi */}
        {featured && (
          <div className="card mt-8 flex gap-4 p-3 sm:p-4">
            <Link to={`/loyiha/${featured.id}`} className="shrink-0">
              {featured.photo_url ? (
                <img
                  src={featured.photo_url}
                  alt={featured.name}
                  className="h-28 w-28 rounded-xl object-cover sm:h-32 sm:w-32"
                />
              ) : (
                <div className="flex h-28 w-28 items-center justify-center rounded-xl bg-page text-ink-3 sm:h-32 sm:w-32">
                  📷
                </div>
              )}
            </Link>
            <div className="flex min-w-0 flex-1 flex-col">
              <div className="flex items-center gap-2">
                {featured.urgent && (
                  <span className="chip bg-accent text-white">Shoshilinch</span>
                )}
                <span className="text-xs font-semibold text-ink-3">
                  Hozir eng zarur
                </span>
              </div>
              <h3 className="mt-1 truncate text-lg font-bold">
                {featured.name} · {featured.age} yosh
              </h3>
              <p className="truncate text-sm text-ink-2">{featured.diagnosis}</p>
              <div className="mt-auto pt-3">
                <ProgressBar value={percent(featured.raised, featured.target)} />
                <div className="mt-1.5 flex items-center justify-between text-xs text-ink-3">
                  <span className="font-semibold text-ink">
                    {formatSom(featured.raised)} so'm
                  </span>
                  <span>{compactSom(featured.target)} so'mdan</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* O'ng: umumiy statistika */}
      <div className="flex flex-col gap-4">
        <div className="card bg-gradient-to-br from-accent to-[#FF8B5A] p-6 text-white">
          <p className="text-sm font-medium opacity-90">Bu oy yig'ildi</p>
          <p className="mt-1 text-4xl font-extrabold">
            {stats ? compactSom(stats.month_raised_som) : "—"}
            <span className="ml-1 text-lg font-bold">so'm</span>
          </p>
          {stats && stats.month_delta_percent !== 0 && (
            <p className="mt-2 inline-flex items-center gap-1 rounded-full bg-white/20 px-2.5 py-1 text-xs font-semibold">
              {stats.month_delta_percent > 0 ? "▲" : "▼"}{" "}
              {Math.abs(stats.month_delta_percent)}% o'tgan oyga nisbatan
            </p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <StatBox label="Yordamchilar" value={stats?.total_donors} />
          <StatBox label="Faol so'rov" value={stats?.active_people} />
          <StatBox label="Bu oy yordam berdi" value={stats?.month_donors} />
          <StatBox label="Yopilgan hikoya" value={stats?.closed_people} accent />
        </div>
      </div>
    </section>
  );
}

function StatBox({
  label,
  value,
  accent,
}: {
  label: string;
  value?: number;
  accent?: boolean;
}) {
  return (
    <div className="card p-4">
      <p className={`text-2xl font-extrabold ${accent ? "text-accent" : "text-ink"}`}>
        {value ?? "—"}
      </p>
      <p className="mt-0.5 text-xs font-medium text-ink-3">{label}</p>
    </div>
  );
}
