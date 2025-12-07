// =============================================================================
// NotificationContext - プッシュ通知管理コンテキスト
// =============================================================================
// Expo Notifications を使用してプッシュ通知を管理します。
// - デバイストークンの取得・登録
// - 通知受信リスナーの管理
// - 通知タップ時のナビゲーション

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  ReactNode,
} from 'react';
import { useNavigation, NavigationProp } from '@react-navigation/native';
import * as Notifications from 'expo-notifications';
import {
  getExpoPushToken,
  registerDeviceToken,
  unregisterDeviceToken,
  addNotificationReceivedListener,
  addNotificationResponseListener,
  extractNotificationData,
  getNavigationTarget,
  initializeNotifications,
  setupAndroidNotificationChannel,
  NotificationData,
} from '../services/notifications';
import { MainTabParamList } from '../navigation/AppNavigator';

// -----------------------------------------------------------------------------
// Types - 型定義
// -----------------------------------------------------------------------------

// 通知コンテキストの型
interface NotificationContextType {
  // 状態
  expoPushToken: string | null;
  isInitialized: boolean;
  lastNotification: Notifications.Notification | null;

  // メソッド
  initialize: () => Promise<boolean>;
  cleanup: () => Promise<void>;
}

// -----------------------------------------------------------------------------
// Context - コンテキスト
// -----------------------------------------------------------------------------

const NotificationContext = createContext<NotificationContextType | undefined>(
  undefined
);

// -----------------------------------------------------------------------------
// Provider - プロバイダーコンポーネント
// -----------------------------------------------------------------------------

interface NotificationProviderProps {
  children: ReactNode;
}

export function NotificationProvider({ children }: NotificationProviderProps) {
  const [expoPushToken, setExpoPushToken] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [lastNotification, setLastNotification] =
    useState<Notifications.Notification | null>(null);

  // ナビゲーション（メインタブ用）
  // 注意: このフックはナビゲーションコンテキスト内で使用する必要があります
  let navigation: NavigationProp<MainTabParamList> | null = null;
  try {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    navigation = useNavigation<NavigationProp<MainTabParamList>>();
  } catch {
    // NavigationContainer の外で使用された場合は null
    navigation = null;
  }

  // ---------------------------------------------------------------------------
  // Handlers - ハンドラー
  // ---------------------------------------------------------------------------

  // 通知タップ時のハンドラー
  const handleNotificationResponse = useCallback(
    (response: Notifications.NotificationResponse) => {
      console.log('通知がタップされました:', response);

      // 通知データを抽出
      const data = extractNotificationData(response);

      // ナビゲーション先を決定
      const target = getNavigationTarget(data);

      if (target && navigation) {
        // 型安全なナビゲーション
        const routeName = target.route as keyof MainTabParamList;
        navigation.navigate(routeName);
      }
    },
    [navigation]
  );

  // フォアグラウンド通知受信時のハンドラー
  const handleNotificationReceived = useCallback(
    (notification: Notifications.Notification) => {
      console.log('フォアグラウンドで通知を受信:', notification);
      setLastNotification(notification);
    },
    []
  );

  // ---------------------------------------------------------------------------
  // Methods - メソッド
  // ---------------------------------------------------------------------------

  // 通知サービスを初期化
  const initialize = useCallback(async (): Promise<boolean> => {
    try {
      console.log('通知サービスを初期化中...');

      // Android チャンネル設定
      await setupAndroidNotificationChannel();

      // Push Token を取得
      const token = await getExpoPushToken();

      if (token) {
        setExpoPushToken(token);

        // バックエンドに登録
        await registerDeviceToken(token);
        console.log('通知サービスの初期化が完了しました');
        setIsInitialized(true);
        return true;
      }

      console.log('Push Token を取得できませんでした（シミュレータの場合は正常）');
      setIsInitialized(true);
      return false;
    } catch (error) {
      console.error('通知サービスの初期化に失敗しました:', error);
      setIsInitialized(true);
      return false;
    }
  }, []);

  // クリーンアップ（ログアウト時に呼び出す）
  const cleanup = useCallback(async (): Promise<void> => {
    try {
      await unregisterDeviceToken();
      setExpoPushToken(null);
      setLastNotification(null);
      console.log('通知サービスのクリーンアップが完了しました');
    } catch (error) {
      console.error('通知サービスのクリーンアップに失敗しました:', error);
    }
  }, []);

  // ---------------------------------------------------------------------------
  // Effects - 副作用
  // ---------------------------------------------------------------------------

  // リスナーの設定
  useEffect(() => {
    // フォアグラウンド通知リスナー
    const receivedSubscription = addNotificationReceivedListener(
      handleNotificationReceived
    );

    // 通知タップリスナー
    const responseSubscription = addNotificationResponseListener(
      handleNotificationResponse
    );

    // クリーンアップ
    return () => {
      receivedSubscription.remove();
      responseSubscription.remove();
    };
  }, [handleNotificationReceived, handleNotificationResponse]);

  // ---------------------------------------------------------------------------
  // Render - レンダリング
  // ---------------------------------------------------------------------------

  const value: NotificationContextType = {
    expoPushToken,
    isInitialized,
    lastNotification,
    initialize,
    cleanup,
  };

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
}

// -----------------------------------------------------------------------------
// Hook - カスタムフック
// -----------------------------------------------------------------------------

/**
 * 通知コンテキストを使用するためのカスタムフック
 *
 * @returns NotificationContextType
 * @throws Error NotificationProvider の外で使用された場合
 */
export function useNotifications() {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error(
      'useNotifications must be used within a NotificationProvider'
    );
  }
  return context;
}
