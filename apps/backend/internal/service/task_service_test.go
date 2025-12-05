// Package service - TaskService Unit Tests
//
// TaskService のユニットテストを提供します。
// テスト対象:
//   - CreateTask: タスク作成の正常系
//   - CompleteTask: タスク完了と繰り返しタスク自動生成
//   - GetOverdueTasks: 期限切れタスク取得
//   - GetTodayTasks: 今日のタスク取得
package service

import (
	"context"
	"testing"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// CreateTask テスト
// =============================================================================

// TestCreateTask_Success はタスク作成の正常系テストです。
// 期待動作:
//   - タスクがDBに保存される
//   - IDが自動採番される
//   - CreatedAt, UpdatedAt が設定される
func TestCreateTask_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// テスト用タスクを作成
	dueDate := time.Now().Add(24 * time.Hour)
	task := &model.Task{
		UserID:      1,
		Title:       "水やり",
		Description: "トマトに水をやる",
		DueDate:     dueDate,
		Priority:    "high",
		Status:      "pending",
	}

	// タスクを作成
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// IDが設定されていることを確認
	if task.ID == 0 {
		t.Error("Expected task ID to be set")
	}

	// タスクがリポジトリに保存されていることを確認
	savedTask, err := svc.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}

	if savedTask.Title != "水やり" {
		t.Errorf("Expected title '水やり', got '%s'", savedTask.Title)
	}

	if savedTask.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", savedTask.Priority)
	}
}

// TestCreateTask_WithRecurrence は繰り返し設定付きタスク作成のテストです。
// 期待動作:
//   - 繰り返し設定が正しく保存される
func TestCreateTask_WithRecurrence(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 繰り返し設定付きタスクを作成
	dueDate := time.Now().Add(24 * time.Hour)
	maxOccurrences := 5
	task := &model.Task{
		UserID:             1,
		Title:              "毎日の水やり",
		DueDate:            dueDate,
		Priority:           "medium",
		Status:             "pending",
		Recurrence:         "daily",
		RecurrenceInterval: 1,
		MaxOccurrences:     &maxOccurrences,
	}

	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// 繰り返し設定が保存されていることを確認
	savedTask, err := svc.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}

	if savedTask.Recurrence != "daily" {
		t.Errorf("Expected recurrence 'daily', got '%s'", savedTask.Recurrence)
	}

	if savedTask.RecurrenceInterval != 1 {
		t.Errorf("Expected recurrence interval 1, got %d", savedTask.RecurrenceInterval)
	}

	if savedTask.MaxOccurrences == nil || *savedTask.MaxOccurrences != 5 {
		t.Error("Expected max occurrences to be 5")
	}
}

// =============================================================================
// CompleteTask テスト
// =============================================================================

// TestCompleteTask_Success はタスク完了の正常系テストです。
// 期待動作:
//   - Status が "completed" に変更される
//   - CompletedAt が現在時刻に設定される
//   - OccurrenceCount がインクリメントされる
func TestCompleteTask_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// タスクを作成
	dueDate := time.Now().Add(24 * time.Hour)
	task := &model.Task{
		UserID:   1,
		Title:    "収穫",
		DueDate:  dueDate,
		Priority: "high",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// タスクを完了
	err = svc.CompleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// 完了状態を確認
	completedTask, err := svc.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}

	if completedTask.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", completedTask.Status)
	}

	if completedTask.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}

	if completedTask.OccurrenceCount != 1 {
		t.Errorf("Expected occurrence count 1, got %d", completedTask.OccurrenceCount)
	}
}

// TestCompleteTask_WithRecurrence_GeneratesNextTask は繰り返しタスク完了時の
// 次回タスク自動生成テストです。
// 期待動作:
//   - 元タスクが完了される
//   - 次回タスクが自動生成される
//   - 次回タスクの DueDate が正しく計算される
//   - 次回タスクの ParentTaskID が設定される
func TestCompleteTask_WithRecurrence_GeneratesNextTask(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 繰り返しタスクを作成（毎日）
	dueDate := time.Now().Truncate(24 * time.Hour)
	task := &model.Task{
		UserID:             1,
		Title:              "毎日の水やり",
		DueDate:            dueDate,
		Priority:           "medium",
		Status:             "pending",
		Recurrence:         "daily",
		RecurrenceInterval: 1,
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	originalTaskID := task.ID

	// タスクを完了
	err = svc.CompleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// ユーザーのタスク一覧を取得して次回タスクを確認
	tasks, err := svc.GetUserTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserTasks failed: %v", err)
	}

	// 2つのタスクがあるはず（完了したタスク + 新規生成されたタスク）
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}

	// 新規生成されたタスクを検証
	var nextTask *model.Task
	for i := range tasks {
		if tasks[i].ID != originalTaskID {
			nextTask = &tasks[i]
			break
		}
	}

	if nextTask == nil {
		t.Fatal("Next task was not generated")
	}

	// 次回タスクの検証
	if nextTask.Status != "pending" {
		t.Errorf("Expected next task status 'pending', got '%s'", nextTask.Status)
	}

	if nextTask.Title != "毎日の水やり" {
		t.Errorf("Expected next task title '毎日の水やり', got '%s'", nextTask.Title)
	}

	// 期限日が翌日であることを確認
	expectedDueDate := dueDate.AddDate(0, 0, 1)
	if !nextTask.DueDate.Equal(expectedDueDate) {
		t.Errorf("Expected due date %v, got %v", expectedDueDate, nextTask.DueDate)
	}

	// ParentTaskID が設定されていることを確認
	if nextTask.ParentTaskID == nil || *nextTask.ParentTaskID != originalTaskID {
		t.Error("Expected ParentTaskID to be set to original task ID")
	}

	// OccurrenceCount が引き継がれていることを確認
	if nextTask.OccurrenceCount != 1 {
		t.Errorf("Expected occurrence count 1, got %d", nextTask.OccurrenceCount)
	}
}

// TestCompleteTask_WithRecurrence_Weekly は週次繰り返しタスクのテストです。
func TestCompleteTask_WithRecurrence_Weekly(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 週次繰り返しタスクを作成
	dueDate := time.Now().Truncate(24 * time.Hour)
	task := &model.Task{
		UserID:             1,
		Title:              "週次レポート",
		DueDate:            dueDate,
		Priority:           "low",
		Status:             "pending",
		Recurrence:         "weekly",
		RecurrenceInterval: 2, // 2週間ごと
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// タスクを完了
	err = svc.CompleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// 次回タスクを確認
	tasks, err := svc.GetUserTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserTasks failed: %v", err)
	}

	var nextTask *model.Task
	for i := range tasks {
		if tasks[i].ID != task.ID {
			nextTask = &tasks[i]
			break
		}
	}

	if nextTask == nil {
		t.Fatal("Next task was not generated")
	}

	// 期限日が2週間後であることを確認
	expectedDueDate := dueDate.AddDate(0, 0, 14) // 2週間 = 14日
	if !nextTask.DueDate.Equal(expectedDueDate) {
		t.Errorf("Expected due date %v, got %v", expectedDueDate, nextTask.DueDate)
	}
}

// TestCompleteTask_WithRecurrence_StopsAtMaxOccurrences は最大回数到達時の
// 繰り返し停止テストです。
// 期待動作:
//   - MaxOccurrences に達した場合、次回タスクは生成されない
func TestCompleteTask_WithRecurrence_StopsAtMaxOccurrences(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 最大回数2回の繰り返しタスクを作成
	dueDate := time.Now().Truncate(24 * time.Hour)
	maxOccurrences := 2
	task := &model.Task{
		UserID:             1,
		Title:              "限定タスク",
		DueDate:            dueDate,
		Priority:           "medium",
		Status:             "pending",
		Recurrence:         "daily",
		RecurrenceInterval: 1,
		MaxOccurrences:     &maxOccurrences,
		OccurrenceCount:    1, // 既に1回実行済み
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// タスクを完了（これが2回目 = 最大回数）
	err = svc.CompleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// タスク一覧を取得
	tasks, err := svc.GetUserTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserTasks failed: %v", err)
	}

	// タスクは1つのまま（次回タスクは生成されない）
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task (no next task generated), got %d", len(tasks))
	}
}

// TestCompleteTask_WithRecurrence_StopsAfterEndDate は終了日到達時の
// 繰り返し停止テストです。
// 期待動作:
//   - RecurrenceEndDate を過ぎた場合、次回タスクは生成されない
func TestCompleteTask_WithRecurrence_StopsAfterEndDate(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 繰り返し終了日が翌日のタスクを作成
	dueDate := time.Now().Truncate(24 * time.Hour)
	endDate := dueDate.Add(12 * time.Hour) // 終了日は今日の12時間後
	task := &model.Task{
		UserID:             1,
		Title:              "期間限定タスク",
		DueDate:            dueDate,
		Priority:           "medium",
		Status:             "pending",
		Recurrence:         "daily",
		RecurrenceInterval: 1,
		RecurrenceEndDate:  &endDate,
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// タスクを完了
	err = svc.CompleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// タスク一覧を取得
	tasks, err := svc.GetUserTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserTasks failed: %v", err)
	}

	// 次回期限日（翌日）が終了日を過ぎているので、タスクは1つのまま
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task (no next task generated due to end date), got %d", len(tasks))
	}
}

// =============================================================================
// GetOverdueTasks テスト
// =============================================================================

// TestGetOverdueTasks_Success は期限切れタスク取得の正常系テストです。
// 期待動作:
//   - 期限日が過去で、ステータスが pending のタスクのみ取得される
func TestGetOverdueTasks_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)
	today := time.Now().Truncate(24 * time.Hour)

	// 期限切れタスク（昨日が期限）
	overdueTask := &model.Task{
		UserID:   userID,
		Title:    "期限切れタスク",
		DueDate:  today.AddDate(0, 0, -1), // 昨日
		Priority: "high",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, overdueTask)
	if err != nil {
		t.Fatalf("CreateTask (overdue) failed: %v", err)
	}

	// 未来のタスク（明日が期限）
	futureTask := &model.Task{
		UserID:   userID,
		Title:    "未来のタスク",
		DueDate:  today.AddDate(0, 0, 1), // 明日
		Priority: "medium",
		Status:   "pending",
	}
	err = svc.CreateTask(ctx, futureTask)
	if err != nil {
		t.Fatalf("CreateTask (future) failed: %v", err)
	}

	// 完了済みの期限切れタスク
	completedOverdueTask := &model.Task{
		UserID:   userID,
		Title:    "完了済み期限切れ",
		DueDate:  today.AddDate(0, 0, -2), // 一昨日
		Priority: "low",
		Status:   "completed",
	}
	err = svc.CreateTask(ctx, completedOverdueTask)
	if err != nil {
		t.Fatalf("CreateTask (completed overdue) failed: %v", err)
	}

	// 期限切れタスクを取得
	overdueTasks, err := svc.GetOverdueTasks(ctx, userID)
	if err != nil {
		t.Fatalf("GetOverdueTasks failed: %v", err)
	}

	// 期限切れかつ pending のタスクのみ取得される
	if len(overdueTasks) != 1 {
		t.Errorf("Expected 1 overdue task, got %d", len(overdueTasks))
	}

	if len(overdueTasks) > 0 && overdueTasks[0].Title != "期限切れタスク" {
		t.Errorf("Expected overdue task title '期限切れタスク', got '%s'", overdueTasks[0].Title)
	}
}

// TestGetOverdueTasks_Empty は期限切れタスクがない場合のテストです。
func TestGetOverdueTasks_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)
	today := time.Now().Truncate(24 * time.Hour)

	// 未来のタスクのみ作成
	futureTask := &model.Task{
		UserID:   userID,
		Title:    "未来のタスク",
		DueDate:  today.AddDate(0, 0, 1),
		Priority: "medium",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, futureTask)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// 期限切れタスクを取得
	overdueTasks, err := svc.GetOverdueTasks(ctx, userID)
	if err != nil {
		t.Fatalf("GetOverdueTasks failed: %v", err)
	}

	// 期限切れタスクはない
	if len(overdueTasks) != 0 {
		t.Errorf("Expected 0 overdue tasks, got %d", len(overdueTasks))
	}
}

// =============================================================================
// GetTodayTasks テスト
// =============================================================================

// TestGetTodayTasks_Success は今日のタスク取得の正常系テストです。
// 期待動作:
//   - 今日が期限で、ステータスが pending のタスクのみ取得される
func TestGetTodayTasks_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)
	today := time.Now().Truncate(24 * time.Hour)

	// 今日のタスク
	todayTask := &model.Task{
		UserID:   userID,
		Title:    "今日のタスク",
		DueDate:  today.Add(12 * time.Hour), // 今日の正午
		Priority: "high",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, todayTask)
	if err != nil {
		t.Fatalf("CreateTask (today) failed: %v", err)
	}

	// 明日のタスク
	tomorrowTask := &model.Task{
		UserID:   userID,
		Title:    "明日のタスク",
		DueDate:  today.AddDate(0, 0, 1),
		Priority: "medium",
		Status:   "pending",
	}
	err = svc.CreateTask(ctx, tomorrowTask)
	if err != nil {
		t.Fatalf("CreateTask (tomorrow) failed: %v", err)
	}

	// 昨日のタスク（期限切れ）
	yesterdayTask := &model.Task{
		UserID:   userID,
		Title:    "昨日のタスク",
		DueDate:  today.AddDate(0, 0, -1),
		Priority: "low",
		Status:   "pending",
	}
	err = svc.CreateTask(ctx, yesterdayTask)
	if err != nil {
		t.Fatalf("CreateTask (yesterday) failed: %v", err)
	}

	// 今日のタスクを取得
	todayTasks, err := svc.GetTodayTasks(ctx, userID)
	if err != nil {
		t.Fatalf("GetTodayTasks failed: %v", err)
	}

	// 今日のタスクのみ取得される
	if len(todayTasks) != 1 {
		t.Errorf("Expected 1 today task, got %d", len(todayTasks))
	}

	if len(todayTasks) > 0 && todayTasks[0].Title != "今日のタスク" {
		t.Errorf("Expected today task title '今日のタスク', got '%s'", todayTasks[0].Title)
	}
}

// =============================================================================
// GetUserTasks テスト
// =============================================================================

// TestGetUserTasks_Success はユーザーの全タスク取得の正常系テストです。
func TestGetUserTasks_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)
	otherUserID := uint(2)
	today := time.Now().Truncate(24 * time.Hour)

	// ユーザー1のタスク
	task1 := &model.Task{
		UserID:   userID,
		Title:    "タスク1",
		DueDate:  today,
		Priority: "high",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, task1)
	if err != nil {
		t.Fatalf("CreateTask (task1) failed: %v", err)
	}

	task2 := &model.Task{
		UserID:   userID,
		Title:    "タスク2",
		DueDate:  today.AddDate(0, 0, 1),
		Priority: "medium",
		Status:   "pending",
	}
	err = svc.CreateTask(ctx, task2)
	if err != nil {
		t.Fatalf("CreateTask (task2) failed: %v", err)
	}

	// 他のユーザーのタスク
	otherTask := &model.Task{
		UserID:   otherUserID,
		Title:    "他のユーザーのタスク",
		DueDate:  today,
		Priority: "low",
		Status:   "pending",
	}
	err = svc.CreateTask(ctx, otherTask)
	if err != nil {
		t.Fatalf("CreateTask (other) failed: %v", err)
	}

	// ユーザー1のタスク一覧を取得
	tasks, err := svc.GetUserTasks(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserTasks failed: %v", err)
	}

	// ユーザー1のタスクのみ取得される
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks for user 1, got %d", len(tasks))
	}

	// 他のユーザーのタスクが含まれていないことを確認
	for _, task := range tasks {
		if task.UserID != userID {
			t.Errorf("Expected all tasks to belong to user %d, got user %d", userID, task.UserID)
		}
	}
}

// =============================================================================
// GetUserTasksByStatus テスト
// =============================================================================

// TestGetUserTasksByStatus_Success はステータスフィルタの正常系テストです。
func TestGetUserTasksByStatus_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)
	today := time.Now().Truncate(24 * time.Hour)

	// pending タスク
	pendingTask := &model.Task{
		UserID:   userID,
		Title:    "未完了タスク",
		DueDate:  today,
		Priority: "high",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, pendingTask)
	if err != nil {
		t.Fatalf("CreateTask (pending) failed: %v", err)
	}

	// completed タスク
	completedTask := &model.Task{
		UserID:   userID,
		Title:    "完了タスク",
		DueDate:  today,
		Priority: "medium",
		Status:   "completed",
	}
	err = svc.CreateTask(ctx, completedTask)
	if err != nil {
		t.Fatalf("CreateTask (completed) failed: %v", err)
	}

	// pending タスクのみ取得
	pendingTasks, err := svc.GetUserTasksByStatus(ctx, userID, "pending")
	if err != nil {
		t.Fatalf("GetUserTasksByStatus (pending) failed: %v", err)
	}

	if len(pendingTasks) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pendingTasks))
	}

	if len(pendingTasks) > 0 && pendingTasks[0].Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", pendingTasks[0].Status)
	}

	// completed タスクのみ取得
	completedTasks, err := svc.GetUserTasksByStatus(ctx, userID, "completed")
	if err != nil {
		t.Fatalf("GetUserTasksByStatus (completed) failed: %v", err)
	}

	if len(completedTasks) != 1 {
		t.Errorf("Expected 1 completed task, got %d", len(completedTasks))
	}
}

// =============================================================================
// DeleteTask テスト
// =============================================================================

// TestDeleteTask_Success はタスク削除の正常系テストです。
func TestDeleteTask_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// タスクを作成
	task := &model.Task{
		UserID:   1,
		Title:    "削除するタスク",
		DueDate:  time.Now(),
		Priority: "low",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// タスクを削除
	err = svc.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// タスクが取得できないことを確認
	_, err = svc.GetTaskByID(ctx, task.ID)
	if err == nil {
		t.Error("Expected error when getting deleted task")
	}
}

// =============================================================================
// UpdateTask テスト
// =============================================================================

// TestUpdateTask_Success はタスク更新の正常系テストです。
func TestUpdateTask_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// タスクを作成
	task := &model.Task{
		UserID:   1,
		Title:    "元のタイトル",
		DueDate:  time.Now(),
		Priority: "low",
		Status:   "pending",
	}
	err := svc.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// タスクを更新
	task.Title = "更新後のタイトル"
	task.Priority = "high"
	err = svc.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// 更新を確認
	updatedTask, err := svc.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}

	if updatedTask.Title != "更新後のタイトル" {
		t.Errorf("Expected title '更新後のタイトル', got '%s'", updatedTask.Title)
	}

	if updatedTask.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", updatedTask.Priority)
	}
}
