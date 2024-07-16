# [ WIP ] ConsoleApplication Operator

The ConsoleApplication Operator is a Kubernetes operator designed to automate the deployment of applications on OpenShift from Git URLs using custom resources. This operator provides a streamlined and efficient approach to manage application lifecycle processes such as building, deploying, and serving applications without needing to manually fill out forms in the OpenShift Web Console.

## Features

- **Automated Deployment**: Automatically create and manage BuildConfig, Deployment, Service, and Route resources from a Git URL specified in a custom resource (CR).
- **CLI Integration**: Apply ConsoleApplication CRs via the CLI to trigger application deployment processes, bypassing the need for web-based forms.
- **Modular Architecture**: Supports various import strategies and can be extended to integrate new tools and technologies.
- **Status Tracking**: Utilizes a database to persist state and track the progress of resource creation.
- **Error Handling**: Ensures proper error handling and status updates for failed operations.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
