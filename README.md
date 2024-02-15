# Kubernetes Service Account Management Tool

This Go program is designed to facilitate the management of Kubernetes service accounts, cluster roles, and kubeconfig files. It provides a command-line interface (CLI) for users to interactively create or delete service accounts, manage cluster roles, and generate kubeconfig files.

## Usage

### Prerequisites

Before using this tool, ensure that you have the following prerequisites installed:

- `kubectl`: The Kubernetes command-line tool used for managing Kubernetes clusters.
- Go environment: This tool is written in Go, so you need to have Go installed on your system.

### Installation

To use this tool, follow these steps:

1. Clone or download the repository containing the source code.
2. Navigate to the directory containing the source code.
3. Build the Go program using the `go build` command:
    ```
    go build -o kubeconfig-generator
    ```
4. Run the compiled binary:
    ```
    ./kubeconfig-generator [flags]
    ```

### Flags

- `-email`: Email address associated with the service account (required).
- `-ip`: Cluster IP address (optional).
- `-cr`: Custom cluster role YAML file path (optional).
- `-skipIP`: Skip providing the cluster IP and use the default from the selected context (optional).
- `-delete`: Skip creating service account and delete only (optional).

### Examples

#### Create a Service Account

To create a service account, run the tool with the `--email` flag:
```
./kubeconfig-generator --email example@mail.com
```
This will prompt you to provide the necessary information interactively.

#### Delete a Service Account

To delete a service account, use the `--delete` flag:
```
./kubeconfig-generator --email example@mail.com --delete
```
This will delete the specified service account.

## Functionality

- **Service Account Management**: Create or delete service accounts associated with the provided email address.
- **Cluster Role Management**: Apply default or custom cluster roles to the service account.
- **Kubeconfig Generation**: Generate a kubeconfig file with the necessary authentication details for the service account.
- **Interactive CLI**: Provides an interactive command-line interface for ease of use.

## Contributing

Contributions to this project are welcome. If you encounter any issues or have suggestions for improvements, please open an issue or submit a pull request on GitHub.
