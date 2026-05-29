import type { ReactNode } from "react";

// Sarlavhali bo'lim — sarlavhada bir so'z serif kursiv bilan urg'ulanadi.
export default function Section({
  id,
  title,
  accentWord,
  subtitle,
  action,
  children,
}: {
  id?: string;
  title: string;
  accentWord?: string;
  subtitle?: string;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <section id={id} className="mx-auto w-full max-w-6xl px-4 py-10 sm:px-6 sm:py-14">
      <div className="mb-7 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
            {title}{" "}
            {accentWord && (
              <span className="font-display italic text-accent">{accentWord}</span>
            )}
          </h2>
          {subtitle && <p className="mt-2 text-ink-2">{subtitle}</p>}
        </div>
        {action}
      </div>
      {children}
    </section>
  );
}
