export default function Toast({ message }: { message: string }) {
  return (
    <div className="fixed bottom-6 left-1/2 z-[100] -translate-x-1/2 animate-fade-up">
      <div className="flex items-center gap-3 rounded-2xl bg-ink px-5 py-3.5 text-white shadow-card">
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-500 text-sm">
          ✓
        </span>
        <span className="text-sm font-medium">{message}</span>
      </div>
    </div>
  );
}
