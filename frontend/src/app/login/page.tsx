'use client'

export default function LoginPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-lg shadow-xl p-8">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
              <div className="text-2xl font-bold text-blue-600">🐳</div>
            </div>
            <h1 className="text-2xl font-bold text-gray-800 mb-2">容器编排管理平台</h1>
            <p className="text-gray-600">Docker 构建测试版本 - 登录页面</p>
          </div>

          <div className="space-y-4">
            <div className="text-center text-gray-600">
              <p>登录功能正在开发中...</p>
              <p className="text-sm text-gray-500 mt-2">
                请等待后续功能完善
              </p>
            </div>
          </div>
        </div>

        <div className="text-center mt-8 text-gray-500 text-sm">
          <p>© 2024 容器编排管理平台. 保留所有权利.</p>
        </div>
      </div>
    </div>
  )
}