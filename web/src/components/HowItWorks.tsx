import Section from "./ui/Section";

const steps = [
  {
    n: "01",
    title: "Tasdiqlangan hikoyani tanlang",
    text: "Har bir so'rov klinika hujjatlari asosida Yaxshilik Izlab jamoasi tomonidan tekshiriladi. Ishonchli hikoyani tanlang.",
    icon: "🔎",
  },
  {
    n: "02",
    title: "Click yoki Payme orqali to'lang",
    text: "Bir bosishda, xavfsiz to'lov. Karta, UzCard, Humo yoki ilova balansi orqali.",
    icon: "💳",
  },
  {
    n: "03",
    title: "Natijani kuzating",
    text: "Pulingiz to'g'ridan-to'g'ri bemorga yetadi. Yig'ilgan summa va natijani ko'rib turasiz.",
    icon: "📈",
  },
];

export default function HowItWorks() {
  return (
    <Section id="qanday" title="Qanday" accentWord="ishlaydi?">
      <div className="grid gap-5 sm:grid-cols-3">
        {steps.map((s) => (
          <div key={s.n} className="card relative p-6">
            <span className="absolute right-5 top-4 font-display text-5xl italic text-line">
              {s.n}
            </span>
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-accent-soft text-2xl">
              {s.icon}
            </div>
            <h3 className="mt-4 text-lg font-bold">{s.title}</h3>
            <p className="mt-2 text-sm leading-relaxed text-ink-2">{s.text}</p>
          </div>
        ))}
      </div>
    </Section>
  );
}
