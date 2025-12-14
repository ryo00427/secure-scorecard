// =============================================================================
// AppNavigator - アプリケーションナビゲーション
// =============================================================================
// デザインファイル: design/stitch_ (4)/screen.png のボトムナビ参照
// 認証状態に応じてナビゲーションを切り替えます。
// ボトムタブ: ホーム、マイプラント、カレンダー、設定

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
import CropsScreen from '../screens/main/CropsScreen';
import CropDetailScreen from '../screens/main/CropDetailScreen';
import AddCropScreen from '../screens/main/AddCropScreen';
import CalendarScreen from '../screens/main/CalendarScreen';
import TasksScreen from '../screens/main/TasksScreen';
import WorkLogScreen from '../screens/main/WorkLogScreen';
import SettingsScreen from '../screens/main/SettingsScreen';

// 認証スタックのパラメータ
export type AuthStackParamList = {
  Login: undefined;
  Register: undefined;
};

// メインスタックのパラメータ
export type MainStackParamList = {
  MainTabs: undefined;
  CropDetail: { cropId: number };
  AddCrop: undefined;
  EditCrop: { cropId: number };
  WorkLog: { cropId?: number };
  AddTask: { date?: string };
  TaskDetail: { taskId: number };
};

// メインタブのパラメータ
export type MainTabParamList = {
  Home: undefined;
  MyPlants: undefined;
  Tasks: undefined;
  Calendar: undefined;
  Settings: undefined;
};

// ルートスタックのパラメータ
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
};

const Stack = createNativeStackNavigator<RootStackParamList>();
const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const MainStack = createNativeStackNavigator<MainStackParamList>();
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
            case 'MyPlants':
              iconName = focused ? 'leaf' : 'leaf-outline';
              break;
            case 'Tasks':
              iconName = focused ? 'checkmark-circle' : 'checkmark-circle-outline';
              break;
            case 'Calendar':
              iconName = focused ? 'calendar' : 'calendar-outline';
              break;
            case 'Settings':
              iconName = focused ? 'settings' : 'settings-outline';
              break;
          }

          return <Ionicons name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: '#22c55e', // emerald-500
        tabBarInactiveTintColor: '#6b7280', // gray-500
        tabBarStyle: {
          backgroundColor: '#ffffff',
          borderTopWidth: 1,
          borderTopColor: '#f3f4f6',
          paddingTop: 4,
          paddingBottom: 4,
          height: 60,
        },
        tabBarLabelStyle: {
          fontSize: 11,
          fontWeight: '500',
        },
        headerShown: false,
      })}
    >
      <Tab.Screen
        name="Home"
        component={HomeScreen}
        options={{ title: 'ホーム' }}
      />
      <Tab.Screen
        name="MyPlants"
        component={CropsScreen}
        options={{ title: 'マイプラント' }}
      />
      <Tab.Screen
        name="Tasks"
        component={TasksScreen}
        options={{ title: 'タスク' }}
      />
      <Tab.Screen
        name="Calendar"
        component={CalendarScreen}
        options={{ title: 'カレンダー' }}
      />
      <Tab.Screen
        name="Settings"
        component={SettingsScreen}
        options={{ title: '設定' }}
      />
    </Tab.Navigator>
  );
}

// メインスタックナビゲーター（タブ + モーダル/詳細画面）
function MainNavigator() {
  return (
    <MainStack.Navigator
      screenOptions={{
        headerShown: false,
      }}
    >
      {/* メインタブ */}
      <MainStack.Screen name="MainTabs" component={MainTabNavigator} />

      {/* 詳細画面 */}
      <MainStack.Screen
        name="CropDetail"
        component={CropDetailScreen}
        options={{
          animation: 'slide_from_right',
        }}
      />

      {/* 作物追加画面 */}
      <MainStack.Screen
        name="AddCrop"
        component={AddCropScreen}
        options={{
          animation: 'slide_from_bottom',
          presentation: 'modal',
        }}
      />

      {/* 作物編集画面（AddCropを再利用） */}
      <MainStack.Screen
        name="EditCrop"
        component={AddCropScreen}
        options={{
          animation: 'slide_from_right',
        }}
      />

      {/* 作業ログ画面 */}
      <MainStack.Screen
        name="WorkLog"
        component={WorkLogScreen}
        options={{
          animation: 'slide_from_bottom',
          presentation: 'modal',
        }}
      />
    </MainStack.Navigator>
  );
}

export default function AppNavigator() {
  const { isLoading, isAuthenticated } = useAuth();

  // 認証情報の読み込み中はローディング表示
  if (isLoading) {
    return (
      <View className="flex-1 items-center justify-center bg-white">
        <ActivityIndicator size="large" color="#22c55e" />
      </View>
    );
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {isAuthenticated ? (
          <Stack.Screen name="Main" component={MainNavigator} />
        ) : (
          <Stack.Screen name="Auth" component={AuthNavigator} />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
