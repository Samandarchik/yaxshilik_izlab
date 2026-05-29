import { Link } from "react-router-dom";
import type { Person } from "../lib/types";
import { percent } from "../lib/format";

// Maqsadiga eng yaqin qolgan 3 ta bemor.
export default function AlmostThere({ people }: { people: Person[] }) {
  const top = [...people]
    .filter((p) => p.target > 0)
    .sort((a, b) => percent(b.raised, b.target) - percent(a.raised, a.target))
    .slice(0, 3);

  if (top.length === 0) return null;

  return (
    <div className="card p-5">
      <h3 className="mb-4 font-bold">Yopilishiga oz qoldi</h3>
      <ul className="space-y-4">
        {top.map((p) => {
          const pct = percent(p.raised, p.target);
          return (
            <li key={p.id}>
              <Link to={`/loyiha/${p.id}`} className="flex items-center gap-3 group">
                {p.photo_url ? (
                  <img src={p.photo_url} alt={p.name} className="h-12 w-12 rounded-xl object-cover" />
                ) : (
                  <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-page">📷</div>
                )}
                <div className="min-w-0 flex-1">
                  <p className="truncate font-semibold group-hover:text-accent">{p.name}</p>
                  <p className="truncate text-xs text-ink-3">{p.region}</p>
                </div>
                <span className="shrink-0 text-sm font-bold text-emerald-600">{pct}%</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
