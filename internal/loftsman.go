// Package internal is for internal loftsman operations
package internal

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Cray-HPE/loftsman/internal/helm"
	"github.com/Cray-HPE/loftsman/internal/interfaces"
	"github.com/Cray-HPE/loftsman/internal/kubernetes"
	"github.com/Cray-HPE/loftsman/internal/logger"
	"github.com/Cray-HPE/loftsman/internal/manifest"
	"github.com/Cray-HPE/loftsman/internal/settings"
)

const (
	// ShipCmd is the cli ship command identifier
	ShipCmd = "ship"
	// ManifestCmd is the cli manifest command identifier
	ManifestCmd = "manifest"
	// CreateCmd is the cli create command identifier
	CreateCmd = "create"
	// ValidateCmd is the cli validate command identifier
	ValidateCmd = "validate"
	// AvastCmd is the cli avast command identifier
	AvastCmd = "avast"

	statusKey             = "status"
	statusActive          = "active"
	statusFailed          = "failed"
	statusSuccess         = "success"
	statusCancelled       = "cancelled"
	statusCrashed         = "crashed"
	statusAvasted         = "avasted"
	configMapNameTemplate = "loftsman-%s"
)

// To reduce the need for always initializing cluster connectivity and internal objects
// like Helm and Kubernetes, we can maintain this simple list of only the commands that
// actually require it
var commandsRequiringClusterConnectivity = []string{
	ShipCmd,
	AvastCmd,
}

// Loftsman is the central object for loftsman operations, settings, data, etc.
type Loftsman struct {
	Settings   *settings.Settings
	reader     io.Reader
	manifest   interfaces.Manifest
	logger     *logger.Logger
	kubernetes interfaces.Kubernetes
	helm       interfaces.Helm
}

// Initialize will go through the process of initializing or setting up common needs/objects across all commands
func (loftsman *Loftsman) Initialize(commandString string) error {
	var err error
	loftsman.logger = logger.New(loftsman.Settings.JSONLog.File, commandString)

	for _, commandRequiringClusterConnectivity := range commandsRequiringClusterConnectivity {
		if commandRequiringClusterConnectivity == commandString {
			kubeconfigUsed := loftsman.Settings.Kubernetes.KubeconfigPath
			if kubeconfigUsed == "" {
				kubeconfigUsed = "(system default)"
			}
			kubeContextUsed := loftsman.Settings.Kubernetes.KubeContext
			if kubeContextUsed == "" {
				kubeContextUsed = "(current-context)"
			}

			loftsman.logger.Info().Msgf("Initializing the connection to the Kubernetes cluster using KUBECONFIG %s, and context %s", kubeconfigUsed, kubeContextUsed)
			if err = loftsman.kubernetes.Initialize(loftsman.Settings.Kubernetes.KubeconfigPath, loftsman.Settings.Kubernetes.KubeContext); err != nil {
				return err
			}
			loftsman.logger.Info().Msg("Initializing helm client object")
			loftsman.Settings.HelmExecConfig.KubeconfigPath = loftsman.Settings.Kubernetes.KubeconfigPath
			loftsman.Settings.HelmExecConfig.KubeContext = loftsman.Settings.Kubernetes.KubeContext
			if err = loftsman.helm.Initialize(loftsman.Settings.HelmExecConfig, loftsman.Settings.ChartsSource); err != nil {
				return err
			}
			break
		}
	}

	if loftsman.Settings.Manifest.Path != "" && loftsman.manifest == nil {
		if err = loftsman.Settings.ValidateManifestPath(); err != nil {
			return err
		}
		loftsman.Settings.Manifest.Content, err = ioutil.ReadFile(loftsman.Settings.Manifest.Path)
		if err != nil {
			return err
		}
		loftsman.manifest, err = manifest.Validate(string(loftsman.Settings.Manifest.Content))
		if err != nil {
			return err
		}
	}
	if loftsman.Settings.Manifest.Name == "" && loftsman.manifest != nil {
		loftsman.Settings.Manifest.Name = loftsman.manifest.GetName()
	}

	return nil
}

// Ship is the main operation to prep and ship out workloads to the cluster via Helm, etc.
func (loftsman *Loftsman) Ship() error {
	var err error

	if err = loftsman.Settings.ValidateChartsSource(); err != nil {
		return loftsman.fail(err)
	}

	loftsman.logger.Header("Shipping your Helm workloads with Loftsman")

	configMapName := fmt.Sprintf(configMapNameTemplate, loftsman.Settings.Manifest.Name)
	configMapData := make(map[string]string)

	loftsman.logger.Info().Msgf("Ensuring that the %s namespace exists", loftsman.Settings.Namespace)
	if err = loftsman.kubernetes.EnsureNamespace(loftsman.Settings.Namespace); err != nil {
		return loftsman.fail(fmt.Errorf("Error ensuring that the %s namespace exists: %s", loftsman.Settings.Namespace, err))
	}
	activeConfigMap, err := loftsman.kubernetes.FindConfigMap(configMapName, loftsman.Settings.Namespace, statusKey, statusActive)
	if err != nil {
		return loftsman.fail(fmt.Errorf("Error determining if another loftsman ship is in progress for manifest %s: %s", loftsman.Settings.Manifest.Name, err))
	}
	if activeConfigMap != nil {
		return loftsman.fail(fmt.Errorf(
			"There's another loftsman ship in progress for manifest %s in this cluster, please wait and try again in a bit, or use `loftsman avast` to cancel it",
			loftsman.Settings.Manifest.Name))
	}

	if loftsman.Settings.ChartsSource.Path != "" {
		loftsman.logger.Info().Msgf("Loftsman will use the packaged charts at %s as the Helm install source", loftsman.Settings.ChartsSource.Path)
	} else if loftsman.Settings.ChartsSource.Repo != "" {
		loftsman.logger.Info().Msgf("Loftsman will use the charts repo at %s as the Helm install source", loftsman.Settings.ChartsSource.Repo)
		if loftsman.Settings.ChartsSource.RepoUsername != "" && loftsman.Settings.ChartsSource.RepoPassword != "" {
			loftsman.logger.Info().Msgf("Charts repo access will authenticate with credentials: %s/*********", loftsman.Settings.ChartsSource.RepoUsername)
		}
	}

	loftsman.logger.Info().Msgf("Running a release for the provided manifest at %s", loftsman.Settings.Manifest.Path)

	configMapData[statusKey] = statusActive
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigChannel
		loftsman.recordShipResult(configMapName, configMapData, statusCancelled)
		os.Exit(0)
	}()
	if _, err := loftsman.kubernetes.InitializeConfigMap(configMapName, loftsman.Settings.Namespace, configMapData); err != nil {
		return loftsman.fail(fmt.Errorf("Error creating ship configmap %s in namespace %s: %s", configMapName, loftsman.Settings.Namespace, err))
	}
	crashHandler := func() {
		if r := recover(); r != nil {
			loftsman.recordShipResult(configMapName, configMapData, statusCrashed)
			loftsman.fail(fmt.Errorf("%v", r))
		}
	}
	defer crashHandler()
	loftsman.manifest.SetLogger(loftsman.logger)
	loftsman.manifest.SetTempDirectory(loftsman.Settings.TempDirectory)
	releaseErrors := loftsman.manifest.Release(loftsman.kubernetes, loftsman.helm)
	releaseStatus := statusSuccess
	if len(releaseErrors) > 0 {
		releaseStatus = statusFailed
	}
	loftsman.recordShipResult(configMapName, configMapData, releaseStatus)

	if len(releaseErrors) > 0 {
		loftsman.logger.ClosingHeader("Encountered errors during the manifest release:")
		for _, releaseError := range releaseErrors {
			loftsman.logger.Error().
				Str("chart", releaseError.Chart).
				Str("version", releaseError.Version).
				Str("namespace", releaseError.Namespace).
				Msg(strings.TrimSpace(releaseError.Error.Error()))
			fmt.Println("")
		}
		return loftsman.fail(errors.New("Some charts did not release successfully, see above and/or the output log file for more info"))
	}
	return nil
}

func (loftsman *Loftsman) recordShipResult(configMapName string, configMapData map[string]string, status string) {
	loftsman.logger.Info().Msgf("Ship status: %s. Recording status, manifest, and log data to configmap %s in namespace %s", status,
		configMapName, loftsman.Settings.Namespace)
	configMapData["manifest.yaml"] = string(loftsman.Settings.Manifest.Content)
	configMapData["loftsman.log"] = loftsman.logger.GetRecord()
	configMapData["status"] = status
	if _, err := loftsman.kubernetes.PatchConfigMap(configMapName, loftsman.Settings.Namespace, configMapData); err != nil {
		loftsman.logger.Error().Err(fmt.Errorf("Error patching configmap %s with result, manifest, and log data to the %s namespace: %s",
			configMapName, loftsman.Settings.Namespace, err)).Msg("")
		fmt.Println("")
	}
}

// ManifestCreate will create a new manifest and output it to stdout
func (loftsman *Loftsman) ManifestCreate() error {
	var err error
	manifestContent, err := manifest.Create(strings.Split(loftsman.Settings.Manifest.ChartNames, ","))
	if err != nil {
		return loftsman.fail(err)
	}
	fmt.Println(manifestContent)
	return nil
}

// ManifestValidate will validate a manifest
func (loftsman *Loftsman) ManifestValidate(args ...string) error {
	var err error
	for _, manifestPath := range args {
		loftsman.Settings.Manifest.Path = manifestPath
		if err = loftsman.Settings.ValidateManifestPath(); err != nil {
			return loftsman.fail(err)
		}
		loftsman.logger.Info().Msgf("%s is valid!", loftsman.Settings.Manifest.Path)
	}
	return nil
}

// Avast will look for an existing, locked manifest being shipped and tell it to halt
// NOTE: this command/method won't actually be able to kill any existing loftsman cli operation
// in progress, this is more about updating the state stored in the cluster about a given ship
// run in the case of that getting stuck by catastrophic failures of the loftsman cli during
// a ship
func (loftsman *Loftsman) Avast() error {
	var err error
	var response string

	if loftsman.Settings.Manifest.Name == "" {
		return loftsman.fail(errors.New("Unable to determine manifest name in order to avast, one of a manifest path or name must be provided"))
	}

	loftsman.logger.Header(fmt.Sprintf("Clearing/halting any ship in progress for manifest: %s", loftsman.Settings.Manifest.Name))

	fmt.Println(fmt.Sprintf(`WARNING: loftsman avast is currently used for unlocking stuck ship runs only, so those that might have
         encountered fatal errors and left the loftsman cluster state stuck in the running position.")
         If you have other loftsman ship runs going for the manifest, %s, please cancel them by simply")
         killing the process. Use avast only to recover from bad loftsman states.`, loftsman.Settings.Manifest.Name))
	fmt.Print("Do you want to continue? (only a response of 'yes' will continue with the avast operation): ")
	_, err = fmt.Fscanln(loftsman.reader, &response)
	if err != nil {
		return loftsman.fail(err)
	}
	if response != "yes" {
		loftsman.logger.Info().Msgf("User did not enter 'yes', not running avast")
		return nil
	}

	configMapName := fmt.Sprintf(configMapNameTemplate, loftsman.Settings.Manifest.Name)

	activeConfigMap, err := loftsman.kubernetes.FindConfigMap(configMapName, loftsman.Settings.Namespace, statusKey, statusActive)
	if err != nil {
		return loftsman.fail(fmt.Errorf("Error determining if another loftsman ship is in progress for manifest %s: %s", loftsman.Settings.Manifest.Name, err))
	}
	if activeConfigMap == nil {
		return loftsman.fail(fmt.Errorf("Couldn't find an active ship in progress for manifest: %s", loftsman.Settings.Manifest.Name))
	}
	activeConfigMap.Data[statusKey] = statusAvasted
	if _, err := loftsman.kubernetes.PatchConfigMap(configMapName, loftsman.Settings.Namespace, activeConfigMap.Data); err != nil {
		loftsman.logger.Error().Err(fmt.Errorf("Error patching configmap %s with avasted status to the %s namespace: %s",
			configMapName, loftsman.Settings.Namespace, err)).Msg("")
		fmt.Println("")
	}

	return nil
}

func (loftsman *Loftsman) fail(err error) error {
	loftsman.logger.Error().Err(err).Msg("")
	return err
}

// NewLoftsman will return the default, initial Loftsman object
func NewLoftsman() *Loftsman {
	return &Loftsman{
		Settings:   settings.New(),
		reader:     os.Stdin,
		manifest:   nil,
		kubernetes: &kubernetes.Kubernetes{},
		helm:       &helm.Helm{},
	}
}
