// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type abapEnvironmentBuildOptions struct {
	CfAPIEndpoint                   string   `json:"cfApiEndpoint,omitempty"`
	CfOrg                           string   `json:"cfOrg,omitempty"`
	CfSpace                         string   `json:"cfSpace,omitempty"`
	CfServiceInstance               string   `json:"cfServiceInstance,omitempty"`
	CfServiceKeyName                string   `json:"cfServiceKeyName,omitempty"`
	Host                            string   `json:"host,omitempty"`
	AbapSourceClient                string   `json:"abapSourceClient,omitempty"`
	Username                        string   `json:"username,omitempty"`
	Password                        string   `json:"password,omitempty"`
	Phase                           string   `json:"phase,omitempty"`
	Values                          string   `json:"values,omitempty"`
	DownloadAllResultFiles          bool     `json:"downloadAllResultFiles,omitempty"`
	DownloadResultFilenames         []string `json:"downloadResultFilenames,omitempty"`
	PublishAllDownloadedResultFiles bool     `json:"publishAllDownloadedResultFiles,omitempty"`
	PublishResultFilenames          []string `json:"publishResultFilenames,omitempty"`
	SubDirectoryForDownload         string   `json:"subDirectoryForDownload,omitempty"`
	FilenamePrefixForDownload       string   `json:"filenamePrefixForDownload,omitempty"`
	TreatWarningsAsError            bool     `json:"treatWarningsAsError,omitempty"`
	MaxRuntimeInMinutes             int      `json:"maxRuntimeInMinutes,omitempty"`
	PollingIntervalInSeconds        int      `json:"pollingIntervalInSeconds,omitempty"`
	CertificateNames                []string `json:"certificateNames,omitempty"`
	CpeValues                       string   `json:"cpeValues,omitempty"`
	UseFieldsOfAddonDescriptor      string   `json:"useFieldsOfAddonDescriptor,omitempty"`
	ConditionOnAddonDescriptor      string   `json:"conditionOnAddonDescriptor,omitempty"`
	StopOnFirstError                bool     `json:"stopOnFirstError,omitempty"`
	AddonDescriptor                 string   `json:"addonDescriptor,omitempty"`
}

type abapEnvironmentBuildCommonPipelineEnvironment struct {
	abap struct {
		buildValues string
	}
}

func (p *abapEnvironmentBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "abap", name: "buildValues", value: p.abap.buildValues},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Piper environment")
	}
}

// AbapEnvironmentBuildCommand Executes builds as defined with the build framework
func AbapEnvironmentBuildCommand() *cobra.Command {
	const STEP_NAME = "abapEnvironmentBuild"

	metadata := abapEnvironmentBuildMetadata()
	var stepConfig abapEnvironmentBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment abapEnvironmentBuildCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createAbapEnvironmentBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes builds as defined with the build framework",
		Long:  `Executes builds as defined with the build framework. Transaction overview /n/BUILD/OVERVIEW`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, err := os.Getwd()
			if err != nil {
				return err
			}
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			vaultClient := config.GlobalVaultClient()
			if vaultClient != nil {
				defer vaultClient.MustRevokeToken()
			}

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.LogStepTelemetryData()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if GeneralConfig.HookConfig.GCPPubSubConfig.Enabled {
					err := gcp.NewGcpPubsubClient(
						vaultClient,
						GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
						GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.OIDCConfig.RoleID,
					).Publish(GeneralConfig.HookConfig.GCPPubSubConfig.Topic, telemetryClient.GetDataBytes())
					if err != nil {
						log.Entry().WithError(err).Warn("event publish failed")
					}
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(STEP_NAME)
			abapEnvironmentBuild(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addAbapEnvironmentBuildFlags(createAbapEnvironmentBuildCmd, &stepConfig)
	return createAbapEnvironmentBuildCmd
}

func addAbapEnvironmentBuildFlags(cmd *cobra.Command, stepConfig *abapEnvironmentBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.CfAPIEndpoint, "cfApiEndpoint", os.Getenv("PIPER_cfApiEndpoint"), "Cloud Foundry API endpoint")
	cmd.Flags().StringVar(&stepConfig.CfOrg, "cfOrg", os.Getenv("PIPER_cfOrg"), "Cloud Foundry target organization")
	cmd.Flags().StringVar(&stepConfig.CfSpace, "cfSpace", os.Getenv("PIPER_cfSpace"), "Cloud Foundry target space")
	cmd.Flags().StringVar(&stepConfig.CfServiceInstance, "cfServiceInstance", os.Getenv("PIPER_cfServiceInstance"), "Cloud Foundry Service Instance")
	cmd.Flags().StringVar(&stepConfig.CfServiceKeyName, "cfServiceKeyName", os.Getenv("PIPER_cfServiceKeyName"), "Cloud Foundry Service Key")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Specifies the host address of the SAP BTP ABAP Environment system")
	cmd.Flags().StringVar(&stepConfig.AbapSourceClient, "abapSourceClient", os.Getenv("PIPER_abapSourceClient"), "Specifies the client of the SAP BTP ABAP Environment system, use only in combination with host")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")
	cmd.Flags().StringVar(&stepConfig.Phase, "phase", os.Getenv("PIPER_phase"), "Phase as specified in the build script in the backend system")
	cmd.Flags().StringVar(&stepConfig.Values, "values", os.Getenv("PIPER_values"), "Input values for the build framework, please enter in the format '[{\"value_id\":\"Id1\",\"value\":\"value1\"},{\"value_id\":\"Id2\",\"value\":\"value2\"}]'")
	cmd.Flags().BoolVar(&stepConfig.DownloadAllResultFiles, "downloadAllResultFiles", false, "If true, all build artefacts are downloaded")
	cmd.Flags().StringSliceVar(&stepConfig.DownloadResultFilenames, "downloadResultFilenames", []string{}, "Only the specified files are downloaded. If downloadAllResultFiles is true, this parameter is ignored")
	cmd.Flags().BoolVar(&stepConfig.PublishAllDownloadedResultFiles, "publishAllDownloadedResultFiles", false, "If true, it publishes all downloaded files")
	cmd.Flags().StringSliceVar(&stepConfig.PublishResultFilenames, "publishResultFilenames", []string{}, "Only the specified files get published, in case the file was not downloaded before an error occures")
	cmd.Flags().StringVar(&stepConfig.SubDirectoryForDownload, "subDirectoryForDownload", os.Getenv("PIPER_subDirectoryForDownload"), "Target directory to store the downloaded files, {buildID} and {taskID} can be used and will be resolved accordingly")
	cmd.Flags().StringVar(&stepConfig.FilenamePrefixForDownload, "filenamePrefixForDownload", os.Getenv("PIPER_filenamePrefixForDownload"), "Filename prefix for the downloaded files, {buildID} and {taskID} can be used and will be resolved accordingly")
	cmd.Flags().BoolVar(&stepConfig.TreatWarningsAsError, "treatWarningsAsError", false, "If a warrning occures, the step will be set to unstable")
	cmd.Flags().IntVar(&stepConfig.MaxRuntimeInMinutes, "maxRuntimeInMinutes", 360, "maximal runtime of the step in minutes")
	cmd.Flags().IntVar(&stepConfig.PollingIntervalInSeconds, "pollingIntervalInSeconds", 60, "wait time in seconds till next status request in the backend system")
	cmd.Flags().StringSliceVar(&stepConfig.CertificateNames, "certificateNames", []string{}, "file names of trusted (self-signed) server certificates - need to be stored in .pipeline/trustStore")
	cmd.Flags().StringVar(&stepConfig.CpeValues, "cpeValues", os.Getenv("PIPER_cpeValues"), "Values taken from the previous step, if a value was also specified in the config file, the value from cpe will be discarded")
	cmd.Flags().StringVar(&stepConfig.UseFieldsOfAddonDescriptor, "useFieldsOfAddonDescriptor", os.Getenv("PIPER_useFieldsOfAddonDescriptor"), "use fields of the addonDescriptor in the cpe as input values. Please enter in the format '[{\"use\":\"Name\",\"renameTo\":\"SWC\"}]'")
	cmd.Flags().StringVar(&stepConfig.ConditionOnAddonDescriptor, "conditionOnAddonDescriptor", os.Getenv("PIPER_conditionOnAddonDescriptor"), "normally if useFieldsOfAddonDescriptor is not initial, a build is triggered for each repository in the addonDescriptor. This can be changed by posing conditions. Please enter in the format '[{\"field\":\"Status\",\"operator\":\"==\",\"value\":\"P\"}]'")
	cmd.Flags().BoolVar(&stepConfig.StopOnFirstError, "stopOnFirstError", false, "If false, it does not stop if an error occured for one repository in the addonDescriptor, but continues with the next repository. However the step is marked as failed in the end if an error occured.")
	cmd.Flags().StringVar(&stepConfig.AddonDescriptor, "addonDescriptor", os.Getenv("PIPER_addonDescriptor"), "Structure in the commonPipelineEnvironment containing information about the Product Version and corresponding Software Component Versions")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("phase")
	cmd.MarkFlagRequired("downloadAllResultFiles")
	cmd.MarkFlagRequired("publishAllDownloadedResultFiles")
	cmd.MarkFlagRequired("treatWarningsAsError")
	cmd.MarkFlagRequired("maxRuntimeInMinutes")
	cmd.MarkFlagRequired("pollingIntervalInSeconds")
}

// retrieve step metadata
func abapEnvironmentBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "abapEnvironmentBuild",
			Aliases:     []config.Alias{},
			Description: "Executes builds as defined with the build framework",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapCredentialsId", Description: "Jenkins credentials ID containing user and password to authenticate to the Cloud Platform ABAP Environment system or the Cloud Foundry API", Type: "jenkins", Aliases: []config.Alias{{Name: "cfCredentialsId", Deprecated: false}, {Name: "credentialsId", Deprecated: false}}},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "cfApiEndpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/apiEndpoint"}},
						Default:     os.Getenv("PIPER_cfApiEndpoint"),
					},
					{
						Name:        "cfOrg",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/org"}},
						Default:     os.Getenv("PIPER_cfOrg"),
					},
					{
						Name:        "cfSpace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/space"}},
						Default:     os.Getenv("PIPER_cfSpace"),
					},
					{
						Name:        "cfServiceInstance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceInstance"}},
						Default:     os.Getenv("PIPER_cfServiceInstance"),
					},
					{
						Name:        "cfServiceKeyName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "cloudFoundry/serviceKey"}, {Name: "cloudFoundry/serviceKeyName"}, {Name: "cfServiceKey"}},
						Default:     os.Getenv("PIPER_cfServiceKeyName"),
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_host"),
					},
					{
						Name:        "abapSourceClient",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapSourceClient"),
					},
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_username"),
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_password"),
					},
					{
						Name:        "phase",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_phase"),
					},
					{
						Name:        "values",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_values"),
					},
					{
						Name:        "downloadAllResultFiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "downloadResultFilenames",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "publishAllDownloadedResultFiles",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "publishResultFilenames",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "subDirectoryForDownload",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_subDirectoryForDownload"),
					},
					{
						Name:        "filenamePrefixForDownload",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_filenamePrefixForDownload"),
					},
					{
						Name:        "treatWarningsAsError",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "maxRuntimeInMinutes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     360,
					},
					{
						Name:        "pollingIntervalInSeconds",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "int",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     60,
					},
					{
						Name:        "certificateNames",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "cpeValues",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "abap/buildValues",
							},
						},
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_cpeValues"),
					},
					{
						Name:        "useFieldsOfAddonDescriptor",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_useFieldsOfAddonDescriptor"),
					},
					{
						Name:        "conditionOnAddonDescriptor",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_conditionOnAddonDescriptor"),
					},
					{
						Name:        "stopOnFirstError",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "addonDescriptor",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "abap/addonDescriptor",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_addonDescriptor"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "cf", Image: "ppiper/cf-cli:latest"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "abap/buildValues"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
