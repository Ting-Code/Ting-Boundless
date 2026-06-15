import { useQuery } from '@tanstack/react-query';
import {
  apiFetch,
  ApiError,
  userPaths,
  type ListUsersResponse,
} from '@ting/api';
import { Alert, Card, Table, Typography } from 'antd';
import { signInPath } from '../config/auth';
import { handleAuthError } from '../hooks/useSession';

export function UsersPage() {
  const usersQuery = useQuery({
    queryKey: ['users', 'list'],
    queryFn: () => apiFetch<ListUsersResponse>(`${userPaths.list}?limit=50`),
    retry: false,
  });

  if (usersQuery.isError && handleAuthError(usersQuery.error)) {
    return null;
  }

  if (usersQuery.isError) {
    const err = usersQuery.error;
    const unauthorized = err instanceof ApiError && err.status === 401;
    const forbidden = err instanceof ApiError && err.status === 403;
    return (
      <Alert
        type={unauthorized || forbidden ? 'warning' : 'error'}
        message={unauthorized ? '未登录' : forbidden ? '无权限' : '加载失败'}
        description={
          unauthorized ? (
            <a href={signInPath('/admin/users')}>前往登录</a>
          ) : err instanceof Error ? (
            err.message
          ) : (
            String(err)
          )
        }
      />
    );
  }

  const users = usersQuery.data?.users ?? [];

  return (
    <>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        租户用户
      </Typography.Title>
      <Card>
        <Table
          rowKey="user_id"
          loading={usersQuery.isLoading}
          dataSource={users}
          pagination={false}
          size="small"
          scroll={{ x: true }}
          columns={[
            { title: '用户 ID', dataIndex: 'user_id', width: 200, ellipsis: true },
            { title: '显示名称', dataIndex: 'display_name' },
            { title: '租户 ID', dataIndex: 'tenant_id', width: 140 },
            { title: '创建时间', dataIndex: 'created_at', width: 200 },
            { title: '更新时间', dataIndex: 'updated_at', width: 200 },
          ]}
        />
      </Card>
    </>
  );
}
