import { useQuery } from '@tanstack/react-query';
import {
  apiFetch,
  ApiError,
  auditPaths,
  type AuditEvent,
  type ListAuditEventsResponse,
} from '@ting/api';
import { Alert, Button, Card, Descriptions, Drawer, Select, Space, Table, Typography } from 'antd';
import { useMemo, useState } from 'react';
import { signInPath } from '../config/auth';
import { handleAuthError } from '../hooks/useSession';

export function AuditPage() {
  const [typeFilter, setTypeFilter] = useState<string | undefined>();
  const [sourceFilter, setSourceFilter] = useState<string | undefined>();
  const [selected, setSelected] = useState<AuditEvent | null>(null);

  const queryPath = useMemo(() => {
    const params = new URLSearchParams({ limit: '50' });
    if (typeFilter) {
      params.set('type', typeFilter);
    }
    if (sourceFilter) {
      params.set('source', sourceFilter);
    }
    return `${auditPaths.events}?${params}`;
  }, [typeFilter, sourceFilter]);

  const eventsQuery = useQuery({
    queryKey: ['audit', 'events', typeFilter, sourceFilter],
    queryFn: () => apiFetch<ListAuditEventsResponse>(queryPath),
    retry: false,
  });

  if (eventsQuery.isError && handleAuthError(eventsQuery.error)) {
    return null;
  }

  if (eventsQuery.isError) {
    const err = eventsQuery.error;
    const unauthorized = err instanceof ApiError && err.status === 401;
    const forbidden = err instanceof ApiError && err.status === 403;
    return (
      <Alert
        type={unauthorized || forbidden ? 'warning' : 'error'}
        message={unauthorized ? '未登录' : forbidden ? '无权限' : '加载失败'}
        description={
          unauthorized ? (
            <a href={signInPath('/admin/audit')}>前往登录</a>
          ) : err instanceof Error ? (
            err.message
          ) : (
            String(err)
          )
        }
      />
    );
  }

  const events = eventsQuery.data?.events ?? [];

  return (
    <>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        审计日志
      </Typography.Title>
      <Card>
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            allowClear
            placeholder="事件类型"
            style={{ minWidth: 220 }}
            value={typeFilter}
            onChange={setTypeFilter}
            options={[
              { value: 'business.item.created', label: 'business.item.created' },
              { value: 'business.item.updated', label: 'business.item.updated' },
              { value: 'business.item.deleted', label: 'business.item.deleted' },
              { value: 'user.login.success', label: 'user.login.success' },
              { value: 'api.access.denied', label: 'api.access.denied' },
            ]}
          />
          <Select
            allowClear
            placeholder="来源"
            style={{ minWidth: 180 }}
            value={sourceFilter}
            onChange={setSourceFilter}
            options={[
              { value: 'business-service', label: 'business-service' },
              { value: 'auth-service', label: 'auth-service' },
              { value: 'gateway', label: 'gateway' },
            ]}
          />
        </Space>
        <Table
          rowKey="id"
          loading={eventsQuery.isLoading}
          dataSource={events}
          pagination={false}
          size="small"
          scroll={{ x: true }}
          columns={[
            { title: '时间', dataIndex: 'time', width: 180 },
            { title: '类型', dataIndex: 'type', width: 200 },
            { title: '来源', dataIndex: 'source', width: 140 },
            { title: '主体', dataIndex: 'subject', ellipsis: true },
            { title: '租户', dataIndex: 'tenant_id', width: 120 },
            { title: '操作者', dataIndex: 'actor_user_id', width: 120 },
            {
              title: 'ID',
              dataIndex: 'id',
              width: 280,
              ellipsis: true,
            },
            {
              title: '',
              key: 'actions',
              width: 72,
              fixed: 'right',
              render: (_: unknown, record: AuditEvent) => (
                <Button type="link" size="small" onClick={() => setSelected(record)}>
                  详情
                </Button>
              ),
            },
          ]}
          onRow={(record) => ({
            onClick: () => setSelected(record),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
      <Drawer
        title="审计事件详情"
        width={560}
        open={selected !== null}
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="ID">{selected.id}</Descriptions.Item>
              <Descriptions.Item label="类型">{selected.type}</Descriptions.Item>
              <Descriptions.Item label="来源">{selected.source}</Descriptions.Item>
              <Descriptions.Item label="时间">{selected.time}</Descriptions.Item>
              <Descriptions.Item label="接收时间">{selected.received_at}</Descriptions.Item>
              <Descriptions.Item label="主体">{selected.subject ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="租户">{selected.tenant_id ?? '—'}</Descriptions.Item>
              <Descriptions.Item label="操作者">{selected.actor_user_id ?? '—'}</Descriptions.Item>
            </Descriptions>
            <Typography.Title level={5} style={{ marginTop: 16 }}>
              data
            </Typography.Title>
            <pre
              style={{
                margin: 0,
                padding: 12,
                background: '#f5f5f5',
                borderRadius: 6,
                overflow: 'auto',
                fontSize: 12,
              }}
            >
              {JSON.stringify(selected.data ?? {}, null, 2)}
            </pre>
          </>
        ) : null}
      </Drawer>
    </>
  );
}
