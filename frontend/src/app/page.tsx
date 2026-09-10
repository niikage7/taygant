"use client";

import Image from "next/image";
import { useState } from "react";

export default function Home() {
  const [response, setResponse] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handlePing() {
    setLoading(true);
    setError(null);
    setResponse(null);
    try {
      const res = await fetch("http://localhost:4000/ping");
      if (!res.ok) {
        throw new Error(`Request failed with status ${res.status}`);
      }
      const text = await res.text();
      setResponse(text);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col flex-1 items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="flex flex-1 w-full max-w-3xl flex-col items-center justify-between py-32 px-16 bg-white dark:bg-black sm:items-start">
        <Image
          className="dark:invert h-5 w-[100px]"
          src="/next.svg"
          alt="Next.js logo"
          width={100}
          height={20}
          priority
        />
        <div className="flex flex-col items-center gap-6 text-center sm:items-start sm:text-left">
          <h1 className="max-w-xs text-3xl font-semibold leading-10 tracking-tight text-black dark:text-zinc-50">
            Backend ping check
          </h1>
          <p className="max-w-md text-lg leading-8 text-zinc-600 dark:text-zinc-400">
            Send a request to{" "}
            <code className="rounded bg-black/[.06] px-1.5 py-0.5 font-mono text-[0.9em] dark:bg-white/[.08]">
              http://localhost:4000/ping
            </code>{" "}
            and see the response below.
          </p>
          <button
            type="button"
            onClick={handlePing}
            disabled={loading}
            className="flex h-12 items-center justify-center gap-2 rounded-full bg-foreground px-6 text-background transition-colors hover:bg-[#383838] disabled:opacity-60 dark:hover:bg-[#ccc]"
          >
            {loading ? "Pinging..." : "Ping backend"}
          </button>
          {response && (
            <p className="rounded bg-black/[.06] px-3 py-2 font-mono text-sm text-black dark:bg-white/[.08] dark:text-zinc-50">
              Response: {response}
            </p>
          )}
          {error && (
            <p className="rounded bg-red-100 px-3 py-2 font-mono text-sm text-red-700 dark:bg-red-950 dark:text-red-300">
              Error: {error}
            </p>
          )}
        </div>
        <div />
      </main>
    </div>
  );
}
