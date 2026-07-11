import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type { AccountStatus } from "../api/types";
import { PageHeader } from "../components/PageHeader";

export function AccountsPage({ onChange }: { onChange?: (rows: AccountStatus[]) => void }) {
  const [rows, setRows] = useState<AccountStatus[] | null>(null), [busy, setBusy] = useState(""), [message, setMessage] = useState(""), [error, setError] = useState("");
  const load = useCallback(() => api.accounts().then((statuses) => { setRows(statuses); onChange?.(statuses); }), [onChange]);
  useEffect(() => { void load().catch(() => setError("账户状态加载失败")); }, [load]);
  async function sync(id: string) { setBusy(id); setMessage(""); setError(""); try { await api.sync(id); await load(); setMessage("同步已完成"); } catch { setError("同步失败，请稍后重试"); } finally { setBusy(""); } }
  return <section><PageHeader title="账户状态" description="上游节点、同步新鲜度与错误状态" />{message ? <div className="notice success" role="status">{message}</div> : null}{error ? <div className="notice error" role="alert">{error}</div> : null}{rows === null ? <div className="loading-block">正在加载账户状态…</div> : <div className="account-list">{rows.map((row) => <article key={row.ID}><div className={`health-dot ${row.LastError ? "bad" : "good"}`} /><div><h3>{row.Name}</h3><p>{row.ID} · {row.CurrentHost || "等待首次连接"}</p></div><dl><div><dt>最近同步</dt><dd>{row.LastSyncAt ? new Date(row.LastSyncAt).toLocaleString() : "尚未同步"}</dd></div><div><dt>状态</dt><dd>{row.LastError || "正常"}</dd></div></dl><button onClick={() => void sync(row.ID)} disabled={busy === row.ID}>{busy === row.ID ? "同步中…" : "立即同步"}</button></article>)}</div>}</section>;
}
