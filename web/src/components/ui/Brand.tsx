// "Yaxshilik Izlab" brend belgisi — logodagi kabi: qora doira ichida
// iz qoldirib ketayotgan odam. SVG bo'lgani uchun har qanday o'lchamda tiniq.

export function BrandMark({ size = 38 }: { size?: number }) {
  return (
    <span
      className="inline-flex shrink-0 items-center justify-center rounded-full bg-ink"
      style={{ width: size, height: size }}
      aria-hidden
    >
      <svg viewBox="0 0 64 64" style={{ width: size * 0.82, height: size * 0.82 }} fill="none">
        {/* tashqi halqa */}
        <circle cx="32" cy="32" r="29" stroke="white" strokeWidth="2.5" />
        {/* yuruvchi odam */}
        <g fill="white">
          <circle cx="34" cy="20" r="3.4" />
          <path
            d="M33 24c1.6 0 2.7 1.1 3 2.6l1.2 6 3.4 3.1c.8.7.9 1.9.2 2.7-.7.8-1.9.9-2.7.2l-3.5-3.2c-.4-.4-.7-.9-.8-1.4l-.4-2-2.1 5.1 2.2 7.6c.3 1-.3 2.1-1.3 2.4-1 .3-2.1-.3-2.4-1.3l-2.4-8.3c-.2-.6-.1-1.2.1-1.7l1.7-4.1-2.7 1.6-1.4 3.3c-.4 1-1.5 1.4-2.5 1-1-.4-1.4-1.5-1-2.5l1.6-3.8c.2-.5.6-.9 1.1-1.2l4.6-2.7c.7-.4 1.4-.9 2.3-.9z"
          />
        </g>
        {/* izlar */}
        <g fill="white" opacity="0.85">
          <ellipse cx="20" cy="46" rx="2" ry="3" transform="rotate(20 20 46)" />
          <ellipse cx="26" cy="49" rx="2" ry="3" transform="rotate(20 26 49)" />
          <ellipse cx="14" cy="49" rx="1.8" ry="2.6" transform="rotate(20 14 49)" />
        </g>
      </svg>
    </span>
  );
}

// To'liq brend bloki: belgi + nom + (ixtiyoriy) shior
export function BrandLockup({ size = 38, showSlogan = false }: { size?: number; showSlogan?: boolean }) {
  return (
    <span className="flex items-center gap-2.5">
      <BrandMark size={size} />
      <span className="flex flex-col leading-none">
        <span className="text-[17px] font-extrabold tracking-tight">
          Yaxshilik <span className="text-accent">Izlab</span>
        </span>
        {showSlogan && (
          <span className="mt-0.5 text-[10px] font-semibold uppercase tracking-wide text-ink-3">
            Bizni kutar, biz kutgan kunlar!
          </span>
        )}
      </span>
    </span>
  );
}
