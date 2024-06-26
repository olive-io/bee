/*
   Copyright 2024 The bee Authors

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library;
*/

package bee

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	json "github.com/json-iterator/go"
	"github.com/olive-io/bpmn/flow"
	"github.com/olive-io/bpmn/flow_node/activity"
	"github.com/olive-io/bpmn/flow_node/activity/script"
	"github.com/olive-io/bpmn/flow_node/activity/service"
	bprocess "github.com/olive-io/bpmn/process"
	"github.com/olive-io/bpmn/process/instance"
	"github.com/olive-io/bpmn/schema"
	"github.com/olive-io/bpmn/tracing"
	"go.uber.org/zap"

	"github.com/olive-io/bee/plugins/callback"
	"github.com/olive-io/bee/plugins/filter"
	"github.com/olive-io/bee/process"
	"github.com/olive-io/bee/stats"
)

func (rt *Runtime) Play(ctx context.Context, pr *process.Process, opts ...RunOption) error {
	definitions, dataObjects, properties, err := pr.Build()
	if err != nil {
		return err
	}

	return rt.RunBpmnProcess(ctx, definitions, dataObjects, properties, opts...)
}

func (rt *Runtime) RunBpmnProcess(ctx context.Context, definitions *schema.Definitions, dataObjects, properties map[string]string, opts ...RunOption) error {

	lg := rt.Logger()

	runOptions := newRunOptions()
	for _, opt := range opts {
		opt(runOptions)
	}

	cb := callback.NewCallBack()
	if runOptions.Callback != nil {
		cb = runOptions.Callback
	}

	ft := filter.NewFilter()
	if runOptions.Filter != nil {
		ft = runOptions.Filter
	}

	var sources []string
	if v, ok := properties["hosts"]; ok {
		sources = strings.Split(v, ",")
	}

	if len(sources) == 0 {
		return fmt.Errorf("missing sources")
	}

	if err := rt.inventory.AddSources(sources...); err != nil {
		return err
	}

	processElement := (*definitions.Processes())[0]
	proc := bprocess.New(&processElement, definitions)

	_properties := map[string]any{}
	for key, value := range _properties {
		_properties[key] = value
	}
	_dataObjects := map[string]any{}
	for key, value := range dataObjects {
		_dataObjects[key] = value
	}
	bpmnOptions := []instance.Option{
		instance.WithVariables(_properties),
		instance.WithDataObjects(_dataObjects),
	}

	ins, err := proc.Instantiate(bpmnOptions...)
	if err != nil {
		return err
	}
	if tracer := runOptions.Tracer; tracer != nil {
		ins.Tracer.SubscribeChannel(tracer)
		defer ins.Tracer.Unsubscribe(tracer)
	}

	traces := ins.Tracer.Subscribe()
	err = ins.StartAll(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to run the instance")
	}
	defer ins.Tracer.Unsubscribe(traces)

	runTasks := make([]process.ITask, 0)

LOOP:
	for {
		var trace tracing.ITrace
		select {
		case <-ctx.Done():
			return context.Canceled
		case trace = <-traces:
		}

		if trace == nil {
			continue
		}

		trace = tracing.Unwrap(trace)
		switch tt := trace.(type) {
		case flow.Trace:
		case *activity.Trace:
			act := tt.GetActivity()
			id, _ := act.Element().Id()

			tProps, tHeaders := ft.OnPreTaskProps(*id, tt.GetProperties(), tt.GetHeaders())
			rspProperties := map[string]any{}

			var aErr error
			switch act.(type) {
			case *service.Node:

				sv := process.DecodeServiceTask(tProps, tHeaders)
				runTasks = append(runTasks, sv)
				hosts := sv.Hosts
				if len(hosts) == 0 {
					hosts = sources
				}

				if caller := rt.opts.caller; caller != nil {
					for _, host := range hosts {
						result := &stats.TaskResult{
							Host: host,
						}

						ropts := append(opts, WithMetadata(tHeaders))
						in, _ := json.Marshal(sv.Args)
						data, err := caller(ctx, host, sv.Action, in, ropts...)
						if err != nil {
							aErr = multierror.Append(aErr, err)
							result.ErrMsg = err.Error()
							cb.RunnerOkFailed(result)
							continue
						}

						stdout := map[string]any{}
						if err = json.Unmarshal(data, &stdout); err != nil {
							aErr = multierror.Append(aErr, err)
							result.ErrMsg = err.Error()
							cb.RunnerOkFailed(result)
							continue
						}

						stdout = ft.OnPostTaskStdout(*id, stdout)
						result.Stdout = stdout
						for key, value := range stdout {
							rspProperties[key] = value
						}

						cb.RunnerOnOk(result)
					}
				}

			case *script.Node:

				task := process.DecodeScriptTask(tProps, tHeaders)
				runTasks = append(runTasks, task)
				hosts := task.Hosts
				if len(hosts) == 0 {
					hosts = sources
				}

				args := make([]string, 0)
				args = append(args, task.Action)
				for name, arg := range task.Args {
					value, _ := json.Marshal(arg)
					args = append(args, name+"="+strings.ReplaceAll(string(value), "\"", ""))
				}
				shell := strings.Join(args, " ")
				for _, host := range hosts {
					result := &stats.TaskResult{
						Host: host,
					}

					ropts := append(opts, WithMetadata(tHeaders))
					data, err := rt.Execute(ctx, host, shell, ropts...)
					if err != nil {
						aErr = multierror.Append(aErr, err)
						result.ErrMsg = err.Error()
						cb.RunnerOkFailed(result)
						continue
					}

					stdout := map[string]any{}
					if err = json.Unmarshal(data, &stdout); err != nil {
						aErr = multierror.Append(aErr, err)
						result.ErrMsg = err.Error()
						cb.RunnerOkFailed(result)
						continue
					}

					stdout = ft.OnPostTaskStdout(*id, stdout)
					result.Stdout = stdout
					for key, value := range stdout {
						rspProperties[key] = value
					}

					cb.RunnerOnOk(result)
				}
			}

			actOpts := make([]activity.DoOption, 0)
			if aErr != nil {
				actOpts = append(actOpts, activity.WithErr(aErr))
			}
			actOpts = append(actOpts, activity.WithProperties(rspProperties))

			tt.Do(actOpts...)
		case tracing.ErrorTrace:
			err = tt.Error
			break LOOP
		case flow.CeaseFlowTrace:
			break LOOP
		default:
			lg.Sugar().Debugf("%#v", tt)
		}
	}
	ins.WaitUntilComplete(ctx)

	if err != nil {
		for i := len(runTasks) - 1; i >= 0; i-- {
			task := runTasks[i]

			caught, ok := task.(process.ICatchTask)
			if !ok {
				continue
			}

			fields := make([]zap.Field, 0)
			if named, ok := task.(process.INamedTask); ok {
				fields = append(fields,
					zap.String("name", named.GetName()),
					zap.String("id", named.GetId()))
			}

			catch := caught.GetCatch()
			if catch == nil {
				lg.Debug("skip task catch", fields...)
				continue
			}

			hosts := caught.GetHosts()
			if len(hosts) == 0 {
				hosts = sources
			}

			fields = append(fields, zap.Stringer("handler", catch))
			lg.Info("handle task catch", fields...)

			_ = rt.handle(ctx, hosts, catch)
		}
	}

	for i := len(runTasks) - 1; i >= 0; i-- {
		task := runTasks[i]

		caught, ok := task.(process.ICatchTask)
		if !ok {
			continue
		}

		fields := make([]zap.Field, 0)
		if named, ok := task.(process.INamedTask); ok {
			fields = append(fields,
				zap.String("name", named.GetName()),
				zap.String("id", named.GetId()))
		}

		finish := caught.GetFinish()
		if finish == nil {
			lg.Debug("skip service catch", fields...)
			continue
		}

		hosts := caught.GetHosts()
		if len(hosts) == 0 {
			hosts = sources
		}

		fields = append(fields, zap.Stringer("handler", finish))
		lg.Info("handle service finish", fields...)

		_ = rt.handle(ctx, hosts, finish)
	}

	return err
}

func (rt *Runtime) handle(ctx context.Context, hosts []string, handler *process.Handler, opts ...RunOption) error {
	switch handler.Kind {
	case process.ServiceKey:
		caller := rt.opts.caller
		if caller == nil {
			return nil
		}

		for _, host := range hosts {
			ropts := append(opts)
			in, _ := json.Marshal(handler.Args)
			data, err := caller(ctx, host, handler.Action, in, ropts...)
			if err != nil {
				return err
			}

			stdout := map[string]any{}
			if err = json.Unmarshal(data, &stdout); err != nil {
			}
		}

	case process.TaskKey:
		args := make([]string, 0)
		args = append(args, handler.Action)
		for name, arg := range handler.Args {
			value, _ := json.Marshal(arg)
			args = append(args, name+"="+strings.ReplaceAll(string(value), "\"", ""))
		}
		shell := strings.Join(args, " ")
		for _, host := range hosts {
			ropts := append(opts)
			data, err := rt.Execute(ctx, host, shell, ropts...)
			if err != nil {
				return err
			}

			stdout := map[string]any{}
			if err = json.Unmarshal(data, &stdout); err != nil {
			}
		}
	}

	return nil
}
