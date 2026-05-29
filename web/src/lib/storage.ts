// localStorage: sevimlilar (favorites).
// Eslatma: yordamlar tarixi endi Telegram ID bo'yicha backend'da saqlanadi
// (qarang: lib/telegram.ts, api.myDonations) — bu yerda faqat sevimlilar.

const FAV_KEY = "inson-uchun:favorites";

export function getFavorites(): number[] {
  try {
    return JSON.parse(localStorage.getItem(FAV_KEY) || "[]");
  } catch {
    return [];
  }
}

export function toggleFavorite(id: number): number[] {
  const set = new Set(getFavorites());
  if (set.has(id)) set.delete(id);
  else set.add(id);
  const arr = [...set];
  localStorage.setItem(FAV_KEY, JSON.stringify(arr));
  return arr;
}
