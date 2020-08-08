# github-dependabot-prometheus

Prometheus exporter for GitHub Pull Requests opened by dependabot

## Metrics

This exporter exposes metrics like below:

### github_dependabot_open_pull_requests

```
github_dependabot_open_pull_requests{directory="frontend",from_version="3.8.2",full_reponame="yuya-takeyama/my-chat-app",language="javascript",library="typescript",reponame="my-chat-app",security="false",to_version="3.8.3",username="yuya-takeyama"} 1
github_dependabot_open_pull_requests{directory="server",from_version="3.9.1",full_reponame="yuya-takeyama/my-chat-app",language="ruby",library="rspec-rails",reponame="my-chat-app",security="false",to_version="4.0.0",username="yuya-takeyama"} 1
```

## Usage

### Datadog x Kubernetes

You can let Datadog agent to collect the metrics using [Autodiscovery](https://docs.datadoghq.com/agent/kubernetes/integrations/?tab=kubernetes) feature and [OpenMetrics](https://docs.datadoghq.com/integrations/openmetrics/) integration of Datadog.

You can deploy using an example manifest.

Please note that you need to replace some environment variables in ConfigMap.

```
$ kubectl apply -n YOUR_NAMESPACE -f manifest.yaml
```

### Prometheus

I'm not familiar with Prometheus :)
