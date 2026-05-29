// Backend API dan keladigan ma'lumotlar tiplari.
// Maydon nomlari Go backend JSON javoblari bilan bir xil.

export interface Person {
  id: number;
  name: string;
  age: number;
  region: string;
  diagnosis: string;
  help: string;
  facility: string;
  facility_verified: boolean;
  org: string;
  story: string;
  photo_url: string;
  target: number; // so'm
  raised: number; // so'm
  donors: number;
  days_left: number;
  urgent: boolean;
  category: string;
  author_name: string;
  author_role: string;
  status: string;
  created_at: string;
}

export interface SuccessStory {
  id: number;
  name: string;
  age: number;
  region: string;
  diagnosis: string;
  photo_url: string;
  raised: number;
  target: number;
}

export interface RecentDonation {
  id: number;
  donor: string;
  anonim: boolean;
  amount_som: number;
  person_id: number;
  person_name: string;
  at: string;
}

export interface PublicStats {
  total_raised_som: number;
  total_people: number;
  active_people: number;
  urgent_people: number;
  closed_people: number;
  total_donors: number;
  month_donors: number;
  month_raised_som: number;
  month_delta_percent: number;
}

export type Provider = "click" | "payme";

export interface CreatePaymentReq {
  person_id: number;
  amount: number; // so'm
  anonim: boolean;
  donor_name?: string;
  donor_phone?: string;
  tg_user_id?: number;
  tg_username?: string;
}

export interface CreatePaymentRes {
  donation_id: number;
  redirect_url: string;
}

// Telegram foydalanuvchisining bitta yordami (backend'dan)
export interface MyDonationItem {
  id: number;
  person_id: number;
  person_name: string;
  provider: string;
  amount_som: number;
  status: string; // pending | prepared | paid | cancelled
  created_at: string;
  paid_at: string | null;
}

export interface MyDonationsRes {
  items: MyDonationItem[];
  total_paid_som: number;
  paid_people: number;
  count: number;
}
