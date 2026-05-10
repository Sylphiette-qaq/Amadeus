## Context

The built-in tool registry currently registers `bash` and `load_skill` from `internal/tool/basetools`. The `bash` tool already provides the desired execution contract: validated `command`, optional `workdir`, bounded `timeout_seconds`, captured stdout/stderr, exit code reporting, and output truncation.

Windows users need an equivalent native command runner because `cmd.exe` has different built-ins, quoting rules, environment behavior, and batch-file compatibility than bash.

## Goals / Non-Goals

**Goals:**
- Register a new built-in `cmd` tool alongside `bash`.
- Keep the `cmd` input and output contract consistent with `bash`.
- Use native Windows command execution via `cmd.exe /C`.
- Preserve existing `bash` behavior.
- Cover the new tool with focused tests.

**Non-Goals:**
- Replace or rename the existing `bash` tool.
- Add PowerShell support.
- Add long-running interactive command support.
- Expand timeout or output-size limits.

## Decisions

- **Reuse the existing command contract.** The `cmd` tool will accept `command`, `workdir`, and `timeout_seconds` to match `bash`, minimizing model-facing complexity and user learning cost. An alternative was to expose Windows-specific parameters, but that would make the two command tools harder to reason about.
- **Execute through `cmd.exe /C`.** This supports native cmd built-ins and batch semantics. Direct process execution would not support built-ins such as `dir` or command chaining in the same way.
- **Share execution plumbing where practical.** Validation, workdir resolution, timeout handling, stdout/stderr capture, exit code formatting, and truncation should remain consistent between `bash` and `cmd`. Duplicating that logic would increase drift risk.
- **Return a clear unsupported-platform error.** If invoked outside Windows, the tool should fail explicitly instead of relying on a missing `cmd.exe` error. This makes behavior predictable and easier for the model to interpret.

## Risks / Trade-offs

- **Platform-specific behavior** -> Guard `cmd` execution with an explicit Windows runtime check and keep tests focused on registration/validation where cross-platform behavior matters.
- **Refactor regression in `bash`** -> Preserve the existing output shape and run existing tests plus new basetool tests.
- **Command execution risk** -> Keep current timeout and output truncation limits; do not add elevated privileges or background execution.
