'use client'

import { useState, useEffect } from 'react'
import {
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Button,
  Upload,
  message,
  Card,
  List,
  Space,
  Tag,
  Tooltip,
  Popconfirm,
  Typography,
  Divider,
  Alert,
} from 'antd'
import {
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
  LinkOutlined,
  FileTextOutlined,
  KeyOutlined,
  CloudServerOutlined,
} from '@ant-design/icons'
import { k8sApi, type K8sConnection } from '../../services/k8s'

interface K8sConnectionManagerProps {
  visible: boolean
  onClose: () => void
  onConnectionChange: (connection: K8sConnection | null) => void
}

export default function K8sConnectionManager({
  visible,
  onClose,
  onConnectionChange,
}: K8sConnectionManagerProps) {
  const [form] = Form.useForm()
  const [connections, setConnections] = useState<K8sConnection[]>([])
  const [loading, setLoading] = useState(false)
  const [configContent, setConfigContent] = useState('')

  useEffect(() => {
    loadConnections()
  }, [])

  const loadConnections = () => {
    // 从本地存储加载连接配置
    const stored = localStorage.getItem('k8s-connections')
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        setConnections(parsed)
        const active = parsed.find((c: K8sConnection) => c.isActive)
        if (active) {
          onConnectionChange(active)
        }
      } catch (error) {
        console.error('Failed to load K8s connections:', error)
      }
    }
  }

  const saveConnections = (newConnections: K8sConnection[]) => {
    localStorage.setItem('k8s-connections', JSON.stringify(newConnections))
    setConnections(newConnections)
  }

  const handleCreateConnection = async (values: any) => {
    try {
      setLoading(true)

      const newConnection: K8sConnection = {
        id: Date.now().toString(),
        ...values,
        isActive: false,
        createdAt: new Date().toISOString(),
      }

      const updatedConnections = [...connections, newConnection]
      saveConnections(updatedConnections)

      message.success('K8s 连接配置已保存')
      form.resetFields()
      setConfigContent('')
    } catch (error) {
      console.error('Failed to create K8s connection:', error)
      message.error('保存连接配置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSetActiveConnection = (connection: K8sConnection) => {
    const updatedConnections = connections.map(c => ({
      ...c,
      isActive: c.id === connection.id,
    }))
    saveConnections(updatedConnections)
    onConnectionChange(connection)
    message.success(`已切换到: ${connection.name}`)
  }

  const handleDeleteConnection = (id: string) => {
    Modal.confirm({
      title: '删除连接配置',
      content: '确定要删除这个 K8s 连接配置吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        const updatedConnections = connections.filter(c => c.id !== id)
        saveConnections(updatedConnections)

        if (connections.find(c => c.id === id)?.isActive) {
          onConnectionChange(null)
        }

        message.success('连接配置已删除')
      },
    })
  }

  const handleTestConnection = async (values: any) => {
    try {
      setLoading(true)
      message.loading('测试连接中...')

      await k8sApi.testConnection(values)

      message.success('连接测试成功')
    } catch (error) {
      console.error('Connection test failed:', error)
      message.error('连接测试失败: ' + (error as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const normFile = (e: any) => {
    if (Array.isArray(e)) {
      return e
    }
    return e?.fileList
  }

  const uploadProps = {
    beforeUpload: (file: File) => {
      const reader = new FileReader()
      reader.onload = (e) => {
        setConfigContent(e.target?.result as string)
      }
      reader.readAsText(file)
      return false // 阻止自动上传
    },
    onChange: (info: any) => {
      if (info.file.status === 'done') {
        form.setFieldsValue({ config: configContent })
      }
    },
  }

  return (
    <Modal
      title="Kubernetes 连接管理"
      open={visible}
      onCancel={onClose}
      footer={null}
      width={900}
    >
      <div className="space-y-6">
        {/* 连接列表 */}
        <div>
          <div className="flex justify-between items-center mb-4">
            <Typography.Title level={5} className="mb-0">
              已配置的连接
            </Typography.Title>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => form.resetFields()}
            >
              新建连接
            </Button>
          </div>

          {connections.length === 0 ? (
            <Alert
              message="暂无 K8s 连接配置"
              description="请先配置一个 Kubernetes 集群连接"
              type="info"
              showIcon
            />
          ) : (
            <List
              dataSource={connections}
              renderItem={(connection) => (
                <List.Item
                  key={connection.id}
                  actions={[
                    <Tooltip title="设为活跃连接">
                      <Button
                        type="text"
                        size="small"
                        icon={<LinkOutlined />}
                        onClick={() => handleSetActiveConnection(connection)}
                        disabled={connection.isActive}
                      />
                    </Tooltip>,
                    <Popconfirm
                      title="删除连接"
                      description="确定要删除这个连接配置吗？"
                      onConfirm={() => handleDeleteConnection(connection.id)}
                      okText="确定"
                      cancelText="取消"
                    >
                      <Button
                        type="text"
                        size="small"
                        danger
                        icon={<DeleteOutlined />}
                      />
                    </Popconfirm>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <div className="flex items-center justify-center w-10 h-10 rounded-full bg-blue-100">
                        {connection.isActive ? (
                          <LinkOutlined className="text-blue-600" />
                        ) : (
                          <CloudServerOutlined className="text-gray-400" />
                        )}
                      </div>
                    }
                    title={
                      <div className="flex items-center space-x-2">
                        <span className="font-medium">{connection.name}</span>
                        {connection.isActive && (
                          <Tag color="green">活跃</Tag>
                        )}
                      </div>
                    }
                    description={
                      <div className="space-y-1">
                        <div className="text-sm text-gray-600">
                          类型: {connection.configType === 'kubeconfig' ? 'Kubeconfig' : 'Token'}
                        </div>
                        <div className="text-sm text-gray-600">
                          端点: {connection.endpoint}
                        </div>
                        <div className="text-sm text-gray-500">
                          创建时间: {new Date(connection.createdAt).toLocaleString()}
                        </div>
                      </div>
                    }
                  />
                </List.Item>
              )}
            />
          )}
        </div>

        <Divider />

        {/* 连接配置表单 */}
        <Card title="新建 K8s 连接">
          <Form
            form={form}
            layout="vertical"
            onFinish={handleCreateConnection}
          >
            <Form.Item
              name="name"
              label="连接名称"
              rules={[{ required: true, message: '请输入连接名称' }]}
            >
              <Input placeholder="我的 K8s 集群" />
            </Form.Item>

            <Form.Item
              name="endpoint"
              label="API 端点"
              rules={[{ required: true, message: '请输入 API 端点' }]}
            >
              <Input placeholder="https://192.168.1.100:6443" />
            </Form.Item>

            <Form.Item
              name="configType"
              label="认证方式"
              rules={[{ required: true, message: '请选择认证方式' }]}
            >
              <Select placeholder="选择认证方式">
                <Select.Option value="kubeconfig">Kubeconfig 文件</Select.Option>
                <Select.Option value="token">Service Account Token</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item noStyle shouldUpdate={(prevValues) => prevValues.configType}>
              {({ getFieldValue }) => {
                const configType = getFieldValue('configType')
                if (configType === 'kubeconfig') {
                  return (
                    <Form.Item
                      name="config"
                      label="Kubeconfig 内容"
                      rules={[{ required: true, message: '请输入或上传 kubeconfig 内容' }]}
                    >
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <Upload {...uploadProps} accept=".yaml,.yml">
                          <Button icon={<FileTextOutlined />}>
                            上传 kubeconfig 文件
                          </Button>
                        </Upload>
                        <Input.TextArea
                          placeholder="粘贴 kubeconfig 内容或使用上方上传按钮"
                          value={configContent}
                          onChange={(e) => setConfigContent(e.target.value)}
                          rows={10}
                        />
                      </Space>
                    </Form.Item>
                  )
                }
                if (configType === 'token') {
                  return (
                    <Form.Item
                      name="token"
                      label="Service Account Token"
                      rules={[{ required: true, message: '请输入 Token' }]}
                    >
                      <Input.TextArea
                        placeholder="粘贴 Service Account Token"
                        rows={6}
                      />
                    </Form.Item>
                  )
                }
                return null
              }}
            </Form.Item>

            <Form.Item
              name="namespace"
              label="默认命名空间"
              initialValue="default"
            >
              <Input placeholder="default" />
            </Form.Item>

            <div className="flex justify-between">
              <Button onClick={() => form.resetFields()}>
                重置
              </Button>
              <Space>
                <Button onClick={() => handleTestConnection(form.getFieldsValue())}>
                  测试连接
                </Button>
                <Button type="primary" htmlType="submit" loading={loading}>
                  保存连接
                </Button>
              </Space>
            </div>
          </Form>
        </Card>
      </div>
    </Modal>
  )
}