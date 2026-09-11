import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Доска задач",
};

export default function BoardLayout({ children }: { children: ReactNode }) {
  return children;
}
