package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) Export(ctx context.Context, req *runtimev1.ExportRequest) (*runtimev1.ExportResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	if req.Limit <= 0 {
		req.Limit = 10000
	}

	r, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	out := fmt.Sprintf("/v1/download?%s=%s", "request", base64.URLEncoding.EncodeToString(r))

	return &runtimev1.ExportResponse{
		DownloadUrlPath: out,
	}, nil
}

func (s *Server) downloadHandler(w http.ResponseWriter, req *http.Request) {
	marshalled, err := base64.URLEncoding.DecodeString(req.URL.Query().Get("request"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	request := &runtimev1.ExportRequest{}
	err = proto.Unmarshal(marshalled, request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %s", err), http.StatusBadRequest)
		return
	}

	if request.Limit > 10000 {
		http.Error(w, "limit must be less than or equal to 10000", http.StatusBadRequest)
		return
	}

	var q runtime.Query
	var filename string
	switch v := request.Request.(type) {
	case *runtimev1.ExportRequest_MetricsViewToplistRequest:
		v.MetricsViewToplistRequest.Limit = int64(request.Limit)
		filename, err = exportFilename(req.Context(), s.runtime, request.InstanceId, v.MetricsViewToplistRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse request: %s", err), http.StatusBadRequest)
			return
		}

		q, err = createToplistQuery(req.Context(), w, v.MetricsViewToplistRequest, request.Format)
	case *runtimev1.ExportRequest_MetricsViewRowsRequest:
		v.MetricsViewRowsRequest.Limit = request.Limit
		filename, err = exportFilename(req.Context(), s.runtime, request.InstanceId, v.MetricsViewRowsRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse request: %s", err), http.StatusBadRequest)
			return
		}

		q, err = createRowsQuery(req.Context(), w, v.MetricsViewRowsRequest, request.Format)
	default:
		http.Error(w, fmt.Sprintf("Unsupported request type: %s", reflect.TypeOf(v).Name()), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !auth.GetClaims(req.Context()).CanInstance(request.InstanceId, auth.ReadMetrics) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	switch request.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename))
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", filename))
	default:
		http.Error(w, fmt.Sprintf("Unsupported format %s", request.Format), http.StatusBadRequest)
		return
	}

	err = q.Export(req.Context(), s.runtime, request.InstanceId, 0, request.Format, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func exportFilename(ctx context.Context, rt *runtime.Runtime, instanceID string, request any) (string, error) {
	pointer := reflect.ValueOf(request)
	elem := pointer.Elem()
	field := elem.FieldByName("MetricsViewName")
	metricsViewName := field.String()

	filter := elem.FieldByName("Filter").Interface().(*runtimev1.MetricsViewFilter)
	timeStart := elem.FieldByName("TimeStart").Interface().(*timestamppb.Timestamp)
	timeEnd := elem.FieldByName("TimeStart").Interface().(*timestamppb.Timestamp)

	timeRange := false
	if timeStart != nil || timeEnd != nil {
		timeRange = true
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return "", err
	}

	filteredString := ""
	if (filter != nil && (len(filter.Exclude) > 0 || len(filter.Include) > 0)) || timeRange {
		filteredString = "_filtered"
	}

	return fmt.Sprintf("%s%s_%s", strings.ReplaceAll(mv.Model, `"`, "_"), filteredString, time.Now().Format("20060102150405")), nil
}

func createToplistQuery(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewToplistRequest, format runtimev1.ExportFormat) (runtime.Query, error) {
	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		InlineMeasures:  req.InlineMeasures,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Filter:          req.Filter,
	}

	return q, nil
}

func createRowsQuery(ctx context.Context, writer http.ResponseWriter, req *runtimev1.MetricsViewRowsRequest, format runtimev1.ExportFormat) (runtime.Query, error) {
	q := &queries.MetricsViewRows{
		MetricsViewName: req.MetricsViewName,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
		Sort:            req.Sort,
		Limit:           req.Limit,
		Offset:          req.Offset,
	}

	return q, nil
}

func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.MetricsView, error) {
	obj, err := rt.GetCatalogEntry(ctx, instanceID, name)
	if err != nil {
		return nil, err
	}

	if obj.GetMetricsView() == nil {
		return nil, err
	}

	return obj.GetMetricsView(), nil
}
