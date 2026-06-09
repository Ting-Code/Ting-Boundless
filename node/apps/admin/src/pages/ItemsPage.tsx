import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { CreateItemRequest, ListItemsResponse } from '@ting/api-types';
import { Alert, Button, Card, Form, Input, Space, Table, Typography, message } from 'antd';
import { apiFetch, ApiError } from '../api/client';
import { signInPath } from '../config/auth';
import { handleAuthError } from '../hooks/useSession';

export function ItemsPage() {
  const queryClient = useQueryClient();
  const [form] = Form.useForm<CreateItemRequest>();

  const itemsQuery = useQuery({
    queryKey: ['business', 'items'],
    queryFn: () => apiFetch<ListItemsResponse>('/v1/business/items'),
  });

  const createMutation = useMutation({
    mutationFn: (body: CreateItemRequest) =>
      apiFetch('/v1/business/items', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      message.success('已创建');
      form.resetFields();
      void queryClient.invalidateQueries({ queryKey: ['business', 'items'] });
    },
    onError: (err: unknown) => {
      if (handleAuthError(err)) {
        return;
      }
      const msg = err instanceof ApiError ? err.message : '创建失败';
      message.error(msg);
    },
  });

  if (itemsQuery.isError && handleAuthError(itemsQuery.error)) {
    return null;
  }

  const error = itemsQuery.error;
  if (error instanceof ApiError && error.status === 401) {
    return (
      <Alert
        type="warning"
        message="未登录"
        description={
          <Button type="link" href={signInPath('/admin/items')}>
            前往登录
          </Button>
        }
      />
    );
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Typography.Title level={3} style={{ margin: 0 }}>
        业务条目
      </Typography.Title>

      <Card title="新建">
        <Form
          form={form}
          layout="inline"
          onFinish={(values) => createMutation.mutate(values)}
        >
          <Form.Item name="title" rules={[{ required: true, message: '请输入标题' }]}>
            <Input placeholder="标题" style={{ width: 240 }} />
          </Form.Item>
          <Form.Item name="body">
            <Input placeholder="备注（可选）" style={{ width: 320 }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={createMutation.isPending}>
              创建
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="列表">
        {itemsQuery.isError && !(itemsQuery.error instanceof ApiError && itemsQuery.error.status === 401) ? (
          <Alert
            type="error"
            message="加载失败"
            description={
              itemsQuery.error instanceof ApiError
                ? itemsQuery.error.message
                : String(itemsQuery.error)
            }
          />
        ) : (
          <Table
            rowKey="id"
            loading={itemsQuery.isLoading}
            dataSource={itemsQuery.data?.items ?? []}
            pagination={false}
            columns={[
              { title: '标题', dataIndex: 'title' },
              { title: '备注', dataIndex: 'body', ellipsis: true },
              { title: '创建人', dataIndex: 'created_by', width: 140 },
              {
                title: '创建时间',
                dataIndex: 'created_at',
                width: 200,
                render: (v: string) => new Date(v).toLocaleString(),
              },
            ]}
          />
        )}
      </Card>
    </Space>
  );
}
