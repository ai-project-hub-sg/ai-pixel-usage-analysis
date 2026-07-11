export function StatusNotice({ children }: { children: string }) {
  return <div className="status-notice" role="status">{children}</div>;
}
