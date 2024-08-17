# Component Development Best Practices

## 1. Component Structure and Organization

### 1.1 Package Organization
- Each component should reside in its own package.
- Use descriptive package names that reflect the component's functionality (e.g., `tcpserver`, `logger`, `redis`).
- For components with multiple implementations or APIs, use subpackages:
  ```
  httpserver/
  ├── http/
  ├── echo/
  └── gin/
  ```

### 1.2 API Definition
- If the component has a public API, define it in a separate `api` or `[component]api` package (e.g., `redisapi`, `dbapi`).
- Keep the API package lightweight, containing only interface definitions and essential types.

### 1.3 File Structure
- Main component logic goes in a file named after the package (e.g., `redis.go` in the `redis` package).
- Use additional files for specific functionalities if needed (e.g., `pid_unix.go`, `pid_windows.go` in the `pidfile` package).

## 2. Naming Conventions

### 2.1 Component Identifier
- Define a `Name` constant for each component:
  ```go
  const Name = "your/repository/path/components/[component-name]"
  ```

### 2.2 Component Struct
- Use a lowercase name ending with "Component" (e.g., `redisComponent`, `loggerComponent`).
- Keep the struct unexported to encapsulate internal details.

### 2.3 Options Struct
- Name it `Options` and keep it exported for configuration.

### 2.4 API Interfaces
- Name interfaces ending with `Component` or `API` in the API package (e.g., `Component` in `redisapi`).

## 3. Component Registration

- Use an `init()` function to register the component:
  ```go
  func init() {
      component.Register(Name, func() component.Component {
          return &componentNameComponent{}
      })
  }
  ```
- Important: The registration function should only create and return the component object. Do not initialize any fields here. All initialization should be done in the `Init` method.

## 4. Options and Configuration

### 4.1 Options Struct
- Define an `Options` struct with all configurable parameters.
- Use meaningful field names and add comments for clarity.
- Use `time.Duration` for time-related configurations.
- For boolean flags, define them so that the default (zero) value is false and represents the most common or safest configuration.
- If a flag needs to be true by default, invert its meaning in the name (e.g., use `DisableFeature` instead of `EnableFeature`).

### 4.2 OnLoaded hook
- Implement `OnLoaded` method to set default values or validator for options:
  ```go
  func (o *Options) OnLoaded() error {
      // Set default values
      // Validate values
  }
  ```

## 5. Component Implementation

### 5.1 Base Component
- Embed `component.BaseComponent[Options]` for basic components.
- Embed `component.BaseComponentWithRefs[Options, Refs]` for components with references to other components.

### 5.2 Interface Compliance
- Ensure the component implements necessary interfaces.
- Use a compile-time check to ensure interface compliance:
  ```go
  var _ componentapi.Component = (*componentNameComponent)(nil)
  ```

### 5.3 Lifecycle Methods
- Implement `Init`, `Start`, `Shutdown`, and `Uninit` methods as needed.
- Use context for cancellation support in these methods.
- Lifecycle method pairs:
  - `Init` and `Uninit`: If `Init` is called, `Uninit` will always be called. If `Init` is not called, `Uninit` will not be called.
  - `Start` and `Shutdown`: Similarly, if `Start` is called, `Shutdown` will always be called. If `Start` is not called, `Shutdown` will not be called.
- In `Init`:
  - Set default options
  - Initialize the component's own resources and fields
  - Do not access or use referenced components in `Init`
- After `Init`, ensure that the component's public API is ready for use by other components
- In `Start`:
  - Begin main component operations
  - Start background goroutines if needed
  - You can now safely use referenced components
- In `Shutdown`:
  - Stop all operations gracefully
  - Release resources
  - Wait for goroutines to finish
- In `Uninit`:
  - Perform final cleanup
  - Release any resources that weren't handled in `Shutdown`
  - This method is called after `Shutdown` and should handle any remaining cleanup tasks

### 5.4 Error Handling
- Return errors from methods instead of panicking.
- Use descriptive error messages, wrapping errors for context when appropriate.
- Create custom error types if needed.

### 5.5 Logging
- Use the logger provided by the base component (`c.Logger()`).
- Log important events, errors, and state changes.
- Use appropriate log levels (Info, Warn, Error).

## 6. Concurrency and Resource Management

### 6.1 Goroutines
- Use goroutines for concurrent operations.
- Implement proper shutdown mechanisms to stop goroutines gracefully.

### 6.2 Synchronization
- Use `sync.Mutex` or `sync.RWMutex` for protecting shared resources.
- Consider using `atomic` operations for simple counters.

### 6.3 Channels
- Use channels for communication between goroutines.
- Use `select` statements with context cancellation for graceful shutdowns.

### 6.4 Resource Cleanup
- Always close opened resources (files, network connections, etc.).
- Use `defer` for cleanup operations where appropriate.

## 7. API Design

### 7.1 Interface Definition
- Define clear, minimal interfaces in the API package.
- Focus on the core functionality of the component.

### 7.2 Method Signatures
- Use context as the first parameter for methods that may be long-running.
- Return errors as the last return value.

### 7.3 Configuration
- Use functional options pattern for complex configurations when appropriate.

## 8. Cross-platform Compatibility

- Use build tags for platform-specific implementations (e.g., `abc_unix.go`, `abc_windows.go`).
- Provide fallback implementations for unsupported platforms when possible.

## 9. Testing

- Write unit tests for component logic.
- Use table-driven tests for multiple test cases.
- Mock external dependencies for isolated testing.

## 10. Documentation

- Provide clear godoc comments for exported types, functions, and methods.
- Include usage examples in package documentation.

## 11. Component Interaction

### 11.1 Component References
- Use `component.Reference` for required component references.
  - After `Init`, these referenced components are guaranteed to be available and their APIs can be used directly.
- Use `component.OptionalReference` for optional component references.
  - Even after `Init`, always check if `Component()` is nil before using optional references.
  - Optional components may not be injected, so your component should handle their absence gracefully.

### 11.2 Event Handling
- Use the `event` package for implementing event-driven components.
- In the package where events are defined, add the following comment:
  ```go
  //go:generate eventer
  ```
  This will generate helper definitions for all types suffixed with `Event`, facilitating easier use of events.
- To install the `eventer` tool, use the following command:
  ```
  go install github.com/gopherd/tools/cmd/eventer@latest
  ```

Remember that the framework automatically handles dependency injection, so you don't need to manually resolve references in the `Init` method.

By following these practices, you'll create components that are consistent, maintainable, and work well within the framework's design. Always ensure that your component's public API is ready for use after `Init`, and only interact with other components from the `Start` method onwards.
