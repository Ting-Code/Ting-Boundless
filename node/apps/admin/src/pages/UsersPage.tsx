import { apiFetch, userPaths, type ListUsersResponse } from '@ting/api';
import { Card, Empty, Table } from 'antd';
import { PageShell } from '../components/PageShell';
import { QueryErrorAlert } from '../components/QueryErrorAlert';
import { useApiQuery } from '../hooks/useApiQuery';
import { handleAuthError } from '../hooks/useSession';
import { formatDateTime } from '../utils/format';

export function UsersPage() {
  const usersQuery = useApiQuery({
    queryKey: ['users', 'list'],
    queryFn: () => apiFetch<ListUsersResponse>(userPaths.listQuery(50)),
    authReturnTo: '/admin/users',
  });

  if (usersQuery.isError && handleAuthError(usersQuery.error, '/admin/users')) {
    return null;
  }

  if (usersQuery.isError) {
    return (
      <PageShell title="租户用户">
        <QueryErrorAlert
          error={usersQuery.error}
          returnTo="/admin/users"
          forbiddenMessage="无权限（需要 admin 角色）"
        />
      </PageShell>
    );
  }

  const users = usersQuery.data?.users ?? [];

  return (
    <PageShell title="租户用户">
      <Card>
        <Table
          rowKey="user_id"
          loading={usersQuery.isLoading}
          dataSource={users}
          pagination={false}
          size="small"
          scroll={{ x: true }}
          locale={{ emptyText: <Empty description="暂无用户" /> }}
          columns={[
            { title: '用户 ID', dataIndex: 'user_id', width: 200, ellipsis: true },
            { title: '显示名称', dataIndex: 'display_name' },
            { title: '租户 ID', dataIndex: 'tenant_id', width: 140 },
            {
              title: '创建时间',
              dataIndex: 'created_at',
              width: 200,
              render: (v: string) => formatDateTime(v),
            },
            {
              title: '更新时间',
              dataIndex: 'updated_at',
              width: 200,
              render: (v: string) => formatDateTime(v),
            },
          ]}
        />
      </Card>
    </PageShell>
  );
}
