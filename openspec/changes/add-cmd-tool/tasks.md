## 1. Command Tool Implementation

- [x] 1.1 Refactor shared shell execution logic so `bash` behavior remains unchanged while another shell can reuse validation, timeout, output capture, and result formatting.
- [x] 1.2 Add a built-in `cmd` tool with the same model-facing parameters as `bash`.
- [x] 1.3 Register `cmd` from `basetools.Load()` alongside `bash` and `load_skill`.
- [x] 1.4 Return a clear unsupported-platform error when `cmd` is invoked outside Windows.

## 2. Verification

- [x] 2.1 Add tests for `cmd` tool registration and metadata.
- [x] 2.2 Add tests for successful Windows command execution, workdir handling, timeout validation, and unsupported-platform behavior where applicable.
- [x] 2.3 Run the Go test suite.
