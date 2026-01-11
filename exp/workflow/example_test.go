package workflow_test

import (
	"context"
	"fmt"

	"github.com/jxskiss/gopkg/v2/exp/workflow"
)

func Example() {
	// 1. 定义任务
	// 任务 1: 获取数据
	taskFetch := workflow.NewTask("fetch_data", func(ctx context.Context, rc workflow.RunContext) (any, error) {
		// 模拟从外部获取数据，这里还可以读取工作流的输入
		input := rc.WorkflowInput()
		userID := input.GetString("user_id")
		fmt.Printf("Fetching data for user: %s\n", userID)
		return "raw_data", nil
	})

	// 任务 2: 处理数据 (依赖 fetch_data)
	taskProcess := workflow.NewTask("process_data", func(ctx context.Context, rc workflow.RunContext) (any, error) {
		// 获取上游任务的输出
		raw, ok := rc.GetTaskOutput("fetch_data")
		if !ok {
			return nil, fmt.Errorf("missing upstream output")
		}
		fmt.Printf("Processing data: %v\n", raw)
		return "processed_" + raw.(string), nil
	}, workflow.DependsOn("fetch_data"))

	// 任务 3: 保存数据 (依赖 process_data)
	taskSave := workflow.NewTask("save_data", func(ctx context.Context, rc workflow.RunContext) (any, error) {
		data, _ := rc.GetTaskOutput("process_data")
		fmt.Printf("Saving data: %v\n", data)
		return true, nil
	}, workflow.DependsOn("process_data"))

	// 2. 创建工作流，并设置初始输入
	wf := workflow.NewWorkflow("data_pipeline",
		workflow.WithInput(map[string]any{
			"user_id": "1001",
		}),
	)

	// 3. 添加任务到工作流
	// 注意：添加顺序不影响执行顺序，执行顺序由依赖关系决定
	// 建议一次性添加所有相关任务
	if err := wf.AddTask(context.Background(), taskFetch, taskProcess, taskSave); err != nil {
		fmt.Printf("Add tasks failed: %v\n", err)
		return
	}

	// 4. 执行工作流
	// 第二个参数可以传入 gopool.GoPool 用于控制并发，传入 nil 使用默认 goroutine
	result, err := wf.Execute(context.Background(), nil)
	if err != nil {
		fmt.Printf("Workflow execution failed: %v\n", err)
		return
	}

	// 5. 查看结果
	saveResult := result.TaskResults["save_data"]
	fmt.Printf("Save task success: %v\n", saveResult.Output)

	// Output:
	// Fetching data for user: 1001
	// Processing data: raw_data
	// Saving data: processed_raw_data
	// Save task success: true
}

func Example_dynamic() {
	// 这个示例演示如何在任务执行过程中动态添加新任务。

	// 定义初始任务，它将决定后续需要执行什么任务
	initTask := workflow.NewTask("planner", func(ctx context.Context, rc workflow.RunContext) (any, error) {
		fmt.Println("Planner started")

		// 动态定义并添加任务 A
		taskA := workflow.NewTask("dynamic_A", func(ctx context.Context, rc workflow.RunContext) (any, error) {
			fmt.Println("Executing Dynamic Task A")
			return "Result A", nil
		}, workflow.DependsOn("planner")) // 依赖当前任务，确保在当前任务完成后执行（虽然在这个闭包里当前任务还没完，但逻辑上依赖关系是生效的）
		// 注意：通常动态添加的任务如果依赖当前任务，它会在当前任务成功返回后才会被调度。

		// 动态定义并添加任务 B，它依赖任务 A
		taskB := workflow.NewTask("dynamic_B", func(ctx context.Context, rc workflow.RunContext) (any, error) {
			resA, _ := rc.GetTaskOutput("dynamic_A")
			fmt.Printf("Executing Dynamic Task B with input: %v\n", resA)
			return "Result B", nil
		}, workflow.DependsOn("dynamic_A"))

		// 将这些任务添加到当前运行的工作流中
		err := rc.AddTask(ctx, taskA, taskB)
		if err != nil {
			return nil, err
		}

		return "Plan created", nil
	})

	wf := workflow.NewWorkflow("dynamic_workflow")
	_ = wf.AddTask(context.Background(), initTask)

	_, err := wf.Execute(context.Background(), nil)
	if err != nil {
		fmt.Printf("Workflow failed: %v\n", err)
	}

	// Output:
	// Planner started
	// Executing Dynamic Task A
	// Executing Dynamic Task B with input: Result A
}
