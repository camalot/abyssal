# Abyssal - Results

## Provider: Argo - App Of Apps

<table>
        <thead>
                <tr>
                        <th>Package</th>
                        <th>Source</th>
                        <th>Repository</th>
                        <th>Current Version</th>
                        <th>Expected Version</th>
                        <th>Status</th>
                </tr>
        </thead>
        <tbody>
                <tr>
                        <td>minecraft</td>
                        <td>sample/env/aws-eus2/values.yaml</td>
                        <td>https://itzg.github.io/minecraft-server-charts/</td>
                        <td>4.26.3</td>
                        <td>4.26.3</td>
                        <td>✅</td>
                </tr>
                <tr>
                        <td>minecraft-bedrock</td>
                        <td>sample/env/aws-eus2/values.yaml</td>
                        <td>https://itzg.github.io/minecraft-server-charts/</td>
                        <td>2.8.4</td>
                        <td>2.8.4</td>
                        <td>✅</td>
                </tr>
                <tr>
                        <td>minecraft-proxy</td>
                        <td>sample/env/aws-eus2/values.yaml</td>
                        <td>https://itzg.github.io/minecraft-server-charts/</td>
                        <td>3.8.1</td>
                        <td>3.9.0</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>rcon-web-admin</td>
                        <td>sample/env/aws-eus2/values.yaml</td>
                        <td>https://itzg.github.io/minecraft-server-charts/</td>
                        <td>1.1.0</td>
                        <td>1.1.0</td>
                        <td>✅</td>
                </tr>
                <tr>
                        <td>mc-router</td>
                        <td>sample/env/aws-eus2/values.yaml</td>
                        <td>https://itzg.github.io/minecraft-server-charts/</td>
                        <td>1.3.0</td>
                        <td>1.4.0</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>argo-cd</td>
                        <td>sample/values.yaml</td>
                        <td>https://argoproj.github.io/argo-helm</td>
                        <td>5.43.0</td>
                        <td>8.0.15</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>jenkins</td>
                        <td>sample/values.yaml</td>
                        <td>https://charts.jenkins.io</td>
                        <td>4.3.0</td>
                        <td>5.8.56</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>cert-manager</td>
                        <td>sample/values.yaml</td>
                        <td>https://charts.jetstack.io</td>
                        <td>1.14.0</td>
                        <td>1.18.0-beta.0</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>nginx-ingress</td>
                        <td>sample/values.yaml</td>
                        <td>https://kubernetes.github.io/ingress-nginx</td>
                        <td></td>
                        <td></td>
                        <td>❗</td>
                </tr>
                <tr>
                        <td colspan="6" style="color: red;">Error: failed to parse expectedVersion 'null': Malformed version: null</td>
                </tr>
                <tr>
                        <td>grafana</td>
                        <td>sample/values.yaml</td>
                        <td>https://grafana.github.io/helm-charts</td>
                        <td>6.60.0</td>
                        <td>9.2.2</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>opentelemetry-kube-stack</td>
                        <td>sample/values.yaml</td>
                        <td>https://open-telemetry.github.io/opentelemetry-helm-charts</td>
                        <td>0.6.1</td>
                        <td>0.6.1</td>
                        <td>✅</td>
                </tr>
                <tr>
                        <td>temporal</td>
                        <td>sample/values.yaml</td>
                        <td>https://go.temporal.io/helm-charts/</td>
                        <td>0.63.0</td>
                        <td>0.63.0</td>
                        <td>✅</td>
                </tr>
                <tr>
                        <td>redis-cluster</td>
                        <td>sample/values.yaml</td>
                        <td>https://charts.bitnami.com/bitnami </td>
                        <td>7.6.4</td>
                        <td>12.0.9</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>aad-pod-identity</td>
                        <td>sample/values.yaml</td>
                        <td>https://raw.githubusercontent.com/Azure/aad-pod-identity/master/charts</td>
                        <td>4.1.12</td>
                        <td>4.1.18</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>consul</td>
                        <td>sample/values.yaml</td>
                        <td>https://helm.releases.hashicorp.com</td>
                        <td>1.1.1</td>
                        <td>1.7.1</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>vault</td>
                        <td>sample/values.yaml</td>
                        <td>https://helm.releases.hashicorp.com</td>
                        <td>0.21.0</td>
                        <td>0.30.0</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>prometheus-consul-exporter</td>
                        <td>sample/values.yaml</td>
                        <td>https://prometheus-community.github.io/helm-charts</td>
                        <td>1.0.0</td>
                        <td>1.0.0</td>
                        <td>✅</td>
                </tr>
                <tr>
                        <td>kube-prometheus-stack</td>
                        <td>sample/values.yaml</td>
                        <td>https://prometheus-community.github.io/helm-charts</td>
                        <td>45.27.2</td>
                        <td>73.2.0</td>
                        <td>🔄</td>
                </tr>
                <tr>
                        <td>prom-stack-grafana-extension</td>
                        <td>sample/values.yaml</td>
                        <td>https://artifactory.imanage.com/artifactory/helm</td>
                        <td>0.1.23</td>
                        <td>0.1.23</td>
                        <td>✅</td>
                </tr></tbody>
</table>

### Status Legend

- ✅ Up to date
- 🔄 Outdated
- ⚠️ Warning
- ❗ Error
- ❌ Failure
- ⏭️ Skipped


---

---

Execution Duration: 6.1638038s
