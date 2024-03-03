package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (tr *taskRunner) buildConfigSecretObject(
	ctx context.Context,
	configs []resolvedConfig,
) (*k8scorev1.Secret, error) {
	stringData := map[string]string{}
	for _, c := range configs {
		stringData[c.Env] = c.Value
	}

	rv := &k8scorev1.Secret{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      tr.configsSecretName(),
			Namespace: tr.objectNamespace(),
			Labels: map[string]string{
				labelKeyTaskName: tr.taskConfig.Name,
			},
		},
		Type:       "Opaque",
		StringData: stringData,
	}

	return rv, nil
}

func (tr *taskRunner) resolveSourceFileContent(
	ctx context.Context,
	source string,
) (string, error) {
	absSourceFilePath, err := filepath.Abs(filepath.Join(tr.sourceBaseDir, source))
	if err != nil {
		return "", fmt.Errorf("invalid source file %q: %w", source, err)
	}

	if !isChildPath(tr.sourceBaseDir, absSourceFilePath) {
		return "", fmt.Errorf("invalid source file %q: not a child of %q", source, tr.sourceBaseDir)
	}

	b, err := os.ReadFile(absSourceFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file %q: %w", source, err)
	}

	return string(b), nil
}

func (tr *taskRunner) buildScriptsConfigMapObject(
	ctx context.Context,
) (*k8scorev1.ConfigMap, error) {
	data := map[string]string{}
	for _, file := range tr.taskConfig.Files {
		sourceFileContent, err := tr.resolveSourceFileContent(ctx, file.Source)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", file.Source, err)
		}
		data[file.Dest] = sourceFileContent
	}

	rv := &k8scorev1.ConfigMap{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      tr.scriptsConfigMapName(),
			Namespace: tr.objectNamespace(),
			Labels: map[string]string{
				labelKeyTaskName: tr.taskConfig.Name,
			},
		},
		Data: data,
	}

	return rv, nil
}
