"use client";

import { useState, useEffect } from "react";
import axios from "axios";

// Define a type for the user data
interface User {
  id: number;
  name: string;
  username: string;
  email: string;
}

export default function Home() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Here we use the knowledge from context7 on how to use axios!
    axios
      .get<User[]>("https://jsonplaceholder.typicode.com/users")
      .then((response) => {
        setUsers(response.data);
        setLoading(false);
      })
      .catch((err) => {
        setError("Failed to fetch users.");
        setLoading(false);
        console.error(err);
      });
  }, []); // The empty dependency array makes this effect run once on mount

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="w-full max-w-4xl rounded-lg bg-white p-8 shadow-md dark:bg-zinc-900">
        <h1 className="mb-6 text-center text-3xl font-semibold tracking-tight text-black dark:text-zinc-50">
          Users from JSONPlaceholder
        </h1>
        <p className="mb-6 text-center text-lg text-zinc-600 dark:text-zinc-400">
          This data was fetched using{" "}
          <code className="rounded bg-zinc-200 px-1 font-mono text-sm dark:bg-zinc-800">
            axios
          </code>
          . We learned how to use it with{" "}
          <code className="rounded bg-zinc-200 px-1 font-mono text-sm dark:bg-zinc-800">
            context7
          </code>
          !
        </p>

        {loading && <p className="text-center text-zinc-500 dark:text-zinc-400">Loading...</p>}
        {error && <p className="text-center text-red-500">{error}</p>}

        {!loading && !error && (
          <ul className="divide-y divide-zinc-200 dark:divide-zinc-700">
            {users.map((user) => (
              <li key={user.id} className="py-4">
                <p className="font-semibold text-zinc-900 dark:text-zinc-100">
                  {user.name} (@{user.username})
                </p>
                <p className="text-sm text-zinc-600 dark:text-zinc-400">
                  {user.email}
                </p>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  );
}
