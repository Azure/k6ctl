# k6ctl

k6ctl = [k6][] + plugins

[k6]: https://github.com/grafana/k6

## Problem Statement

K6 is helpful for performing large scale load test with easy-to-write test spec.
However, we captured a few challenges for using k6 to load test our internal infrastructure
under microservice fashion:

1. **Internal Configurations**: There are gaps in getting/setting the correct internal configurations, such as authentication tokens, within the test payload. K6 provides a way to extend its functionalities through [xk6][xk6], a framework for creating k6 extensions. However, this approach requires rebuilding the binary and image for changes, potentially enlarging the feedback cycle.
2. **Test Orchestration**: Although the test runs can be orchestrated by k6-operator, it's likely that we will
need to run the test load under the same Kubernetes cluster as the target service. Not every environment is capable for
deploying the operator.

To address these challenges, we developed k6ctl based on the following principles:

- **Composite with External Plugins**: To allow extending the functionality without paying the cost of rebuilding everything,
`k6ctl` leverages [`hashicorp/go-plugin`][hashcorp/go-plugin] for loading configurations from plugin binaries at runtime.
This design enables users to implement application-side logic and integrate it into the test script via environment variables.
- **Client-Side Only**: `k6ctl` assumes each test run can be isolated in pod level, and that the orchestration can be pre-calculated on the client side.
For instance, test loads can be scaled up by increasing the replicas and scheduled to different nodes by setting pod anti-affinity overrides.
These setup can be easily archived in client-side configurations and overrides, eliminating the need for maintaining the k6-operator.

[xk6]: https://k6.io/docs/extensions/
[hashicorp/go-plugin]: https://github.com/hashicorp/go-plugin

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
