"use client";

import { useQuery } from "@tanstack/react-query";

import { PageError, PageLoading } from "@/components/app/page-state";
import { NewProjectForm } from "@/components/project/new-project-form";
import { usersService } from "@/services";

export default function NewProjectPage() {
  const users = useQuery({ queryKey: ["users"], queryFn: () => usersService.list() });
  const me = useQuery({ queryKey: ["users", "me"], queryFn: () => usersService.getMe() });

  if (users.error) return <PageError error={users.error} />;
  if (!users.data) return <PageLoading label="Загружаем список пользователей…" />;

  return (
    <div className="mx-auto max-w-[1200px]">
      <NewProjectForm users={users.data} currentUser={me.data ?? null} />
    </div>
  );
}
