// Barcha backend so'rovlari shu yerda — bitta joyda, tushunarli.
import type {
  Person,
  SuccessStory,
  RecentDonation,
  PublicStats,
  Provider,
  CreatePaymentReq,
  CreatePaymentRes,
  MyDonationsRes,
} from "./types";
import { authHeaders, authQuery } from "./telegram";

// Same-origin: React build Go backend ichidan beriladi, shuning uchun nisbiy yo'l.
const BASE = "/api";

async function get<T>(path: string, headers?: HeadersInit): Promise<T> {
  const res = await fetch(BASE + path, headers ? { headers } : undefined);
  if (!res.ok) throw new Error(`So'rov xatosi: ${res.status}`);
  return res.json() as Promise<T>;
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(BASE + path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    let msg = `So'rov xatosi: ${res.status}`;
    try {
      const data = await res.json();
      if (data?.error) msg = data.error;
    } catch {
      /* ignore */
    }
    throw new Error(msg);
  }
  return res.json() as Promise<T>;
}

export const api = {
  people: () => get<Person[]>("/people"),
  person: (id: number) => get<Person>(`/people/${id}`),
  successStories: (limit = 8) => get<SuccessStory[]>(`/success-stories?limit=${limit}`),
  recentDonations: (limit = 10) => get<RecentDonation[]>(`/donations/recent?limit=${limit}`),
  stats: () => get<PublicStats>("/stats/public"),

  createPayment: (provider: Provider, body: CreatePaymentReq) =>
    post<CreatePaymentRes>(`/${provider}/create`, body),

  // Joriy Telegram foydalanuvchisining yordamlari (header yoki ?tg_id= bilan)
  myDonations: () => get<MyDonationsRes>(`/my/donations${authQuery()}`, authHeaders()),
};
