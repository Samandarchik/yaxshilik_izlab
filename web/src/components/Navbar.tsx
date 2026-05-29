import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { getTgUser } from "../lib/telegram";
import Avatar from "./ui/Avatar";
import { BrandLockup } from "./ui/Brand";

const links = [
  { label: "Kashf et", href: "/" },
  { label: "Shoshilinch", href: "/#shoshilinch" },
  { label: "Hikoyalar", href: "/#hikoyalar" },
  { label: "Qanday ishlaydi", href: "/#qanday" },
];

function Logo() {
  return (
    <Link to="/">
      <BrandLockup size={38} />
    </Link>
  );
}

export default function Navbar() {
  const [open, setOpen] = useState(false);
  const loc = useLocation();
  const tg = getTgUser();
  const tgName = [tg.first_name, tg.last_name].filter(Boolean).join(" ");

  return (
    <header className="sticky top-0 z-50 border-b border-line bg-white/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
        <Logo />

        {/* Desktop menyu */}
        <nav className="hidden items-center gap-1 md:flex">
          {links.map((l) => (
            <a
              key={l.label}
              href={l.href}
              className="rounded-lg px-3.5 py-2 text-sm font-semibold text-ink-2 transition hover:bg-page hover:text-ink"
            >
              {l.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          {/* Telegram foydalanuvchisi — login o'rniga */}
          <Link
            to="/mening-yordamlarim"
            className={`flex items-center gap-2 rounded-full border px-2 py-1 pr-3 transition ${
              loc.pathname === "/mening-yordamlarim"
                ? "border-accent/40 bg-accent-soft"
                : "border-line bg-white hover:border-accent/40"
            }`}
            title="Mening yordamlarim"
          >
            <Avatar name={tgName} src={tg.photo_url} size={28} />
            <span className="hidden max-w-[110px] truncate text-sm font-semibold sm:block">
              {tg.first_name}
            </span>
          </Link>
          <a href="/#shoshilinch" className="hidden btn-primary !px-4 !py-2.5 text-sm sm:inline-flex">
            Yordam ber
          </a>

          {/* Mobil menyu tugmasi */}
          <button
            onClick={() => setOpen((v) => !v)}
            className="flex h-10 w-10 items-center justify-center rounded-lg border border-line md:hidden"
            aria-label="Menyu"
          >
            <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2">
              {open ? <path d="M6 6l12 12M18 6L6 18" /> : <path d="M4 7h16M4 12h16M4 17h16" />}
            </svg>
          </button>
        </div>
      </div>

      {/* Mobil ochiladigan menyu */}
      {open && (
        <div className="border-t border-line bg-white px-4 py-3 md:hidden">
          {links.map((l) => (
            <a
              key={l.label}
              href={l.href}
              onClick={() => setOpen(false)}
              className="block rounded-lg px-3 py-2.5 font-semibold text-ink-2 hover:bg-page"
            >
              {l.label}
            </a>
          ))}
          <Link
            to="/mening-yordamlarim"
            onClick={() => setOpen(false)}
            className="block rounded-lg px-3 py-2.5 font-semibold text-ink-2 hover:bg-page"
          >
            Mening yordamlarim
          </Link>
          <a
            href="/#shoshilinch"
            onClick={() => setOpen(false)}
            className="mt-2 btn-primary w-full"
          >
            Yordam ber
          </a>
        </div>
      )}
    </header>
  );
}
