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
`k6ctl` leverages [`hashicorp/go-plugin`][hashicorp/go-plugin] for loading configurations from plugin binaries at runtime.
This design enables users to implement application-side logic and integrate it into the test script via environment variables.
- **Client-Side Only**: `k6ctl` assumes each test run can be isolated in pod level, and that the orchestration can be pre-calculated on the client side.
For instance, test loads can be scaled up by increasing the replicas and scheduled to different nodes by setting pod anti-affinity overrides.
These setup can be easily archived in client-side configurations and overrides, eliminating the need for maintaining the k6-operator.

[xk6]: https://k6.io/docs/extensions/
[hashicorp/go-plugin]: https://github.com/hashicorp/go-plugin

## Getting Started

### Installing

`k6ctl` can be installed via go install:

```
$ go install github.com/Azure/k6ctl/cmd/k6ctl@latest
$ k6ctl version
<version output>
```

### Running Test

We will run a test in a Kubernetes cluster with default setup. The test setup would look like this:

```
$ ls sample/helloworld
k6ctl.yaml      run.js
   |             |------------------> k6 test scenario
   |---> k6ctl configuration file
```

To invoke the test, we can:

```
$ k6ctl run -d sample/helloworld run.js
Please input value for parameter "message"
? hello, world
Please input value for parameter "level"
? info
```
<details>
<summary>sample logs output</summary>

```
Following logs of default/k6ctl-job-helloworld-qqxrm...

          /\      |‾‾| /‾‾/   /‾‾/   
     /\  /  \     |  |/  /   /  /    
    /  \/    \    |     (   /   ‾‾\  
   /          \   |  |\  \ |  (‾)  | 
  / __________ \  |__| \__\ \_____/ .io

     execution: local
        script: run.js
        output: -

     scenarios: (100.00%) 1 scenario, 500 max VUs, 1m10s max duration (incl. graceful stop):
              * default: Up to 500 looping VUs for 40s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)


Init      [  60% ] 298/500 VUs initialized
default   [   0% ]
time="2024-03-21T21:50:40Z" level=info msg="requesting https://k6.io/?message=hello%2C+world&level=info" source=console
time="2024-03-21T21:50:40Z" level=info msg="requesting https://k6.io/?message=hello%2C+world&level=info" source=console

running (0m00.2s), 002/500 VUs, 0 complete and 0 interrupted iterations
default   [   0% ] 002/500 VUs  00.2s/40.0s
time="2024-03-21T21:50:40Z" level=info msg="requesting https://k6.io/?message=hello%2C+world&level=info" source=console
time="2024-03-21T21:50:40Z" level=info msg="requesting https://k6.io/?message=hello%2C+world&level=info" source=console
time="2024-03-21T21:50:40Z" level=info msg="requesting https://k6.io/?message=hello%2C+world&level=info" source=console
# omitted logs
running (0m43.2s), 079/500 VUs, 4049 complete and 0 interrupted iterations
default ↓ [ 100% ] 500/500 VUs  40s

     █ teardown

     data_received..................: 2.1 GB 48 MB/s
     data_sent......................: 1.0 MB 24 kB/s
     http_req_blocked...............: avg=339.59ms min=500ns   med=800ns   max=7.89s    p(90)=16.3ms   p(95)=3.62s   
     http_req_connecting............: avg=679.55µs min=0s      med=0s      max=125.24ms p(90)=266.54µs p(95)=849.64µs
     http_req_duration..............: avg=2.12s    min=731.4µs med=2.1s    max=8.96s    p(90)=4.08s    p(95)=5.02s   
       { expected_response:true }...: avg=2.12s    min=731.4µs med=2.1s    max=8.96s    p(90)=4.08s    p(95)=5.02s   
     http_req_failed................: 0.00%  ✓ 0         ✗ 4129 
     http_req_receiving.............: avg=446.01ms min=111µs   med=8.31ms  max=5.76s    p(90)=1.76s    p(95)=2.68s   
     http_req_sending...............: avg=1.74ms   min=36.9µs  med=146.8µs max=758.8ms  p(90)=3.2ms    p(95)=7.22ms  
     http_req_tls_handshaking.......: avg=338.7ms  min=0s      med=0s      max=7.88s    p(90)=14.16ms  p(95)=3.62s   
     http_req_waiting...............: avg=1.67s    min=0s      med=1.87s   max=3.82s    p(90)=2.86s    p(95)=3.15s   
     http_reqs......................: 4129   94.857783/s
     iteration_duration.............: avg=3.49s    min=1.72ms  med=3.18s   max=16.87s   p(90)=5.52s    p(95)=8.24s   
     iterations.....................: 4128   94.83481/s
     vus............................: 79     min=0       max=500
     vus_max........................: 500    min=299     max=500
```

</details>

For writing k6 test scenario, please refer k6's [official doc][k6-doc]. The `k6ctl.yaml` defines the test run settings, below is a minimum setup:

```yaml
# name specifics the test run name
name: helloworld
# files section maps the local files to the pod via configmap.
# Test scenario files should be included in the list. Extra files like gRPC protobuf definitions can be included too.
files:
- source: run.js
  dest: run.js
# k6 section defines the k6 test run settings like target namespace, k6 base image to use.
k6:
  namespace: default
  image: ghcr.io/grafana/k6@sha256:8cd78f9d0de5f50bc8821cceecf356d5d9e839e6611c226a3fcf13c591080fbd
# configs section defines the configurations to be used for the test. Each config can be injected to the pod's environment
# with the name declared by `env`. Each config can be generated by a "provider", either built-in or from external plugin.
configs:
- provider:
    # `parameter` provider is the simplest way for providing value. The value can be specified either via the `run` command
    # CLI flag `-p/--parameters KEY=VALUE` or through input prompt.
    name: parameter
    params:
      name: "message"
  env: MESSAGE
- provider:
    name: parameter
    params:
      name: "level"
  env: LEVEL
```

[k6-doc]: https://grafana.com/docs/k6/latest/using-k6/

<!-- TODO
## Plugins

## Configurations
-->

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
