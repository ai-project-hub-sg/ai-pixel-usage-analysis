import type { ReactNode } from "react";

export type PageName = "overview" | "usage" | "ledger" | "accounts";
const items: Array<[PageName, string, string]> = [["overview","总览","⌁"],["usage","请求分析","↗"],["ledger","余额流水","⇄"],["accounts","账户状态","◎"]];

export function AppShell({ page, setPage, username, logout, children }: { page: PageName; setPage: (p: PageName) => void; username: string; logout: () => void; children: ReactNode }) {
  return <div className="app-layout">
    <aside className="sidebar">
      <div className="brand"><span className="brand-mark small">AP</span><strong>Pixel Lens</strong></div>
      <nav>{items.map(([id,label,icon]) => <button key={id} className={page===id?"active":""} onClick={() => setPage(id)}><span aria-hidden="true">{icon}</span>{label}</button>)}</nav>
      <div className="sidebar-user"><span>{username.slice(0,1).toUpperCase()}</span><div><strong>{username}</strong><button onClick={logout}>退出登录</button></div></div>
    </aside>
    <main className="workspace">{children}</main>
  </div>;
}
