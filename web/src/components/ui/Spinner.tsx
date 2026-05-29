export default function Spinner({ label }: { label?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16 text-ink-3">
      <div className="h-8 w-8 animate-spin rounded-full border-[3px] border-line border-t-accent" />
      {label && <p className="text-sm">{label}</p>}
    </div>
  );
}
