// =============================================================================
// Notification Service - プッシュ通知サービス
// =============================================================================
// Expo Notifications を使用してプッシュ通知を管理します。
// - デバイストークンの取得と登録
// - 通知パーミッションの管理
// - フォアグラウンド/バックグラウンド通知ハンドリング

import * as Notifications from 'expo-notifications';
import * as Device from 'expo-device';
import Constants from 'expo-constants';
import { Platform } from 'react-native';
import { api } from './api';

// =============================================================================
// 通知設定
// =============================================================================

// フォアグラウンド通知の表示設定
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
  }),
});

// =============================================================================
// 型定義
// =============================================================================

export interface NotificationData {
  type?: string;
  taskId?: number;
  cropId?: number;
  [key: string]: unknown;
}

export interface PushNotificationState {
  expoPushToken: string | null;
  notification: Notifications.Notification | null;
}

// =============================================================================
// デバイストークン取得
// =============================================================================

/**
 * Expo Push Token を取得します。
 * 物理デバイスでのみ動作します（シミュレータでは動作しません）。
 *
 * @returns Expo Push Token または null
 */
export async function getExpoPushToken(): Promise<string | null> {
  // 物理デバイスかチェック
  if (!Device.isDevice) {
    console.log('プッシュ通知は物理デバイスでのみ利用可能です');
    return null;
  }

  try {
    // 既存のパーミッションをチェック
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;

    // パーミッションがない場合はリクエスト
    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }

    if (finalStatus !== 'granted') {
      console.log('プッシュ通知のパーミッションが拒否されました');
      return null;
    }

    // Expo Push Token を取得
    const projectId = Constants.expoConfig?.extra?.eas?.projectId;
    const tokenData = await Notifications.getExpoPushTokenAsync({
      projectId: projectId,
    });

    return tokenData.data;
  } catch (error) {
    console.error('Push Token の取得に失敗しました:', error);
    return null;
  }
}

/**
 * FCM トークンを取得します（Android用）。
 * Expo Managed Workflow では Expo Push Token を使用するため、
 * この関数は将来の bare workflow 移行時に使用します。
 *
 * @returns FCM Token または null
 */
export async function getFCMToken(): Promise<string | null> {
  if (Platform.OS !== 'android') {
    return null;
  }

  try {
    const token = await Notifications.getDevicePushTokenAsync();
    return token.data;
  } catch (error) {
    console.error('FCM Token の取得に失敗しました:', error);
    return null;
  }
}

// =============================================================================
// デバイストークン登録
// =============================================================================

/**
 * デバイストークンをバックエンドに登録します。
 *
 * @param token - Push Token
 * @returns 登録成功の場合 true
 */
export async function registerDeviceToken(token: string): Promise<boolean> {
  try {
    const platform = Platform.OS === 'ios' ? 'ios' : 'android';

    await api.post('/notifications/device-token', {
      token,
      platform,
      device_id: Device.modelId || undefined,
    });

    console.log('デバイストークンを登録しました');
    return true;
  } catch (error) {
    console.error('デバイストークンの登録に失敗しました:', error);
    return false;
  }
}

/**
 * デバイストークンをバックエンドから削除します。
 * ログアウト時に呼び出します。
 *
 * @returns 削除成功の場合 true
 */
export async function unregisterDeviceToken(): Promise<boolean> {
  try {
    const platform = Platform.OS === 'ios' ? 'ios' : 'android';

    await api.delete(`/notifications/device-token?platform=${platform}`);

    console.log('デバイストークンを削除しました');
    return true;
  } catch (error) {
    console.error('デバイストークンの削除に失敗しました:', error);
    return false;
  }
}

// =============================================================================
// 通知パーミッション
// =============================================================================

/**
 * 通知パーミッションの状態を確認します。
 *
 * @returns パーミッションが許可されている場合 true
 */
export async function checkNotificationPermission(): Promise<boolean> {
  const { status } = await Notifications.getPermissionsAsync();
  return status === 'granted';
}

/**
 * 通知パーミッションをリクエストします。
 *
 * @returns パーミッションが許可された場合 true
 */
export async function requestNotificationPermission(): Promise<boolean> {
  const { status } = await Notifications.requestPermissionsAsync();
  return status === 'granted';
}

// =============================================================================
// 通知リスナー
// =============================================================================

/**
 * 通知受信リスナーを登録します。
 * フォアグラウンドで通知を受信した際に呼び出されます。
 *
 * @param callback - 通知受信時のコールバック
 * @returns リスナーの購読解除関数
 */
export function addNotificationReceivedListener(
  callback: (notification: Notifications.Notification) => void
): Notifications.Subscription {
  return Notifications.addNotificationReceivedListener(callback);
}

/**
 * 通知タップリスナーを登録します。
 * ユーザーが通知をタップした際に呼び出されます。
 *
 * @param callback - 通知タップ時のコールバック
 * @returns リスナーの購読解除関数
 */
export function addNotificationResponseListener(
  callback: (response: Notifications.NotificationResponse) => void
): Notifications.Subscription {
  return Notifications.addNotificationResponseReceivedListener(callback);
}

// =============================================================================
// 通知データ解析
// =============================================================================

/**
 * 通知レスポンスからデータを抽出します。
 *
 * @param response - 通知レスポンス
 * @returns 通知データ
 */
export function extractNotificationData(
  response: Notifications.NotificationResponse
): NotificationData {
  const data = response.notification.request.content.data as NotificationData;
  return data || {};
}

/**
 * 通知タイプに基づいて画面遷移先を決定します。
 *
 * @param data - 通知データ
 * @returns 遷移先のルート名とパラメータ
 */
export function getNavigationTarget(
  data: NotificationData
): { route: string; params?: Record<string, unknown> } | null {
  switch (data.type) {
    case 'task_due_reminder':
    case 'task_overdue_alert':
      return {
        route: 'Tasks',
        params: data.taskId ? { taskId: data.taskId } : undefined,
      };

    case 'harvest_reminder':
      return {
        route: 'Crops',
        params: data.cropId ? { cropId: data.cropId } : undefined,
      };

    default:
      return null;
  }
}

// =============================================================================
// Android チャンネル設定
// =============================================================================

/**
 * Android 用の通知チャンネルを設定します。
 * アプリ起動時に一度呼び出します。
 */
export async function setupAndroidNotificationChannel(): Promise<void> {
  if (Platform.OS !== 'android') {
    return;
  }

  // タスクリマインダーチャンネル
  await Notifications.setNotificationChannelAsync('task-reminders', {
    name: 'タスクリマインダー',
    description: 'タスクの期限に関する通知',
    importance: Notifications.AndroidImportance.HIGH,
    vibrationPattern: [0, 250, 250, 250],
    lightColor: '#16a34a',
  });

  // 収穫リマインダーチャンネル
  await Notifications.setNotificationChannelAsync('harvest-reminders', {
    name: '収穫リマインダー',
    description: '収穫時期に関する通知',
    importance: Notifications.AndroidImportance.DEFAULT,
  });

  // 一般通知チャンネル
  await Notifications.setNotificationChannelAsync('default', {
    name: '一般',
    description: 'その他の通知',
    importance: Notifications.AndroidImportance.DEFAULT,
  });
}

// =============================================================================
// 初期化
// =============================================================================

/**
 * 通知サービスを初期化します。
 * アプリ起動時に呼び出します。
 *
 * @returns 初期化成功の場合 true
 */
export async function initializeNotifications(): Promise<boolean> {
  try {
    // Android チャンネル設定
    await setupAndroidNotificationChannel();

    // Push Token を取得
    const token = await getExpoPushToken();

    if (token) {
      // バックエンドに登録
      await registerDeviceToken(token);
      return true;
    }

    return false;
  } catch (error) {
    console.error('通知サービスの初期化に失敗しました:', error);
    return false;
  }
}
