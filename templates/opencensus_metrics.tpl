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
  "go.opencensus.io/plugin/ochttp"
  "go.opencensus.io/stats"
  "go.opencensus.io/stats/view"
  "go.opencensus.io/tag"
)

{{ $decorator := (or .Vars.DecoratorName (printf "%sWithMetrics" .Interface.Name)) }}
{{ $metric_name := (or .Vars.MetricName (printf "%s_duration_ms" (down .Interface.Name))) }}
{{ $packageName := (or .Vars.PackageName (printf "%s" (down .Interface.Type))) }}

var (
  {{down .Interface.Name}}HistogramResultTag = tag.MustNewKey("result")
  {{down .Interface.Name}}HistogramMethodNameTag = tag.MustNewKey("method_name")
  {{down .Interface.Name}}HistogramInstanceNameTag = tag.MustNewKey("instance_name")

  {{down .Interface.Name}}Histogram = stats.Float64("{{$packageName}}_{{$metric_name}}", "Time spent in ms on running an operation of type <{{.Interface.Name}}> in package <{{$packageName}}>", stats.UnitMilliseconds)
)

// {{$decorator}} implements {{.Interface.Type}} interface with all methods wrapped
// with metrics
type {{$decorator}} struct { //nolint:golint
  base {{.Interface.Type}}
  instanceName string
}

// NewWithMetrics returns an instance of the {{.Interface.Type}} decorated with metrics
func NewWithMetrics(base {{.Interface.Type}}, instanceName string) {{.Interface.Name}} {
  if err := view.Register(
      &view.View{
          Name:        {{down .Interface.Name}}Histogram.Name(),
          Description: {{down .Interface.Name}}Histogram.Description(),
          Measure:     {{down .Interface.Name}}Histogram,
          TagKeys:     []tag.Key{
              {{down .Interface.Name}}HistogramResultTag,
              {{down .Interface.Name}}HistogramMethodNameTag,
              {{down .Interface.Name}}HistogramInstanceNameTag,
          },
          Aggregation: ochttp.DefaultLatencyDistribution,
      },
  ); err != nil {
      log.Fatalf("could not register view (%v): %v", {{down .Interface.Name}}Histogram.Name(), err)
  }

  return {{$decorator}} {
    base: base,
    instanceName: instanceName,
  }
}

{{range $method := .Interface.Methods}}
  // {{$method.Name}} implements {{$.Interface.Type}}
  func (_d {{$decorator}}) {{$method.Declaration}} {
  	  _since := time.Now()
      defer func() {
        result := "ok"
        {{- if $method.ReturnsError}}
          if err != nil {
            result = "error"
          }
        {{end}}

        _ctx, err := tag.New(context.Background(),
          tag.Insert({{down $.Interface.Name}}HistogramInstanceNameTag, _d.instanceName),
          tag.Insert({{down $.Interface.Name}}HistogramMethodNameTag, "{{$method.Name}}"),
          tag.Insert({{down $.Interface.Name}}HistogramResultTag, result),
        )
        if err != nil {
          log.Printf("could not create tag with context for instance (%v) method (%v): %v",
            _d.instanceName,
            "{{$method.Name}}",
            err,
          )
          return
        }
        stats.Record(
          _ctx,
          {{down $.Interface.Name}}Histogram.M(float64(time.Since(_since)/time.Millisecond)),
        )
      }()

      {{$method.Pass "_d.base."}}
  }
{{end}}
