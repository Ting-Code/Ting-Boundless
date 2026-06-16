import {
  useQuery,
  type QueryKey,
  type UseQueryOptions,
  type UseQueryResult,
} from '@tanstack/react-query';

type ApiQueryOptions<
  TQueryFnData,
  TError = Error,
  TData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey,
> = UseQueryOptions<TQueryFnData, TError, TData, TQueryKey> & {
  /** Gateway BFF return_to when this query returns 401. */
  authReturnTo?: string;
};

/** useQuery with optional authReturnTo meta for global 401 redirect. */
export function useApiQuery<
  TQueryFnData,
  TError = Error,
  TData = TQueryFnData,
  TQueryKey extends QueryKey = QueryKey,
>(
  options: ApiQueryOptions<TQueryFnData, TError, TData, TQueryKey>,
): UseQueryResult<TData, TError> {
  const { authReturnTo, meta, ...rest } = options;
  return useQuery({
    ...rest,
    meta: authReturnTo ? { ...meta, authReturnTo } : meta,
  });
}
