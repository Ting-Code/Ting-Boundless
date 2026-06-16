import { Link } from 'react-router-dom';
import { Alert, Card, Col, Row, Spin, Statistic, Typography } from 'antd';
import { PageShell } from '../components/PageShell';
import { QueryErrorAlert } from '../components/QueryErrorAlert';
import { useDashboardData } from '../hooks/useDashboardData';
import { handleAuthError } from '../hooks/useSession';

const quickLinks = [
  { to: '/items', label: '业务条目' },
  { to: '/files', label: '文件' },
  { to: '/users', label: '租户用户' },
  { to: '/audit', label: '审计日志' },
  { to: '/account', label: '账户' },
];

export function DashboardPage() {
  const { meQuery, pingQuery, itemsQuery, filesQuery, usersQuery, isAdmin } = useDashboardData();

  if (meQuery.isError && handleAuthError(meQuery.error, '/admin')) {
    return null;
  }

  const ping = pingQuery.data;
  const pingOk = ping?.ok === true;

  return (
    <PageShell title="概览">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card title="business-service">
            {pingQuery.isError ? (
              <QueryErrorAlert error={pingQuery.error} returnTo="/admin" />
            ) : (
              <>
                <Statistic
                  title="状态"
                  value={pingQuery.isLoading ? '检查中…' : pingOk ? '正常' : '异常'}
                  valueStyle={{ color: pingOk ? '#3f8600' : '#cf1322' }}
                />
                {ping?.service ? (
                  <Typography.Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
                    {ping.service}
                  </Typography.Text>
                ) : null}
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Link to="/items">
            <Card title="业务条目" hoverable>
              {itemsQuery.isLoading ? (
                <Spin size="small" />
              ) : itemsQuery.isError ? (
                <Typography.Text type="secondary">—</Typography.Text>
              ) : (
                <Statistic title="当前租户" value={itemsQuery.data?.items.length ?? 0} suffix="条" />
              )}
            </Card>
          </Link>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Link to="/files">
            <Card title="我的文件" hoverable>
              {filesQuery.isLoading ? (
                <Spin size="small" />
              ) : filesQuery.isError ? (
                <Typography.Text type="secondary">—</Typography.Text>
              ) : (
                <Statistic title="最近 50 条" value={filesQuery.data?.files.length ?? 0} suffix="个" />
              )}
            </Card>
          </Link>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          {isAdmin ? (
            <Link to="/users">
              <Card title="租户用户" hoverable>
                {usersQuery.isLoading ? (
                  <Spin size="small" />
                ) : usersQuery.isError ? (
                  <Typography.Text type="secondary">—</Typography.Text>
                ) : (
                  <Statistic title="最近 50 人" value={usersQuery.data?.users.length ?? 0} suffix="人" />
                )}
              </Card>
            </Link>
          ) : (
            <Card title="租户用户">
              <Typography.Text type="secondary">需 admin 角色</Typography.Text>
            </Card>
          )}
        </Col>
        <Col xs={24}>
          <Card title="快捷入口">
            <Row gutter={[8, 8]}>
              {quickLinks.map((link) => (
                <Col key={link.to}>
                  <Link to={link.to}>{link.label}</Link>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>
      <Alert
        style={{ marginTop: 16 }}
        type="info"
        showIcon
        message="API 经 Gateway 代理"
        description="所有请求使用 HttpOnly cookie（BFF）或开发环境 /sign-in/dev；业务服务不直接暴露给浏览器。"
      />
    </PageShell>
  );
}
