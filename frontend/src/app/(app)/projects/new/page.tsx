import type { Metadata } from "next";

import { NewProjectForm } from "@/components/project/new-project-form";
import { demoMembers } from "@/data/demo";

export const metadata: Metadata = {
  title: "Создание проекта",
  description: "Мастер инициации проекта: параметры, команда и настройки диаграммы Ганта",
};

export default function NewProjectPage() {
  const [owner, ...members] = demoMembers;

  return (
    <div className="mx-auto max-w-[1200px]">
      <NewProjectForm owner={owner} members={members} />
    </div>
  );
}
