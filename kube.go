package k6ctl

import "github.com/Azure/k6ctl/internal/kubelib"

type KubeClientFactory = kubelib.KubeClientFactory

var CreateKubeClientFromKubeConfig = kubelib.CreateKubeClientFromKubeConfig
