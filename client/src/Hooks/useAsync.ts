import {
  useState,
  useCallback,
  useEffect,
  useRef,
  DependencyList,
} from "react";

type AsyncFn<T, P extends unknown[] = unknown[]> = (...params: P) => Promise<T>;

export type UseAsyncState<T> = {
  loading: boolean;
  error: Error | undefined;
  value: T | undefined;
};

export type UseAsyncReturn<T, P extends unknown[]> = UseAsyncState<T> & {
  execute: (...params: P) => Promise<T>;
};

type UseAsyncAutoReturn<T> = Omit<UseAsyncReturn<T, []>, "execute">;

export function useAsync<T>(
  fn: AsyncFn<T, []>,
  deps: DependencyList = [],
): UseAsyncAutoReturn<T> {
  const { loading, error, value, execute } = useAsyncInternal(fn, deps, true);

  useEffect(() => {
    void execute();
  }, [execute]);

  return { loading, error, value };
}

export function useAsyncFn<T, P extends unknown[] = unknown[]>(
  fn: AsyncFn<T, P>,
  deps: DependencyList = [],
): UseAsyncReturn<T, P> {
  return useAsyncInternal(fn, deps, false);
}

function useAsyncInternal<T, P extends unknown[]>(
  fn: AsyncFn<T, P>,
  deps: DependencyList,
  startInLoading: boolean,
): UseAsyncReturn<T, P> {
  const [state, setState] = useState<UseAsyncState<T>>({
    loading: startInLoading,
    error: undefined,
    value: undefined,
  });

  const isMounted = useRef(true);
  useEffect(
    () => () => {
      isMounted.current = false;
    },
    [],
  );

  const execute = useCallback(
    async (...params: P) => {
      setState((s) => ({ ...s, loading: true }));

      try {
        const data = await fn(...params);
        if (isMounted.current)
          setState({ loading: false, error: undefined, value: data });
        return data;
      } catch (e) {
        const err = e instanceof Error ? e : new Error(String(e));
        if (isMounted.current)
          setState({ loading: false, error: err, value: undefined });
        return Promise.reject(err);
      }
    },
    [fn, ...deps],
  ) as (...params: P) => Promise<T>;

  return { ...state, execute };
}
