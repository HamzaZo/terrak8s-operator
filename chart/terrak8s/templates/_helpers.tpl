{{/*
Expand the name of the chart.
*/}}
{{- define "terrak8s.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "terrak8s.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "terrak8s.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "terrak8s.labels" -}}
helm.sh/chart: {{ include "terrak8s.chart" . }}
{{ include "terrak8s.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "terrak8s.selectorLabels" -}}
app.kubernetes.io/name: {{ include "terrak8s.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Add extraVolumeMount
*/}}
{{- define "terrak8s.extraVolumeMount.tpl" -}}
{{- range  .Values.extraMountVolumes}}
- mountPath: {{ .mountPath }}
  name: {{ .name }}
{{- end }}
{{- end }}

{{/*
Add extravolume
*/}}
{{- define "terrak8s.extraVolume.tpl" -}}
{{- range .Values.extraMountVolumes }}
- name: {{ .name }}
  {{- if .secret }}
  secret:
    secretName: {{ .secret }}
  {{- else if .configMap }}
  configMap:
    name: {{ .configMap }}
  {{- end }}
  {{- if .items }}
    items:
    - key: {{ .items.key }}
      path: {{ .items.path }}
  {{- end }}
  {{- if .emptyDir }}
  emptyDir:
    medium: "{{ .emptyDir }}"
  {{- end }}
{{- end }}
{{- end }}
