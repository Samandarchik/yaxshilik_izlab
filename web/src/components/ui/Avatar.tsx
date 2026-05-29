import { initials } from "../../lib/format";

// Rasm bo'lsa rasm, bo'lmasa ism bosh harflari bilan rangli doira.
export default function Avatar({
  name,
  src,
  size = 40,
}: {
  name: string;
  src?: string;
  size?: number;
}) {
  if (src) {
    return (
      <img
        src={src}
        alt={name}
        style={{ width: size, height: size }}
        className="rounded-full object-cover"
      />
    );
  }
  // Ismdan barqaror rang tanlash
  const colors = ["#FF6B35", "#5B8DEF", "#34D399", "#A78BFA", "#F472B6", "#FBBF24"];
  const idx = name.charCodeAt(0) % colors.length;
  return (
    <div
      style={{ width: size, height: size, background: colors[idx], fontSize: size * 0.4 }}
      className="flex items-center justify-center rounded-full font-bold text-white"
    >
      {initials(name)}
    </div>
  );
}
