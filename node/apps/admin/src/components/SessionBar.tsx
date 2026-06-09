import { Alert, Button, Space, Typography } from 'antd';
import { signInPath } from '../config/auth';
import { useSession } from '../hooks/useSession';
import { ApiError } from '../api/client';

export function SessionBar() {
  const me = useSession();

  if (me.isLoading) {
    return <Typography.Text type="secondary">正在检查登录状态…</Typography.Text>;
  }

  if (me.isError) {
    const unauthorized = me.error instanceof ApiError && me.error.status === 401;
    if (unauthorized) {
      return (
        <Alert
          type="warning"
          message="未登录"
          action={
            <Button size="small" href={signInPath('/admin/items')}>
              {import.meta.env.VITE_DEV_LOGIN === 'true' ? '开发环境登录' : '登录'}
            </Button>
          }
        />
      );
    }
    return (
      <Alert
        type="error"
        message="无法获取会话"
        description={me.error instanceof Error ? me.error.message : String(me.error)}
      />
    );
  }

  const user = me.data;
  if (!user?.user_id) {
    return (
      <Alert
        type="warning"
        message="未登录"
        action={
          <Button size="small" href={signInPath('/admin/items')}>
            {import.meta.env.VITE_DEV_LOGIN === 'true' ? '开发环境登录' : '登录'}
          </Button>
        }
      />
    );
  }

  return (
    <Space size="middle">
      <Typography.Text type="secondary">
        用户 <Typography.Text code>{user.user_id}</Typography.Text>
        {user.tenant_id ? (
          <>
            {' '}
            · 租户 <Typography.Text code>{user.tenant_id}</Typography.Text>
          </>
        ) : null}
      </Typography.Text>
      <Button size="small" href="/sign-out?return_to=/admin/items">
        退出
      </Button>
    </Space>
  );
}
