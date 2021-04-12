// Package cmd is the entrypoint for cli commands
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/Cray-HPE/loftsman/internal"
	"github.com/Cray-HPE/loftsman/internal/logger"
)

const (
	helmCmdHelp         = "loftsman helm is no longer active, simply use the helm v3 CLI directly, this help will be removed in a coming release"
	manifestPathArgName = "manifest-path"
)

// Version is the CLI tool version, this should be overridden at build time for anything to be distributed
var Version = "dev"
var loftsman = internal.NewLoftsman()

var rootCmd = &cobra.Command{
	Use:   "loftsman",
	Short: "Managed releases of Helm chart workloads to a Kubernetes clusters",
	Long: fmt.Sprintf(`%s
Define, organize, and ship your Kubernetes workloads with Helm charts easily

Requirements:
  • helm v3 binary installed on the machine running this tool
  • A kubeconfig and context available locally with admin access to your cluster
  • A container images registry hosting your container images. If it's private,
    your Kubernetes cluster CRI should be configured to have access to pull from
    the registry, fill in your manifest value overrides to point to it.

"Loftsmen at the mould lofts of shipyards were responsible for taking
 the dimensions and details from drawings and plans, and translating
 this information into templates, battens, ordinates, cutting sketches,
 profiles, margins and other data."
   -- https://en.wikipedia.org/wiki/Lofting

Example Usage:

A pretty standard loftsman case of running shape and then ship:
  loftsman ship \
    --charts-repo https://charts.your-domain.com \
    --%s ./manifest.yaml

Certain components, files or programs contained within this package or product are
Copyright 2019 Cray/HPE Inc. All rights reserved.`, logger.GetHelpLogo(), manifestPathArgName),
}

var manifestCmd = &cobra.Command{
	Use:   internal.ManifestCmd,
	Short: "Operations related to creating and maintaining Loftsman manifests",
	Long: fmt.Sprintf(`%s
Provides an easy way to initialize/create new manifests, validate, and other related operations`, logger.GetHelpLogo()),
}

var manifestCreateCmd = &cobra.Command{
	Use:   internal.CreateCmd,
	Short: "Create a new manifest source file",
	Long: fmt.Sprintf(`%s
Create a new manifest source file with the basic structure in place for the current, default schema`, logger.GetHelpLogo()),
	PreRunE: commonPreRun,
	Run:     runManifestCreate,
}

var manifestValidateCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s PATH [PATH ...]", internal.ValidateCmd),
	Short: "Validates a manifest file against its schema",
	Long: fmt.Sprintf(`%s
Will take a manifest file in, identify the schema version, and validate it against its schema version`, logger.GetHelpLogo()),
	Args:    cobra.MinimumNArgs(1),
	PreRunE: commonPreRun,
	Run:     runManifestValidate,
}

var shipCmd = &cobra.Command{
	Use:   internal.ShipCmd,
	Short: "Ship out your Helm chart workloads to run in your Kubernetes cluster",
	Long: fmt.Sprintf(`%s
Shipping will prep your cluster and then ship out your Helm charts to install or upgrade your workloads in the cluster`, logger.GetHelpLogo()),
	PreRunE: commonPreRun,
	Run:     runShip,
}

var avastCmd = &cobra.Command{
	Use:   internal.AvastCmd,
	Short: "Halt or clear an existing ship command that's stuck",
	Long: fmt.Sprintf(`%s
Shipping a given manifest will prevent others from running over it, use this command if you've had a failure that leaves
a loftsman ship in the "running" state, halting that operation`, logger.GetHelpLogo()),
	PreRunE: commonPreRun,
	Run:     runAvast,
}

var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: fmt.Sprintf("REMOVED: %s", helmCmdHelp),
	Long: fmt.Sprintf(`%s
REMOVED: %s`, logger.GetHelpLogo(), helmCmdHelp),
	PreRunE: commonPreRun,
	Run:     runHelm,
}

func init() {
	rootCmd.Version = strings.TrimSpace(Version)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&loftsman.Settings.JSONLog.Path, "json-log-path", "", loftsman.Settings.JSONLog.Path,
		"Path to the file where JSON/machine-readable logs written, note that this is in addition to the stdout logs")
	rootCmd.PersistentFlags().StringVarP(&loftsman.Settings.Kubernetes.KubeconfigPath, "kubeconfig", "", loftsman.Settings.Kubernetes.KubeconfigPath,
		"Path to the Kubernetes config file to use (default is the system default)")
	rootCmd.PersistentFlags().StringVarP(&loftsman.Settings.Kubernetes.KubeContext, "kube-context", "", loftsman.Settings.Kubernetes.KubeContext,
		"The name of the Kubernetes config context to use (default is the current-context in kubeconfig used)")
	rootCmd.PersistentFlags().StringVarP(&loftsman.Settings.HelmExecConfig.Binary, "helm-binary", "", loftsman.Settings.HelmExecConfig.Binary,
		"The Helm binary to use, helpful in being able to have Helm 3 installed alternatively")
	rootCmd.PersistentFlags().StringVarP(&loftsman.Settings.Namespace, "loftsman-namespace", "", loftsman.Settings.Namespace,
		"The namespace where loftsman records are stored: manifests, logs, etc.")

	manifestCreateCmd.PersistentFlags().StringVarP(&loftsman.Settings.Manifest.ChartNames, "chart-names", "", "",
		"A comma-delimited list of charts to initialize in the manifest")

	shipCmd.PersistentFlags().StringVarP(&loftsman.Settings.ChartsSource.Repo, "charts-repo", "", "",
		"The root URL for an external helm chart repo to use for installing/upgrading charts. (required if not using charts-path)")
	shipCmd.PersistentFlags().StringVarP(&loftsman.Settings.ChartsSource.Path, "charts-path", "", "",
		"Local path to a directory containing helm-packaged charts, e.g. files like my-chart-0.1.0.tgz (required if not using charts-repo)")
	shipCmd.PersistentFlags().StringVarP(&loftsman.Settings.ChartsSource.RepoUsername, "charts-repo-username", "", "",
		"The username for charts-repo, if applicable")
	shipCmd.PersistentFlags().StringVarP(&loftsman.Settings.ChartsSource.RepoPassword, "charts-repo-password", "", "",
		"The password for charts-repo, if applicable")

	shipCmd.PersistentFlags().StringVarP(&loftsman.Settings.Manifest.Path, manifestPathArgName, "", "",
		"Local path to the Loftsman YAML manifest file, instruction on what charts to install and how to install them. See \n"+
			"loftsman manifest --help for more info (required)")

	avastCmd.PersistentFlags().StringVarP(&loftsman.Settings.Manifest.Path, manifestPathArgName, "", "",
		"Local path to the Loftsman YAML mainfest file, by name it will determine the existing loftsman ship to halt (required if not using manifest-name)")
	avastCmd.PersistentFlags().StringVarP(&loftsman.Settings.Manifest.Name, "manifest-name", "", "",
		fmt.Sprintf("The name of the manifest ship operation you want to halt (required if not using %s)", manifestPathArgName))

	manifestCmd.AddCommand(manifestCreateCmd, manifestValidateCmd)
	helmCmd.Flags().SetInterspersed(false)
	rootCmd.AddCommand(manifestCmd, shipCmd, avastCmd, helmCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer cleanup()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("loftsman")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
}

func cleanup() {
	loftsman.Settings.JSONLog.File.Close()
	if err := os.RemoveAll(loftsman.Settings.TempDirectory); err != nil {
		fmt.Println(fmt.Sprintf("Couldn't remove temp directory: %s", loftsman.Settings.TempDirectory))
	}
}

func commonPreRun(cmd *cobra.Command, args []string) error {
	var err error
	loftsman.Settings.RunID = fmt.Sprintf("%v", time.Now().Unix())
	loftsman.Settings.TempDirectory = filepath.Join(os.TempDir(), fmt.Sprintf("loftsman-%s", loftsman.Settings.RunID))
	if err = os.Mkdir(loftsman.Settings.TempDirectory, 0755); err != nil {
		fmt.Println(fmt.Sprintf("couldn't create temp directory at %s: %s", loftsman.Settings.TempDirectory, err))
		return err
	}
	loftsman.Settings.JSONLog.File, err = os.OpenFile(loftsman.Settings.JSONLog.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	commandString := cmd.Name()
	for cmd.HasParent() {
		if cmd.Parent().Name() != cmd.Root().Name() {
			commandString = fmt.Sprintf("%s %s", cmd.Parent().Name(), commandString)
		}
		cmd = cmd.Parent()
	}
	return loftsman.Initialize(commandString)
}

func runManifestCreate(cmd *cobra.Command, args []string) {
	if err := loftsman.ManifestCreate(); err != nil {
		os.Exit(1)
	}
}

func runManifestValidate(cmd *cobra.Command, args []string) {
	if err := loftsman.ManifestValidate(args...); err != nil {
		os.Exit(1)
	}
}

func runShip(cmd *cobra.Command, args []string) {
	if err := loftsman.Ship(); err != nil {
		os.Exit(1)
	}
}

func runAvast(cmd *cobra.Command, args []string) {
	if err := loftsman.Avast(); err != nil {
		os.Exit(1)
	}
}

func runHelm(cmd *cobra.Command, args []string) {
	fmt.Println(helmCmdHelp)
}
