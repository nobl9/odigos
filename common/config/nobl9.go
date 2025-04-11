package config

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/odigos-io/odigos/common"
)

const (
	nobl9SLICEnvironment  = "NOBL9_SLIC_ENVIRONMENT"
	nobl9SLICClientID     = "NOBL9_SLIC_CLIENT_ID"
	nobl9SLICClientSecret = "NOBL9_SLIC_CLIENT_SECRET"
)

var (
	nobl9SLICURLs = map[string]string{
		"dev":      "https://suym.nobl9.dev",
		"prod-app": "https://api.slic.nobl9.com",
	}
)

type Nobl9SLIC struct{}

func (n *Nobl9SLIC) DestType() common.DestinationType {
	return common.Nobl9DestinationType
}

func (n *Nobl9SLIC) ModifyConfig(dest ExporterConfigurer, currentConfig *Config) ([]string, error) {
	environment, exists := dest.GetConfig()[nobl9SLICEnvironment]
	if !exists {
		return nil, errors.New("OTLP http endpoint not specified, gateway will not be configured for otlp http")
	}

	url := nobl9SLICURLs[environment]
	endpoint, err := parseOtlpHttpEndpoint(url, "", "/opentelemetry")
	if err != nil {
		return nil, errors.Join(err, errors.New("otlp http endpoint invalid, gateway will not be configured for otlp http"))
	}

	token, err := n.getToken(dest)
	if err != nil {
		return nil, err
	}
	uniqueUri := environment + "-" + dest.GetID()
	exporterName := "nobl9/" + uniqueUri
	exporterConfig := GenericMap{
		"endpoint": endpoint,
		"headers": GenericMap{
			"Authorization": "Basic " + token,
		},
	}

	currentConfig.Exporters[exporterName] = exporterConfig
	var pipelineNames []string
	if isTracingEnabled(dest) {
		tracesPipelineName := "traces/" + uniqueUri
		currentConfig.Service.Pipelines[tracesPipelineName] = Pipeline{
			Exporters: []string{exporterName},
		}
		pipelineNames = append(pipelineNames, tracesPipelineName)
	}

	if isMetricsEnabled(dest) {
		metricsPipelineName := "metrics/" + uniqueUri
		currentConfig.Service.Pipelines[metricsPipelineName] = Pipeline{
			Exporters: []string{exporterName},
		}
		pipelineNames = append(pipelineNames, metricsPipelineName)
	}

	return pipelineNames, nil
}

func (n *Nobl9SLIC) getToken(dest ExporterConfigurer) (string, error) {
	clientID, exists := dest.GetConfig()[nobl9SLICClientID]
	if !exists {
		return "", errors.New("Nobl9 SLIC client ID not specified, token will not be generated")
	}
	clientSecret, exists := dest.GetConfig()[nobl9SLICClientSecret]
	if !exists {
		return "", errors.New("Nobl9 SLIC client secret not specified, token will not be generated")
	}

	data := fmt.Sprintf("%s:%s", clientID, clientSecret)
	token := base64.StdEncoding.EncodeToString([]byte(data))
	return token, nil
}
