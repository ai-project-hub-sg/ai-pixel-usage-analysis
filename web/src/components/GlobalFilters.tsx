import type { AccountStatus } from "../api/types";

export function AccountFilter({ accounts, value, onChange }: { accounts: AccountStatus[]; value: string; onChange: (value: string) => void }) {
  return <label>账户<select value={value} onChange={(event) => onChange(event.target.value)}><option value="">全部账户</option>{accounts.filter((account) => account.Enabled).map((account) => <option key={account.ID} value={account.ID}>{account.Name}</option>)}</select></label>;
}

export function GranularityFilter({ value, onChange }: { value: "hour" | "day"; onChange: (value: "hour" | "day") => void }) {
  return <label>聚合粒度<select value={value} onChange={(event) => onChange(event.target.value as "hour" | "day")}><option value="hour">每小时</option><option value="day">每日</option></select></label>;
}
