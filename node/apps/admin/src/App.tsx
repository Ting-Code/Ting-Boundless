import { Layout, Menu, Typography } from 'antd';
import { Link, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { SessionBar } from './components/SessionBar';
import { ItemsPage } from './pages/ItemsPage';

const { Header, Content, Sider } = Layout;

export function App() {
  const location = useLocation();
  const selected = location.pathname.startsWith('/items') ? ['items'] : [];

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
          items={[{ key: 'items', label: <Link to="/items">业务条目</Link> }]}
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
          </Routes>
        </Content>
      </Layout>
    </Layout>
  );
}
