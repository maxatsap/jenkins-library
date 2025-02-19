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

type gcpPublishEventOptions struct {
	VaultNamespace                  string `json:"vaultNamespace,omitempty"`
	VaultServerURL                  string `json:"vaultServerUrl,omitempty"`
	OIDCToken                       string `json:"OIDCToken,omitempty"`
	GcpProjectNumber                string `json:"gcpProjectNumber,omitempty"`
	GcpWorkloadIDentityPool         string `json:"gcpWorkloadIdentityPool,omitempty"`
	GcpWorkloadIDentityPoolProvider string `json:"gcpWorkloadIdentityPoolProvider,omitempty"`
	Topic                           string `json:"topic,omitempty"`
	EventSource                     string `json:"eventSource,omitempty"`
	EventType                       string `json:"eventType,omitempty"`
	EventData                       string `json:"eventData,omitempty"`
	AdditionalEventData             string `json:"additionalEventData,omitempty"`
}

// GcpPublishEventCommand Publishes an event to GCP using OIDC authentication (beta)
func GcpPublishEventCommand() *cobra.Command {
	const STEP_NAME = "gcpPublishEvent"

	metadata := gcpPublishEventMetadata()
	var stepConfig gcpPublishEventOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createGcpPublishEventCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Publishes an event to GCP using OIDC authentication (beta)",
		Long: `This step is in beta.
Authentication to GCP is handled by an OIDC token received from, for example, Vault.`,
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
			gcpPublishEvent(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGcpPublishEventFlags(createGcpPublishEventCmd, &stepConfig)
	return createGcpPublishEventCmd
}

func addGcpPublishEventFlags(cmd *cobra.Command, stepConfig *gcpPublishEventOptions) {
	cmd.Flags().StringVar(&stepConfig.VaultNamespace, "vaultNamespace", os.Getenv("PIPER_vaultNamespace"), "")
	cmd.Flags().StringVar(&stepConfig.VaultServerURL, "vaultServerUrl", os.Getenv("PIPER_vaultServerUrl"), "")
	cmd.Flags().StringVar(&stepConfig.OIDCToken, "OIDCToken", os.Getenv("PIPER_OIDCToken"), "")
	cmd.Flags().StringVar(&stepConfig.GcpProjectNumber, "gcpProjectNumber", os.Getenv("PIPER_gcpProjectNumber"), "")
	cmd.Flags().StringVar(&stepConfig.GcpWorkloadIDentityPool, "gcpWorkloadIdentityPool", os.Getenv("PIPER_gcpWorkloadIdentityPool"), "A workload identity pool is an entity that lets you manage external identities.")
	cmd.Flags().StringVar(&stepConfig.GcpWorkloadIDentityPoolProvider, "gcpWorkloadIdentityPoolProvider", os.Getenv("PIPER_gcpWorkloadIdentityPoolProvider"), "A workload identity pool provider is an entity that describes a relationship between Google Cloud and your IdP.")
	cmd.Flags().StringVar(&stepConfig.Topic, "topic", os.Getenv("PIPER_topic"), "The pubsub topic to which the message is published.")
	cmd.Flags().StringVar(&stepConfig.EventSource, "eventSource", os.Getenv("PIPER_eventSource"), "The events source as defined by CDEvents.")
	cmd.Flags().StringVar(&stepConfig.EventType, "eventType", os.Getenv("PIPER_eventType"), "")
	cmd.Flags().StringVar(&stepConfig.EventData, "eventData", os.Getenv("PIPER_eventData"), "Data to be merged with the generated data for the cloud event data field (JSON)")
	cmd.Flags().StringVar(&stepConfig.AdditionalEventData, "additionalEventData", os.Getenv("PIPER_additionalEventData"), "Data (formatted as JSON string) to add to eventData. This can be used to enrich eventData that comes from the pipeline environment.")

}

// retrieve step metadata
func gcpPublishEventMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gcpPublishEvent",
			Aliases:     []config.Alias{},
			Description: "Publishes an event to GCP using OIDC authentication (beta)",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "vaultNamespace",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_vaultNamespace"),
					},
					{
						Name:        "vaultServerUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_vaultServerUrl"),
					},
					{
						Name:        "OIDCToken",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_OIDCToken"),
					},
					{
						Name:        "gcpProjectNumber",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_gcpProjectNumber"),
					},
					{
						Name:        "gcpWorkloadIdentityPool",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_gcpWorkloadIdentityPool"),
					},
					{
						Name:        "gcpWorkloadIdentityPoolProvider",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_gcpWorkloadIdentityPoolProvider"),
					},
					{
						Name:        "topic",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_topic"),
					},
					{
						Name:        "eventSource",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_eventSource"),
					},
					{
						Name:        "eventType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_eventType"),
					},
					{
						Name: "eventData",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/eventData",
							},
						},
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_eventData"),
					},
					{
						Name:        "additionalEventData",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_additionalEventData"),
					},
				},
			},
		},
	}
	return theMetaData
}
