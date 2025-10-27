'use client'

import { Card, Row, Col, Statistic, Progress, List, Avatar, Tag, Space, Button, Divider } from 'antd'
import {
  ContainerOutlined,
  CloudServerOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
  SettingOutlined
} from '@ant-design/icons'

export default function SimpleDashboard() {
  // 模拟数据
  const clusterInfo = {
    name: 'default-cluster',
    version: 'v1.28.0',
    nodeCount: 3,
    namespaceCount: 5,
    podCount: 12,
    serviceCount: 8,
    resourceQuota: {
      cpu: { used: '2.5', total: '8' },
      memory: { used: '4.2Gi', total: '16Gi' },
      storage: { used: '120Gi', total: '500Gi' }
    }
  }

  const nodes = [
    {
      name: 'master-node',
      status: 'Ready' as const,
      roles: ['master', 'control-plane'],
      version: 'v1.28.0',
      internalIP: '192.168.1.10',
      os: 'Ubuntu 20.04.6 LTS',
      kernelVersion: '5.4.0-174-generic',
      containerRuntime: 'containerd://1.6.24',
      resources: {
        cpu: { allocatable: '2', capacity: '2', allocated: '0.5', usage: '25%' },
        memory: { allocatable: '8Gi', capacity: '8Gi', allocated: '2Gi', usage: '25%' },
        pods: { capacity: 110, allocated: 15, running: 12 }
      },
      conditions: [],
      createdAt: new Date().toISOString()
    }
  ]

  const recentActivities = [
    {
      id: '1',
      type: 'container' as const,
      action: 'create' as const,
      resource: 'web-server',
      namespace: 'default',
      user: 'admin',
      status: 'success' as const,
      timestamp: new Date().toISOString()
    }
  ]

  const statistics = {
    totalContainers: 15,
    runningContainers: 12,
    stoppedContainers: 2,
    errorContainers: 1,
    totalServices: 8,
    totalVolumes: 6,
    totalNamespaces: 5,
  }

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      {/* 页面标题和操作按钮 */}
      <div className="mb-6 flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">仪表板</h1>
          <p className="text-gray-600 mt-1">
            集群概览和实时监控 - Docker 构建测试版本
          </p>
        </div>

        <Space>
          <Button icon={<ReloadOutlined />}>
            刷新
          </Button>
          <Button type="default">
            自动刷新: 开启
          </Button>
        </Space>
      </div>

      {/* 集群信息卡片 */}
      <Card
        title={
          <div className="flex items-center">
            <CloudServerOutlined className="mr-2" />
            集群信息
          </div>
        }
        className="mb-6"
        extra={
          <Button icon={<SettingOutlined />} size="small">
            配置
          </Button>
        }
      >
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="集群名称"
              value={clusterInfo.name}
              prefix={<CloudServerOutlined />}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="版本"
              value={clusterInfo.version}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="节点数量"
              value={clusterInfo.nodeCount}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="命名空间"
              value={clusterInfo.namespaceCount}
            />
          </Col>
        </Row>

        <Divider />

        <Row gutter={[16, 16]}>
          <Col xs={24} md={8}>
            <div className="text-center">
              <div className="text-sm text-gray-600 mb-2">CPU 使用率</div>
              <Progress
                type="circle"
                percent={parseFloat(clusterInfo.resourceQuota.cpu.used) / parseFloat(clusterInfo.resourceQuota.cpu.total) * 100}
                size={80}
                format={() => `${clusterInfo.resourceQuota.cpu.used} / ${clusterInfo.resourceQuota.cpu.total}`}
              />
            </div>
          </Col>
          <Col xs={24} md={8}>
            <div className="text-center">
              <div className="text-sm text-gray-600 mb-2">内存使用率</div>
              <Progress
                type="circle"
                percent={parseFloat(clusterInfo.resourceQuota.memory.used) / parseFloat(clusterInfo.resourceQuota.memory.total) * 100}
                size={80}
                format={() => `${clusterInfo.resourceQuota.memory.used} / ${clusterInfo.resourceQuota.memory.total}`}
              />
            </div>
          </Col>
          <Col xs={24} md={8}>
            <div className="text-center">
              <div className="text-sm text-gray-600 mb-2">存储使用率</div>
              <Progress
                type="circle"
                percent={parseFloat(clusterInfo.resourceQuota.storage.used) / parseFloat(clusterInfo.resourceQuota.storage.total) * 100}
                size={80}
                format={() => `${clusterInfo.resourceQuota.storage.used} / ${clusterInfo.resourceQuota.storage.total}`}
              />
            </div>
          </Col>
        </Row>
      </Card>

      <Row gutter={[16, 16]}>
        {/* 节点状态 */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <div className="flex items-center">
                <CloudServerOutlined className="mr-2" />
                节点状态
              </div>
            }
            extra={<Tag color="green">{nodes.length} 个节点</Tag>}
          >
            <List
              dataSource={nodes}
              renderItem={(node) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={
                      <Avatar
                        icon={<CloudServerOutlined />}
                        style={{
                          backgroundColor: node.status === 'Ready' ? '#52c41a' : '#ff4d4f'
                        }}
                      />
                    }
                    title={
                      <div className="flex justify-between items-center">
                        <span>{node.name}</span>
                        <Tag color={node.status === 'Ready' ? 'green' : 'red'}>
                          {node.status}
                        </Tag>
                      </div>
                    }
                    description={
                      <div>
                        <div>版本: {node.version}</div>
                        <div>IP: {node.internalIP}</div>
                        <div>角色: {node.roles.join(', ')}</div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* 最近活动 */}
        <Col xs={24} lg={12}>
          <Card
            title={
              <div className="flex items-center">
                <ClockCircleOutlined className="mr-2" />
                最近活动
              </div>
            }
            extra={
              <Button type="link" size="small">
                查看全部
              </Button>
            }
          >
            <List
              dataSource={recentActivities}
              renderItem={(activity) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={
                      <Avatar
                        icon={<ContainerOutlined />}
                        style={{
                          backgroundColor: activity.status === 'success' ? '#52c41a' : '#ff4d4f'
                        }}
                      />
                    }
                    title={
                      <div className="flex justify-between items-center">
                        <span>{activity.resource}</span>
                        <Tag color={activity.status === 'success' ? 'green' : 'red'}>
                          {activity.status}
                        </Tag>
                      </div>
                    }
                    description={
                      <div>
                        <div>
                          {activity.action} - {activity.type}
                        </div>
                        <div className="text-xs text-gray-500">
                          {activity.user} · {new Date(activity.timestamp).toLocaleString()}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 容器统计 */}
      <Card
        title={
          <div className="flex items-center">
            <ContainerOutlined className="mr-2" />
            容器统计
          </div>
        }
        className="mt-6"
      >
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="总容器数"
              value={statistics.totalContainers}
              prefix={<ContainerOutlined />}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="运行中"
              value={statistics.runningContainers}
              valueStyle={{ color: '#3f8600' }}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="已停止"
              value={statistics.stoppedContainers}
              valueStyle={{ color: '#8c8c8c' }}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Statistic
              title="错误"
              value={statistics.errorContainers}
              valueStyle={{ color: '#cf1322' }}
              prefix={<ExclamationCircleOutlined />}
            />
          </Col>
        </Row>
      </Card>
    </div>
  )
}