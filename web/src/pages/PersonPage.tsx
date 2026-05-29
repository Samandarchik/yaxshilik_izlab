import { useParams, Link } from "react-router-dom";
import { api } from "../lib/api";
import { useApi } from "../hooks/useApi";
import { formatSom, percent, compactSom } from "../lib/format";
import { useDonate } from "../context";
import ProgressBar from "../components/ui/ProgressBar";
import Spinner from "../components/ui/Spinner";

export default function PersonPage() {
  const { id } = useParams();
  const openDonate = useDonate();
  const { data: p, loading, error } = useApi(() => api.person(Number(id)), [id]);

  if (loading) return <Spinner label="Yuklanmoqda..." />;
  if (error || !p)
    return (
      <div className="mx-auto max-w-3xl px-4 py-20 text-center">
        <p className="text-ink-2">Hikoya topilmadi.</p>
        <Link to="/" className="btn-soft mt-4">
          Bosh sahifa
        </Link>
      </div>
    );

  const pct = percent(p.raised, p.target);

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6">
      <Link to="/" className="mb-5 inline-flex items-center gap-1 text-sm font-semibold text-ink-2 hover:text-accent">
        ← Barcha so'rovlar
      </Link>

      <div className="grid gap-8 lg:grid-cols-[1.4fr_1fr]">
        {/* Chap: rasm + hikoya */}
        <div>
          <div className="relative overflow-hidden rounded-2xl bg-page">
            {p.photo_url ? (
              <img src={p.photo_url} alt={p.name} className="aspect-[16/11] w-full object-cover" />
            ) : (
              <div className="flex aspect-[16/11] items-center justify-center text-ink-3">📷</div>
            )}
            <div className="absolute left-4 top-4 flex gap-2">
              {p.urgent && <span className="chip bg-accent text-white shadow-glow">Shoshilinch</span>}
              <span className="chip bg-white/90 backdrop-blur">Hikoya № {p.id}</span>
            </div>
          </div>

          <h1 className="mt-6 text-3xl font-bold tracking-tight">
            {p.name} <span className="text-ink-3">· {p.age} yosh</span>
          </h1>
          <p className="mt-1 text-ink-2">📍 {p.region}</p>

          <div className="mt-5 grid gap-3 sm:grid-cols-2">
            <Info label="Tashxis" value={p.diagnosis} />
            <Info label="Yordam turi" value={p.help} />
            <Info label="Tibbiy muassasa" value={p.facility} verified={p.facility_verified} />
          </div>

          {p.story && (
            <div className="mt-7">
              <h2 className="text-lg font-bold">Bemorning hikoyasi</h2>
              <p className="mt-3 whitespace-pre-line leading-relaxed text-ink-2">{p.story}</p>
            </div>
          )}

          {p.author_name && (
            <p className="mt-6 text-sm text-ink-3">
              — {p.author_name}
              {p.author_role && `, ${p.author_role}`}
            </p>
          )}
        </div>

        {/* O'ng: yig'ilgan + yordam tugmasi (sticky) */}
        <div>
          <div className="card sticky top-20 p-6">
            <p className="text-3xl font-extrabold">{formatSom(p.raised)} so'm</p>
            <p className="text-sm text-ink-3">{compactSom(p.target)} so'm maqsaddan</p>

            <div className="mt-4">
              <ProgressBar value={pct} />
              <div className="mt-2 flex items-center justify-between text-sm">
                <span className="font-semibold text-accent">{pct}%</span>
                <span className="text-ink-3">{p.donors} yordamchi</span>
              </div>
            </div>

            <div className="mt-4 rounded-xl bg-page p-3 text-center text-sm text-ink-2">
              Qolgan: <b className="text-ink">{formatSom(Math.max(0, p.target - p.raised))} so'm</b>
            </div>

            <button onClick={() => openDonate(p)} className="btn-primary mt-5 w-full !py-4 text-base">
              Yordam ber
            </button>
            <p className="mt-3 text-center text-xs text-ink-3">
              🔒 Click yoki Payme orqali xavfsiz to'lov
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function Info({ label, value, verified }: { label: string; value: string; verified?: boolean }) {
  if (!value) return null;
  return (
    <div className="card p-4">
      <p className="text-xs font-semibold uppercase tracking-wide text-ink-3">{label}</p>
      <p className="mt-1 font-semibold">
        {value} {verified && <span className="text-emerald-500">✓</span>}
      </p>
    </div>
  );
}
