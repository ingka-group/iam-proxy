// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
  "go.opentelemetry.io/otel"
  "go.opentelemetry.io/otel/codes"
)

{{ $decorator := (or .Vars.DecoratorName (printf "%sWithTracing" .Interface.Name)) }}

// {{$decorator}} implements {{.Interface.Type}} interface with all methods wrapped
// with Trace
type {{$decorator}} struct { //nolint:golint
  base {{.Interface.Type}}
  instanceName string
}

// NewWithTracing returns an instance of the {{.Interface.Type}} decorated with tracing
func NewWithTracing(base {{.Interface.Type}}, instanceName string) {{.Interface.Name}} {
  return {{$decorator}} {
    base: base,
    instanceName: instanceName,
  }
}

{{range $method := .Interface.Methods}}
  // {{$method.Name}} implements {{$.Interface.Type}}
  func (_d {{$decorator}}) {{$method.Declaration}} {
    {{- if and (and (ne $method.Name "Health") (ne $method.Name "Ready")) (ne $method.Name "Ping") }}
    {{- if $method.AcceptsContext}}
      ctx, span := otel.Tracer(_d.instanceName).Start(ctx, "{{$method.Name}}")
  	  {{else}}
  	  _, span := otel.Tracer(_d.instanceName).Start(context.Background(), "{{$method.Name}}")
  	  {{end}}

      {{- if $method.ReturnsError}}
        defer func() {
          if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
          }
          span.End()
        }()
      {{else}}
        defer span.End()
      {{end}}
      {{end}}

      {{$method.Pass "_d.base."}}
  }
{{end}}
