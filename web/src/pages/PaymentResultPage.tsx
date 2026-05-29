import { Link, useSearchParams } from "react-router-dom";

// To'lovdan qaytgandan keyingi sahifa: /tolov/natija?paid=1
// To'lovning haqiqiy holati backend webhook orqali aniqlanadi va
// "Mening yordamlarim"da ko'rinadi — bu sahifa faqat xabar beradi.
export default function PaymentResultPage() {
  const [params] = useSearchParams();
  const paid = params.get("paid") === "1" || params.get("status") === "success";

  return (
    <div className="mx-auto flex max-w-lg flex-col items-center px-4 py-24 text-center">
      <div
        className={`flex h-20 w-20 items-center justify-center rounded-full text-4xl text-white ${
          paid ? "bg-emerald-500" : "bg-amber-500"
        }`}
      >
        {paid ? "✓" : "!"}
      </div>
      <h1 className="mt-6 text-3xl font-bold">
        {paid ? "Rahmat sizga!" : "To'lov yakunlanmadi"}
      </h1>
      <p className="mt-3 text-ink-2">
        {paid
          ? "Yordamingiz qabul qilindi. Sizning hissangiz kimningdir hayotini o'zgartiradi. To'lov tasdiqlangach, u \"Mening yordamlarim\"da ko'rinadi."
          : "To'lov amalga oshmadi yoki bekor qilindi. Qayta urinib ko'rishingiz mumkin."}
      </p>

      <div className="mt-8 flex flex-wrap justify-center gap-3">
        <Link to="/" className="btn-primary">
          Bosh sahifa
        </Link>
        <Link to="/mening-yordamlarim" className="btn-ghost">
          Mening yordamlarim
        </Link>
      </div>
    </div>
  );
}
