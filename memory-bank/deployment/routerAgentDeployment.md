# Router Agent (router_agent) Cross-Compilation for arm64 Linux

This document details the steps to cross-compile the `router_agent` Go project for a `linux/arm64` target (commonly referred to as `aarch64` at the OS level, but `arm64` for Go's `GOARCH`).

## Target Router Configuration
*   **CPU Architecture:** `arm64` (Go's `GOARCH` value, corresponds to `aarch64` at the OS level, e.g., from `uname -m` output like `aarch64`)
*   **Operating System:** `Linux`

## Prerequisites
*   Go programming language (version 1.20 or later recommended) installed on the development machine.
*   `protoc` compiler and `protoc-gen-go`, `protoc-gen-go-grpc` plugins installed and compatible with your Go and gRPC library versions.
*   Source code of `router_agent` available, located in the `router_agent/` directory.

## Key Configuration for Successful Compilation

Before attempting to compile, ensure the following configurations are correct within the `router_agent/` directory:

1.  **`capture_agent.proto`:**
    *   The `option go_package` should be set to `".;main"` to ensure generated Go files are part of `package main` and placed in the current directory.
        ```protobuf
        syntax = "proto3";

        package router_agent; // This is the protobuf package name

        option go_package = ".;main"; // Ensures Go package is 'main'

        // ... rest of the proto file
        ```

2.  **`main.go` (and other `.go` files intended for the executable):**
    *   The package declaration at the top of `main.go` (and any other `.go` files that are part of the agent executable) must be `package main`.

3.  **`go.mod`:**
    *   The Go version should be `1.20` or later to support features used by newer `protobuf` or `grpc` generated code (e.g., `unsafe.StringData`).
    *   The `google.golang.org/grpc` version should be compatible with the `protoc-gen-go-grpc` plugin used. As of recent generations, this might be `v1.64.0` or later.
    *   The `google.golang.org/protobuf` version should also be up-to-date.
    Example `go.mod` snippet:
    ```mod
    module wifi-pcap-demo/router_agent

    go 1.20 // Or a newer compatible version

    require (
        google.golang.org/grpc v1.64.0 // Or a newer compatible version
        google.golang.org/protobuf v1.33.0 // Or a newer compatible version
        // ... other dependencies
    )
    ```

## Cross-Compilation Steps (on Development Machine)

1.  **Open a Terminal:**
    Navigate to your project's root directory.

2.  **Change to the Agent's Source Code Directory:**
    The `router_agent` code is located in `router_agent/`.
    ```bash
    cd /Users/jiangzheyu/Desktop/project/wifi-pcap-demo/router_agent
    ```

3.  **Generate/Re-generate Protobuf Code (if necessary):**
    If you've made changes to `capture_agent.proto` or need to ensure the generated files are correct:
    ```bash
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           capture_agent.proto
    ```

4.  **Tidy Go Modules:**
    Ensure all dependencies are correctly listed and downloaded:
    ```bash
    go mod tidy
    ```

5.  **Set Environment Variables and Build:**
    The key environment variables for Go cross-compilation are `GOOS` (target operating system) and `GOARCH` (target architecture).
    Execute the following command to compile the agent:
    ```bash
    GOOS=linux GOARCH=arm64 go build -o router_agent_arm64 .
    ```
    *   `GOOS=linux`: Specifies the target operating system as Linux.
    *   `GOARCH=arm64`: Specifies the target CPU architecture as ARM 64-bit. **Note:** This is the correct `GOARCH` value for `aarch64` targets.
    *   `-o router_agent_arm64`: Specifies the output filename for the compiled binary.
    *   `.`: Instructs Go to build the package in the current directory (which should be `package main`).

6.  **Verify the Output:**
    After successful compilation, a binary file named `router_agent_arm64` will be created in the `router_agent/` directory.
    You can verify its properties using the `file` command (if on a Linux/macOS development machine):
    ```bash
    file router_agent_arm64
    ```
    The output should indicate it's an "ELF 64-bit LSB executable, ARM aarch64".

## Transferring to the Router

Once compiled, you need to transfer the `router_agent_arm64` binary to the target router. `scp` (Secure Copy Protocol) is a common method.

1.  **SCP Command:**
    ```bash
    scp router_agent_arm64 USER@ROUTER_IP:/path/on/router/router_agent_arm64
    ```
    *   Replace `USER` with your username on the router.
    *   Replace `ROUTER_IP` with the IP address of your router.
    *   Replace `/path/on/router/` with the desired directory on the router (e.g., `/usr/local/bin/`, `/tmp/`).

2.  **Make it Executable (on the router):**
    After transferring the file, log in to the router via SSH and make the binary executable:
    ```bash
    ssh USER@ROUTER_IP
    chmod +x /path/on/router/router_agent_arm64
    ```
    Now you should be able to run the agent from that location on the router.

## Important Notes
*   **Go Version in `go.mod`:** Ensure the `go` directive in your `go.mod` file (e.g., `go 1.20`) matches a Go version that supports features used by your dependencies and generated code.
*   **gRPC and Protobuf Versions:** Mismatches between `protoc-gen-go-grpc` plugin version, the `google.golang.org/grpc` library version in `go.mod`, and the `google.golang.org/protobuf` library can lead to compilation errors (like `undefined: grpc.SupportPackageIsVersionX`). Keep these aligned and up-to-date.
*   **CGO Dependencies:** If the `router_agent` project uses CGO, cross-compilation becomes more complex and may require a C cross-compiler toolchain for `aarch64-linux-gnu`. These instructions assume a pure Go project or managed CGO.
*   **Network and Permissions:** Ensure network connectivity for `scp` and necessary permissions on the router.
*   **Router's Environment:** The router must have a compatible Linux kernel. Go binaries are generally self-contained.