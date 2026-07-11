import { createElement } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import type { AccountStatus, LedgerRow } from "../api/types";
import { LedgerPage } from "./LedgerPage";
import { OverviewPage } from "./OverviewPage";
import { UsagePage } from "./UsagePage";

const accounts: AccountStatus[] = [
  { ID: "primary", Name: "主账户", Enabled: true, CurrentHost: "", LastSyncAt: "", LastError: "" },
  { ID: "secondary", Name: "备用账户", Enabled: true, CurrentHost: "", LastSyncAt: "", LastError: "timeout" }
];

test("overview switches account and exposes independent comparison toggles", async () => {
  const load = vi.fn().mockResolvedValue({ buckets: [] });
  render(createElement(OverviewPage as never, { accounts, load }));
  expect(await screen.findByText("暂无该时段数据")).toBeInTheDocument();
  await userEvent.selectOptions(screen.getByLabelText("账户"), "secondary");
  await userEvent.selectOptions(screen.getByLabelText("聚合粒度"), "day");
  expect(load.mock.lastCall?.[0]).toEqual(expect.objectContaining({ accountID: "secondary", granularity: "day" }));
  const yesterday = screen.getByRole("checkbox", { name: "昨天同期" });
  const lastWeek = screen.getByRole("checkbox", { name: "上周同期" });
  expect(yesterday).toBeChecked();
  expect(lastWeek).toBeChecked();
  await userEvent.click(yesterday);
  expect(yesterday).not.toBeChecked();
  expect(lastWeek).toBeChecked();
});

test("usage sends account model endpoint and api key dimensions", async () => {
  const load = vi.fn().mockResolvedValue([]);
  render(createElement(UsagePage as never, { accounts, load }));
  await screen.findByText("暂无请求记录");
  await userEvent.selectOptions(screen.getByLabelText("账户"), "primary");
  await userEvent.type(screen.getByLabelText("模型"), "gpt-4.1");
  await userEvent.type(screen.getByLabelText("端点"), "/v1/responses");
  await userEvent.type(screen.getByLabelText("API Key ID"), "101");
  await userEvent.click(screen.getByRole("button", { name: "应用筛选" }));
  expect(load.mock.lastCall?.[0]).toEqual(expect.objectContaining({
    accountID: "primary",
    model: "gpt-4.1",
    endpoint: "/v1/responses",
    apiKeyID: "101"
  }));
});

test("ledger sends all dimensions and expands normalized remark metadata", async () => {
  const row: LedgerRow = {
    id: 1, AccountID: "primary", Direction: "debit", Amount: "1.25", BalanceAfter: "8.75",
    Reason: "usage_charge", Category: "usage", Remark: "请求 req-001 的模型调用扣费",
    RefType: "request", RefID: 9001, Metadata: { request_id: "req-001", model: "gpt-4.1" },
    CreatedAt: "2026-07-11T04:00:00Z"
  };
  const load = vi.fn().mockResolvedValue([row]);
  render(createElement(LedgerPage as never, { accounts, load }));
  await screen.findByText(row.Remark);
  await userEvent.selectOptions(screen.getByLabelText("账户"), "primary");
  await userEvent.selectOptions(screen.getByLabelText("流水方向"), "debit");
  await userEvent.type(screen.getByLabelText("原始流水类型"), "usage_charge");
  await userEvent.type(screen.getByLabelText("业务分类"), "usage");
  await userEvent.type(screen.getByLabelText("备注关键词"), "req-001");
  await userEvent.type(screen.getByLabelText("引用类型"), "request");
  await userEvent.type(screen.getByLabelText("引用 ID"), "9001");
  await userEvent.click(screen.getByRole("button", { name: "应用筛选" }));
  expect(load.mock.lastCall?.[0]).toEqual(expect.objectContaining({
    accountID: "primary", direction: "debit", reason: "usage_charge", category: "usage",
    remark: "req-001", refType: "request", refID: "9001"
  }));
  await userEvent.click(screen.getByRole("button", { name: "查看备注详情" }));
  expect(screen.getByText(/"request_id": "req-001"/)).toBeVisible();
});
