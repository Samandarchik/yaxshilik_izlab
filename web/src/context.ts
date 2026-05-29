import { createContext, useContext } from "react";
import type { Person } from "./lib/types";

// "Yordam ber" oynasini ochish funksiyasi (butun ilovada mavjud)
export const DonateContext = createContext<(p: Person) => void>(() => {});
export const useDonate = () => useContext(DonateContext);

// Qisqa xabar (toast) ko'rsatish
export const ToastContext = createContext<(msg: string) => void>(() => {});
export const useToast = () => useContext(ToastContext);
