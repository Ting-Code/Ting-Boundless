import { Alert, Button, Space, Typography } from 'antd';
import { useLocation } from 'react-router-dom';
import { adminReturnTo, signInPath, signOutPath } from '../config/auth';
import { useSession } from '../hooks/useSession';
import { isApiError } from '@ting/api';

export function SessionBar() {
  const location = useLocation();
  const returnTo = adminReturnTo(location.pathname);
  const me = useSession(returnTo);

  const devLogin = import.meta.env.VITE_DEV_LOGIN === 'true';
  const loginLabel = devLogin ? '开发环境登录' : '登录';

  if (me.isLoading) {
    return <Typography.Text type="secondary">正在检查登录状态…</Typography.Text>;
  }

  if (me.isError) {
    const unauthorized = isApiError(me.error) && me.error.status === 401;
    if (unauthorized) {
      return (
        <Alert
          type="warning"
          message="未登录"
          action={
            <Button size="small" href={signInPath(returnTo)}>
              {loginLabel}
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
          <Button size="small" href={signInPath(returnTo)}>
            {loginLabel}
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
        {user.roles?.length ? (
          <>
            {' '}
            · 角色 <Typography.Text code>{user.roles.join(', ')}</Typography.Text>
          </>
        ) : null}
      </Typography.Text>
      <Button size="small" href={signOutPath(returnTo)}>
        退出
      </Button>
    </Space>
  );
}
