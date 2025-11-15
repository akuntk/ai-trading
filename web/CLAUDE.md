[根目录](../../CLAUDE.md) > **web**

# Web模块 - 前端用户界面

## 模块职责

Web模块是NOFX系统的**用户交互界面**，提供直观、响应式的Web应用，支持多语言、实时监控、配置管理和交易状态可视化，是用户与AI交易系统交互的主要入口。

## 核心功能
- 🎨 **现代化UI**：React 18 + TypeScript + TailwindCSS
- 🌍 **多语言支持**：中英文界面切换
- 📊 **实时监控**：交易状态、性能指标、市场数据
- ⚙️ **配置管理**：AI模型、交易所、交易员完整配置
- 📱 **响应式设计**：桌面端和移动端适配

## 入口与启动

### 主入口文件
- **`src/App.tsx`** - 应用程序主组件
- **`src/main.tsx`** - React应用启动入口
- **`package.json`** - 项目依赖和脚本配置

### 应用结构
```tsx
function App() {
  return (
    <LanguageProvider>
      <AuthProvider>
        <ConfirmDialogProvider>
          <AppContent />
        </ConfirmDialogProvider>
      </AuthProvider>
    </LanguageProvider>
  )
}
```

## 技术栈详细

### 核心框架
- **React 18** - 用户界面框架
- **TypeScript** - 类型安全的JavaScript
- **Vite** - 快速构建工具
- **React Router v6** - 客户端路由

### 状态管理
- **Zustand** - 轻量级状态管理
- **React Query/SWR** - 服务器状态管理
- **Context API** - 全局状态共享

### 样式与UI
- **TailwindCSS** - 实用优先的CSS框架
- **Headless UI** - 无样式组件库
- **Heroicons** - 图标库
- **Chart.js/Recharts** - 图表库

### 开发工具
- **ESLint** - 代码质量检查
- **Prettier** - 代码格式化
- **Husky** - Git钩子
- **Vitest** - 单元测试框架

## 项目结构

```
web/
├── public/                # 静态资源
│   ├── icons/            # 应用图标
│   └── index.html        # HTML模板
├── src/
│   ├── components/       # 可复用组件
│   │   ├── common/       # 通用组件
│   │   ├── forms/        # 表单组件
│   │   └── charts/       # 图表组件
│   ├── contexts/         # React Context
│   │   ├── AuthContext.tsx
│   │   └── LanguageContext.tsx
│   ├── hooks/            # 自定义Hooks
│   │   ├── useAuth.ts
│   │   ├── useSystemConfig.ts
│   │   └── useWebSocket.ts
│   ├── i18n/             # 国际化
│   │   └── translations.ts
│   ├── lib/              # 工具库
│   │   ├── api.ts        # API客户端
│   │   ├── auth.ts       # 认证工具
│   │   └── utils.ts      # 通用工具
│   ├── pages/            # 页面组件
│   │   ├── Dashboard.tsx
│   │   ├── Traders.tsx
│   │   ├── AIModules.tsx
│   │   ├── Exchanges.tsx
│   │   ├── SystemConfig.tsx
│   │   └── Login.tsx
│   ├── routes/           # 路由配置
│   │   └── index.tsx
│   ├── types/            # TypeScript类型定义
│   │   ├── api.ts
│   │   ├── trader.ts
│   │   └── config.ts
│   ├── App.tsx           # 主应用组件
│   └── main.tsx          # 应用入口
├── package.json          # 项目配置
├── tsconfig.json         # TypeScript配置
├── tailwind.config.js    # TailwindCSS配置
├── vite.config.ts        # Vite构建配置
└── CLAUDE.md            # 本文档
```

## 核心组件

### 认证系统
```tsx
// AuthContext
interface AuthContextType {
  user: User | null
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  isLoading: boolean
}

// 登录组件
function LoginPage() {
  const { login, isLoading } = useAuth()

  const handleSubmit = async (formData: FormData) => {
    try {
      await login(formData.get('email'), formData.get('password'))
      // 登录成功后重定向
    } catch (error) {
      // 处理错误
    }
  }
}
```

### 语言切换
```tsx
// LanguageContext
interface LanguageContextType {
  language: 'zh' | 'en'
  setLanguage: (lang: 'zh' | 'en') => void
  t: (key: string, lang?: 'zh' | 'en') => string
}

// 翻译函数示例
const translations = {
  dashboard: {
    zh: '仪表板',
    en: 'Dashboard'
  },
  traders: {
    zh: '交易员',
    en: 'Traders'
  }
}
```

### 实时数据获取
```tsx
// SWR Hook示例
function useTraders() {
  const { data, error, mutate } = useSWR('/api/traders', fetcher, {
    refreshInterval: 5000, // 5秒刷新一次
    revalidateOnFocus: true
  })

  return {
    traders: data,
    isLoading: !error && !data,
    error,
    mutate
  }
}
```

## 页面组件

### 仪表板 (Dashboard)
- 系统概览和关键指标
- 实时交易状态监控
- 性能图表和统计
- 快速操作入口

### 交易员管理 (Traders)
- 交易员列表和状态
- 创建/编辑/删除交易员
- 启动/停止控制
- 性能分析和日志查看

### AI模型配置 (AIModules)
- AI模型列表
- API密钥配置
- 模型参数设置
- 连接测试功能

### 交易所配置 (Exchanges)
- 交易所列表
- API配置管理
- 连接状态监控
- 测试网络切换

### 系统配置 (SystemConfig)
- 全局系统参数
- 风险控制设置
- 日志和监控配置
- 内测码管理

## 状态管理策略

### 全局状态 (Zustand)
```typescript
interface AppState {
  // 用户状态
  user: User | null
  isAuthenticated: boolean

  // 系统配置
  systemConfig: SystemConfig

  // UI状态
  sidebarOpen: boolean
  theme: 'light' | 'dark'

  // Actions
  setUser: (user: User | null) => void
  setSystemConfig: (config: SystemConfig) => void
  toggleSidebar: () => void
}

const useAppStore = create<AppState>((set) => ({
  user: null,
  isAuthenticated: false,
  systemConfig: {},
  sidebarOpen: true,
  theme: 'light',

  setUser: (user) => set({ user, isAuthenticated: !!user }),
  setSystemConfig: (config) => set({ systemConfig: config }),
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen }))
}))
```

### 服务器状态 (SWR)
```typescript
// API客户端配置
const fetcher = async (url: string) => {
  const token = localStorage.getItem('authToken')
  const response = await fetch(`/api${url}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  })

  if (!response.ok) {
    throw new Error('API request failed')
  }

  return response.json()
}

// 自定义Hook
function useAPI<T>(url: string) {
  const { data, error, mutate } = useSWR<T>(url, fetcher)

  return {
    data,
    isLoading: !error && !data,
    error,
    mutate
  }
}
```

## 路由配置

```typescript
// routes/index.tsx
import { createBrowserRouter } from 'react-router-dom'
import { ProtectedRoute } from '../components/ProtectedRoute'

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />
  },
  {
    path: '/',
    element: <ProtectedRoute><Layout /></ProtectedRoute>,
    children: [
      {
        index: true,
        element: <Dashboard />
      },
      {
        path: 'traders',
        element: <Traders />
      },
      {
        path: 'ai-modules',
        element: <AIModules />
      },
      {
        path: 'exchanges',
        element: <Exchanges />
      },
      {
        path: 'system-config',
        element: <SystemConfig />
      }
    ]
  }
])
```

## API集成

### API客户端
```typescript
// lib/api.ts
class APIClient {
  private baseURL: string
  private token: string | null = null

  constructor(baseURL: string) {
    this.baseURL = baseURL
  }

  setToken(token: string) {
    this.token = token
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`

    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(this.token && { 'Authorization': `Bearer ${this.token}` }),
        ...options.headers
      }
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    return response.json()
  }

  // 便捷方法
  get<T>(endpoint: string) {
    return this.request<T>(endpoint, { method: 'GET' })
  }

  post<T>(endpoint: string, data?: any) {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    })
  }

  put<T>(endpoint: string, data?: any) {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  }

  delete<T>(endpoint: string) {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

export const apiClient = new APIClient('/api')
```

## 类型定义

### API响应类型
```typescript
// types/api.ts
export interface APIResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

export interface User {
  id: string
  email: string
  otp_verified: boolean
  created_at: string
  updated_at: string
}

export interface TraderConfig {
  id: string
  name: string
  ai_model_id: string
  exchange_id: string
  initial_balance: number
  is_running: boolean
  btc_eth_leverage: number
  altcoin_leverage: number
  // ... 更多字段
}
```

## 样式系统

### TailwindCSS配置
```javascript
// tailwind.config.js
module.exports = {
  content: ['./src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          900: '#1e3a8a'
        },
        success: {
          500: '#10b981'
        },
        warning: {
          500: '#f59e0b'
        },
        error: {
          500: '#ef4444'
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif']
      }
    }
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography')
  ]
}
```

### 主题系统
```typescript
// contexts/ThemeContext.tsx
interface ThemeContextType {
  theme: 'light' | 'dark'
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<'light' | 'dark'>('light')

  const toggleTheme = () => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light')
  }

  useEffect(() => {
    document.documentElement.className = theme
  }, [theme])

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}
```

## 性能优化

### 代码分割
```typescript
// 路由级别的代码分割
import { lazy, Suspense } from 'react'

const Dashboard = lazy(() => import('../pages/Dashboard'))
const Traders = lazy(() => import('../pages/Traders'))

// 使用Suspense包装
function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/traders" element={<Traders />} />
      </Routes>
    </Suspense>
  )
}
```

### 虚拟化长列表
```typescript
// 使用react-window进行长列表优化
import { FixedSizeList as List } from 'react-window'

function TraderList({ traders }: { traders: TraderConfig[] }) {
  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style}>
      <TraderItem trader={traders[index]} />
    </div>
  )

  return (
    <List
      height={600}
      itemCount={traders.length}
      itemSize={80}
    >
      {Row}
    </List>
  )
}
```

## 测试策略

### 单元测试 (Vitest)
```typescript
// components/TradingButton.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { TradingButton } from './TradingButton'

describe('TradingButton', () => {
  it('renders with correct label', () => {
    render(<TradingButton label="Start Trading" />)
    expect(screen.getByText('Start Trading')).toBeInTheDocument()
  })

  it('calls onClick when clicked', () => {
    const handleClick = vi.fn()
    render(<TradingButton label="Start" onClick={handleClick} />)

    fireEvent.click(screen.getByText('Start'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })
})
```

### 集成测试
```typescript
// e2e tests with Playwright
import { test, expect } from '@playwright/test'

test('user can login and view dashboard', async ({ page }) => {
  await page.goto('/login')

  await page.fill('[data-testid=email]', 'user@example.com')
  await page.fill('[data-testid=password]', 'password123')
  await page.click('[data-testid=login-button]')

  await expect(page).toHaveURL('/')
  await expect(page.locator('h1')).toContainText('Dashboard')
})
```

## 部署配置

### 构建脚本
```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest",
    "test:e2e": "playwright test",
    "lint": "eslint src --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "lint:fix": "eslint src --ext ts,tsx --fix"
  }
}
```

### Docker配置
```dockerfile
# Dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=0 /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 常见问题 (FAQ)

### Q: 如何处理国际化？
A: 使用React Context配合翻译对象，支持动态语言切换和文本插值。

### Q: 如何实现实时数据更新？
A: 使用SWR进行定时轮询，或者集成WebSocket进行实时推送。

### Q: 如何优化大表格性能？
A: 使用虚拟化技术（react-window）和分页加载来处理大量数据。

## 相关文件清单

```
web/
├── src/
│   ├── App.tsx              # 主应用组件
│   ├── main.tsx             # 应用入口
│   ├── components/          # 组件库
│   ├── contexts/            # React Context
│   ├── hooks/               # 自定义Hooks
│   ├── pages/               # 页面组件
│   ├── lib/                 # 工具库
│   ├── types/               # 类型定义
│   └── routes/              # 路由配置
├── public/                  # 静态资源
├── package.json             # 项目配置
├── vite.config.ts           # 构建配置
└── CLAUDE.md               # 本文档
```

## 变更记录 (Changelog)

### 2025-11-15 06:49:04 - 模块文档创建
- ✅ 完成前端架构分析
- ✅ 组件和状态管理文档
- ✅ 构建和部署配置说明