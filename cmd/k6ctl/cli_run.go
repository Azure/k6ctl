package main

import (
	"context"
	"path/filepath"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/target"
	"github.com/Azure/k6ctl/internal/task"
)

type CLIRun struct {
	Kubeconfig string `required:"" type:"existingfile" env:"KUBECONFIG" long:"kubeconfig" description:"Path to the kubeconfig file to use for CLI requests"`
	TaskConfig string `required:"" default:"k6ctl.yaml" type:"existingfile" short:"c" long:"config" description:"Path to the task config file to use for CLI requests"`
	BaseDir    string `required:"" default:"." type:"existingdir" short:"d" long:"base-dir" description:"Base directory to use for relative paths"`
	Script     string `arg:"" default:"script.js" description:"Script to run"`
}

func (c *CLIRun) Run() error {
	t := &target.StaticTarget{
		Kubeconfig: c.Kubeconfig,
	}

	taskConfig, err := task.LoadSchemaFromFile(c.TaskConfig)
	if err != nil {
		return err
	}

	baseDir, err := filepath.Abs(filepath.Clean(c.BaseDir))
	if err != nil {
		return err
	}

	cpRegistry := config.NewRegistry()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopConfigPlugins, err := task.LoadConfigPlugins(
		ctx,
		cpRegistry,
		taskConfig.K6,
	)
	if err != nil {
		return err
	}
	defer stopConfigPlugins()

	if err := task.RunTask(
		ctx,
		t,
		cpRegistry.GetByName,
		taskConfig,
		baseDir,
		c.Script,
	); err != nil {
		return err
	}

	return nil
}