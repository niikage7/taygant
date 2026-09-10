import { redirect } from "next/navigation";

/**
 * Корень приложения — точки входа как таковой нет: рабочее пространство
 * начинается с обзора проекта. Неавторизованного пользователя оболочка
 * `(app)/layout.tsx` сама отправит на `/login`, поэтому проверять токен здесь
 * (он всё равно лежит в localStorage и на сервере недоступен) не нужно.
 */
export default function RootPage() {
  redirect("/overview");
}
