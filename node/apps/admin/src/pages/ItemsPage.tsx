import { useQueryClient } from '@tanstack/react-query';
import {
  apiFetch,
  businessPaths,
  isApiError,
  type BusinessItem,
  type CreateItemRequest,
  type ListItemsResponse,
  type UpdateItemRequest,
} from '@ting/api';
import {
  Button,
  Card,
  Descriptions,
  Drawer,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  message,
} from 'antd';
import { useState } from 'react';
import { PageShell } from '../components/PageShell';
import { QueryErrorAlert } from '../components/QueryErrorAlert';
import { useApiQuery } from '../hooks/useApiQuery';
import { useAuthMutation } from '../hooks/useAuthMutation';
import { handleAuthError } from '../hooks/useSession';
import { formatDateTime } from '../utils/format';

export function ItemsPage() {
  const queryClient = useQueryClient();
  const [form] = Form.useForm<CreateItemRequest>();
  const [editForm] = Form.useForm<UpdateItemRequest>();
  const [editing, setEditing] = useState<BusinessItem | null>(null);
  const [viewing, setViewing] = useState<BusinessItem | null>(null);

  const itemsQuery = useApiQuery({
    queryKey: ['business', 'items'],
    queryFn: () => apiFetch<ListItemsResponse>(businessPaths.items),
    authReturnTo: '/admin/items',
  });

  const createMutation = useAuthMutation('/admin/items', {
    mutationFn: (body: CreateItemRequest) =>
      apiFetch(businessPaths.items, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      message.success('已创建');
      form.resetFields();
      void queryClient.invalidateQueries({ queryKey: ['business', 'items'] });
    },
    onError: (err: unknown) => {
      message.error(isApiError(err) ? err.message : '创建失败');
    },
  });

  const updateMutation = useAuthMutation('/admin/items', {
    mutationFn: ({ id, body }: { id: string; body: UpdateItemRequest }) =>
      apiFetch(businessPaths.item(id), {
        method: 'PATCH',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      message.success('已更新');
      setEditing(null);
      void queryClient.invalidateQueries({ queryKey: ['business', 'items'] });
    },
    onError: (err: unknown) => {
      message.error(isApiError(err) ? err.message : '更新失败');
    },
  });

  const deleteMutation = useAuthMutation('/admin/items', {
    mutationFn: (id: string) =>
      apiFetch(businessPaths.item(id), { method: 'DELETE' }),
    onSuccess: () => {
      message.success('已删除');
      void queryClient.invalidateQueries({ queryKey: ['business', 'items'] });
    },
    onError: (err: unknown) => {
      message.error(isApiError(err) ? err.message : '删除失败');
    },
  });

  const openEdit = (item: BusinessItem) => {
    setEditing(item);
    editForm.setFieldsValue({ title: item.title, body: item.body });
  };

  if (itemsQuery.isError && handleAuthError(itemsQuery.error, '/admin/items')) {
    return null;
  }

  if (itemsQuery.isError) {
    return (
      <PageShell title="业务条目">
        <QueryErrorAlert error={itemsQuery.error} returnTo="/admin/items" />
      </PageShell>
    );
  }

  return (
    <PageShell title="业务条目">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>

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
              render: (v: string) => formatDateTime(v),
            },
            {
              title: '操作',
              width: 180,
              render: (_: unknown, row: BusinessItem) => (
                <Space size="small" onClick={(e) => e.stopPropagation()}>
                  <Button type="link" size="small" onClick={() => setViewing(row)}>
                    详情
                  </Button>
                  <Button type="link" size="small" onClick={() => openEdit(row)}>
                    编辑
                  </Button>
                  <Popconfirm
                    title="删除此条目？"
                    onConfirm={() => deleteMutation.mutate(row.id)}
                  >
                    <Button type="link" size="small" danger loading={deleteMutation.isPending}>
                      删除
                    </Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
          onRow={(record) => ({
            onClick: () => setViewing(record),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>

      <Drawer
        title="条目详情"
        width={480}
        open={viewing !== null}
        onClose={() => setViewing(null)}
      >
        {viewing ? (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="ID">{viewing.id}</Descriptions.Item>
            <Descriptions.Item label="标题">{viewing.title}</Descriptions.Item>
            <Descriptions.Item label="备注">{viewing.body || '—'}</Descriptions.Item>
            <Descriptions.Item label="租户">{viewing.tenant_id || '—'}</Descriptions.Item>
            <Descriptions.Item label="创建人">{viewing.created_by}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(viewing.created_at)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDateTime(viewing.updated_at)}</Descriptions.Item>
          </Descriptions>
        ) : null}
      </Drawer>

      <Modal
        title="编辑条目"
        open={editing !== null}
        onCancel={() => setEditing(null)}
        onOk={() => {
          void editForm.validateFields().then((values) => {
            if (editing) {
              updateMutation.mutate({ id: editing.id, body: values });
            }
          });
        }}
        confirmLoading={updateMutation.isPending}
        destroyOnClose
      >
        <Form form={editForm} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="body" label="备注">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
      </Space>
    </PageShell>
  );
}
