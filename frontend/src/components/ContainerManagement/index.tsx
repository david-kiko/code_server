'use client'

import { useState, useEffect } from 'react'
import {
  Card,
  Row,
  Col,
  Button,
  Table,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  Select,
  message,
  Dropdown,
  Menu,
  Tooltip,
  Popconfirm,
  Typography,
  Divider,
  Alert,
  Spin,
} from 'antd'
import {
  PlusOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  DeleteOutlined,
  SettingOutlined,
  MoreOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  StopOutlined,
  EyeOutlined,
  EditOutlined,
  CloudServerOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import K8sConnectionManager from '../K8sConnection'
import { k8sApi, type Container, type K8sConnection, type CreateContainerRequest } from '../../services/k8s'

export default function ContainerManagement() {
  const [containers, setContainers] = useState<Container[]>([])
  const [loading, setLoading] = useState(false)
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [connectionModalVisible, setConnectionModalVisible] = useState(false)
  const [activeConnection, setActiveConnection] = useState<K8sConnection | null>(null)
  const [k8sConnections, setK8sConnections] = useState<K8sConnection[]>([])
  const [form] = Form.useForm()
  const [createForm] = Form.useForm()

  // 模拟数据 - 后续会替换为真实 API 调用
  const mockContainers: Container[] = [
    {
      name: 'nginx-container',
      namespace: 'default',
      image: 'nginx:latest',
      status: 'Running',
      podName: 'nginx-container-pod-xyz',
      restartCount: 0,
      age: '2d',
      node: 'worker-node-1',
      labels: { app: 'nginx' },
      containerId: 'docker://abc123',
    },
    {
      name: 'redis-cache',
      namespace: 'default',
      image: 'redis:7-alpine',
      status: 'Running',
      podName: 'redis-cache-pod-abc',
      restartCount: 1,
      age: '5h',
      node: 'worker-node-2',
      labels: { app: 'redis' },
      containerId: 'docker://def456',
    },
    {
      name: 'failed-container',
      namespace: 'default',
      image: 'busybox:latest',
      status: 'Failed',
      podName: 'failed-container-pod-def',
      restartCount: 3,
      age: '30m',
      node: 'worker-node-1',
      labels: { app: 'test' },
      containerId: 'docker://ghi789',
    },
  ]

  useEffect(() => {
    loadContainers()
    loadK8sConnections()
  }, [activeConnection])

  const loadContainers = async () => {
    setLoading(true)
    try {
      if (!activeConnection) {
        message.warning('请先配置 K8s 连接')
        return
      }

      const namespace = activeConnection.namespace || 'default'
      const containers = await k8sApi.getContainers(namespace)
      setContainers(containers)
    } catch (error) {
      console.error('Failed to load containers:', error)
      message.error('加载容器列表失败: ' + (error as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const loadK8sConnections = () => {
    // 从本地存储加载 K8s 连接配置
    const stored = localStorage.getItem('k8s-connections')
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        setK8sConnections(parsed)
        const active = parsed.find((c: K8sConnection) => c.isActive)
        if (active) {
          setActiveConnection(active)
        }
      } catch (error) {
        console.error('Failed to load K8s connections:', error)
      }
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Running': return 'success'
      case 'Pending': return 'processing'
      case 'Failed': return 'error'
      case 'Succeeded': return 'default'
      default: return 'default'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'Running': return <CheckCircleOutlined />
      case 'Pending': return <ClockCircleOutlined />
      case 'Failed': return <ExclamationCircleOutlined />
      case 'Succeeded': return <CheckCircleOutlined />
      default: return <ClockCircleOutlined />
    }
  }

  const handleContainerAction = async (container: Container, action: string) => {
    try {
      message.loading(`${action}容器 ${container.name}...`)

      switch (action) {
        case '启动':
          await k8sApi.startContainer(container.namespace, container.podName)
          break
        case '停止':
          await k8sApi.stopContainer(container.namespace, container.podName)
          break
        case '重启':
          await k8sApi.restartContainer(container.namespace, container.podName)
          break
        case '删除':
          await k8sApi.deleteContainer(container.namespace, container.podName)
          break
        default:
          throw new Error(`未知的操作: ${action}`)
      }

      message.success(`${action}容器 ${container.name} 成功`)
      loadContainers() // 重新加载列表
    } catch (error) {
      console.error(`Failed to ${action} container:`, error)
      message.error(`${action}容器失败: ` + (error as Error).message)
    }
  }

  const handleCreateContainer = async (values: CreateContainerRequest) => {
    try {
      message.loading('创建容器中...')

      await k8sApi.createContainer(values)

      message.success('创建容器成功')
      setCreateModalVisible(false)
      createForm.resetFields()
      loadContainers()
    } catch (error) {
      console.error('Failed to create container:', error)
      message.error('创建容器失败: ' + (error as Error).message)
    }
  }

  const handleConnectionChange = (connection: K8sConnection | null) => {
    setActiveConnection(connection)
    if (connection) {
      loadContainers()
    }
  }

  const columns: ColumnsType<Container> = [
    {
      title: '容器名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)} icon={getStatusIcon(status)}>
          {status}
        </Tag>
      ),
    },
    {
      title: '镜像',
      dataIndex: 'image',
      key: 'image',
    },
    {
      title: '命名空间',
      dataIndex: 'namespace',
      key: 'namespace',
    },
    {
      title: '节点',
      dataIndex: 'node',
      key: 'node',
    },
    {
      title: '重启次数',
      dataIndex: 'restartCount',
      key: 'restartCount',
      render: (count: number) => (
        <Tag color={count > 0 ? 'warning' : 'default'}>{count}</Tag>
      ),
    },
    {
      title: '运行时间',
      dataIndex: 'age',
      key: 'age',
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record: Container) => (
        <Space>
          <Tooltip title="查看详情">
            <Button type="text" icon={<EyeOutlined />} size="small" />
          </Tooltip>

          {record.status === 'Running' ? (
            <>
              <Tooltip title="停止">
                <Button
                  type="text"
                  icon={<StopOutlined />}
                  size="small"
                  style={{ color: '#faad14' }}
                  onClick={() => handleContainerAction(record, '停止')}
                />
              </Tooltip>
              <Tooltip title="重启">
                <Button
                  type="text"
                  icon={<ReloadOutlined />}
                  size="small"
                  onClick={() => handleContainerAction(record, '重启')}
                />
              </Tooltip>
            </>
          ) : record.status === 'Failed' || record.status === 'Succeeded' ? (
            <Tooltip title="启动">
              <Button
                type="text"
                icon={<PlayCircleOutlined />}
                size="small"
                style={{ color: '#52c41a' }}
                onClick={() => handleContainerAction(record, '启动')}
              />
            </Tooltip>
          ) : null}

          <Popconfirm
            title="确定要删除这个容器吗？"
            description="此操作不可恢复"
            onConfirm={() => handleContainerAction(record, '删除')}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button
                type="text"
                icon={<DeleteOutlined />}
                size="small"
                danger
              />
            </Tooltip>
          </Popconfirm>

          <Dropdown
            menu={{
              items: [
                {
                  key: 'edit',
                  label: '编辑配置',
                  icon: <EditOutlined />,
                },
                {
                  key: 'logs',
                  label: '查看日志',
                  icon: <EyeOutlined />,
                },
                {
                  key: 'exec',
                  label: '进入容器',
                  icon: <SettingOutlined />,
                },
              ],
            }}
          >
            <Button type="text" icon={<MoreOutlined />} size="small" />
          </Dropdown>
        </Space>
      ),
    },
  ]

  return (
    <div className="p-6 bg-gray-50 min-h-screen">
      {/* 页面标题 */}
      <div className="mb-6">
        <Row justify="space-between" align="middle">
          <Col>
            <h1 className="text-2xl font-bold text-gray-800 mb-2">容器管理</h1>
            <p className="text-gray-600">Kubernetes 容器生命周期管理</p>
          </Col>
          <Col>
            <Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setCreateModalVisible(true)}
              >
                创建容器
              </Button>
              <Button icon={<ReloadOutlined />} onClick={loadContainers}>
                刷新
              </Button>
            </Space>
          </Col>
        </Row>
      </div>

      {/* K8s 连接信息 */}
      <Card className="mb-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <CloudServerOutlined className="mr-2 text-blue-500" />
            <span className="font-medium">集群连接: </span>
            <Tag color="blue" className="ml-2">
              {activeConnection?.name || '未连接'}
            </Tag>
          </div>
          <Button
            size="small"
            icon={<SettingOutlined />}
            onClick={() => setConnectionModalVisible(true)}
          >
            配置连接
          </Button>
        </div>
      </Card>

      {/* 容器列表 */}
      <Card>
        <Table
          columns={columns}
          dataSource={containers}
          rowKey="name"
          loading={loading}
          pagination={{
            total: containers.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
        />
      </Card>

      {/* 创建容器弹窗 */}
      <Modal
        title="创建容器"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setCreateModalVisible(false)}>
            取消
          </Button>,
          <Button key="submit" type="primary" onClick={() => createForm.submit()}>
            创建
          </Button>,
        ]}
        width={600}
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreateContainer}
        >
          <Form.Item
            name="name"
            label="容器名称"
            rules={[{ required: true, message: '请输入容器名称' }]}
          >
            <Input placeholder="my-container" />
          </Form.Item>

          <Form.Item
            name="namespace"
            label="命名空间"
            rules={[{ required: true, message: '请输入命名空间' }]}
          >
            <Select placeholder="default" defaultValue="default">
              <Select.Option value="default">default</Select.Option>
              <Select.Option value="kube-system">kube-system</Select.Option>
              <Select.Option value="monitoring">monitoring</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="image"
            label="镜像"
            rules={[{ required: true, message: '请输入容器镜像' }]}
          >
            <Input placeholder="nginx:latest" />
          </Form.Item>

          <Form.Item
            name="command"
            label="启动命令"
          >
            <Input placeholder="[]" />
          </Form.Item>

          <Form.Item
            name="ports"
            label="端口映射"
          >
            <Input placeholder="80:8080" />
          </Form.Item>

          <Form.Item
            name="env"
            label="环境变量"
          >
            <Input.TextArea placeholder="KEY1=value1&#10;KEY2=value2" rows={4} />
          </Form.Item>

          <Form.Item
            name="resources"
            label="资源限制"
          >
            <Input placeholder="cpu: 500m, memory: 256Mi" />
          </Form.Item>
        </Form>
      </Modal>

      {/* K8s 连接管理模态框 */}
      <K8sConnectionManager
        visible={connectionModalVisible}
        onClose={() => setConnectionModalVisible(false)}
        onConnectionChange={handleConnectionChange}
      />
    </div>
  )
}