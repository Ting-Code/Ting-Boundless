import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import {
  apiFetch,
  businessPaths,
  isApiError,
  userPaths,
  type BusinessMeResponse,
  type UserMeResponse,
} from '@ting/api';
import { Alert, Button, Card, Descriptions, Form, Input, message } from 'antd';
import { signInPath } from '../config/auth';
import { PageShell } from '../components/PageShell';
import { QueryErrorAlert } from '../components/QueryErrorAlert';
import { useApiQuery } from '../hooks/useApiQuery';
import { useAuthMutation } from '../hooks/useAuthMutation';
import { handleAuthError } from '../hooks/useSession';

export function AccountPage() {
  const queryClient = useQueryClient();
  const [form] = Form.useForm<{ display_name: string }>();

  const me = useApiQuery({
    queryKey: ['business', 'me'],
    queryFn: () => apiFetch<BusinessMeResponse>(businessPaths.me),
    authReturnTo: '/admin/account',
  });

  const profile = useApiQuery({
    queryKey: ['user', 'me'],
    queryFn: () => apiFetch<UserMeResponse>(userPaths.me),
    enabled: Boolean(me.data?.user_id),
    authReturnTo: '/admin/account',
  });

  useEffect(() => {
    if (profile.data?.display_name !== undefined) {
      form.setFieldsValue({ display_name: profile.data.display_name });
    }
  }, [form, profile.data?.display_name]);

  const saveProfile = useAuthMutation('/admin/account', {
    mutationFn: (display_name: string) =>
      apiFetch<UserMeResponse>(userPaths.me, {
        method: 'PATCH',
        body: JSON.stringify({ display_name }),
      }),
    onSuccess: (data) => {
      queryClient.setQueryData(['user', 'me'], data);
      message.success('显示名称已更新');
    },
    onError: (err: unknown) => {
      message.error(isApiError(err) ? err.message : '更新失败');
    },
  });

  if (me.isError && handleAuthError(me.error, '/admin/account')) {
    return null;
  }

  if (me.isLoading) {
    return <PageShell title="账户">加载中…</PageShell>;
  }

  if (me.isError) {
    return (
      <PageShell title="账户">
        <QueryErrorAlert error={me.error} returnTo="/admin/account" />
      </PageShell>
    );
  }

  const user = me.data;
  if (!user?.user_id) {
    return (
      <PageShell title="账户">
        <Alert
          type="warning"
          message="未登录"
          description={<a href={signInPath('/admin/account')}>前往登录</a>}
        />
      </PageShell>
    );
  }

  return (
    <PageShell title="账户">
      <Card title="身份" style={{ marginBottom: 16 }}>
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="显示名称">
            {profile.data?.display_name || '—'}
          </Descriptions.Item>
          <Descriptions.Item label="用户 ID">{user.user_id}</Descriptions.Item>
          <Descriptions.Item label="租户 ID">{user.tenant_id || '—'}</Descriptions.Item>
          <Descriptions.Item label="Subject">{user.subject || '—'}</Descriptions.Item>
          <Descriptions.Item label="角色">
            {user.roles?.length ? user.roles.join(', ') : '—'}
          </Descriptions.Item>
          <Descriptions.Item label="Scopes">
            {user.scopes?.length ? user.scopes.join(', ') : '—'}
          </Descriptions.Item>
          <Descriptions.Item label="Request ID">{user.request_id || '—'}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card title="个人资料" loading={profile.isLoading}>
        {profile.isError ? (
          <QueryErrorAlert error={profile.error} returnTo="/admin/account" />
        ) : (
          <Form
            form={form}
            layout="vertical"
            onFinish={(values) => saveProfile.mutate(values.display_name)}
            style={{ maxWidth: 400 }}
          >
            <Form.Item
              label="显示名称"
              name="display_name"
              rules={[{ required: true, message: '请输入显示名称' }, { max: 200 }]}
            >
              <Input />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={saveProfile.isPending}>
              保存
            </Button>
          </Form>
        )}
      </Card>
    </PageShell>
  );
}
