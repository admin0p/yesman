# YesMan
 
Yep this is AI generated ...:)

YesMan is a Go-based task delegation and worker pool system inspired by corporate dynamics. The project demonstrates a manager (the "Yes Man") who delegates tasks to a pool of workers, never refusing any task, and efficiently managing worker resources.

## Project Structure

- **main.go**: Entry point. Sets up the YesMan manager, creates tasks, and pushes them to the manager for execution.
- **boss/**: Contains the manager logic.
  - `manager.go`: Implements the YesManManager, which manages workers and delegates tasks.
  - `poolMaster.go`: Implements the Busybody, a helper that tracks idle and active workers and manages the worker pool.
- **worker/**: Contains the worker logic.
  - `worker.go`: Defines the Worker and Task interfaces, and the logic for assigning and running tasks.

## How It Works

1. **YesManManager** is initialized with a minimum and maximum number of workers.
2. On `Start()`, it creates the minimum number of workers and listens for incoming tasks.
3. Tasks are pushed to the manager via `PushTask()`. The manager assigns tasks to available workers.
4. **Busybody** tracks idle and active workers, providing an available worker or creating a new one if under the max limit.
5. Each **Worker** executes the assigned task and signals completion.
6. The manager waits for all tasks to finish before shutting down.

## Example Usage

```go
func main() {
    yesMan := master.NewYesMan(1, 3, nil)
    yesMan.Start()
    for i := 0; i < 10; i++ {
        t := NewTask(i, fmt.Sprintf(" :) %d", i))
        yesMan.PushTask(t)
    }
    yesMan.Stop()
}
```

## Key Concepts

- **Task Interface**: Any task must implement `Exec()`, `RetryExec()`, and `GetIdentifier()`.
- **Worker**: Executes tasks and reports completion.
- **YesManManager**: Delegates tasks, manages worker pool.
- **Busybody**: Tracks and manages idle/active workers.

## Requirements
- Go 1.18+
- [github.com/rs/xid](https://github.com/rs/xid) for unique worker IDs

## Running the Project

1. Clone the repository.
2. Run `go mod tidy` to install dependencies.
3. Run the project:
   ```sh
   go run main.go
   ```

## License
MIT
