import { Layout, Menu, Typography } from 'antd';
import { Link, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { SessionBar } from './components/SessionBar';
import { AccountPage } from './pages/AccountPage';
import { AuditPage } from './pages/AuditPage';
import { FilesPage } from './pages/FilesPage';
import { ItemsPage } from './pages/ItemsPage';
import { UsersPage } from './pages/UsersPage';

const { Header, Content, Sider } = Layout;

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
          : [];

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
          <Routes>
            <Route path="/" element={<Navigate to="/items" replace />} />
            <Route path="/items" element={<ItemsPage />} />
            <Route path="/files" element={<FilesPage />} />
            <Route path="/audit" element={<AuditPage />} />
            <Route path="/users" element={<UsersPage />} />
            <Route path="/account" element={<AccountPage />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  );
}
