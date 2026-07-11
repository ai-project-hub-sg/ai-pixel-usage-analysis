import { type FormEvent, useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type { AccountStatus, LedgerRow } from "../api/types";
import { AccountFilter } from "../components/GlobalFilters";
import { PageHeader } from "../components/PageHeader";

export type LedgerFilter = { accountID: string; direction: string; reason: string; category: string; remark: string; refType: string; refID: string };
type Loader = (filter: LedgerFilter, signal?: AbortSignal) => Promise<LedgerRow[]>;
const defaultLoad: Loader = (filter, signal) => {
  const end = new Date(), start = new Date(end.getTime() - 31 * 864e5), query = new URLSearchParams({ start: start.toISOString(), end: end.toISOString() });
  const keys: Record<keyof LedgerFilter, string> = { accountID: "account_id", direction: "direction", reason: "reason", category: "category", remark: "remark", refType: "ref_type", refID: "ref_id" };
  (Object.keys(keys) as Array<keyof LedgerFilter>).forEach((key) => filter[key] && query.set(keys[key], filter[key]));
  return api.ledger(query, signal);
};

export function LedgerPage({ accounts = [], load = defaultLoad }: { accounts?: AccountStatus[]; load?: Loader }) {
  const [filter, setFilter] = useState<LedgerFilter>({ accountID: "", direction: "", reason: "", category: "", remark: "", refType: "", refID: "" });
  const [rows, setRows] = useState<LedgerRow[] | null>(null), [expanded, setExpanded] = useState<number | null>(null), [error, setError] = useState("");
  const run = useCallback((signal?: AbortSignal) => { setError(""); return load(filter, signal).then(setRows).catch((reason) => { if (reason?.name !== "AbortError") setError("流水加载失败，请重试"); }); }, [filter, load]);
  useEffect(() => { const controller = new AbortController(); void run(controller.signal); return () => controller.abort(); }, []);
  function submit(event: FormEvent) { event.preventDefault(); void run(); }
  return <section><PageHeader title="余额流水" description="入账、出账、原始类型、业务分类与备注内容分析" />
    <form className="filter-bar ledger-filters" onSubmit={submit}><AccountFilter accounts={accounts} value={filter.accountID} onChange={(accountID) => setFilter((value) => ({ ...value, accountID }))} /><label>流水方向<select value={filter.direction} onChange={(event) => setFilter({ ...filter, direction: event.target.value })}><option value="">全部</option><option value="credit">入账</option><option value="debit">出账</option></select></label><label>原始流水类型<input value={filter.reason} onChange={(event) => setFilter({ ...filter, reason: event.target.value })} placeholder="usage_charge" /></label><label>业务分类<input value={filter.category} onChange={(event) => setFilter({ ...filter, category: event.target.value })} placeholder="usage" /></label><label>备注关键词<input value={filter.remark} onChange={(event) => setFilter({ ...filter, remark: event.target.value })} placeholder="请求 ID、群组…" /></label><label>引用类型<input value={filter.refType} onChange={(event) => setFilter({ ...filter, refType: event.target.value })} placeholder="request" /></label><label>引用 ID<input inputMode="numeric" value={filter.refID} onChange={(event) => setFilter({ ...filter, refID: event.target.value })} placeholder="9001" /></label><button>应用筛选</button></form>
    {error ? <div className="notice error">{error}<button onClick={() => void run()}>重试</button></div> : null}
    {rows === null ? <div className="loading-block">正在加载余额流水…</div> : rows.length === 0 ? <div className="empty-state">暂无匹配流水</div> : <div className="table-wrap"><table><thead><tr><th>时间</th><th>账户</th><th>类型</th><th>备注</th><th>方向</th><th>金额</th><th>余额</th><th>详情</th></tr></thead><tbody>{rows.flatMap((row) => {
      const base = <tr key={`row-${row.id}`}><td>{new Date(row.CreatedAt).toLocaleString()}</td><td>{row.AccountID}</td><td><code>{row.Reason}</code><small>{row.Category}</small></td><td>{row.Remark}</td><td><span className={`direction ${row.Direction}`}>{row.Direction === "credit" ? "入账" : "出账"}</span></td><td>{row.Direction === "credit" ? "+" : "-"}{row.Amount}</td><td>{row.BalanceAfter}</td><td><button className="details-button" aria-expanded={expanded === row.id} onClick={() => setExpanded(expanded === row.id ? null : row.id)}>查看备注详情</button></td></tr>;
      return expanded === row.id ? [base, <tr className="metadata-row" key={`metadata-${row.id}`}><td colSpan={8}><strong>{row.RefType || "无引用"}{row.RefID ? ` #${row.RefID}` : ""}</strong><pre>{JSON.stringify(row.Metadata, null, 2)}</pre></td></tr>] : [base];
    })}</tbody></table></div>}
  </section>;
}
