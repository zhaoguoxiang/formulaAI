export function extractErrorMessage(err: unknown, fallback = '未知错误'): string {
  if (!err) return fallback;
  if (typeof err === 'string') return err;
  if (err instanceof Error) return err.message;
  const e = err as { error?: { message?: string }; message?: string };
  if (e.error?.message) return e.error.message;
  if (e.message) return e.message;
  return fallback;
}
