import { Link } from "react-router-dom";
import type { Person } from "../lib/types";
import { compactSom, formatSom, percent } from "../lib/format";
import { useDonate } from "../context";
import ProgressBar from "./ui/ProgressBar";

export default function PersonCard({
  person,
  isFav,
  onToggleFav,
}: {
  person: Person;
  isFav: boolean;
  onToggleFav: (id: number) => void;
}) {
  const openDonate = useDonate();
  const pct = percent(person.raised, person.target);

  return (
    <div className="card group flex flex-col overflow-hidden transition hover:-translate-y-1 hover:shadow-card">
      {/* Rasm */}
      <div className="relative aspect-[16/10] overflow-hidden bg-page">
        <Link to={`/loyiha/${person.id}`}>
          {person.photo_url ? (
            <img
              src={person.photo_url}
              alt={person.name}
              loading="lazy"
              className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-ink-3">
              Rasm yo'q
            </div>
          )}
        </Link>

        {/* Yuqori belgilar */}
        <div className="absolute left-3 top-3 flex gap-2">
          {person.urgent && (
            <span className="chip bg-accent text-white shadow-glow">Shoshilinch</span>
          )}
          <span className="chip bg-white/90 text-ink backdrop-blur">№ {person.id}</span>
        </div>

        {/* Sevimli tugmasi */}
        <button
          onClick={() => onToggleFav(person.id)}
          aria-label="Sevimli"
          className="absolute right-3 top-3 flex h-9 w-9 items-center justify-center rounded-full bg-white/90 backdrop-blur transition hover:scale-110"
        >
          <svg
            viewBox="0 0 24 24"
            className={`h-5 w-5 transition ${isFav ? "animate-pop fill-accent text-accent" : "fill-none text-ink-3"}`}
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M12 21s-7.5-4.6-10-9.3C.4 8.4 2 5 5.3 5c2 0 3.4 1.1 4.2 2.4C10.3 6.1 11.7 5 13.7 5 17 5 18.6 8.4 17 11.7 14.5 16.4 12 21 12 21z" />
          </svg>
        </button>
      </div>

      {/* Matn qismi */}
      <div className="flex flex-1 flex-col p-4">
        {person.facility_verified && (
          <div className="mb-1.5 flex items-center gap-1 text-xs font-semibold text-emerald-600">
            <span>✓</span> Tasdiqlangan
          </div>
        )}

        <Link to={`/loyiha/${person.id}`} className="hover:text-accent">
          <h3 className="text-lg font-bold leading-tight">
            {person.name} <span className="text-ink-3">· {person.age} yosh</span>
          </h3>
        </Link>

        <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-sm text-ink-2">
          <span className="inline-flex items-center gap-1">📍 {person.region}</span>
          {person.days_left > 0 && (
            <span className="inline-flex items-center gap-1">⏳ {person.days_left} kun qoldi</span>
          )}
        </div>

        {/* Teglar */}
        <div className="mt-3 flex flex-wrap gap-2">
          {person.diagnosis && (
            <span className="chip bg-amber-50 text-amber-700">{person.diagnosis}</span>
          )}
          {person.help && (
            <span className="chip bg-page text-ink-2">{person.help}</span>
          )}
        </div>

        {/* Progress */}
        <div className="mt-4">
          <div className="mb-1.5 flex items-center justify-between text-sm">
            <span className="font-bold text-ink">{formatSom(person.raised)} so'm</span>
            <span className="font-semibold text-accent">{pct}%</span>
          </div>
          <ProgressBar value={pct} />
          <div className="mt-1.5 flex items-center justify-between text-xs text-ink-3">
            <span>{compactSom(person.target)} so'mdan</span>
            <span>{person.donors} yordamchi</span>
          </div>
        </div>

        {/* Tugma */}
        <button
          onClick={() => openDonate(person)}
          className="btn-primary mt-4 w-full"
        >
          Yordam ber
        </button>
      </div>
    </div>
  );
}
