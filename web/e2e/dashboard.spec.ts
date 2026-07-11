import os from "node:os";
import path from "node:path";
import { expect, test } from "@playwright/test";

const screenshots = path.join(os.tmpdir(), "ai-pixel-usage-analysis-screenshots");
const consoleIssues = new WeakMap<object, string[]>();

test.beforeEach(async ({ page }) => {
  const issues: string[] = [];
  consoleIssues.set(page, issues);
  page.on("console", (message) => {
    if (message.type() === "error" && message.text().includes("401 (Unauthorized)")) return;
    if (["error", "warning"].includes(message.type())) issues.push(`${message.type()}: ${message.text()}`);
  });
  page.on("pageerror", (error) => issues.push(`pageerror: ${error.message}`));
  await page.addInitScript(() => {
    const fixed = new Date("2026-07-11T07:00:00Z").valueOf();
    const NativeDate = Date;
    class FixedDate extends NativeDate {
      constructor(value?: string | number | Date) { super(value instanceof NativeDate ? value.valueOf() : value ?? fixed); }
      static now() { return fixed; }
    }
    globalThis.Date = FixedDate as DateConstructor;
  });
  await page.goto("/");
  await expect(page).toHaveTitle(/AI Pixel/);
  await page.getByLabel("用户名").fill("e2e-admin");
  await page.getByLabel("密码").fill("e2e-password");
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page.getByRole("heading", { name: "用量总览" })).toBeVisible();
});

test("authenticated analytics flow supports all required dimensions", async ({ page, context }, testInfo) => {
  const cookie = (await context.cookies()).find((item) => item.name === "ai_pixel_session");
  expect(cookie?.httpOnly).toBe(true);
  expect(cookie?.sameSite).toBe("Lax");
  expect((cookie?.expires ?? 0) - Date.now() / 1000).toBeGreaterThan(23.5 * 60 * 60);
  expect((cookie?.expires ?? 0) - Date.now() / 1000).toBeLessThan(24.5 * 60 * 60);

  await expect(page.getByRole("status")).toContainText("部分账户同步异常");
  await page.getByLabel("账户").selectOption("secondary");
  await expect(page.getByText("备用账户视图")).toBeVisible();
  await page.getByLabel("账户").selectOption("");
  await page.getByLabel("聚合粒度").selectOption("day");
  await page.getByLabel("昨天同期").uncheck();
  await expect(page.getByLabel("上周同期")).toBeChecked();

  await page.getByRole("button", { name: "请求分析" }).click();
  await page.getByLabel("账户").selectOption("primary");
  await page.getByLabel("模型").fill("gpt-4.1");
  await page.getByLabel("端点").fill("/v1/responses");
  await page.getByLabel("API Key ID").fill("101");
  const usageResponse = page.waitForResponse((response) => {
    const url = new URL(response.url());
    return url.pathname === "/api/usage/records" && url.searchParams.get("account_id") === "primary" && url.searchParams.get("api_key_id") === "101";
  });
  await page.getByRole("button", { name: "应用筛选" }).click();
  expect((await usageResponse).ok()).toBe(true);
  await expect(page.getByRole("cell", { name: "gpt-4.1" }).first()).toBeVisible();

  await page.getByRole("button", { name: "余额流水" }).click();
  await page.getByLabel("账户").selectOption("primary");
  await page.getByLabel("流水方向").selectOption("debit");
  await page.getByLabel("原始流水类型").fill("usage_charge");
  await page.getByLabel("业务分类").fill("usage");
  await page.getByLabel("备注关键词").fill("req-001");
  await page.getByLabel("引用类型").fill("request");
  await page.getByLabel("引用 ID").fill("9001");
  const ledgerResponse = page.waitForResponse((response) => {
    const url = new URL(response.url());
    return url.pathname === "/api/ledger/entries" && url.searchParams.get("reason") === "usage_charge" && url.searchParams.get("ref_id") === "9001";
  });
  await page.getByRole("button", { name: "应用筛选" }).click();
  expect((await ledgerResponse).ok()).toBe(true);
  await expect(page.getByText(/request_id=req-001/)).toBeVisible();
  await expect(page.getByRole("button", { name: "查看备注详情" })).toHaveCount(1);
  await page.getByRole("button", { name: "查看备注详情" }).click();
  await expect(page.locator(".metadata-row pre")).toContainText('"request_id": "req-001"');

  await page.getByRole("button", { name: "账户状态" }).click();
  await page.getByRole("button", { name: "立即同步" }).first().click();
  await expect(page.getByText("同步已完成")).toBeVisible();

  const suffix = testInfo.project.name.includes("mobile") ? "mobile" : "desktop";
  await page.screenshot({ path: path.join(screenshots, `dashboard-${suffix}.png`), fullPage: true });
  await page.getByRole("button", { name: "退出登录" }).click();
  await expect(page.getByRole("heading", { name: "登录分析中心" })).toBeVisible();
  expect(consoleIssues.get(page)).toEqual([]);
});
