import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Новый проект",
};

export default function NewProjectLayout({ children }: { children: ReactNode }) {
  return children;
}
