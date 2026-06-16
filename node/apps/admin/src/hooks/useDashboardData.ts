import {
  apiFetch,
  businessPaths,
  filePaths,
  userPaths,
  type BusinessMeResponse,
  type BusinessPingResponse,
  type ListFilesResponse,
  type ListItemsResponse,
  type ListUsersResponse,
} from '@ting/api';
import { useApiQuery } from './useApiQuery';

export function useDashboardData() {
  const meQuery = useApiQuery({
    queryKey: ['business', 'me'],
    queryFn: () => apiFetch<BusinessMeResponse>(businessPaths.me),
    authReturnTo: '/admin',
  });

  const pingQuery = useApiQuery({
    queryKey: ['business', 'ping'],
    queryFn: () => apiFetch<BusinessPingResponse>(businessPaths.ping),
  });

  const isAdmin = meQuery.data?.roles?.includes('admin') ?? false;

  const itemsQuery = useApiQuery({
    queryKey: ['business', 'items'],
    queryFn: () => apiFetch<ListItemsResponse>(businessPaths.items),
    enabled: Boolean(meQuery.data?.user_id),
    authReturnTo: '/admin',
  });

  const filesQuery = useApiQuery({
    queryKey: ['files'],
    queryFn: () => apiFetch<ListFilesResponse>(filePaths.listQuery(50)),
    enabled: Boolean(meQuery.data?.user_id),
    authReturnTo: '/admin',
  });

  const usersQuery = useApiQuery({
    queryKey: ['users', 'list'],
    queryFn: () => apiFetch<ListUsersResponse>(userPaths.listQuery(50)),
    enabled: isAdmin,
    authReturnTo: '/admin/users',
  });

  return { meQuery, pingQuery, itemsQuery, filesQuery, usersQuery, isAdmin };
}
