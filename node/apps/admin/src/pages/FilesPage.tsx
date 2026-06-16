import { useQueryClient } from '@tanstack/react-query';
import { Button, Card, Descriptions, Drawer, Popconfirm, Space, Table, Upload, message } from 'antd';
import type { UploadProps } from 'antd';
import { useState } from 'react';
import {
  apiFetch,
  apiUpload,
  filePaths,
  isApiError,
  type FileRow,
  type ListFilesResponse,
  type PresignResponse,
} from '@ting/api';
import { PageShell } from '../components/PageShell';
import { QueryErrorAlert } from '../components/QueryErrorAlert';
import { useApiQuery } from '../hooks/useApiQuery';
import { useAuthMutation } from '../hooks/useAuthMutation';
import { handleAuthError } from '../hooks/useSession';
import { basename, formatBytes, formatDateTime } from '../utils/format';

export function FilesPage() {
  const queryClient = useQueryClient();
  const [viewing, setViewing] = useState<FileRow | null>(null);

  const filesQuery = useApiQuery({
    queryKey: ['files'],
    queryFn: () => apiFetch<ListFilesResponse>(filePaths.listQuery(50)),
    authReturnTo: '/admin/files',
  });

  const uploadMutation = useAuthMutation('/admin/files', {
    mutationFn: (file: File) => apiUpload(filePaths.list, file),
    onSuccess: () => {
      message.success('上传成功');
      void queryClient.invalidateQueries({ queryKey: ['files'] });
    },
    onError: (err: unknown) => {
      message.error(isApiError(err) ? err.message : '上传失败');
    },
  });

  const deleteMutation = useAuthMutation('/admin/files', {
    mutationFn: (fileId: string) =>
      apiFetch(filePaths.item(fileId), { method: 'DELETE' }),
    onSuccess: () => {
      message.success('已删除');
      setViewing(null);
      void queryClient.invalidateQueries({ queryKey: ['files'] });
    },
    onError: (err: unknown) => {
      message.error(isApiError(err) ? err.message : '删除失败');
    },
  });

  const download = async (fileId: string) => {
    try {
      const presign = await apiFetch<PresignResponse>(filePaths.url(fileId));
      window.open(presign.url, '_blank', 'noopener,noreferrer');
    } catch (err) {
      if (handleAuthError(err, '/admin/files')) {
        return;
      }
      const msg = isApiError(err) ? err.message : '获取下载链接失败';
      message.error(msg);
    }
  };

  const uploadProps: UploadProps = {
    showUploadList: false,
    beforeUpload: (file) => {
      uploadMutation.mutate(file);
      return false;
    },
  };

  if (filesQuery.isError && handleAuthError(filesQuery.error, '/admin/files')) {
    return null;
  }

  if (filesQuery.isError) {
    return (
      <PageShell title="文件">
        <QueryErrorAlert error={filesQuery.error} returnTo="/admin/files" />
      </PageShell>
    );
  }

  return (
    <PageShell title="文件">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>

      <Card title="上传">
        <Upload {...uploadProps}>
          <Button loading={uploadMutation.isPending}>选择文件</Button>
        </Upload>
      </Card>

      <Card title="我的文件">
        <Table
          rowKey="file_id"
          loading={filesQuery.isLoading}
          dataSource={filesQuery.data?.files ?? []}
          pagination={false}
          columns={[
            {
              title: '名称',
              dataIndex: 'object_key',
              render: (key: string) => basename(key),
            },
            { title: '类型', dataIndex: 'content_type', width: 160 },
            {
              title: '大小',
              dataIndex: 'size_bytes',
              width: 100,
              render: (n: number) => formatBytes(n),
            },
            {
              title: '上传时间',
              dataIndex: 'created_at',
              width: 200,
              render: (v: string) => formatDateTime(v),
            },
            {
              title: '操作',
              width: 180,
              render: (_: unknown, row: FileRow) => (
                <Space size="small" onClick={(e) => e.stopPropagation()}>
                  <Button type="link" size="small" onClick={() => setViewing(row)}>
                    详情
                  </Button>
                  <Button type="link" size="small" onClick={() => void download(row.file_id)}>
                    下载
                  </Button>
                  <Popconfirm
                    title="删除此文件？"
                    onConfirm={() => deleteMutation.mutate(row.file_id)}
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
        title="文件详情"
        width={520}
        open={viewing !== null}
        onClose={() => setViewing(null)}
      >
        {viewing ? (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="文件 ID">{viewing.file_id}</Descriptions.Item>
            <Descriptions.Item label="名称">{basename(viewing.object_key)}</Descriptions.Item>
            <Descriptions.Item label="对象键">{viewing.object_key}</Descriptions.Item>
            <Descriptions.Item label="类型">{viewing.content_type}</Descriptions.Item>
            <Descriptions.Item label="大小">{formatBytes(viewing.size_bytes)}</Descriptions.Item>
            <Descriptions.Item label="桶">{viewing.bucket}</Descriptions.Item>
            <Descriptions.Item label="租户">{viewing.tenant_id || '—'}</Descriptions.Item>
            <Descriptions.Item label="所有者">{viewing.owner_id}</Descriptions.Item>
            <Descriptions.Item label="上传时间">{formatDateTime(viewing.created_at)}</Descriptions.Item>
          </Descriptions>
        ) : null}
      </Drawer>
      </Space>
    </PageShell>
  );
}
