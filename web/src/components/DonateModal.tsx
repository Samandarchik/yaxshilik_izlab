import { useEffect, useState } from "react";
import type { Person, Provider } from "../lib/types";
import { formatSom } from "../lib/format";
import { api } from "../lib/api";
import { getTgUser } from "../lib/telegram";
import { useToast } from "../context";

const QUICK = [50_000, 100_000, 500_000, 1_000_000];
const MIN = 1_000;
const MAX = 10_000_000;

export default function DonateModal({
  person,
  onClose,
}: {
  person: Person;
  onClose: () => void;
}) {
  const toast = useToast();
  const [amount, setAmount] = useState(100_000);
  const [anonim, setAnonim] = useState(false);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [loading, setLoading] = useState<Provider | null>(null);
  const [err, setErr] = useState<string | null>(null);

  // ESC bilan yopish + fon scroll'ini to'xtatish
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [onClose]);

  const valid = amount >= MIN && amount <= MAX;

  async function pay(provider: Provider) {
    if (!valid) {
      setErr(`Miqdor ${formatSom(MIN)} – ${formatSom(MAX)} so'm orasida bo'lishi kerak`);
      return;
    }
    setErr(null);
    setLoading(provider);
    try {
      const tg = getTgUser();
      // Ism kiritilmagan bo'lsa, Telegram ismini ishlatamiz (anonim bo'lmasa)
      const donorName = anonim ? "" : name.trim() || [tg.first_name, tg.last_name].filter(Boolean).join(" ");
      const res = await api.createPayment(provider, {
        person_id: person.id,
        amount,
        anonim,
        donor_name: donorName,
        donor_phone: phone,
        tg_user_id: tg.id,
        tg_username: tg.username || "",
      });
      // Yordam backend'da 'pending' sifatida saqlandi — to'lov sahifasiga o'tamiz.
      // To'lov tasdiqlangach (webhook) "Mening yordamlarim"da 'paid' bo'lib ko'rinadi.
      window.location.href = res.redirect_url;
    } catch (e) {
      setErr((e as Error).message);
      setLoading(null);
      toast("To'lovni boshlashda xatolik");
    }
  }

  return (
    <div
      className="fixed inset-0 z-[90] flex items-end justify-center bg-ink/40 p-0 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md animate-fade-up rounded-t-3xl bg-white p-6 shadow-card sm:rounded-3xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Sarlavha */}
        <div className="flex items-start justify-between">
          <div>
            <h3 className="text-xl font-bold">Yordam berish</h3>
            <p className="mt-0.5 text-sm text-ink-2">
              {person.name} · {person.age} yosh
            </p>
          </div>
          <button
            onClick={onClose}
            className="flex h-9 w-9 items-center justify-center rounded-full bg-page text-ink-2 hover:text-ink"
            aria-label="Yopish"
          >
            ✕
          </button>
        </div>

        {/* Tez miqdorlar */}
        <div className="mt-5 grid grid-cols-4 gap-2">
          {QUICK.map((q) => (
            <button
              key={q}
              onClick={() => setAmount(q)}
              className={`rounded-xl border py-2.5 text-sm font-bold transition ${
                amount === q
                  ? "border-accent bg-accent text-white"
                  : "border-line bg-white text-ink-2 hover:border-accent/40"
              }`}
            >
              {q >= 1_000_000 ? "1 mln" : `${q / 1000}k`}
            </button>
          ))}
        </div>

        {/* Boshqa miqdor */}
        <label className="mt-4 block">
          <span className="text-sm font-medium text-ink-2">Boshqa miqdor (so'm)</span>
          <div className="mt-1.5 flex items-center rounded-xl border border-line bg-page px-4 focus-within:border-accent">
            <input
              type="text"
              inputMode="numeric"
              value={formatSom(amount)}
              onChange={(e) => {
                const n = parseInt(e.target.value.replace(/\D/g, ""), 10);
                setAmount(isNaN(n) ? 0 : n);
              }}
              className="w-full bg-transparent py-3 text-lg font-bold outline-none"
            />
            <span className="text-ink-3">so'm</span>
          </div>
        </label>

        {/* Anonim + ism */}
        <label className="mt-4 flex cursor-pointer items-center gap-3">
          <input
            type="checkbox"
            checked={anonim}
            onChange={(e) => setAnonim(e.target.checked)}
            className="h-5 w-5 accent-accent"
          />
          <span className="text-sm text-ink-2">Anonim ko'rinishda yuborish</span>
        </label>

        {!anonim && (
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Ismingiz (ixtiyoriy)"
            className="mt-3 w-full rounded-xl border border-line bg-page px-4 py-3 outline-none focus:border-accent"
          />
        )}
        <input
          type="tel"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          placeholder="Telefon (ixtiyoriy)"
          className="mt-3 w-full rounded-xl border border-line bg-page px-4 py-3 outline-none focus:border-accent"
        />

        {err && <p className="mt-3 text-sm font-medium text-red-500">{err}</p>}

        {/* To'lov usullari */}
        <div className="mt-5 grid grid-cols-2 gap-3">
          <button
            disabled={loading !== null}
            onClick={() => pay("click")}
            className="btn rounded-xl bg-[#0EA5E9] py-3.5 font-bold text-white hover:opacity-90 disabled:opacity-50"
          >
            {loading === "click" ? "..." : "Click orqali"}
          </button>
          <button
            disabled={loading !== null}
            onClick={() => pay("payme")}
            className="btn rounded-xl bg-[#1FB36B] py-3.5 font-bold text-white hover:opacity-90 disabled:opacity-50"
          >
            {loading === "payme" ? "..." : "Payme orqali"}
          </button>
        </div>

        <p className="mt-3 text-center text-xs text-ink-3">
          🔒 To'lov xavfsiz, shifrlangan ulanish orqali amalga oshiriladi
        </p>
      </div>
    </div>
  );
}
