package subdb

import "sync"

// AUTO GENERATED, DO NOT EDIT

{{- define "defArgs" -}}
	{{- range . -}}
		{{ .name }} {{ ((empty .spread) | ternary false .spread) | ternary "..." "" }}{{ .type }}
	{{ end -}}
{{- end }}

{{- define "inpArgs" -}}
	{{- range . -}}
		{{ .name }}{{ ((empty .spread) | ternary false .spread) | ternary "..." "" }}
	{{ end -}}
{{- end }}

{{- define "args" -}}
	{{ include .t .v | indent 0 | trim | splitList "\n" | join ", " }}
{{- end }}

{{ range . -}}
type {{ .name }}Func[IDType IDConstraint] func({{ include "args" (dict "t" "defArgs" "v" .args) }}){{ empty .returns | ternary "" (print " " .returns) }}
{{ end }}

{{ range . -}}
type BackendWith{{ .name }}Func[IDType IDConstraint] interface {
	{{ .name }}({{ include "args" (dict "t" "defArgs" "v" .args) }}){{ empty .returns | ternary "" (print " " .returns) }}
}
{{ end }}

type BackendWithEverything[IDType IDConstraint] interface {
	BackendWithInsertFunc[IDType]
	BackendWithDeleteIDFunc[IDType]
	BackendWithDeleteQueryFunc[IDType]
	BackendWithReadFunc[IDType]
	BackendWithReadIDFunc[IDType]
}

type Hooks[IDType IDConstraint] struct {
{{- range . }}
	{{ .name }}	[]{{ .name }}Func[IDType]
{{- end }}
}

{{- define "asyncFunc" }}
	func (h *Hooks[IDType]) Do{{ .name }}(cb chan bool, {{ include "args" (dict "t" "defArgs" "v" .args) }}) {
		l := &sync.WaitGroup{}
		l.Add(len(h.{{ .name }}))

		for _, h := range h.{{ .name }} {
			go func () {
				h({{ include "args" (dict "t" "inpArgs" "v" .args)}})
				l.Done()
			}()
		}

		if cb != nil {
			go func() {
				l.Wait()
				cb <- true
			}()
		}
	}
{{ end -}}

{{- define "syncFunc" }}
	func (h *Hooks[IDType]) Do{{ .name }}({{ include "args" (dict "t" "defArgs" "v" .args) }}){{ empty .returns | ternary "" (print " " .returns) }} {
		return Hooks{{ .name }}(h.{{ .name }}, {{ include "args" (dict "t" "inpArgs" "v" .args) }})
	}
{{ end -}}

{{ range . }}
	{{- $fType := "syncFunc" -}}
	{{- if (empty .returns) -}}
		{{- $fType = "asyncFunc" -}}
	{{- end -}}

	{{- include $fType . | indent 0 -}}
{{- end }}
