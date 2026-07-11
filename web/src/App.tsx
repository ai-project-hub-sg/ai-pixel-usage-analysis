import { useEffect, useState } from "react";
import { api } from "./api/client";
import { AppShell, PageName } from "./components/AppShell";
import { LoginPage } from "./pages/LoginPage";

export function App() {
  const [username, setUsername] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState<PageName>("overview");
  useEffect(() => { api.session().then(s => setUsername(s.username)).catch(() => setUsername(null)).finally(() => setLoading(false)); }, []);
  if (loading) return <div className="boot-screen"><div className="brand-mark">AP</div><span>正在载入分析中心</span></div>;
  if (!username) return <LoginPage login={api.login} onSuccess={() => api.session().then(s => setUsername(s.username))} />;
  return <AppShell page={page} setPage={setPage} username={username} logout={() => api.logout().finally(() => setUsername(null))}>
    <div className="page-placeholder"><h1>{page === "overview" ? "用量总览" : page === "usage" ? "请求分析" : page === "ledger" ? "余额流水" : "账户状态"}</h1><p>数据正在连接。</p></div>
  </AppShell>;
}
