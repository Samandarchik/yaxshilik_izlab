// Raqam va sanalarni chiroyli ko'rsatish uchun yordamchilar.

// 1500000 -> "1 500 000"
export function formatSom(n: number): string {
  return Math.round(n)
    .toString()
    .replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

// 1500000 -> "1.5 mln", 250000 -> "250 ming"
export function compactSom(n: number): string {
  if (n >= 1_000_000) {
    const m = n / 1_000_000;
    return `${m % 1 === 0 ? m : m.toFixed(1)} mln`;
  }
  if (n >= 1_000) return `${Math.round(n / 1000)} ming`;
  return `${n}`;
}

// Foiz (0..100), maksimal 100 da to'xtaydi
export function percent(raised: number, target: number): number {
  if (target <= 0) return 0;
  return Math.min(100, Math.round((raised / target) * 100));
}

// "3 daqiqa oldin", "2 soat oldin", "kecha" ...
export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  if (isNaN(then)) return "";
  const sec = Math.floor((Date.now() - then) / 1000);
  if (sec < 60) return "hozir";
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min} daqiqa oldin`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr} soat oldin`;
  const day = Math.floor(hr / 24);
  if (day === 1) return "kecha";
  if (day < 30) return `${day} kun oldin`;
  const mon = Math.floor(day / 30);
  return `${mon} oy oldin`;
}

// Ismni qisqartirish: "Aziz Karimov" -> "Aziz K."
export function shortName(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length < 2) return name;
  return `${parts[0]} ${parts[1][0]}.`;
}

// Ismdan bosh harflar (avatar uchun): "Aziz Karimov" -> "AK"
export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[1][0]).toUpperCase();
}
