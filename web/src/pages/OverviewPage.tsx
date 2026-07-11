import { useCallback, useEffect, useMemo, useState } from "react";
import { api } from "../api/client";
import type { AccountStatus, Overview } from "../api/types";
import { AccountFilter, GranularityFilter } from "../components/GlobalFilters";
import { KpiCard } from "../components/KpiCard";
import { PageHeader } from "../components/PageHeader";
import { TrendChart } from "../components/TrendChart";

export type OverviewFilter = { accountID: string; granularity: "hour" | "day" };
type Loader = (filter: OverviewFilter, signal?: AbortSignal) => Promise<Overview>;
const defaultLoad: Loader = (filter, signal) => {
  const end = new Date(), start = new Date(end.getTime() - 8 * 864e5);
  const query = new URLSearchParams({ start: start.toISOString(), end: end.toISOString(), granularity: filter.granularity });
  if (filter.accountID) query.set("account_id", filter.accountID);
  return api.overview(query, signal);
};
function initialFilter(): OverviewFilter {
  const query = new URLSearchParams(window.location.search);
  return { accountID: query.get("account_id") ?? "", granularity: query.get("granularity") === "day" ? "day" : "hour" };
}

export function OverviewPage({ accounts = [], load = defaultLoad }: { accounts?: AccountStatus[]; load?: Loader }) {
  const [filter, setFilter] = useState(initialFilter);
  const [data, setData] = useState<Overview | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [yesterday, setYesterday] = useState(true), [lastWeek, setLastWeek] = useState(true);
  const run = useCallback((signal?: AbortSignal) => {
    setLoading(true); setError("");
    return load(filter, signal).then(setData).catch((reason) => { if (reason?.name !== "AbortError") setError("数据加载失败，请重试"); }).finally(() => setLoading(false));
  }, [filter, load]);
  useEffect(() => {
    const controller = new AbortController(), query = new URLSearchParams(window.location.search);
    filter.accountID ? query.set("account_id", filter.accountID) : query.delete("account_id");
    query.set("granularity", filter.granularity);
    window.history.replaceState(null, "", `${window.location.pathname}?${query}`);
    void run(controller.signal);
    return () => controller.abort();
  }, [filter, run]);
  const accountName = useMemo(() => accounts.find((account) => account.ID === filter.accountID)?.Name, [accounts, filter.accountID]);
  const current = data?.buckets.at(-1);
  return <section>
    <PageHeader title="用量总览" description={accountName ? `${accountName}视图` : "全部账户汇总视图"} />
    <div className="global-filters"><AccountFilter accounts={accounts} value={filter.accountID} onChange={(accountID) => setFilter((value) => ({ ...value, accountID }))} /><GranularityFilter value={filter.granularity} onChange={(granularity) => setFilter((value) => ({ ...value, granularity }))} /></div>
    {error ? <div className="notice error">{error}<button onClick={() => void run()}>重试</button></div> : null}
    <div className="kpi-row"><KpiCard label={filter.granularity === "day" ? "当前日请求" : "当前小时请求"} value={String(current?.requests ?? 0)} /><KpiCard label="实际成本" value={`$${(current?.actual_cost ?? 0).toFixed(4)}`} /><KpiCard label="余额入账" value={`+${(current?.credit ?? 0).toFixed(2)}`} tone="green" /><KpiCard label="余额出账" value={`-${(current?.debit ?? 0).toFixed(2)}`} tone="red" /></div>
    <div className="data-section"><div className="section-title"><div><h2>{filter.granularity === "day" ? "每日请求趋势" : "每小时请求趋势"}</h2><p>当前时段与历史同期</p></div><div className="compare-controls"><label><input type="checkbox" checked={yesterday} onChange={(event) => setYesterday(event.target.checked)} />昨天同期</label><label><input type="checkbox" checked={lastWeek} onChange={(event) => setLastWeek(event.target.checked)} />上周同期</label></div></div>{loading && !data ? <div className="loading-block">正在聚合数据…</div> : data ? <TrendChart buckets={data.buckets} yesterday={yesterday} lastWeek={lastWeek} /> : null}</div>
  </section>;
}
