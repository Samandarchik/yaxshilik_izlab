export default function ProgressBar({ value }: { value: number }) {
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-line">
      <div
        className="bar-fill h-full rounded-full bg-gradient-to-r from-accent to-[#FF9D6F]"
        style={{ width: `${Math.max(2, value)}%` }}
      />
    </div>
  );
}
