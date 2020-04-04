# kubelet-rubber-stamp

[![Build Status](https://cloud.drone.io/api/badges/kontena/kubelet-rubber-stamp/status.svg)](https://cloud.drone.io/kontena/kubelet-rubber-stamp)

kubelet-rubber-stamp is simple CSR auto approver operator to help bootstrapping kubelet serving certificates easily.

The logic used follows the same logic used when auto-approving kubelet client certificates in kubelet [TLS bootstrap](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet-tls-bootstrapping/#approval) phase.

So basically the flow is:
1. kubelet gets the client cert (see [TLS bootstrap](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet-tls-bootstrapping/#approval))
2. Kubelet creates a [CSR](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/#create-a-certificate-signing-request-object-to-send-to-the-kubernetes-api)
3. kubelet-rubber-stamp reacts to the creation of a CSR
    - validates that it's a valid request for kubelet serving certificate
    - validates that the requestor (the kubelet/node) has sufficient authorization
    - approve the CSR
4. Kubelet fetches the certificate
5. Kubelet auto-rotates certs, goto 2 :)

## Kubelet configuration

To enable kubelet serving certificate bootstrapping you need to set `serverTLSBootstrap: true` in kubelet [configuration](https://kubernetes.io/docs/tasks/administer-cluster/kubelet-config-file/) file. That will enable kubelet to start the automated bootsrapping process.

## Validation


Once kubelet, api-server and kubelet-rubber-stamp have all played nice, kubelet should have gotten hold of the automatically bootsrapped serving certificate.

Kubelet stores the certificates it uses on disk, under `/var/lib/kubelet/pki`:
```sh
root@cluster-worker-4:~# ls -lah /var/lib/kubelet/pki/kubelet-server-*
-rw------- 1 root root 1.2K Mar  6 09:46 /var/lib/kubelet/pki/kubelet-server-2019-03-06-09-46-48.pem
lrwxrwxrwx 1 root root   59 Mar  6 09:46 /var/lib/kubelet/pki/kubelet-server-current.pem -> /var/lib/kubelet/pki/kubelet-server-2019-03-06-09-46-48.pem
```

You can see the certificate details e.g. with:
```
root@cluster-worker-4:~# openssl x509 -in /var/lib/kubelet/pki/kubelet-server-current.pem -text -noout
```

Check that the output details match the expected ones, especially that the certificate is signed by your Kubernetes CA.

Another way of validating is to peek into the kubelet API to see which certificate it offers:
```sh
root@cluster-worker-4:~# openssl s_client -showcerts -connect localhost:10250 </dev/null
...
```

Check that the certificate kubelet offers in it's API is signed by control plane.

## Deploying

`deploy` directory has the needed skeleton YAMLs for deploying to a cluster. They should though be taken more as reference, so adapt to your cluster setup.

### Image

Image is available on `quay.io/kontena/kubelet-rubber-stamp-<arch>:<tag>`. We bake both `amd64` and `arm64` images automatically. We've also automated several different tags:
- `latest` - Always up-to-date with `master` branch. Might not be 100% stable.
- `x.y.z` - Actual releases from git tags. Creates also semver "aliases", so releasing e.g. `0.1.3` will make `0.1` tag point to latest `0.1.z` patch release

## Troubleshooting

Use the usual troubleshooting steps as for any app running on Kubernetes:
- Is the pod running? `kubectl get pod --all-namespaces`
- Check the pod/container logs
    For this you probably need to ssh into the node as if kubelet does not have proper TLS cert, API server will not allow to proxy the logs.
- Make sure your RBAC rules allow kubelet to create proper CSRs
- Check kubelet logs for any errors related to TLS bootstrapping
- Check that the CSRs in the API look correct. `kubectl get csr ...`

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/kontena/kubelet-rubber-stamp

## License

Copyright (c) 2019 Kontena, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
