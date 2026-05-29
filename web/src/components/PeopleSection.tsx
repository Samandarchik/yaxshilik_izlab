import { useMemo, useState } from "react";
import type { Person } from "../lib/types";
import { percent } from "../lib/format";
import { getFavorites, toggleFavorite } from "../lib/storage";
import { useToast } from "../context";
import PersonCard from "./PersonCard";
import CategoryFilter from "./CategoryFilter";

type Sort = "yangi" | "shoshilinch" | "oz_qoldi" | "katta" | "ommabop";

const SORTS: { key: Sort; label: string }[] = [
  { key: "yangi", label: "Yangilari" },
  { key: "shoshilinch", label: "Shoshilinchlari" },
  { key: "oz_qoldi", label: "Yopilishiga oz qoldi" },
  { key: "katta", label: "Eng katta miqdor" },
  { key: "ommabop", label: "Eng ko'p yordamchili" },
];

export default function PeopleSection({ people }: { people: Person[] }) {
  const toast = useToast();
  const [cat, setCat] = useState("all");
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<Sort>("yangi");
  const [onlyFav, setOnlyFav] = useState(false);
  const [favs, setFavs] = useState<number[]>(getFavorites);

  function onFav(id: number) {
    const next = toggleFavorite(id);
    setFavs(next);
    if (next.includes(id)) toast("Sevimliga qo'shildi");
  }

  const list = useMemo(() => {
    let r = [...people];

    if (cat !== "all") r = r.filter((p) => (p.category || "").toLowerCase() === cat);
    if (onlyFav) r = r.filter((p) => favs.includes(p.id));
    if (query.trim()) {
      const q = query.toLowerCase();
      r = r.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          p.diagnosis.toLowerCase().includes(q) ||
          p.region.toLowerCase().includes(q),
      );
    }

    switch (sort) {
      case "shoshilinch":
        r.sort((a, b) => Number(b.urgent) - Number(a.urgent));
        break;
      case "oz_qoldi":
        r.sort((a, b) => percent(b.raised, b.target) - percent(a.raised, a.target));
        break;
      case "katta":
        r.sort((a, b) => b.target - a.target);
        break;
      case "ommabop":
        r.sort((a, b) => b.donors - a.donors);
        break;
      default:
        r.sort((a, b) => +new Date(b.created_at) - +new Date(a.created_at));
    }
    return r;
  }, [people, cat, onlyFav, favs, query, sort]);

  const urgentCount = people.filter((p) => p.urgent).length;

  return (
    <section id="shoshilinch" className="mx-auto w-full max-w-6xl px-4 py-10 sm:px-6 sm:py-14">
      <div className="mb-6">
        <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Yordam <span className="font-display italic text-accent">kutmoqdalar</span>
        </h2>
        <p className="mt-2 text-ink-2">
          {people.length} ta so'rov · {urgentCount} ta shoshilinch
        </p>
      </div>

      {/* Kategoriyalar */}
      <CategoryFilter active={cat} onChange={setCat} />

      {/* Qidiruv + saralash + sevimli */}
      <div className="mt-4 flex flex-wrap items-center gap-3">
        <div className="flex min-w-[220px] flex-1 items-center gap-2 rounded-xl border border-line bg-white px-4 py-2.5 focus-within:border-accent">
          <span className="text-ink-3">🔍</span>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Ism, tashxis yoki viloyat..."
            className="w-full bg-transparent text-sm outline-none"
          />
        </div>

        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as Sort)}
          className="rounded-xl border border-line bg-white px-4 py-2.5 text-sm font-semibold outline-none focus:border-accent"
        >
          {SORTS.map((s) => (
            <option key={s.key} value={s.key}>
              {s.label}
            </option>
          ))}
        </select>

        <button
          onClick={() => setOnlyFav((v) => !v)}
          className={`chip border px-4 py-2.5 text-sm transition ${
            onlyFav ? "border-accent bg-accent text-white" : "border-line bg-white text-ink-2"
          }`}
        >
          ♡ Sevimli
        </button>
      </div>

      {/* Natijalar */}
      {list.length === 0 ? (
        <div className="card mt-8 flex flex-col items-center gap-3 py-16 text-center">
          <p className="text-4xl">🔍</p>
          <p className="font-semibold">Hech narsa topilmadi</p>
          <button
            onClick={() => {
              setCat("all");
              setQuery("");
              setOnlyFav(false);
            }}
            className="btn-soft"
          >
            Filtrlarni tozalash
          </button>
        </div>
      ) : (
        <div className="mt-7 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {list.map((p) => (
            <PersonCard key={p.id} person={p} isFav={favs.includes(p.id)} onToggleFav={onFav} />
          ))}
        </div>
      )}
    </section>
  );
}
