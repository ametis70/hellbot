{{/*
Expand the name of the chart.
*/}}
{{- define "hellbot.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Truncate at 63 chars because some Kubernetes name fields are limited to this.
*/}}
{{- define "hellbot.fullname" -}}
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
Create chart label value.
*/}}
{{- define "hellbot.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to every resource.
*/}}
{{- define "hellbot.labels" -}}
helm.sh/chart: {{ include "hellbot.chart" . }}
{{ include "hellbot.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels — used in Deployment selector and Service selector.
*/}}
{{- define "hellbot.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hellbot.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name.
*/}}
{{- define "hellbot.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hellbot.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Container image reference. Uses .Values.image.tag when set, otherwise falls
back to the chart's appVersion.
*/}}
{{- define "hellbot.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
hellbot.isExternalSecret — returns "true" if a field value is an
existingSecret reference rather than a plain string.

Usage: {{- if eq (include "hellbot.isExternalSecret" .value) "true" }}

A value is considered an external secret when it is a dict (map) containing
an "existingSecret" key.
*/}}
{{- define "hellbot.isExternalSecret" -}}
{{- if and (kindIs "map" .) .existingSecret }}true{{- else }}false{{- end }}
{{- end }}

{{/*
hellbot.hasValue — returns "true" if a field has any value set (inline string
or existingSecret reference). Used to decide whether to emit a _file line in
the config and add an entry to the projected volume.
*/}}
{{- define "hellbot.hasValue" -}}
{{- if kindIs "map" . }}
  {{- if .existingSecret }}true{{- end }}
{{- else if . }}true{{- end }}
{{- end }}

{{/*
hellbot.secretFields — collects every sensitive field across all notifiers and
the valkey store into a flat list of objects with the shape:
  { filePath, inlineValue, existingSecret, existingSecretKey }

filePath      — the filename under /run/secrets/ (also the key in the
                chart-managed Secret)
inlineValue   — set when the user provided a plain string; chart creates the
                Secret entry
existingSecret / existingSecretKey — set when the user provided an
                existingSecret reference

This list is used by both secret.yaml (to build stringData) and
deployment.yaml (to build the projected volume sources).

Because Helm templates cannot return structured data, we emit this as a YAML
list that callers parse with `fromYaml`. The template is called with the root
context (.).
*/}}
{{- define "hellbot.secretFields" -}}
{{- $fields := list }}

{{- /* ── notifiers ── */}}
{{- range .Values.notifiers }}
  {{- $id := .id }}
  {{- $type := .type }}
  {{- $opts := .options | default dict }}

  {{- if eq $type "discord" }}
    {{- range $field, $fileSlug := dict "token" "token" "channel_id" "channel-id" "guild_id" "guild-id" }}
      {{- $val := index $opts $field }}
      {{- if include "hellbot.hasValue" $val }}
        {{- $fp := printf "discord-%s-%s" $id $fileSlug }}
        {{- if eq (include "hellbot.isExternalSecret" $val) "true" }}
          {{- $fields = append $fields (dict "filePath" $fp "existingSecret" $val.existingSecret "existingSecretKey" $val.existingSecretKey) }}
        {{- else }}
          {{- $fields = append $fields (dict "filePath" $fp "inlineValue" ($val | toString)) }}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- end }}

  {{- if eq $type "telegram" }}
    {{- range $field, $fileSlug := dict "token" "token" "chat_id" "chat-id" }}
      {{- $val := index $opts $field }}
      {{- if include "hellbot.hasValue" $val }}
        {{- $fp := printf "telegram-%s-%s" $id $fileSlug }}
        {{- if eq (include "hellbot.isExternalSecret" $val) "true" }}
          {{- $fields = append $fields (dict "filePath" $fp "existingSecret" $val.existingSecret "existingSecretKey" $val.existingSecretKey) }}
        {{- else }}
          {{- $fields = append $fields (dict "filePath" $fp "inlineValue" ($val | toString)) }}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- end }}

  {{- if eq $type "webhook" }}
    {{- range $field, $fileSlug := dict "url" "url" "secret_value" "secret-value" }}
      {{- $val := index $opts $field }}
      {{- if include "hellbot.hasValue" $val }}
        {{- $fp := printf "webhook-%s-%s" $id $fileSlug }}
        {{- if eq (include "hellbot.isExternalSecret" $val) "true" }}
          {{- $fields = append $fields (dict "filePath" $fp "existingSecret" $val.existingSecret "existingSecretKey" $val.existingSecretKey) }}
        {{- else }}
          {{- $fields = append $fields (dict "filePath" $fp "inlineValue" ($val | toString)) }}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}

{{- /* ── valkey password ── */}}
{{- $valkeyPass := .Values.store.valkey.password }}
{{- if include "hellbot.hasValue" $valkeyPass }}
  {{- if eq (include "hellbot.isExternalSecret" $valkeyPass) "true" }}
    {{- $fields = append $fields (dict "filePath" "valkey-password" "existingSecret" $valkeyPass.existingSecret "existingSecretKey" $valkeyPass.existingSecretKey) }}
  {{- else }}
    {{- $fields = append $fields (dict "filePath" "valkey-password" "inlineValue" ($valkeyPass | toString)) }}
  {{- end }}
{{- end }}

{{- toJson (dict "items" $fields) }}
{{- end }}

{{/*
hellbot.hasInlineSecrets — returns "true" when at least one field uses an
inline value (so a chart-managed Secret needs to be created).
*/}}
{{- define "hellbot.hasInlineSecrets" -}}
{{- $data := include "hellbot.secretFields" . | fromJson }}
{{- range $data.items }}
  {{- if .inlineValue }}true{{- end }}
{{- end }}
{{- end }}

{{/*
hellbot.hasAnySecrets — returns "true" when there is at least one sensitive
field with any value (inline or external). Used to decide whether to mount
the /run/secrets volume at all.
*/}}
{{- define "hellbot.hasAnySecrets" -}}
{{- $data := include "hellbot.secretFields" . | fromJson }}
{{- if $data.items }}true{{- end }}
{{- end }}
