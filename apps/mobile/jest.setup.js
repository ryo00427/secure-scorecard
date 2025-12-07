// Jest setup for React Native + Expo
import '@testing-library/react-native/extend-expect';

// Mock expo modules
jest.mock('expo-status-bar', () => ({
  StatusBar: 'StatusBar',
}));

jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
  MaterialIcons: 'MaterialIcons',
  FontAwesome: 'FontAwesome',
}));

// expo-notifications モック
jest.mock('expo-notifications', () => ({
  setNotificationHandler: jest.fn(),
  getExpoPushTokenAsync: jest.fn().mockResolvedValue({
    data: 'ExponentPushToken[xxxxxxxxxxxxxx]',
  }),
  getDevicePushTokenAsync: jest.fn().mockResolvedValue({
    data: 'fcm-token-xxx',
  }),
  getPermissionsAsync: jest.fn().mockResolvedValue({
    status: 'granted',
  }),
  requestPermissionsAsync: jest.fn().mockResolvedValue({
    status: 'granted',
  }),
  addNotificationReceivedListener: jest.fn().mockReturnValue({
    remove: jest.fn(),
  }),
  addNotificationResponseReceivedListener: jest.fn().mockReturnValue({
    remove: jest.fn(),
  }),
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

// react-native-chart-kit モック
jest.mock('react-native-chart-kit', () => ({
  BarChart: 'BarChart',
  PieChart: 'PieChart',
  LineChart: 'LineChart',
}));

// react-native-svg モック
jest.mock('react-native-svg', () => ({
  Svg: 'Svg',
  Circle: 'Circle',
  Ellipse: 'Ellipse',
  G: 'G',
  Text: 'Text',
  TSpan: 'TSpan',
  TextPath: 'TextPath',
  Path: 'Path',
  Polygon: 'Polygon',
  Polyline: 'Polyline',
  Line: 'Line',
  Rect: 'Rect',
  Use: 'Use',
  Image: 'Image',
  Symbol: 'Symbol',
  Defs: 'Defs',
  LinearGradient: 'LinearGradient',
  RadialGradient: 'RadialGradient',
  Stop: 'Stop',
  ClipPath: 'ClipPath',
  Pattern: 'Pattern',
  Mask: 'Mask',
}));

// Suppress console warnings during tests
global.console = {
  ...console,
  warn: jest.fn(),
  error: jest.fn(),
};
