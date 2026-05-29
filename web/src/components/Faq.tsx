import { useState } from "react";
import Section from "./ui/Section";

const faqs = [
  {
    q: "Pulim haqiqatdan bemorga yetib boradimi?",
    a: "Ha. To'lovlar Click va Payme orqali bevosita o'tadi va yig'ilgan summa to'liq bemor uchun yo'naltiriladi. Jarayon shaffof — yig'ilayotgan summani va natijani kuzatib borishingiz mumkin.",
  },
  {
    q: "Bemor hikoyalari qanday tekshiriladi?",
    a: "Har bir so'rov tibbiy muassasa hujjatlari asosida Yaxshilik Izlab jamoasi tomonidan tekshiriladi va tasdiqlanadi. Tasdiqlangan so'rovlarda ✓ belgisi bo'ladi.",
  },
  {
    q: "Eng kichik yordam miqdori qancha?",
    a: "Eng kam summa — 1 000 so'm. Istalgan miqdorda yordam berishingiz mumkin.",
  },
  {
    q: "Soliq imtiyozi olish mumkinmi?",
    a: "Yuridik shaxslar uchun xayriya to'lovlari bo'yicha imtiyozlar amal qilishi mumkin. Batafsil ma'lumot uchun biz bilan bog'laning.",
  },
  {
    q: "Anonim yordam berishim mumkinmi?",
    a: "Albatta. Yordam berishda 'Anonim' belgisini tanlasangiz, ismingiz ko'rsatilmaydi.",
  },
];

export default function Faq() {
  const [open, setOpen] = useState<number | null>(0);

  return (
    <div className="bg-white">
      <Section title="Tez-tez beriladigan" accentWord="savollar">
        <div className="mx-auto max-w-3xl space-y-3">
          {faqs.map((f, i) => {
            const isOpen = open === i;
            return (
              <div
                key={i}
                className={`card overflow-hidden transition ${isOpen ? "ring-1 ring-accent/30" : ""}`}
              >
                <button
                  onClick={() => setOpen(isOpen ? null : i)}
                  className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left"
                >
                  <span className="font-semibold">{f.q}</span>
                  <span
                    className={`shrink-0 text-2xl text-accent transition-transform ${isOpen ? "rotate-45" : ""}`}
                  >
                    +
                  </span>
                </button>
                {isOpen && (
                  <p className="px-5 pb-5 text-ink-2 leading-relaxed animate-fade-up">{f.a}</p>
                )}
              </div>
            );
          })}
        </div>
      </Section>
    </div>
  );
}
