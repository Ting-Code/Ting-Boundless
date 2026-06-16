import type { ReactNode } from 'react';
import { Typography } from 'antd';

type PageShellProps = {
  title: string;
  children: ReactNode;
};

export function PageShell({ title, children }: PageShellProps) {
  return (
    <>
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        {title}
      </Typography.Title>
      {children}
    </>
  );
}
