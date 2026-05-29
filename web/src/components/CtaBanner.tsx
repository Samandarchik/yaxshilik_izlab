export default function CtaBanner() {
  return (
    <section className="mx-auto w-full max-w-6xl px-4 py-10 sm:px-6">
      <div className="card overflow-hidden bg-gradient-to-br from-accent to-[#FF8B5A] p-8 text-white sm:p-12">
        <h2 className="max-w-2xl text-2xl font-bold leading-snug sm:text-3xl">
          Oyiga 50 ming so'mdan boshlab — har oy{" "}
          <span className="font-display italic">birovning hayotini</span> o'zgartiring
        </h2>
        <div className="mt-6 flex flex-wrap gap-3">
          <a
            href="#shoshilinch"
            className="btn rounded-xl bg-white px-6 py-3 font-bold text-accent hover:bg-white/90"
          >
            Doimiy yordamchi bo'lish
          </a>
          <a
            href="#qanday"
            className="btn rounded-xl border border-white/40 px-6 py-3 font-bold text-white hover:bg-white/10"
          >
            Batafsil
          </a>
        </div>
      </div>
    </section>
  );
}
