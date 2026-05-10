## ADDED Requirements

### Requirement: Native Windows cmd execution tool
The system SHALL provide a built-in `cmd` tool that executes local Windows commands through native `cmd.exe` command semantics.

#### Scenario: Execute a successful cmd command
- **WHEN** the model invokes the `cmd` tool with a valid `command`
- **THEN** the system executes the command using `cmd.exe /C` on Windows and returns command metadata, exit code, stdout, and stderr

#### Scenario: Use an explicit working directory
- **WHEN** the model invokes the `cmd` tool with valid `command` and `workdir` parameters
- **THEN** the system executes the command from the resolved working directory and includes that directory in the result

### Requirement: Cmd tool safety bounds
The `cmd` tool MUST enforce the same timeout and output-size bounds as the existing `bash` tool.

#### Scenario: Reject invalid timeout
- **WHEN** the model invokes the `cmd` tool with `timeout_seconds` less than or equal to zero or greater than the maximum allowed value
- **THEN** the system rejects the tool invocation with a clear validation error

#### Scenario: Command times out
- **WHEN** a `cmd` command runs longer than the configured timeout
- **THEN** the system stops the command and returns the captured output with a timeout error

### Requirement: Unsupported platform behavior
The `cmd` tool MUST fail explicitly when invoked on a non-Windows platform.

#### Scenario: Invoke cmd outside Windows
- **WHEN** the model invokes the `cmd` tool on a non-Windows platform
- **THEN** the system returns a clear unsupported-platform error without attempting to run `cmd.exe`
