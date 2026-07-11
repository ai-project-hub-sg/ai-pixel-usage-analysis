import { type FormEvent, useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type { AccountStatus, UsageRow } from "../api/types";
import { AccountFilter } from "../components/GlobalFilters";
import { PageHeader } from "../components/PageHeader";

export type UsageFilter = { accountID: string; model: string; endpoint: string; apiKeyID: string };
type Loader = (filter: UsageFilter, signal?: AbortSignal) => Promise<UsageRow[]>;
const defaultLoad: Loader = (filter, signal) => {
  const end = new Date(), start = new Date(end.getTime() - 7 * 864e5), query = new URLSearchParams({ start: start.toISOString(), end: end.toISOString() });
  if (filter.accountID) query.set("account_id", filter.accountID);
  if (filter.model) query.set("model", filter.model);
  if (filter.endpoint) query.set("endpoint", filter.endpoint);
  if (filter.apiKeyID) query.set("api_key_id", filter.apiKeyID);
  return api.usage(query, signal);
};

export function UsagePage({ accounts = [], load = defaultLoad }: { accounts?: AccountStatus[]; load?: Loader }) {
  const [filter, setFilter] = useState<UsageFilter>({ accountID: "", model: "", endpoint: "", apiKeyID: "" });
  const [rows, setRows] = useState<UsageRow[] | null>(null), [error, setError] = useState("");
  const run = useCallback((signal?: AbortSignal) => { setError(""); return load(filter, signal).then(setRows).catch((reason) => { if (reason?.name !== "AbortError") setError("请求记录加载失败，请重试"); }); }, [filter, load]);
  useEffect(() => { const controller = new AbortController(); void run(controller.signal); return () => controller.abort(); }, []);
  function submit(event: FormEvent) { event.preventDefault(); const query = new URLSearchParams(); Object.entries(filter).forEach(([key, value]) => value && query.set(key, value)); window.history.replaceState(null, "", `${window.location.pathname}?${query}`); void run(); }
  return <section><PageHeader title="请求分析" description="按账户、模型、端点和 API Key 分析 Token、成本与延迟" />
    <form className="filter-bar" onSubmit={submit}><AccountFilter accounts={accounts} value={filter.accountID} onChange={(accountID) => setFilter((value) => ({ ...value, accountID }))} /><label>模型<input value={filter.model} onChange={(event) => setFilter({ ...filter, model: event.target.value })} placeholder="全部模型" /></label><label>端点<input value={filter.endpoint} onChange={(event) => setFilter({ ...filter, endpoint: event.target.value })} placeholder="/v1/responses" /></label><label>API Key ID<input inputMode="numeric" value={filter.apiKeyID} onChange={(event) => setFilter({ ...filter, apiKeyID: event.target.value })} placeholder="全部密钥" /></label><button>应用筛选</button></form>
    {error ? <div className="notice error">{error}<button onClick={() => void run()}>重试</button></div> : null}
    {rows === null ? <div className="loading-block">正在加载请求记录…</div> : rows.length === 0 ? <div className="empty-state">暂无请求记录</div> : <div className="table-wrap"><table><thead><tr><th>时间</th><th>账户</th><th>模型</th><th>端点</th><th>API Key</th><th>Token</th><th>实际成本</th></tr></thead><tbody>{rows.map((row) => <tr key={row.id}><td>{new Date(row.CreatedAt).toLocaleString()}</td><td>{row.AccountID}</td><td>{row.Model}</td><td><code>{row.Endpoint || "—"}</code></td><td>{row.APIKeyID || "—"}</td><td>{row.InputTokens + row.OutputTokens}</td><td>${row.ActualCost}</td></tr>)}</tbody></table></div>}
  </section>;
}
