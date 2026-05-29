import { useEffect, useState } from "react";
import { api } from "../lib/api";
import type { RecentDonation } from "../lib/types";
import { compactSom, timeAgo } from "../lib/format";
import Avatar from "./ui/Avatar";

export default function LiveFeed() {
  const [items, setItems] = useState<RecentDonation[]>([]);

  useEffect(() => {
    let alive = true;
    const load = () =>
      api
        .recentDonations(8)
        .then((d) => alive && setItems(d))
        .catch(() => {});
    load();
    // Har 15 soniyada yangilab turamiz
    const t = setInterval(load, 15000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, []);

  return (
    <div className="card p-5">
      <div className="mb-4 flex items-center gap-2">
        <span className="h-2.5 w-2.5 animate-pulse2 rounded-full bg-emerald-500" />
        <h3 className="font-bold">Jonli · hozir</h3>
      </div>

      {items.length === 0 ? (
        <p className="py-6 text-center text-sm text-ink-3">Hali to'lovlar yo'q</p>
      ) : (
        <ul className="space-y-3">
          {items.map((d) => (
            <li key={d.id} className="flex items-center gap-3 animate-fade-up">
              <Avatar name={d.donor} size={36} />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm">
                  <span className="font-semibold">{d.donor}</span>{" "}
                  <span className="text-ink-3">→</span>{" "}
                  <span className="text-ink-2">{d.person_name}</span>
                </p>
                <p className="text-xs text-ink-3">{timeAgo(d.at)}</p>
              </div>
              <span className="shrink-0 text-sm font-bold text-accent">
                {compactSom(d.amount_som)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
