export interface LoadingState<T> {
  loading: boolean;
  error: string | null;
  data: T | null;
}

export function initialLoadingState<T>(): LoadingState<T> {
  return { loading: true, error: null, data: null };
}

export function loadingComplete<T>(data: T | null): LoadingState<T> {
  return { loading: false, error: null, data };
}

export function loadingFailed<T>(error: string): LoadingState<T> {
  return { loading: false, error, data: null };
}
