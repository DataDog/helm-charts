install-agent:
	@echo "Installing agent..."
	helm install dd -f values.yaml ./datadog  --set-file datadog.otelcollector.config=config.yaml --set datadog.apiKey=${DD_API_KEY}
	@chmod +x /usr/local/bin/agent
	@echo "Agent installed successfully."

uninstall-agent:
	@echo "Uninstalling agent..."
	helm uninstall dd
	@echo "Agent uninstalled successfully."
