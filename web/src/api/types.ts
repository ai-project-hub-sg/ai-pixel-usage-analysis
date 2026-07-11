export type Session = { username: string };
export type Comparison = { requests: number; actual_cost: number; Credit: number; Debit: number; Net: number };
export type Bucket = {
  start: string; requests: number; InputTokens: number; OutputTokens: number;
  actual_cost: number; Credit: number; Debit: number; Net: number;
  yesterday: Comparison | null; last_week: Comparison | null;
};
export type Overview = { buckets: Bucket[] };
