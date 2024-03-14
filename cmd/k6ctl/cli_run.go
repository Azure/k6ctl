package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/Azure/k6ctl/internal/config"
	coreconfig "github.com/Azure/k6ctl/internal/config/core"
	"github.com/Azure/k6ctl/internal/target"
	"github.com/Azure/k6ctl/internal/task"
)

const defaultTaskConfigFile = "k6ctl.yaml"

type CLIRun struct {
	Kubeconfig   string            `required:"" type:"existingfile" env:"KUBECONFIG" long:"kubeconfig" help:"Path to the kubeconfig file to use for CLI requests"`
	TaskConfig   string            `type:"existingfile" short:"c" long:"config" help:"Path to the task config file to use for CLI requests"`
	BaseDir      string            `required:"" default:"." type:"existingdir" short:"d" long:"base-dir" help:"Base directory to use for relative paths"`
	Script       string            `arg:"" default:"script.js" help:"Script to run"`
	NoFollowLogs bool              `default:"false" long:"no-follow-logs" help:"Do not follow logs"`
	Parameters   map[string]string `short:"p" long:"parameter" help:"Parameters to pass to the script (can be used multiple times)"`
}

func (c *CLIRun) resolveTaskConfig(baseDir string, taskConfigFile string) (*task.Schema, error) {
	if taskConfigFile == "" {
		// defaults to k6ctl.yaml in the base directory
		taskConfigFile = filepath.Join(baseDir, defaultTaskConfigFile)
	}

	stat, err := os.Stat(taskConfigFile)
	if err != nil {
		return nil, fmt.Errorf("no task config file found at %q", taskConfigFile)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("task config file %q is a directory", taskConfigFile)
	}

	return task.LoadSchemaFromFile(taskConfigFile)
}

func (c *CLIRun) Run() error {
	t := &target.StaticTarget{
		Kubeconfig: c.Kubeconfig,
	}

	baseDir, err := filepath.Abs(filepath.Clean(c.BaseDir))
	if err != nil {
		return err
	}

	taskConfig, err := c.resolveTaskConfig(baseDir, c.TaskConfig)
	if err != nil {
		return err
	}

	cpRegistry := config.NewRegistry()

	if err := coreconfig.RegisterProviders(
		cpRegistry,
		taskConfig.Configs,
		c.Parameters,
	); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
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
		task.WithFollowLogs(!c.NoFollowLogs),
	); err != nil {
		return err
	}

	return nil
}
