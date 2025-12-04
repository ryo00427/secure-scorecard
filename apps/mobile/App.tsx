import './global.css';

import { StatusBar } from 'expo-status-bar';
import { Text, View } from 'react-native';

export default function App() {
  return (
    <View className="flex-1 items-center justify-center bg-white">
      <Text className="text-2xl font-bold text-primary-600">家庭菜園管理アプリ</Text>
      <Text className="mt-2 text-gray-500">Home Garden Management</Text>
      <StatusBar style="auto" />
    </View>
  );
}
