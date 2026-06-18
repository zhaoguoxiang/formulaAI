export function extractErrorMessage(err: unknown, fallback = '未知错误'): string {
  if (!err) return fallback;
  if (typeof err === 'string') return err;
  if (err instanceof Error) return err.message;
  const e = err as { error?: { error?: string; details?: string; errors?: string[]; message?: string }; message?: string };
  if (e.error?.errors?.length) return e.error.errors.join('；');
  if (e.error?.details) return e.error.details;
  if (e.error?.error) return e.error.error;
  if (e.error?.message) return e.error.message;
  if (e.message) return e.message;
  return fallback;
}
