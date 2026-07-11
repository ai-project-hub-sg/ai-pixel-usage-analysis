import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { OverviewPage } from "./OverviewPage";
import { LedgerPage } from "./LedgerPage";

test("overview exposes independent comparison toggles", async () => {
  render(<OverviewPage load={vi.fn().mockResolvedValue({ buckets: [] })} />);
  expect(await screen.findByText("暂无该时段数据")).toBeInTheDocument();
  const yesterday = screen.getByRole("checkbox", { name: "昨天同期" });
  const lastWeek = screen.getByRole("checkbox", { name: "上周同期" });
  expect(yesterday).toBeChecked(); expect(lastWeek).toBeChecked();
  await userEvent.click(yesterday);
  expect(yesterday).not.toBeChecked(); expect(lastWeek).toBeChecked();
});

test("ledger sends raw type direction and remark filters", async () => {
  const load = vi.fn().mockResolvedValue([]);
  render(<LedgerPage load={load} />);
  await screen.findByText("暂无匹配流水");
  await userEvent.selectOptions(screen.getByLabelText("流水方向"), "debit");
  await userEvent.type(screen.getByLabelText("原始流水类型"), "usage_charge");
  await userEvent.type(screen.getByLabelText("备注关键词"), "request");
  await userEvent.click(screen.getByRole("button", { name: "应用筛选" }));
  expect(load).toHaveBeenLastCalledWith(expect.objectContaining({ direction: "debit", reason: "usage_charge", remark: "request" }));
});
