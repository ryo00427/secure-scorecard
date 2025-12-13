// =============================================================================
// AppNavigator - アプリケーションナビゲーション
// =============================================================================
// 認証状態に応じてナビゲーションを切り替えます。
// 未認証: 認証スタック（ログイン/登録画面）
// 認証済み: メインタブナビゲーション

import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { ActivityIndicator, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../context/AuthContext';

// スクリーンのインポート
import LoginScreen from '../screens/auth/LoginScreen';
import RegisterScreen from '../screens/auth/RegisterScreen';
import HomeScreen from '../screens/main/HomeScreen';
import TasksScreen from '../screens/main/TasksScreen';
import CropsScreen from '../screens/main/CropsScreen';
import AnalyticsScreen from '../screens/main/AnalyticsScreen';
import SettingsScreen from '../screens/main/SettingsScreen';

// 認証スタックのパラメータ
export type AuthStackParamList = {
  Login: undefined;
  Register: undefined;
};

// メインタブのパラメータ
export type MainTabParamList = {
  Home: undefined;
  Tasks: undefined;
  Crops: undefined;
  Analytics: undefined;
  Settings: undefined;
};

// ルートスタックのパラメータ
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
};

const Stack = createNativeStackNavigator<RootStackParamList>();
const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const Tab = createBottomTabNavigator<MainTabParamList>();

// 認証ナビゲーター
function AuthNavigator() {
  return (
    <AuthStack.Navigator
      screenOptions={{
        headerShown: false,
      }}
    >
      <AuthStack.Screen name="Login" component={LoginScreen} />
      <AuthStack.Screen name="Register" component={RegisterScreen} />
    </AuthStack.Navigator>
  );
}

// メインタブナビゲーター
function MainTabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        tabBarIcon: ({ focused, color, size }) => {
          let iconName: keyof typeof Ionicons.glyphMap = 'home';

          switch (route.name) {
            case 'Home':
              iconName = focused ? 'home' : 'home-outline';
              break;
            case 'Tasks':
              iconName = focused ? 'checkbox' : 'checkbox-outline';
              break;
            case 'Crops':
              iconName = focused ? 'leaf' : 'leaf-outline';
              break;
            case 'Analytics':
              iconName = focused ? 'analytics' : 'analytics-outline';
              break;
            case 'Settings':
              iconName = focused ? 'settings' : 'settings-outline';
              break;
          }

          return <Ionicons name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: '#16a34a', // primary-600
        tabBarInactiveTintColor: 'gray',
        headerStyle: {
          backgroundColor: '#16a34a',
        },
        headerTintColor: '#fff',
        headerTitleStyle: {
          fontWeight: 'bold',
        },
      })}
    >
      <Tab.Screen
        name="Home"
        component={HomeScreen}
        options={{ title: 'ホーム' }}
      />
      <Tab.Screen
        name="Tasks"
        component={TasksScreen}
        options={{ title: 'タスク' }}
      />
      <Tab.Screen
        name="Crops"
        component={CropsScreen}
        options={{ title: '作物' }}
      />
      <Tab.Screen
        name="Analytics"
        component={AnalyticsScreen}
        options={{ title: '分析' }}
      />
      <Tab.Screen
        name="Settings"
        component={SettingsScreen}
        options={{ title: '設定' }}
      />
    </Tab.Navigator>
  );
}

export default function AppNavigator() {
  const { isLoading, isAuthenticated } = useAuth();

  // 認証情報の読み込み中はローディング表示
  if (isLoading) {
    return (
      <View className="flex-1 items-center justify-center bg-white">
        <ActivityIndicator size="large" color="#16a34a" />
      </View>
    );
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {isAuthenticated ? (
          <Stack.Screen name="Main" component={MainTabNavigator} />
        ) : (
          <Stack.Screen name="Auth" component={AuthNavigator} />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
