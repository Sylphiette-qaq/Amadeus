## Why

Amadeus currently exposes a local `bash` base tool, but Windows users often need native `cmd.exe` semantics for commands, built-ins, quoting, and batch-script behavior. Adding a dedicated `cmd` tool improves local Windows operability without changing the existing `bash` tool.

## What Changes

- Add a new built-in `cmd` tool for executing local Windows `cmd.exe /C` commands.
- Keep the command interface aligned with `bash`: `command`, optional `workdir`, and optional `timeout_seconds`.
- Return structured command execution output with command, workdir, exit code, stdout, and stderr.
- Fail clearly when the `cmd` tool is invoked on a non-Windows platform.

## Capabilities

### New Capabilities
- `windows-cmd-tool`: Provides native Windows command execution as a built-in tool.

### Modified Capabilities

## Impact

- Affected code: `internal/tool/basetools`.
- Tool registry behavior changes by exposing one additional built-in tool.
- No external dependencies or breaking changes.
