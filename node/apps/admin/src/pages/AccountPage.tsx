import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  apiFetch,
  ApiError,
  businessPaths,
  userPaths,
  type BusinessMeResponse,
  type UserMeResponse,
} from '@ting/api';
import { Alert, Button, Card, Descriptions, Form, Input, Typography, message } from 'antd';
import { signInPath } from '../config/auth';
import { handleAuthError } from '../hooks/useSession';

export function AccountPage() {
  const queryClient = useQueryClient();
  const [form] = Form.useForm<{ display_name: string }>();

  const me = useQuery({
    queryKey: ['business', 'me'],
    queryFn: () => apiFetch<BusinessMeResponse>(businessPaths.me),
    retry: false,
  });

  const profile = useQuery({
    queryKey: ['user', 'me'],
    queryFn: () => apiFetch<UserMeResponse>(userPaths.me),
    enabled: Boolean(me.data?.user_id),
    retry: false,
  });

  const saveProfile = useMutation({
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
      message.error(err instanceof Error ? err.message : '更新失败');
    },
  });

  if (me.isError && handleAuthError(me.error)) {
    return null;
  }

  if (me.isLoading) {
    return <Typography.Text type="secondary">加载中…</Typography.Text>;
  }

  if (me.isError) {
    const unauthorized = me.error instanceof ApiError && me.error.status === 401;
    return (
      <Alert
        type={unauthorized ? 'warning' : 'error'}
        message={unauthorized ? '未登录' : '加载失败'}
        description={
          unauthorized ? (
            <a href={signInPath('/admin/account')}>前往登录</a>
          ) : me.error instanceof Error ? (
            me.error.message
          ) : (
            String(me.error)
          )
        }
      />
    );
  }

  const user = me.data;
  if (!user?.user_id) {
    return (
      <Alert
        type="warning"
        message="未登录"
        description={<a href={signInPath('/admin/account')}>前往登录</a>}
      />
    );
  }

  const profileData = profile.data;

  useEffect(() => {
    if (profileData?.display_name !== undefined) {
      form.setFieldsValue({ display_name: profileData.display_name });
    }
  }, [form, profileData?.display_name]);

  return (
    <>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        账户
      </Typography.Title>
      <Card title="身份" style={{ marginBottom: 16 }}>
        <Descriptions column={1} bordered size="small">
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
          <Alert type="error" message="无法加载个人资料" />
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
    </>
  );
}
