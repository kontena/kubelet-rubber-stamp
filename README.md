# kubelet-rubber-stamp

kubelet-rubber-stamp is simple CSR auto approver operator to help bootstrapping kubelet serving certificates easily.

The logic used follows the same logic used when auto-approving kubelet client certificates in kubelet [TLS bootstrap](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet-tls-bootstrapping/#approval) phase.

So basically the flow is:
1. kubelet gets the client cert (see [TLS bootstrap](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet-tls-bootstrapping/#approval))
2. Kubelet creates a [CSR](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/#create-a-certificate-signing-request-object-to-send-to-the-kubernetes-api)
3. kubelet-rubber-stamp reacts to the creationg of a CSR
    - validates that it's a valid request for kubelet serving certificate
    - validates that the requestor (the kubelet/node) has sufficient authorization
    - approve the CSR
4. Kubelet fetches the certificate
5. Kubelet auto-rotates certs, goto 2 :)

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/kontena/kubelet-rubber-stamp

## License

Copyright (c) 2019 Kontena, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.