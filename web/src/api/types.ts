export type Session = { username: string };
export type Comparison = { requests: number; actual_cost: number; credit: number; debit: number; net: number };
export type Bucket = {
  start: string; requests: number; input_tokens: number; output_tokens: number;
  actual_cost: number; credit: number; debit: number; net: number;
  yesterday: Comparison | null; last_week: Comparison | null;
};
export type Overview = { buckets: Bucket[] };
export type UsageRow = { id:number; AccountID:string; Model:string; Endpoint:string; APIKeyID:number; InputTokens:number; OutputTokens:number; ActualCost:string; CreatedAt:string };
export type LedgerRow = { id:number; AccountID:string; Direction:"credit"|"debit"; Amount:string; BalanceAfter:string; Reason:string; Category:string; Remark:string; RefType:string; RefID:number; Metadata:Record<string,unknown>; CreatedAt:string };
export type AccountStatus = { ID:string; Name:string; Enabled:boolean; CurrentHost:string; LastSyncAt:string; LastError:string };
