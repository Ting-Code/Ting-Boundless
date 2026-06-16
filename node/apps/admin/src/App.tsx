import { lazy, Suspense } from 'react';
import { Layout, Menu, Spin, Typography } from 'antd';
import { Link, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { SessionBar } from './components/SessionBar';

const DashboardPage = lazy(() =>
  import('./pages/DashboardPage').then((m) => ({ default: m.DashboardPage })),
);
const ItemsPage = lazy(() => import('./pages/ItemsPage').then((m) => ({ default: m.ItemsPage })));
const FilesPage = lazy(() => import('./pages/FilesPage').then((m) => ({ default: m.FilesPage })));
const AuditPage = lazy(() => import('./pages/AuditPage').then((m) => ({ default: m.AuditPage })));
const UsersPage = lazy(() => import('./pages/UsersPage').then((m) => ({ default: m.UsersPage })));
const AccountPage = lazy(() =>
  import('./pages/AccountPage').then((m) => ({ default: m.AccountPage })),
);

const { Header, Content, Sider } = Layout;

function PageFallback() {
  return (
    <div style={{ padding: 48, textAlign: 'center' }}>
      <Spin size="large" />
    </div>
  );
}

export function App() {
  const location = useLocation();
  const selected = location.pathname.startsWith('/items')
    ? ['items']
    : location.pathname.startsWith('/files')
      ? ['files']
      : location.pathname.startsWith('/audit')
        ? ['audit']
        : location.pathname.startsWith('/users')
          ? ['users']
          : location.pathname.startsWith('/account')
            ? ['account']
            : ['dashboard'];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth="0">
        <div style={{ padding: 16 }}>
          <Typography.Text strong style={{ color: '#fff' }}>
            Ting Admin
          </Typography.Text>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selected}
          items={[
            { key: 'dashboard', label: <Link to="/">概览</Link> },
            { key: 'items', label: <Link to="/items">业务条目</Link> },
            { key: 'files', label: <Link to="/files">文件</Link> },
            { key: 'audit', label: <Link to="/audit">审计</Link> },
            { key: 'users', label: <Link to="/users">用户</Link> },
            { key: 'account', label: <Link to="/account">账户</Link> },
          ]}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', paddingInline: 24 }}>
          <SessionBar />
        </Header>
        <Content style={{ margin: 24 }}>
          <Suspense fallback={<PageFallback />}>
            <Routes>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/items" element={<ItemsPage />} />
              <Route path="/files" element={<FilesPage />} />
              <Route path="/audit" element={<AuditPage />} />
              <Route path="/users" element={<UsersPage />} />
              <Route path="/account" element={<AccountPage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </Content>
      </Layout>
    </Layout>
  );
}
