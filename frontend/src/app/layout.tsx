import '@/styles/globals.css'
import { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { Providers } from './providers'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: '容器编排管理平台',
  description: '可视化的容器编排管理平台，支持Kubernetes环境下的容器生命周期管理',
  keywords: ['容器', 'Kubernetes', '编排', '管理', 'Docker'],
  authors: [{ name: '容器管理平台团队' }],
  viewport: 'width=device-width, initial-scale=1',
  themeColor: '#1890ff',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body className={inter.className}>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}