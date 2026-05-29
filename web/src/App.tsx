import { Routes, Route } from "react-router-dom";
import { useState, useCallback } from "react";
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import DonateModal from "./components/DonateModal";
import Toast from "./components/ui/Toast";
import HomePage from "./pages/HomePage";
import PersonPage from "./pages/PersonPage";
import MyDonationsPage from "./pages/MyDonationsPage";
import PaymentResultPage from "./pages/PaymentResultPage";
import { DonateContext, ToastContext } from "./context";
import type { Person } from "./lib/types";

export default function App() {
  // Yordam berish oynasi (modal) butun ilova bo'ylab ishlaydi
  const [donateTarget, setDonateTarget] = useState<Person | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const openDonate = useCallback((p: Person) => setDonateTarget(p), []);
  const showToast = useCallback((msg: string) => {
    setToast(msg);
    window.setTimeout(() => setToast(null), 3500);
  }, []);

  return (
    <ToastContext.Provider value={showToast}>
      <DonateContext.Provider value={openDonate}>
        <div className="flex min-h-screen flex-col">
          <Navbar />
          <main className="flex-1">
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/loyiha/:id" element={<PersonPage />} />
              <Route path="/mening-yordamlarim" element={<MyDonationsPage />} />
              <Route path="/tolov/natija" element={<PaymentResultPage />} />
              <Route path="*" element={<HomePage />} />
            </Routes>
          </main>
          <Footer />

          {donateTarget && (
            <DonateModal person={donateTarget} onClose={() => setDonateTarget(null)} />
          )}
          {toast && <Toast message={toast} />}
        </div>
      </DonateContext.Provider>
    </ToastContext.Provider>
  );
}
