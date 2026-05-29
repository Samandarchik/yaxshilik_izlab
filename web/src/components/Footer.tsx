import { BrandLockup } from "./ui/Brand";

const groups = [
  {
    title: "Platforma",
    items: ["Kashf et", "Shoshilinch so'rovlar", "Hikoyalar", "Doimiy obuna", "Hisobotlar"],
  },
  {
    title: "Bilish",
    items: ["Qanday ishlaydi", "Tekshiruv jarayoni", "Soliq imtiyozlari", "FAQ", "Blog"],
  },
  {
    title: "Aloqa",
    items: ["+998 71 200-00-00", "hello@yaxshilikizlab.uz", "Toshkent, O'zbekiston", "Press-kit"],
  },
];

export default function Footer() {
  return (
    <footer className="mt-10 border-t border-line bg-white">
      <div className="mx-auto grid max-w-6xl gap-10 px-4 py-12 sm:px-6 md:grid-cols-[1.4fr_1fr_1fr_1fr]">
        <div>
          <BrandLockup size={40} showSlogan />
          <p className="mt-4 max-w-xs text-sm leading-relaxed text-ink-2">
            Muhtojlarga bevosita, ishonchli va shaffof yordam platformasi. Har bir so'm to'g'ridan-to'g'ri bemorga yetadi.
          </p>
        </div>

        {groups.map((g) => (
          <div key={g.title}>
            <h4 className="mb-3 text-sm font-bold">{g.title}</h4>
            <ul className="space-y-2">
              {g.items.map((it) => (
                <li key={it}>
                  <span className="cursor-pointer text-sm text-ink-2 transition hover:text-accent">
                    {it}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="border-t border-line">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-5 text-xs text-ink-3 sm:px-6">
          <p>© {new Date().getFullYear()} Yaxshilik Izlab. Barcha huquqlar himoyalangan.</p>
          <div className="flex gap-4">
            <span className="cursor-pointer hover:text-ink-2">Ommaviy oferta</span>
            <span className="cursor-pointer hover:text-ink-2">Maxfiylik siyosati</span>
            <span className="cursor-pointer hover:text-ink-2">Cookie</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
