package shitdb

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

{{ range . }}
{{- if (empty .returns) -}}
{{- include "asyncFunc" . | indent 0 -}}
{{- end -}}
{{- end }}

func (h *Hooks[IDType]) DoReadID(ids ...IDType) []Group[IDType] {
	o := []Group[IDType]{}

	for _, f := range h.ReadID {
		buff := f(ids...)
		o = append(o, buff...)

		if len(o) == len(ids) {
			break
		}
	}

	return o
}

func (h *Hooks[IDType]) DoRead(idPointer *IDPointer[IDType], oldToNew bool, f Filter[IDType]) ([]Group[IDType], bool) {
	o := []Group[IDType]{}
	cutFirst := 0

	for _, h := range h.Read {
		buf, exitEarly := h(idPointer, oldToNew, f)
		o = append(o, buf[cutFirst:]...)
		if exitEarly {
			return o, true
		}
		if len(buf) != 0 {
			idPointer = &IDPointer[IDType]{
				ID:   buf[len(buf)-1].GetID(),
				ApproximationBehavior: APPROXIMATE_NEWEST,
			}
			cutFirst = 1
		}
	}

	return o, false
}
