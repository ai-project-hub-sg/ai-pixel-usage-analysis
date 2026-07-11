import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { LoginPage } from "./LoginPage";

test("submits dashboard credentials and reports success", async () => {
  const login = vi.fn().mockResolvedValue(undefined);
  const success = vi.fn();
  render(<LoginPage login={login} onSuccess={success} />);
  await userEvent.type(screen.getByLabelText("用户名"), "admin");
  await userEvent.type(screen.getByLabelText("密码"), "secret");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));
  expect(login).toHaveBeenCalledWith("admin", "secret");
  expect(success).toHaveBeenCalled();
});

test("shows a generic authentication error", async () => {
  const login = vi.fn().mockRejectedValue(new Error("denied"));
  render(<LoginPage login={login} onSuccess={() => undefined} />);
  await userEvent.type(screen.getByLabelText("用户名"), "admin");
  await userEvent.type(screen.getByLabelText("密码"), "wrong");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("用户名或密码错误");
});
