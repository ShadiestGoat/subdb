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
{{ range . -}}
	BackendWith{{ .name }}Func[IDType]
{{ end }}
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
		{{ $async := default (dict) .async }}
		{{ default "" $async.extraLogic }}
		{{ $args := .args }}

		{{ range $tpmI, $v := (default (dict) $async.argOverride) }}
			{{ $i := int $tpmI }}
			{{ $newV := deepCopy (index $args $i) }}
			{{ $_ := set $newV "name" $v }}
			{{ $args = concat (slice $args 0 $i) (list $newV) (slice $args (add $i 1)) }}
		{{ end }}

		for {{ default "_" $async.iVar }}, h := range h.{{ .name }} {
			go func (h {{ .name }}Func[IDType] {{- if $async.iVar}}, {{ $async.iVar}} int {{- end }}) {
				h({{ include "args" (dict "t" "inpArgs" "v" $args)}})
				l.Done()
			}(h {{- if $async.iVar }}, {{ $async.iVar }} {{- end }})
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
