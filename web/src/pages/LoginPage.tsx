import { FormEvent, useState } from "react";

type Props = {
  login: (username: string, password: string) => Promise<unknown>;
  onSuccess: () => void;
};

export function LoginPage({ login, onSuccess }: Props) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    setBusy(true); setError("");
    try { await login(username, password); onSuccess(); }
    catch { setError("用户名或密码错误"); }
    finally { setBusy(false); }
  }

  return <div className="login-layout">
    <section className="login-story">
      <div className="brand-mark">AP</div>
      <h1>看清每一次<br />AI 调用的成本</h1>
      <p>多账户请求、余额流水与同期变化，在同一处持续更新。</p>
      <div className="story-lines" aria-hidden="true"><i /><i /><i /></div>
    </section>
    <main className="login-panel">
      <form className="login-form" onSubmit={submit}>
        <h2>登录分析中心</h2>
        <p>使用服务器 .env 中的管理凭据</p>
        <label>用户名<input autoComplete="username" value={username} onChange={e => setUsername(e.target.value)} required /></label>
        <label>密码<input type="password" autoComplete="current-password" value={password} onChange={e => setPassword(e.target.value)} required /></label>
        {error ? <div className="form-error" role="alert">{error}</div> : null}
        <button className="primary-button" disabled={busy}>{busy ? "正在登录…" : "登录"}</button>
      </form>
    </main>
  </div>;
}
