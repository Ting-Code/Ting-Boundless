import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Card, Space, Table, Typography, Upload, message } from 'antd';
import type { UploadProps } from 'antd';
import {
  apiFetch,
  apiUpload,
  ApiError,
  filePaths,
  type FileRow,
  type ListFilesResponse,
  type PresignResponse,
} from '@ting/api';
import { signInPath } from '../config/auth';
import { handleAuthError } from '../hooks/useSession';

function formatBytes(n: number): string {
  if (n < 1024) {
    return `${n} B`;
  }
  if (n < 1024 * 1024) {
    return `${(n / 1024).toFixed(1)} KiB`;
  }
  return `${(n / (1024 * 1024)).toFixed(1)} MiB`;
}

function basename(key: string): string {
  const parts = key.split('/');
  return parts[parts.length - 1] || key;
}

export function FilesPage() {
  const queryClient = useQueryClient();

  const filesQuery = useQuery({
    queryKey: ['files'],
    queryFn: () => apiFetch<ListFilesResponse>(filePaths.list),
  });

  const uploadMutation = useMutation({
    mutationFn: (file: File) => apiUpload(filePaths.list, file),
    onSuccess: () => {
      message.success('上传成功');
      void queryClient.invalidateQueries({ queryKey: ['files'] });
    },
    onError: (err: unknown) => {
      if (handleAuthError(err)) {
        return;
      }
      const msg = err instanceof ApiError ? err.message : '上传失败';
      message.error(msg);
    },
  });

  const download = async (fileId: string) => {
    try {
      const presign = await apiFetch<PresignResponse>(filePaths.url(fileId));
      window.open(presign.url, '_blank', 'noopener,noreferrer');
    } catch (err) {
      if (handleAuthError(err)) {
        return;
      }
      const msg = err instanceof ApiError ? err.message : '获取下载链接失败';
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

  if (filesQuery.isError && handleAuthError(filesQuery.error)) {
    return null;
  }

  if (filesQuery.error instanceof ApiError && filesQuery.error.status === 401) {
    return (
      <Alert
        type="warning"
        message="未登录"
        description={
          <Button type="link" href={signInPath('/admin/files')}>
            前往登录
          </Button>
        }
      />
    );
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Typography.Title level={3} style={{ margin: 0 }}>
        文件
      </Typography.Title>

      <Card title="上传">
        <Upload {...uploadProps}>
          <Button loading={uploadMutation.isPending}>选择文件</Button>
        </Upload>
      </Card>

      <Card title="我的文件">
        {filesQuery.isError && !(filesQuery.error instanceof ApiError && filesQuery.error.status === 401) ? (
          <Alert
            type="error"
            message="加载失败"
            description={
              filesQuery.error instanceof ApiError
                ? filesQuery.error.message
                : String(filesQuery.error)
            }
          />
        ) : (
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
                render: (v: string) => new Date(v).toLocaleString(),
              },
              {
                title: '操作',
                width: 100,
                render: (_: unknown, row: FileRow) => (
                  <Button type="link" onClick={() => void download(row.file_id)}>
                    下载
                  </Button>
                ),
              },
            ]}
          />
        )}
      </Card>
    </Space>
  );
}
