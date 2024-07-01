install-agent-prod:
	@echo "Installing agent..."
	helm install dd  --set datadog.apiKey=${DD_API_KEY} -f values.yaml datadog/datadog
	@echo "Agent installed successfully."

install-agent:
	@echo "Installing agent..."
	helm install dd -f values.yaml ./datadog  --set datadog.otelcollector.enabled=true --set-file datadog.otelcollector.config=config.yaml --set datadog.apiKey=${DD_API_KEY}
	@echo "Agent installed successfully."

uninstall-agent:
	@echo "Uninstalling agent..."
	helm uninstall dd
	@echo "Agent uninstalled successfully."
