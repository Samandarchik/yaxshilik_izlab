import { useEffect } from "react";
import { api } from "../lib/api";
import { useApi } from "../hooks/useApi";
import Hero from "../components/Hero";
import SuccessStories from "../components/SuccessStories";
import PeopleSection from "../components/PeopleSection";
import LiveFeed from "../components/LiveFeed";
import AlmostThere from "../components/AlmostThere";
import MyImpactCard from "../components/MyImpactCard";
import HowItWorks from "../components/HowItWorks";
import Faq from "../components/Faq";
import CtaBanner from "../components/CtaBanner";
import Spinner from "../components/ui/Spinner";

export default function HomePage() {
  const { data: people, loading, error } = useApi(() => api.people(), []);
  const { data: stats } = useApi(() => api.stats(), []);

  // "/#bo'lim" havolasi bilan kelinganda o'sha joyga aylantiramiz
  useEffect(() => {
    if (window.location.hash) {
      const el = document.querySelector(window.location.hash);
      if (el) setTimeout(() => el.scrollIntoView({ behavior: "smooth" }), 200);
    }
  }, [people]);

  const featured = people?.find((p) => p.urgent) ?? people?.[0] ?? null;

  return (
    <>
      <Hero featured={featured} stats={stats} />

      <SuccessStories />

      {loading && <Spinner label="Yuklanmoqda..." />}
      {error && (
        <p className="py-16 text-center text-red-500">Ma'lumotlarni yuklab bo'lmadi: {error}</p>
      )}

      {people && <PeopleSection people={people} />}

      {/* Jamoa: mening ta'sirim + jonli oqim + oz qolganlar */}
      {people && (
        <section className="mx-auto grid max-w-6xl gap-5 px-4 pb-6 sm:px-6 md:grid-cols-3">
          <MyImpactCard />
          <LiveFeed />
          <AlmostThere people={people} />
        </section>
      )}

      <HowItWorks />
      <Faq />
      <CtaBanner />
    </>
  );
}
