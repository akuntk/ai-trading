import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Progress } from './ui/progress';
import { Alert, AlertDescription } from './ui/alert';
import {
  Download,
  RefreshCw,
  AlertTriangle,
  CheckCircle,
  Clock,
  Zap,
  Settings,
  History,
  RotateCcw,
  Database,
  ArrowUp,
  ArrowDown,
  Save
} from 'lucide-react';

interface VersionInfo {
  version: string;
  build_time: string;
  release_date: string;
  release_notes: string;
  platform: string;
  published_at: string;
}

interface UpdateStatus {
  has_update: boolean;
  current_ver: string;
  latest_ver?: string;
  update_info?: VersionInfo;
  last_check: string;
  download_url?: string;
  auto_update_enabled: boolean;
}

interface UpdateProgress {
  status: string;
  progress: number;
  message: string;
  speed?: number;
  total_size?: number;
  downloaded?: number;
  eta?: number;
}

interface RestartStatus {
  is_counting_down: boolean;
  countdown_time: number;
  restart_time: string;
  reason: string;
  can_cancel: boolean;
  message: string;
  last_restart?: string;
}

interface UpdateHistory {
  restart_time: string;
  reason: string;
  version: string;
  success: boolean;
}

interface MigrationStatus {
  current_version: string;
  pending_migrations: number;
  total_migrations: number;
  needs_migration: boolean;
  pending_details?: Array<{
    version: string;
    name: string;
    description: string;
    is_critical: boolean;
  }>;
}

interface PendingMigration {
  version: string;
  name: string;
  description: string;
  up_sql: string;
  down_sql: string;
  author: string;
  created_at: string;
  is_critical: boolean;
}

interface BackupResult {
  backup_path: string;
  timestamp: string;
}

const VersionUpdate: React.FC = () => {
  const [currentVersion, setCurrentVersion] = useState<VersionInfo | null>(null);
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null);
  const [updateProgress, setUpdateProgress] = useState<UpdateProgress | null>(null);
  const [restartStatus, setRestartStatus] = useState<RestartStatus | null>(null);
  const [updateHistory, setUpdateHistory] = useState<UpdateHistory[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'overview' | 'history' | 'settings' | 'migration'>('overview');
  const [error, setError] = useState<string | null>(null);

  // 迁移相关状态
  const [migrationStatus, setMigrationStatus] = useState<MigrationStatus | null>(null);
  const [pendingMigrations, setPendingMigrations] = useState<PendingMigration[]>([]);
  const [isMigrationLoading, setIsMigrationLoading] = useState(false);

  // 获取当前版本
  const fetchCurrentVersion = useCallback(async () => {
    try {
      const response = await fetch('/api/version/current');
      const data = await response.json();
      if (data.success) {
        setCurrentVersion(data.data);
      }
    } catch (error) {
      console.error('获取当前版本失败:', error);
    }
  }, []);

  // 检查更新
  const checkForUpdates = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/check');
      const data = await response.json();
      if (data.success) {
        setUpdateStatus(data.data);
      } else {
        setError(data.error || '检查更新失败');
      }
    } catch (error) {
      setError('网络错误，请稍后重试');
      console.error('检查更新失败:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // 获取更新进度
  const fetchUpdateProgress = useCallback(async () => {
    try {
      const response = await fetch('/api/version/progress');
      const data = await response.json();
      if (data.success) {
        setUpdateProgress(data.data);
      }
    } catch (error) {
      console.error('获取更新进度失败:', error);
    }
  }, []);

  // 获取重启状态
  const fetchRestartStatus = useCallback(async () => {
    try {
      const response = await fetch('/api/version/status');
      const data = await response.json();
      if (data.success) {
        setRestartStatus(data.data);
      }
    } catch (error) {
      console.error('获取重启状态失败:', error);
    }
  }, []);

  // 下载更新
  const downloadUpdate = useCallback(async (autoRestart = false) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/download', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          auto_restart: autoRestart,
          backup: true,
        }),
      });
      const data = await response.json();
      if (data.success) {
        // 开始轮询进度
        const progressInterval = setInterval(fetchUpdateProgress, 1000);
        setTimeout(() => clearInterval(progressInterval), 60000); // 1分钟后停止轮询
      } else {
        setError(data.error || '下载失败');
      }
    } catch (error) {
      setError('下载失败，请稍后重试');
      console.error('下载更新失败:', error);
    } finally {
      setIsLoading(false);
    }
  }, [fetchUpdateProgress]);

  // 安装更新
  const installUpdate = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/install', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          backup: true,
        }),
      });
      const data = await response.json();
      if (data.success) {
        // 开始轮询进度
        const progressInterval = setInterval(fetchUpdateProgress, 1000);
        setTimeout(() => clearInterval(progressInterval), 120000); // 2分钟后停止轮询
      } else {
        setError(data.error || '安装失败');
      }
    } catch (error) {
      setError('安装失败，请稍后重试');
      console.error('安装更新失败:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // 重启应用
  const restartApplication = useCallback(async (delaySeconds = 10) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/restart', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          delay_seconds: delaySeconds,
          reason: '手动重启',
        }),
      });
      const data = await response.json();
      if (data.success) {
        // 开始轮询重启状态
        const restartInterval = setInterval(fetchRestartStatus, 1000);
        setTimeout(() => clearInterval(restartInterval), 120000); // 2分钟后停止轮询
      } else {
        setError(data.error || '重启失败');
      }
    } catch (error) {
      setError('重启失败，请稍后重试');
      console.error('重启失败:', error);
    } finally {
      setIsLoading(false);
    }
  }, [fetchRestartStatus]);

  // 取消重启
  const cancelRestart = useCallback(async () => {
    try {
      const response = await fetch('/api/version/status', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          action: 'cancel_restart',
        }),
      });
      const data = await response.json();
      if (data.success) {
        fetchRestartStatus();
      } else {
        setError(data.error || '取消重启失败');
      }
    } catch (error) {
      setError('取消重启失败');
      console.error('取消重启失败:', error);
    }
  }, [fetchRestartStatus]);

  // 切换自动更新
  const toggleAutoUpdate = useCallback(async (enabled: boolean) => {
    try {
      const response = await fetch('/api/version/auto-update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          enabled: enabled,
        }),
      });
      const data = await response.json();
      if (data.success) {
        checkForUpdates();
      } else {
        setError(data.error || '设置失败');
      }
    } catch (error) {
      setError('设置失败');
      console.error('切换自动更新失败:', error);
    }
  }, [checkForUpdates]);

  // 获取更新历史
  const fetchUpdateHistory = useCallback(async () => {
    try {
      const response = await fetch('/api/version/history');
      const data = await response.json();
      if (data.success) {
        setUpdateHistory(data.data);
      }
    } catch (error) {
      console.error('获取更新历史失败:', error);
    }
  }, []);

  // 获取迁移状态
  const fetchMigrationStatus = useCallback(async () => {
    try {
      const response = await fetch('/api/version/migration/status');
      const data = await response.json();
      if (data.success) {
        setMigrationStatus(data.data);
      }
    } catch (error) {
      console.error('获取迁移状态失败:', error);
    }
  }, []);

  // 获取待执行迁移
  const fetchPendingMigrations = useCallback(async () => {
    try {
      const response = await fetch('/api/version/migration/pending');
      const data = await response.json();
      if (data.success) {
        setPendingMigrations(data.data || []);
      }
    } catch (error) {
      console.error('获取待执行迁移失败:', error);
    }
  }, []);

  // 执行迁移
  const executeMigration = useCallback(async (version: string, autoBackup = true) => {
    setIsMigrationLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/migration/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          version,
          auto_backup: autoBackup,
        }),
      });
      const data = await response.json();
      if (data.success) {
        await fetchMigrationStatus();
        await fetchPendingMigrations();
      } else {
        setError(data.error || '执行迁移失败');
      }
    } catch (error) {
      setError('执行迁移失败，请稍后重试');
      console.error('执行迁移失败:', error);
    } finally {
      setIsMigrationLoading(false);
    }
  }, [fetchMigrationStatus, fetchPendingMigrations]);

  // 回滚迁移
  const rollbackMigration = useCallback(async (targetVersion: string) => {
    setIsMigrationLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/migration/rollback', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          target_version: targetVersion,
        }),
      });
      const data = await response.json();
      if (data.success) {
        await fetchMigrationStatus();
        await fetchPendingMigrations();
      } else {
        setError(data.error || '回滚迁移失败');
      }
    } catch (error) {
      setError('回滚迁移失败，请稍后重试');
      console.error('回滚迁移失败:', error);
    } finally {
      setIsMigrationLoading(false);
    }
  }, [fetchMigrationStatus, fetchPendingMigrations]);

  // 创建备份
  const createBackup = useCallback(async () => {
    setIsMigrationLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/version/migration/backup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({}),
      });
      const data = await response.json();
      if (data.success) {
        alert(`备份创建成功！\n备份路径: ${data.data.backup_path}\n备份时间: ${data.data.timestamp}`);
      } else {
        setError(data.error || '创建备份失败');
      }
    } catch (error) {
      setError('创建备份失败，请稍后重试');
      console.error('创建备份失败:', error);
    } finally {
      setIsMigrationLoading(false);
    }
  }, []);

  // 格式化文件大小
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // 格式化时间
  const formatTime = (seconds: number): string => {
    if (seconds < 60) return `${seconds}秒`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
    return `${Math.floor(seconds / 3600)}小时${Math.floor((seconds % 3600) / 60)}分`;
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'downloading':
      case 'installing':
        return <Download className="w-4 h-4 text-blue-500" />;
      case 'failed':
        return <AlertTriangle className="w-4 h-4 text-red-500" />;
      default:
        return <Clock className="w-4 h-4 text-gray-500" />;
    }
  };

  useEffect(() => {
    fetchCurrentVersion();
    checkForUpdates();
    fetchRestartStatus();
    fetchMigrationStatus();

    // 设置定时检查
    const interval = setInterval(() => {
      fetchUpdateProgress();
      fetchRestartStatus();
    }, 5000);

    return () => clearInterval(interval);
  }, [fetchCurrentVersion, checkForUpdates, fetchUpdateProgress, fetchRestartStatus, fetchMigrationStatus]);

  useEffect(() => {
    if (activeTab === 'history') {
      fetchUpdateHistory();
    } else if (activeTab === 'migration') {
      fetchMigrationStatus();
      fetchPendingMigrations();
    }
  }, [activeTab, fetchUpdateHistory, fetchMigrationStatus, fetchPendingMigrations]);

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Zap className="w-8 h-8 text-blue-500" />
          版本更新管理
        </h1>
        <div className="flex gap-2">
          <Button
            variant={activeTab === 'overview' ? 'default' : 'outline'}
            onClick={() => setActiveTab('overview')}
            className={activeTab === 'overview' ? 'bg-binance-yellow text-black hover:bg-binance-yellow-light' : 'border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10'}
          >
            总览
          </Button>
          <Button
            variant={activeTab === 'history' ? 'default' : 'outline'}
            onClick={() => setActiveTab('history')}
            className={activeTab === 'history' ? 'bg-binance-yellow text-black hover:bg-binance-yellow-light' : 'border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10'}
          >
            <History className="w-4 h-4 mr-2" />
            历史记录
          </Button>
          <Button
            variant={activeTab === 'settings' ? 'default' : 'outline'}
            onClick={() => setActiveTab('settings')}
            className={activeTab === 'settings' ? 'bg-binance-yellow text-black hover:bg-binance-yellow-light' : 'border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10'}
          >
            <Settings className="w-4 h-4 mr-2" />
            设置
          </Button>
          <Button
            variant={activeTab === 'migration' ? 'default' : 'outline'}
            onClick={() => setActiveTab('migration')}
            className={activeTab === 'migration' ? 'bg-binance-yellow text-black hover:bg-binance-yellow-light' : 'border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10'}
          >
            <Database className="w-4 h-4 mr-2" />
            迁移管理
          </Button>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {activeTab === 'overview' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 当前版本信息 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                当前版本
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {currentVersion && (
                <>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">版本号</span>
                    <Badge variant="secondary">{currentVersion.version}</Badge>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">构建时间</span>
                    <span className="text-sm text-gray-600">{currentVersion.build_time}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">平台</span>
                    <span className="text-sm text-gray-600">{currentVersion.platform}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">发布时间</span>
                    <span className="text-sm text-gray-600">{currentVersion.release_date}</span>
                  </div>
                </>
              )}
            </CardContent>
          </Card>

          {/* 更新状态 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span className="flex items-center gap-2">
                  <RefreshCw className="w-5 h-5 text-blue-500" />
                  更新状态
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={checkForUpdates}
                  disabled={isLoading}
                  className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                >
                  <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                  检查更新
                </Button>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {updateStatus && (
                <>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">最新版本</span>
                    <div className="flex items-center gap-2">
                      {updateStatus.has_update && (
                        <Badge variant="destructive">有更新</Badge>
                      )}
                      <span className="text-sm">
                        {updateStatus.latest_ver || updateStatus.current_ver}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">自动更新</span>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => toggleAutoUpdate(!updateStatus.auto_update_enabled)}
                      className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                    >
                      {updateStatus.auto_update_enabled ? '已启用' : '已禁用'}
                    </Button>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">上次检查</span>
                    <span className="text-sm text-gray-600">
                      {new Date(updateStatus.last_check).toLocaleString()}
                    </span>
                  </div>
                </>
              )}
            </CardContent>
          </Card>

          {/* 更新进度 */}
          {updateProgress && updateProgress.status !== 'idle' && (
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  {getStatusIcon(updateProgress.status)}
                  更新进度
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium">{updateProgress.message}</span>
                    <span className="text-sm text-gray-600">{updateProgress.progress.toFixed(1)}%</span>
                  </div>
                  <Progress value={updateProgress.progress} className="w-full" />
                </div>

                {updateProgress.speed && (
                  <div className="flex items-center justify-between text-sm">
                    <span>下载速度</span>
                    <span>{formatFileSize(updateProgress.speed)}/s</span>
                  </div>
                )}

                {updateProgress.total_size && updateProgress.downloaded && (
                  <div className="flex items-center justify-between text-sm">
                    <span>下载进度</span>
                    <span>
                      {formatFileSize(updateProgress.downloaded)} / {formatFileSize(updateProgress.total_size)}
                    </span>
                  </div>
                )}

                {updateProgress.eta && (
                  <div className="flex items-center justify-between text-sm">
                    <span>预计剩余时间</span>
                    <span>{formatTime(updateProgress.eta)}</span>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* 重启状态 */}
          {restartStatus && restartStatus.is_counting_down && (
            <Card className="lg:col-span-2 border-orange-200 bg-orange-50">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-orange-600">
                  <RotateCcw className="w-5 h-5" />
                  系统重启倒计时
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="font-medium">重启原因</span>
                  <span className="text-sm">{restartStatus.reason}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-medium">倒计时</span>
                  <span className="text-lg font-mono text-orange-600">{restartStatus.countdown_time}秒</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-medium">重启时间</span>
                  <span className="text-sm">{new Date(restartStatus.restart_time).toLocaleString()}</span>
                </div>
                {restartStatus.can_cancel && (
                  <Button
                    variant="outline"
                    onClick={cancelRestart}
                    className="w-full border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                  >
                    取消重启
                  </Button>
                )}
              </CardContent>
            </Card>
          )}

          {/* 更新操作 */}
          {updateStatus && updateStatus.has_update && (
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Download className="w-5 h-5" />
                  更新操作
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {updateStatus.update_info && (
                  <div>
                    <h4 className="font-medium mb-2">更新说明</h4>
                    <p className="text-sm text-gray-600 whitespace-pre-line">
                      {updateStatus.update_info.release_notes}
                    </p>
                  </div>
                )}

                <div className="flex gap-3">
                  <Button
                    onClick={() => downloadUpdate(false)}
                    disabled={isLoading || (updateProgress && updateProgress.status !== 'idle')}
                  >
                    <Download className="w-4 h-4 mr-2" />
                    下载更新
                  </Button>
                  <Button
                    onClick={() => downloadUpdate(true)}
                    variant="outline"
                    disabled={isLoading || (updateProgress && updateProgress.status !== 'idle')}
                    className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                  >
                    <Download className="w-4 h-4 mr-2" />
                    下载并重启
                  </Button>
                </div>

                <div className="flex gap-3">
                  <Button
                    onClick={installUpdate}
                    disabled={isLoading || (updateProgress && updateProgress.status !== 'idle')}
                  >
                    <Zap className="w-4 h-4 mr-2" />
                    安装更新
                  </Button>
                  <Button
                    onClick={() => restartApplication()}
                    variant="outline"
                    disabled={isLoading || (restartStatus && restartStatus.is_counting_down)}
                    className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                  >
                    <RotateCcw className="w-4 h-4 mr-2" />
                    重启应用
                  </Button>
                </div>

                <Alert className="mt-4">
                  <AlertTriangle className="h-4 w-4" />
                  <AlertDescription>
                    <strong>自动更新包含：</strong>
                    如果新版本需要数据库结构变更，系统会自动创建备份并执行数据库迁移，无需手动操作。
                  </AlertDescription>
                </Alert>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {activeTab === 'history' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <History className="w-5 h-5" />
              更新历史
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {updateHistory.map((item, index) => (
                <div key={index} className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <div className="font-medium">版本 {item.version}</div>
                    <div className="text-sm text-gray-600">{item.reason}</div>
                    <div className="text-xs text-gray-500">
                      {new Date(item.restart_time).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {item.success ? (
                      <CheckCircle className="w-5 h-5 text-green-500" />
                    ) : (
                      <AlertTriangle className="w-5 h-5 text-red-500" />
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {activeTab === 'migration' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 迁移状态 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="w-5 h-5 text-blue-500" />
                数据库迁移状态
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {migrationStatus && (
                <>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">当前版本</span>
                    <Badge variant="secondary">{migrationStatus.current_version}</Badge>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">待执行迁移</span>
                    <div className="flex items-center gap-2">
                      {migrationStatus.needs_migration && (
                        <Badge variant="destructive">需要迁移</Badge>
                      )}
                      <span className="text-sm">{migrationStatus.pending_migrations} 个</span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">总迁移数</span>
                    <span className="text-sm text-gray-600">{migrationStatus.total_migrations}</span>
                  </div>
                </>
              )}
              <div className="flex gap-3 pt-4">
                <Button
                  onClick={() => fetchMigrationStatus()}
                  variant="outline"
                  className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                >
                  <RefreshCw className="w-4 h-4 mr-2" />
                  刷新状态
                </Button>
                <Button
                  onClick={createBackup}
                  disabled={isMigrationLoading}
                  variant="outline"
                  className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                >
                  <Save className="w-4 h-4 mr-2" />
                  创建备份
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* 迁移操作 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <ArrowUp className="w-5 h-5 text-green-500" />
                迁移操作
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-gray-600">
                执行数据库迁移前，系统会自动创建备份以确保数据安全。
              </p>
              <div className="space-y-3">
                {pendingMigrations && pendingMigrations.length > 0 ? (
                  pendingMigrations.map((migration, index) => (
                    <div key={index} className="border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <h4 className="font-medium">{migration.name}</h4>
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">{migration.version}</Badge>
                          {migration.is_critical && (
                            <Badge variant="destructive">关键</Badge>
                          )}
                        </div>
                      </div>
                      <p className="text-sm text-gray-600 mb-3">{migration.description}</p>
                      <div className="flex gap-2">
                        <Button
                          onClick={() => executeMigration(migration.version)}
                          disabled={isMigrationLoading}
                          className="flex-1"
                        >
                          <ArrowUp className="w-4 h-4 mr-2" />
                          执行迁移
                        </Button>
                        <Button
                          onClick={() => {
                            if (confirm(`确定要回滚到版本 ${migration.version} 吗？\n\n此操作将撤销该版本之后的所有迁移。`)) {
                              rollbackMigration(migration.version);
                            }
                          }}
                          disabled={isMigrationLoading}
                          variant="outline"
                          className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                        >
                          <ArrowDown className="w-4 h-4 mr-2" />
                          回滚
                        </Button>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    <CheckCircle className="w-12 h-12 mx-auto mb-4 text-green-500" />
                    <p>数据库已是最新版本，无需执行迁移。</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* 迁移历史详情 */}
          {migrationStatus && migrationStatus.pending_details && migrationStatus.pending_details.length > 0 && (
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <History className="w-5 h-5 text-purple-500" />
                  待执行迁移详情
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {migrationStatus.pending_details.map((detail, index) => (
                    <div key={index} className="border-l-4 border-purple-300 pl-4">
                      <div className="flex items-center justify-between mb-2">
                        <h4 className="font-medium flex items-center gap-2">
                          {detail.name}
                          {detail.is_critical && (
                            <Badge variant="destructive" className="text-xs">关键迁移</Badge>
                          )}
                        </h4>
                        <Badge variant="outline">{detail.version}</Badge>
                      </div>
                      <p className="text-sm text-gray-600">{detail.description}</p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {activeTab === 'settings' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>自动更新设置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="font-medium">启用自动更新</span>
                <Button
                  variant="outline"
                  onClick={() => updateStatus && toggleAutoUpdate(!updateStatus.auto_update_enabled)}
                  className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10"
                >
                  {updateStatus?.auto_update_enabled ? '已启用' : '已禁用'}
                </Button>
              </div>
              <p className="text-sm text-gray-600">
                启用后，系统会自动检查并下载更新。关键更新会自动应用。
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>手动重启</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-gray-600">
                手动重启应用，通常用于解决临时问题或应用配置更改。
              </p>
              <div className="flex gap-3">
                <Button onClick={() => restartApplication(10)}>
                  立即重启
                </Button>
                <Button onClick={() => restartApplication(60)} variant="outline" className="border-binance-yellow text-binance-yellow hover:bg-binance-yellow hover:bg-opacity-10">
                  1分钟后重启
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
};

export default VersionUpdate;