import type { ReactNode } from 'react';
import { Alert } from 'antd';
import { isApiError } from '@ting/api';
import { signInPath } from '../config/auth';

type QueryErrorAlertProps = {
  error: unknown;
  returnTo: string;
  forbiddenMessage?: string;
};

/** Renders 401/403/load error for list pages. */
export function QueryErrorAlert({
  error,
  returnTo,
  forbiddenMessage = '无权限',
}: QueryErrorAlertProps): ReactNode {
  if (!isApiError(error) && !(error instanceof Error)) {
    return <Alert type="error" message="加载失败" description={String(error)} />;
  }

  if (isApiError(error) && error.status === 401) {
    return (
      <Alert
        type="warning"
        message="未登录"
        description={<a href={signInPath(returnTo)}>前往登录</a>}
      />
    );
  }

  if (isApiError(error) && error.status === 403) {
    return (
      <Alert type="warning" message={forbiddenMessage} description={error.message} />
    );
  }

  return (
    <Alert
      type="error"
      message="加载失败"
      description={error instanceof Error ? error.message : String(error)}
    />
  );
}
