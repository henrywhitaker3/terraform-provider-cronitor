// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/henrywhitaker3/terraform-provider-cronitor/pkg/cronitor"
)

type BaseMonitorModel struct {
	Key               types.String `tfsdk:"key"`
	Name              types.String `tfsdk:"name"`
	Disabled          types.Bool   `tfsdk:"disabled"`
	Paused            types.Bool   `tfsdk:"paused"`
	Schedule          types.String `tfsdk:"schedule"`
	Notify            types.List   `tfsdk:"notify"`
	ScheduleTolerance types.Int32  `tfsdk:"schedule_tolerance"`
	FailureTolerance  types.Int32  `tfsdk:"failure_tolerance"`
	GraceSeconds      types.Int32  `tfsdk:"grace_seconds"`
	RealertInterval   types.String `tfsdk:"realert_interval"`
	Timezone          types.String `tfsdk:"timezone"`
	Tags              types.List   `tfsdk:"tags"`
	Environments      types.List   `tfsdk:"environments"`
	Group             types.String `tfsdk:"group"`
}

type HttpMonitorModel struct {
	BaseMonitorModel

	Url             types.String `tfsdk:"url"`
	Headers         types.Map    `tfsdk:"headers"`
	Cookies         types.Map    `tfsdk:"cookies"`
	Body            types.String `tfsdk:"body"`
	Method          types.String `tfsdk:"method"`
	TimeoutSeconds  types.Int32  `tfsdk:"timeout_seconds"`
	Regions         types.List   `tfsdk:"regions"`
	FollowRedirects types.Bool   `tfsdk:"follow_redirects"`
	VerifySsl       types.Bool   `tfsdk:"verify_ssl"`
	Assertions      types.List   `tfsdk:"assertions"`
}

type HeartbeatMonitorModel struct {
	BaseMonitorModel

	TelemetryUrl types.String `tfsdk:"telemetry_url"`
}

type NotificationListModel struct {
	Name      types.String `tfsdk:"name"`
	Key       types.String `tfsdk:"key"`
	Emails    types.List   `tfsdk:"emails"`
	Slack     types.List   `tfsdk:"slack"`
	Pagerduty types.List   `tfsdk:"pagerduty"`
	Phones    types.List   `tfsdk:"phones"`
	Webhooks  types.List   `tfsdk:"webhooks"`
}

func processSlice[T, U any](in []T, t attr.Type, c func(T) U) types.List {
	if len(in) == 0 {
		return types.ListNull(t)
	}

	elems := []U{}
	for _, e := range in {
		elems = append(elems, c(e))
	}
	list, _ := types.ListValueFrom(context.Background(), t, elems)
	return list
}

func stringSlice(in []string) types.List {
	return processSlice(in, types.StringType, types.StringValue)
}

func toStringSlice(in types.List) []string {
	temp := []types.String{}
	in.ElementsAs(context.Background(), &temp, false)
	out := []string{}
	for _, e := range temp {
		out = append(out, e.ValueString())
	}
	return out
}

func toStringMap(in types.Map) map[string]string {
	temp := map[string]types.String{}
	in.ElementsAs(context.Background(), &temp, false)
	out := map[string]string{}
	for key, val := range temp {
		out[key] = val.ValueString()
	}
	return out
}

func toHttpMonitor(m *cronitor.Monitor) HttpMonitorModel {
	out := HttpMonitorModel{
		BaseMonitorModel: BaseMonitorModel{
			Key:             types.StringValue(*m.Key),
			Name:            types.StringValue(m.Name),
			Disabled:        types.BoolValue(m.Disabled),
			Paused:          types.BoolValue(m.Paused),
			Schedule:        types.StringValue(m.Schedule),
			Notify:          stringSlice(m.Notify),
			Tags:            stringSlice(m.Tags),
			RealertInterval: types.StringValue(m.RealertInterval),
			Environments:    stringSlice(m.Environments),
		},
		Assertions:      stringSlice(m.Assertions),
		Url:             types.StringValue(m.Request.URL),
		Method:          types.StringValue(m.Request.Method),
		Headers:         types.MapNull(types.StringType),
		Cookies:         types.MapNull(types.StringType),
		Body:            types.StringNull(),
		TimeoutSeconds:  types.Int32Value(int32(m.Request.TimeoutSeconds)),
		Regions:         stringSlice(m.Request.Regions),
		FollowRedirects: types.BoolValue(m.Request.FollowRedirects),
		VerifySsl:       types.BoolValue(m.Request.VerifySsl),
	}

	if m.Timezone != nil {
		out.Timezone = types.StringValue(*m.Timezone)
	}
	if m.ScheduleTolerance != nil {
		out.ScheduleTolerance = types.Int32Value(int32(*m.ScheduleTolerance))
	}
	if m.FailureTolerance != nil {
		out.FailureTolerance = types.Int32Value(int32(*m.FailureTolerance))
	}
	if m.GraceSeconds != nil {
		out.GraceSeconds = types.Int32Value(int32(*m.GraceSeconds))
	}
	if m.Group != nil {
		out.Group = types.StringValue(*m.Group)
	}

	if len(m.Request.Headers) > 0 {
		elems := map[string]attr.Value{}
		for key, val := range m.Request.Headers {
			elems[key] = types.StringValue(val)
		}
		out.Headers = types.MapValueMust(types.StringType, elems)
	}
	if len(m.Request.Cookies) > 0 {
		elems := map[string]attr.Value{}
		for key, val := range m.Request.Cookies {
			elems[key] = types.StringValue(val)
		}
		out.Cookies = types.MapValueMust(types.StringType, elems)
	}

	return out
}

func httpToMonitorRequest(data HttpMonitorModel) *cronitor.Monitor {
	out := &cronitor.Monitor{
		Name:         data.Name.ValueString(),
		Assertions:   toStringSlice(data.Assertions),
		Disabled:     data.Disabled.ValueBool(),
		Paused:       data.Disabled.ValueBool(),
		Notify:       toStringSlice(data.Notify),
		Tags:         toStringSlice(data.Tags),
		Environments: toStringSlice(data.Environments),
		Type:         "check",
		Platform:     "http",
		Request: &cronitor.Request{
			URL:             data.Url.ValueString(),
			Method:          data.Method.ValueString(),
			Headers:         toStringMap(data.Headers),
			Cookies:         toStringMap(data.Cookies),
			Body:            data.Body.ValueString(),
			Regions:         toStringSlice(data.Regions),
			TimeoutSeconds:  int(data.TimeoutSeconds.ValueInt32()),
			FollowRedirects: data.FollowRedirects.ValueBool(),
			VerifySsl:       data.VerifySsl.ValueBool(),
		},
	}
	if out.RealertInterval == "" {
		out.RealertInterval = "every 8 hours"
	}
	if data.Schedule.ValueString() != "" {
		out.Schedule = data.Schedule.ValueString()
	}

	g := int(data.GraceSeconds.ValueInt32())
	out.GraceSeconds = &g
	st := int(data.ScheduleTolerance.ValueInt32())
	out.ScheduleTolerance = &st
	ft := int(data.FailureTolerance.ValueInt32())
	out.FailureTolerance = &ft
	if data.Timezone.ValueString() != "" {
		tz := data.Timezone.ValueString()
		out.Timezone = &tz
	}
	if data.Group.ValueString() != "" {
		grp := data.Group.ValueString()
		out.Group = &grp
	}

	return out
}

func toHeartbeatMonitor(m *cronitor.Monitor) HeartbeatMonitorModel {
	out := HeartbeatMonitorModel{
		BaseMonitorModel: BaseMonitorModel{
			Key:             types.StringValue(*m.Key),
			Name:            types.StringValue(m.Name),
			Disabled:        types.BoolValue(m.Disabled),
			Paused:          types.BoolValue(m.Paused),
			Schedule:        types.StringValue(m.Schedule),
			Notify:          stringSlice(m.Notify),
			Tags:            stringSlice(m.Tags),
			RealertInterval: types.StringValue(m.RealertInterval),
			Environments:    stringSlice(m.Environments),
		},
	}

	if m.Timezone != nil {
		out.Timezone = types.StringValue(*m.Timezone)
	}
	if m.ScheduleTolerance != nil {
		out.ScheduleTolerance = types.Int32Value(int32(*m.ScheduleTolerance))
	}
	if m.FailureTolerance != nil {
		out.FailureTolerance = types.Int32Value(int32(*m.FailureTolerance))
	}
	if m.GraceSeconds != nil {
		out.GraceSeconds = types.Int32Value(int32(*m.GraceSeconds))
	}
	if m.Group != nil {
		out.Group = types.StringValue(*m.Group)
	}

	return out
}

func heartbeatToMonitorRequest(data HeartbeatMonitorModel) *cronitor.Monitor {
	out := &cronitor.Monitor{
		Name:         data.Name.ValueString(),
		Disabled:     data.Disabled.ValueBool(),
		Paused:       data.Disabled.ValueBool(),
		Notify:       toStringSlice(data.Notify),
		Tags:         toStringSlice(data.Tags),
		Environments: toStringSlice(data.Environments),
		Type:         "heartbeat",
		Platform:     "linux",
	}
	if out.RealertInterval == "" {
		out.RealertInterval = "every 8 hours"
	}

	if data.Schedule.ValueString() != "" {
		out.Schedule = data.Schedule.ValueString()
	}

	g := int(data.GraceSeconds.ValueInt32())
	out.GraceSeconds = &g
	st := int(data.ScheduleTolerance.ValueInt32())
	out.ScheduleTolerance = &st
	ft := int(data.FailureTolerance.ValueInt32())
	out.FailureTolerance = &ft
	if data.Timezone.ValueString() != "" {
		tz := data.Timezone.ValueString()
		out.Timezone = &tz
	}
	if data.Group.ValueString() != "" {
		grp := data.Group.ValueString()
		out.Group = &grp
	}

	return out
}

func toNotificationList(l *cronitor.NotificationList) NotificationListModel {
	return NotificationListModel{
		Name:      types.StringValue(l.Name),
		Key:       types.StringValue(l.Key),
		Emails:    stringSlice(l.Notifications.Emails),
		Slack:     stringSlice(l.Notifications.Slack),
		Pagerduty: stringSlice(l.Notifications.Pagerduty),
		Phones:    stringSlice(l.Notifications.Phones),
		Webhooks:  stringSlice(l.Notifications.Webhooks),
	}
}

func listToListRequest(data NotificationListModel) *cronitor.NotificationList {
	return &cronitor.NotificationList{
		Name: data.Name.ValueString(),
		Key:  data.Key.ValueString(),
		Notifications: cronitor.Notifications{
			Emails:    toStringSlice(data.Emails),
			Slack:     toStringSlice(data.Slack),
			Pagerduty: toStringSlice(data.Pagerduty),
			Phones:    toStringSlice(data.Phones),
			Webhooks:  toStringSlice(data.Webhooks),
		},
	}
}

func fixSliceOrder[T comparable](correct []T, incorrect *[]T) {
	if incorrect == nil {
		*incorrect = []T{}
	}

	if len(correct) != len(*incorrect) {
		return
	}

	if correct == nil {
		*incorrect = nil
		return
	}

	for _, i := range *incorrect {
		if !slices.Contains(correct, i) {
			return
		}
	}

	new := []T{}
	new = append(new, correct...)
	*incorrect = new
}
