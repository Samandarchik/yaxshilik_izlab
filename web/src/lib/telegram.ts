// Telegram Mini App bilan ishlash. Login o'rniga foydalanuvchini
// Telegram'ning o'zi beradi (id, ism, username, rasm).
// Hujjat: https://core.telegram.org/bots/webapps

export interface TgUser {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
}

interface TelegramWebApp {
  initData: string;
  initDataUnsafe: { user?: TgUser };
  ready: () => void;
  expand: () => void;
  colorScheme: "light" | "dark";
}

declare global {
  interface Window {
    Telegram?: { WebApp?: TelegramWebApp };
  }
}

const wa = typeof window !== "undefined" ? window.Telegram?.WebApp : undefined;

// Haqiqiy Telegram ichida ochilganmi? (initData signed bo'lsa — ha)
export const isTelegram = !!(wa && wa.initData);

// Mini App ishga tushganda chaqiriladi
export function initTelegram() {
  if (wa) {
    try {
      wa.ready();
      wa.expand();
    } catch {
      /* ignore */
    }
  }
}

// Dev (oddiy brauzer) uchun barqaror soxta foydalanuvchi —
// shunda "Mening yordamlarim" lokal testda ham ishlaydi.
function devUser(): TgUser {
  let id = Number(localStorage.getItem("dev-tg-id") || 0);
  if (!id) {
    id = Math.floor(100000 + Math.random() * 900000);
    localStorage.setItem("dev-tg-id", String(id));
  }
  return { id, first_name: "Mehmon", username: "dev_user" };
}

// Joriy foydalanuvchi (Telegram'dan yoki dev-fallback)
export function getTgUser(): TgUser {
  return wa?.initDataUnsafe?.user ?? devUser();
}

// Backendga yuboriladigan signed initData (Telegram'da) yoki bo'sh (dev)
export function getInitData(): string {
  return wa?.initData ?? "";
}

// /api/my/... so'rovlari uchun autentifikatsiya:
// Telegram'da — signed header; dev'da — ?tg_id= (backend non-release'da qabul qiladi).
export function authHeaders(): HeadersInit {
  const data = getInitData();
  return data ? { "X-Telegram-Init-Data": data } : {};
}

export function authQuery(): string {
  if (getInitData()) return "";
  return `?tg_id=${getTgUser().id}`;
}
