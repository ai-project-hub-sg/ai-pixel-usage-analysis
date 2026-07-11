import type { Overview, Session } from "./types";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body?.error?.message ?? "请求失败");
  return body.data as T;
}

export const api = {
  session: () => request<Session>("/api/auth/session"),
  login: (username: string, password: string) =>
    request("/api/auth/login", { method: "POST", body: JSON.stringify({ username, password }) }),
  logout: () => request("/api/auth/logout", { method: "POST", headers: { Origin: window.location.origin } }),
  overview: (query: URLSearchParams) => request<Overview>(`/api/overview?${query}`)
};
