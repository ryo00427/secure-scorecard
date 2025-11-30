@startuml
left to right direction
skinparam packageStyle rectangle
hide circle

actor "ユーザー" as User
actor "通知サービス" as Notifier

rectangle "家庭菜園アプリ" {
package "基本機能 (MVP)" {
usecase "ユーザー登録/ログイン" as UC1 #LightGreen
usecase "畑・プランター管理" as UC2 #LightGreen
usecase "作物登録/編集" as UC3 #LightGreen
usecase "ケアログ（水やり/肥料/収穫）" as UC4 #LightGreen
usecase "写真アップロード" as UC5 #LightGreen
usecase "タイムライン表示" as UC6 #LightGreen
}

package "拡張機能 (優先度中)" {
usecase "カレンダー表示" as UC7 #LightYellow
usecase "リマインド通知" as UC8 #LightYellow
usecase "収穫目安の提示" as UC9 #LightYellow
}

package "拡張機能 (優先度低)" {
usecase "共有/エクスポート" as UC10 #LightPink
usecase "簡易レポート/統計" as UC11 #LightPink
}
}

User --> UC1
User --> UC2
User --> UC3
User --> UC4
User --> UC5
User --> UC6
User --> UC7
User --> UC9
User --> UC10
User --> UC11

Notifier --> UC8

UC7 ..> UC4 : include
UC8 ..> UC7 : extend
UC9 ..> UC3 : extend
UC6 ..> UC3 : include

note right of UC6 #LightGreen
タイムライン/詳細表示

- ケアログの時系列表示
- 写真の一覧
  end note

note right of UC8 #LightYellow
リマインド例

- 水やり間隔で通知
- 収穫目安の通知
  end note
  @enduml
