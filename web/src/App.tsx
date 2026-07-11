import { useCallback, useEffect, useState } from "react";
import { api } from "./api/client";
import type { AccountStatus } from "./api/types";
import { AppShell, type PageName } from "./components/AppShell";
import { StatusNotice } from "./components/StatusNotice";
import { AccountsPage } from "./pages/AccountsPage";
import { LedgerPage } from "./pages/LedgerPage";
import { LoginPage } from "./pages/LoginPage";
import { OverviewPage } from "./pages/OverviewPage";
import { UsagePage } from "./pages/UsagePage";

export function App() {
  const [username, setUsername] = useState<string | null>(null);
  const [accounts, setAccounts] = useState<AccountStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState<PageName>("overview");
  const restore = useCallback(async () => {
    const [session, statuses] = await Promise.all([api.session(), api.accounts()]);
    setUsername(session.username);
    setAccounts(statuses);
  }, []);
  useEffect(() => { restore().catch(() => { setUsername(null); setAccounts([]); }).finally(() => setLoading(false)); }, [restore]);
  if (loading) return <div className="boot-screen"><div className="brand-mark">AP</div><span>正在载入分析中心</span></div>;
  if (!username) return <LoginPage login={api.login} onSuccess={restore} />;
  const partial = accounts.some((account) => account.LastError);
  return <AppShell page={page} setPage={setPage} username={username} logout={() => api.logout().finally(() => setUsername(null))}>
    {partial ? <StatusNotice>部分账户同步异常，汇总数据可能不完整。请到账户状态页查看详情。</StatusNotice> : null}
    {page === "overview" ? <OverviewPage accounts={accounts} /> : page === "usage" ? <UsagePage accounts={accounts} /> : page === "ledger" ? <LedgerPage accounts={accounts} /> : <AccountsPage onChange={setAccounts} />}
  </AppShell>;
}
