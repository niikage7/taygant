import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { LoaderCircle } from "lucide-react";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-control font-medium whitespace-nowrap transition-colors outline-none cursor-pointer focus-visible:focus-ring disabled:pointer-events-none disabled:opacity-60 disabled:cursor-not-allowed [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        primary: "bg-brand text-white hover:bg-brand-hover active:bg-brand-active",
        secondary: "bg-surface text-ink border border-line hover:bg-surface-subtle",
        soft: "bg-brand-tint text-brand hover:bg-brand-soft-alt",
        ghost: "text-ink-muted hover:bg-surface-muted hover:text-ink",
        danger: "bg-danger text-white hover:bg-danger/90",
        link: "text-brand underline underline-offset-2 hover:text-brand-hover",
      },
      size: {
        sm: "h-8 px-3 text-13 [&_svg]:size-4",
        md: "h-9 px-4 text-sm [&_svg]:size-4",
        lg: "h-10 px-5 text-sm [&_svg]:size-4",
        /** Квадратная кнопка-иконка (32px). Без видимого текста — обязателен `aria-label`. */
        icon: "size-8 [&_svg]:size-4",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "md",
    },
  },
);

type ButtonProps = ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    /** Отрисовать стили кнопки на дочернем элементе (например, на `<Link>`). */
    asChild?: boolean;
    /**
     * Состояние выполнения асинхронного действия: спиннер перед содержимым,
     * кнопка недоступна для повторного клика (`disabled` + `aria-busy`).
     * Не действует вместе с `asChild` — Slot требует ровно одного дочернего
     * элемента, вставить в него спиннер нельзя.
     */
    loading?: boolean;
  };

export function Button({
  className,
  variant,
  size,
  asChild = false,
  loading = false,
  disabled,
  children,
  ...props
}: ButtonProps) {
  if (asChild) {
    return (
      <Slot data-slot="button" className={cn(buttonVariants({ variant, size }), className)} {...props}>
        {children}
      </Slot>
    );
  }

  return (
    <button
      data-slot="button"
      className={cn(buttonVariants({ variant, size }), className)}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading ? (
        <LoaderCircle className="animate-spin motion-reduce:animate-none" aria-hidden />
      ) : null}
      {children}
    </button>
  );
}

export { buttonVariants };
export type { ButtonProps };
