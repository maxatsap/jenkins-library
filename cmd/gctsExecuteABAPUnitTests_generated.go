// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type gctsExecuteABAPUnitTestsOptions struct {
	Username             string                 `json:"username,omitempty"`
	Password             string                 `json:"password,omitempty"`
	Host                 string                 `json:"host,omitempty"`
	Repository           string                 `json:"repository,omitempty"`
	Client               string                 `json:"client,omitempty"`
	AUnitTest            bool                   `json:"aUnitTest,omitempty"`
	AtcCheck             bool                   `json:"atcCheck,omitempty"`
	AtcVariant           string                 `json:"atcVariant,omitempty"`
	Scope                string                 `json:"scope,omitempty"`
	Commit               string                 `json:"commit,omitempty"`
	Workspace            string                 `json:"workspace,omitempty"`
	AtcResultsFileName   string                 `json:"atcResultsFileName,omitempty"`
	AUnitResultsFileName string                 `json:"aUnitResultsFileName,omitempty"`
	QueryParameters      map[string]interface{} `json:"queryParameters,omitempty"`
	SkipSSLVerification  bool                   `json:"skipSSLVerification,omitempty"`
}

// GctsExecuteABAPUnitTestsCommand Runs ABAP unit tests and ATC (ABAP Test Cockpit) checks for a specified object scope.
func GctsExecuteABAPUnitTestsCommand() *cobra.Command {
	const STEP_NAME = "gctsExecuteABAPUnitTests"

	metadata := gctsExecuteABAPUnitTestsMetadata()
	var stepConfig gctsExecuteABAPUnitTestsOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createGctsExecuteABAPUnitTestsCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Runs ABAP unit tests and ATC (ABAP Test Cockpit) checks for a specified object scope.",
		Long:  `This step executes ABAP unit test and ATC checks for a specified scope of objects that exist in a local Git repository on an ABAP system.`,
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
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
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
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME, GeneralConfig.HookConfig.PendoConfig.Token)
			gctsExecuteABAPUnitTests(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGctsExecuteABAPUnitTestsFlags(createGctsExecuteABAPUnitTestsCmd, &stepConfig)
	return createGctsExecuteABAPUnitTestsCmd
}

func addGctsExecuteABAPUnitTestsFlags(cmd *cobra.Command, stepConfig *gctsExecuteABAPUnitTestsOptions) {
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User that authenticates to the ABAP system. **Note** - Don´t provide this parameter directly. Either set it in the environment, or in the Jenkins credentials store, and provide the ID as value of the `abapCredentialsId` parameter.")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password of the ABAP user that authenticates to the ABAP system. **Note** - Don´t provide this parameter directly. Either set it in the environment, or in the Jenkins credentials store, and provide the ID as value of the `abapCredentialsId` parameter.")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Protocol and host of the ABAP system, including the port. Please provide it in the format `<protocol>://<host>:<port>`. Supported protocols are `http` and `https`.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Name (ID) of the local repository on the ABAP system")
	cmd.Flags().StringVar(&stepConfig.Client, "client", os.Getenv("PIPER_client"), "Client of the ABAP system in which you want to execute the checks")
	cmd.Flags().BoolVar(&stepConfig.AUnitTest, "aUnitTest", true, "Indication whether you want to execute the ABAP Unit tests. If the ABAP Unit tests fail, the results are mapped as follows to the statuses available in the Jenkins plugin. (Successful ABAP unit test results are not displayed.)\n<br/>\n- Failed with severity `fatal` is displayed as an `Error`.\n<br/>\n- Failed with severity `critical` is displayed as an `Error`.\n<br/>\n- Failed with severity `warning` is displayed as a `Warning`.\n")
	cmd.Flags().BoolVar(&stepConfig.AtcCheck, "atcCheck", true, "Indication whether you want to execute the ATC checks. If the ATC checks result in errors with priorities 1 and 2, they are considered as failed. The results are mapped as follows to the statuses available in the Jenkins plugin. (Successful ATC check results are not displayed.)\n<br/>\n- Priorities 1 and 2 are displayed as an `Error`.\n<br/>\n- Priority 3 is displayed as a `Warning`.\n")
	cmd.Flags().StringVar(&stepConfig.AtcVariant, "atcVariant", `DEFAULT`, "Variant for ATC checks")
	cmd.Flags().StringVar(&stepConfig.Scope, "scope", `repository`, "Scope of objects for which you want to execute the checks:\n<br/>\n- `localChangedObjects`: The object scope is derived from the last activity in the local repository. The checks are executed for the individual objects.\n<br/>\n- `remoteChangedObjects`: The object scope is the delta between the commit that triggered the pipeline and the current commit in the remote repository. The checks are executed for the individual objects.\n<br/>\n- `localChangedPackages`: The object scope is derived from the last activity in the local repository. All objects are resolved into packages. The checks are executed for the packages.\n<br/>\n- `remoteChangedPackages`: The object scope is the delta between the commit that triggered the pipeline and the current commit in the remote repository. All objects are resolved into packages. The checks are executed for the packages.\n<br/>\n- `repository`: The object scope comprises all objects that are part of the local repository. The checks are executed for the individual objects. Packages (DEVC) are excluded. This is the default scope.\n<br/>\n- `packages`: The object scope comprises all packages that are part of the local repository. The checks are executed for the packages.\n")
	cmd.Flags().StringVar(&stepConfig.Commit, "commit", os.Getenv("PIPER_commit"), "ID of the commit that triggered the pipeline or any other commit used to calculate the object scope. Specifying a commit is mandatory for the `remoteChangedObjects` and `remoteChangedPackages` scopes.")
	cmd.Flags().StringVar(&stepConfig.Workspace, "workspace", os.Getenv("PIPER_workspace"), "Absolute path to the directory that contains the source code that your CI/CD tool checks out. For example, in Jenkins, the workspace parameter is `/var/jenkins_home/workspace/<jobName>/`. As an alternative, you can use Jenkins's predefined environmental variable `WORKSPACE`.")
	cmd.Flags().StringVar(&stepConfig.AtcResultsFileName, "atcResultsFileName", `ATCResults.xml`, "Specifies an output file name for the results of the ATC checks.")
	cmd.Flags().StringVar(&stepConfig.AUnitResultsFileName, "aUnitResultsFileName", `AUnitResults.xml`, "Specifies an output file name for the results of the ABAP Unit tests.")

	cmd.Flags().BoolVar(&stepConfig.SkipSSLVerification, "skipSSLVerification", false, "Skip the verification of SSL (Secure Socket Layer) certificates when using HTTPS. This parameter is **not recommended** for productive environments.")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("client")
	cmd.MarkFlagRequired("workspace")
}

// retrieve step metadata
func gctsExecuteABAPUnitTestsMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gctsExecuteABAPUnitTests",
			Aliases:     []config.Alias{},
			Description: "Runs ABAP unit tests and ATC (ABAP Test Cockpit) checks for a specified object scope.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapCredentialsId", Description: "ID taken from the Jenkins credentials store containing user name and password of the user that authenticates to the ABAP system on which you want to execute the checks.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_username"),
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_password"),
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_host"),
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_repository"),
					},
					{
						Name:        "client",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_client"),
					},
					{
						Name:        "aUnitTest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "atcCheck",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "atcVariant",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `DEFAULT`,
					},
					{
						Name:        "scope",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `repository`,
					},
					{
						Name:        "commit",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_commit"),
					},
					{
						Name:        "workspace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_workspace"),
					},
					{
						Name:        "atcResultsFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `ATCResults.xml`,
					},
					{
						Name:        "aUnitResultsFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `AUnitResults.xml`,
					},
					{
						Name:        "queryParameters",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "skipSSLVerification",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
				},
			},
		},
	}
	return theMetaData
}
