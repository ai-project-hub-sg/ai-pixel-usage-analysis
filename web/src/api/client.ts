import type { AccountStatus, LedgerRow, Overview, Session, UsageRow } from "./types";

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
  overview: (query: URLSearchParams, signal?: AbortSignal) => request<Overview>(`/api/overview?${query}`, { signal })
  ,usage: (query: URLSearchParams, signal?: AbortSignal) => request<UsageRow[]>(`/api/usage/records?${query}`, { signal })
  ,ledger: (query: URLSearchParams, signal?: AbortSignal) => request<LedgerRow[]>(`/api/ledger/entries?${query}`, { signal })
  ,accounts: () => request<AccountStatus[]>("/api/accounts/status")
  ,sync: (id: string) => request(`/api/accounts/${encodeURIComponent(id)}/sync`, { method:"POST", headers:{ Origin:window.location.origin } })
};
