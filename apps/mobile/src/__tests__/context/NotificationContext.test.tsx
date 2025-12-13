// =============================================================================
// NotificationContext テスト
// =============================================================================
// 通知コンテキストの機能をテストします。

import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react-native';
import { Text, View } from 'react-native';
import {
  NotificationProvider,
  useNotifications,
} from '../../context/NotificationContext';

// expo-notifications モック
const mockGetExpoPushTokenAsync = jest.fn().mockResolvedValue({
  data: 'ExponentPushToken[xxxxxxxxxxxxxx]',
});

const mockGetPermissionsAsync = jest.fn().mockResolvedValue({
  status: 'granted',
});

const mockAddNotificationReceivedListener = jest.fn().mockReturnValue({
  remove: jest.fn(),
});

const mockAddNotificationResponseReceivedListener = jest.fn().mockReturnValue({
  remove: jest.fn(),
});

jest.mock('expo-notifications', () => ({
  setNotificationHandler: jest.fn(),
  getExpoPushTokenAsync: () => mockGetExpoPushTokenAsync(),
  getPermissionsAsync: () => mockGetPermissionsAsync(),
  requestPermissionsAsync: jest.fn().mockResolvedValue({ status: 'granted' }),
  addNotificationReceivedListener: (cb: unknown) =>
    mockAddNotificationReceivedListener(cb),
  addNotificationResponseReceivedListener: (cb: unknown) =>
    mockAddNotificationResponseReceivedListener(cb),
  setNotificationChannelAsync: jest.fn().mockResolvedValue(undefined),
  AndroidImportance: {
    HIGH: 4,
    DEFAULT: 3,
  },
}));

// expo-device モック
jest.mock('expo-device', () => ({
  isDevice: true,
  modelId: 'test-device-id',
}));

// expo-constants モック
jest.mock('expo-constants', () => ({
  expoConfig: {
    extra: {
      eas: {
        projectId: 'test-project-id',
      },
    },
  },
}));

// API モック
jest.mock('../../services/api', () => ({
  api: {
    post: jest.fn().mockResolvedValue({ message: 'success' }),
    delete: jest.fn().mockResolvedValue({ message: 'success' }),
  },
}));

// react-navigation モック
jest.mock('@react-navigation/native', () => ({
  useNavigation: () => ({
    navigate: jest.fn(),
  }),
}));

// notifications service モック
jest.mock('../../services/notifications', () => ({
  getExpoPushToken: jest.fn().mockResolvedValue('ExponentPushToken[xxx]'),
  registerDeviceToken: jest.fn().mockResolvedValue(true),
  unregisterDeviceToken: jest.fn().mockResolvedValue(true),
  addNotificationReceivedListener: jest
    .fn()
    .mockReturnValue({ remove: jest.fn() }),
  addNotificationResponseListener: jest
    .fn()
    .mockReturnValue({ remove: jest.fn() }),
  extractNotificationData: jest.fn().mockReturnValue({}),
  getNavigationTarget: jest.fn().mockReturnValue(null),
  initializeNotifications: jest.fn().mockResolvedValue(true),
  setupAndroidNotificationChannel: jest.fn().mockResolvedValue(undefined),
}));

function TestComponent() {
  const { expoPushToken, isInitialized, initialize } = useNotifications();

  return (
    <View>
      <Text testID="token">{expoPushToken || 'no-token'}</Text>
      <Text testID="initialized">{isInitialized ? 'yes' : 'no'}</Text>
    </View>
  );
}

describe('NotificationContext', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('NotificationProviderがレンダリングされる', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('token')).toBeTruthy();
    });
  });

  it('初期状態ではトークンがない', async () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('token').props.children).toBe('no-token');
    });
  });

  it('通知リスナーが設定される', async () => {
    const { getExpoPushToken } = require('../../services/notifications');

    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('token')).toBeTruthy();
    });

    // リスナーが設定されていることを確認
    const {
      addNotificationReceivedListener,
      addNotificationResponseListener,
    } = require('../../services/notifications');
    expect(addNotificationReceivedListener).toHaveBeenCalled();
    expect(addNotificationResponseListener).toHaveBeenCalled();
  });

  it('NotificationProviderの外でuseNotificationsを使うとエラー', () => {
    // コンソールエラーを抑制
    const consoleSpy = jest
      .spyOn(console, 'error')
      .mockImplementation(() => {});

    expect(() => {
      render(<TestComponent />);
    }).toThrow('useNotifications must be used within a NotificationProvider');

    consoleSpy.mockRestore();
  });
});
