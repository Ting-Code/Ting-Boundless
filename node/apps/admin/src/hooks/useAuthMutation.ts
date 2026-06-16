import {
  useMutation,
  type UseMutationOptions,
  type UseMutationResult,
} from '@tanstack/react-query';
import { handleAuthError } from './authError';

/** useMutation that redirects to sign-in on 401 before calling onError. */
export function useAuthMutation<
  TData = unknown,
  TError = Error,
  TVariables = void,
  TContext = unknown,
>(
  authReturnTo: string,
  options: UseMutationOptions<TData, TError, TVariables, TContext>,
): UseMutationResult<TData, TError, TVariables, TContext> {
  const { onError, ...rest } = options;
  return useMutation({
    ...rest,
    onError: (err, variables, context, mutation) => {
      if (handleAuthError(err, authReturnTo)) {
        return;
      }
      onError?.(err, variables, context, mutation);
    },
  });
}
