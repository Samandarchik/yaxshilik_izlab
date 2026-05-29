import { useEffect, useState } from "react";
import { api } from "../lib/api";
import type { SuccessStory } from "../lib/types";
import { compactSom } from "../lib/format";
import Section from "./ui/Section";

export default function SuccessStories() {
  const [items, setItems] = useState<SuccessStory[]>([]);

  useEffect(() => {
    api.successStories(8).then(setItems).catch(() => {});
  }, []);

  if (items.length === 0) return null;

  return (
    <div id="hikoyalar" className="bg-white">
      <Section
        title="Yopilgan"
        accentWord="hikoyalar"
        subtitle="Sizning yordamingiz bilan maqsadiga yetgan insonlar"
      >
        <div className="no-scrollbar -mx-4 flex gap-4 overflow-x-auto px-4 pb-2 sm:mx-0 sm:px-0">
          {items.map((s) => (
            <div
              key={s.id}
              className="card w-60 shrink-0 overflow-hidden"
            >
              <div className="relative aspect-square bg-page">
                {s.photo_url ? (
                  <img src={s.photo_url} alt={s.name} className="h-full w-full object-cover" />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-ink-3">📷</div>
                )}
                <span className="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500 text-white shadow-soft">
                  ✓
                </span>
              </div>
              <div className="p-4">
                <h3 className="font-bold">
                  {s.name} <span className="text-ink-3">· {s.age}</span>
                </h3>
                <p className="truncate text-sm text-ink-2">{s.diagnosis}</p>
                <p className="mt-2 text-sm font-semibold text-emerald-600">
                  {compactSom(s.raised)} so'm yig'ildi
                </p>
              </div>
            </div>
          ))}
        </div>
      </Section>
    </div>
  );
}
