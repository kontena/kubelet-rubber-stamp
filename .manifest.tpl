image: docker.io/kontenapharos/kubelet-rubber-stamp:{{#if build.tag}}{{trimPrefix "v" build.tag}}{{else}}latest{{/if}}
{{#if build.tags}}
tags:
{{#each build.tags}}
  - {{this}}
{{/each}}
{{/if}}
manifests:
  -
    image: quay.io/kontena/kubelet-rubber-stamp-amd64:{{#if build.tag}}{{trimPrefix "v" build.tag}}{{else}}latest{{/if}}
    platform:
      architecture: amd64
      os: linux
  -
    image: quay.io/kontena/kubelet-rubber-stamp-arm64:{{#if build.tag}}{{trimPrefix "v" build.tag}}{{else}}latest{{/if}}
    platform:
      architecture: arm64
      os: linux
