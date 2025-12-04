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

// Suppress console warnings during tests
global.console = {
  ...console,
  warn: jest.fn(),
  error: jest.fn(),
};
