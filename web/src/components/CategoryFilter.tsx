export interface Category {
  key: string;
  label: string;
  icon: string;
}

// Backend'dagi `category` qiymatlariga mos keladi (bo'sh bo'lsa "barchasi")
export const CATEGORIES: Category[] = [
  { key: "all", label: "Barchasi", icon: "∞" },
  { key: "yurak", label: "Yurak", icon: "♡" },
  { key: "onkologiya", label: "Onkologiya", icon: "⊚" },
  { key: "bolalar", label: "Bolalar", icon: "◐" },
  { key: "operatsiya", label: "Operatsiya", icon: "✚" },
  { key: "imkoniyat", label: "Imkoniyat", icon: "▴" },
];

export default function CategoryFilter({
  active,
  onChange,
}: {
  active: string;
  onChange: (key: string) => void;
}) {
  return (
    <div className="no-scrollbar -mx-4 flex gap-2 overflow-x-auto px-4 sm:mx-0 sm:flex-wrap sm:px-0">
      {CATEGORIES.map((c) => (
        <button
          key={c.key}
          onClick={() => onChange(c.key)}
          className={`chip shrink-0 border px-4 py-2.5 text-sm transition ${
            active === c.key
              ? "border-accent bg-accent text-white shadow-glow"
              : "border-line bg-white text-ink-2 hover:border-accent/40 hover:text-accent"
          }`}
        >
          <span className="text-base">{c.icon}</span> {c.label}
        </button>
      ))}
    </div>
  );
}
