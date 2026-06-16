import { QueryCache, QueryClient } from '@tanstack/react-query';
import { isApiError } from '@ting/api';
import { redirectToSignIn } from '../config/auth';

export function createQueryClient(): QueryClient {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (error, query) => {
        const returnTo = query.meta?.authReturnTo;
        if (typeof returnTo === 'string' && isApiError(error) && error.status === 401) {
          redirectToSignIn(returnTo);
        }
      },
    }),
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
      },
    },
  });
}
