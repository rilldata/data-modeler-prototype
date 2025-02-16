package admin

import (
	"context"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handle) GetReportMetadata(ctx context.Context, reportName, ownerID, explore, canvas string, webOpenMode adminv1.ReportOptions_OpenMode, emailRecipients []string, anonRecipients bool, executionTime time.Time) (*drivers.ReportMetadata, error) {
	var resources []*adminv1.ResourceName
	resources = append(resources, &adminv1.ResourceName{
		Type: runtime.ResourceKindReport,
		Name: reportName,
	})

	if explore != "" {
		resources = append(resources, &adminv1.ResourceName{
			Type: runtime.ResourceKindExplore,
			Name: explore,
		})
	}

	if canvas != "" {
		resources = append(resources, &adminv1.ResourceName{
			Type: runtime.ResourceKindCanvas,
			Name: canvas,
		})
	}

	res, err := h.admin.GetReportMeta(ctx, &adminv1.GetReportMetaRequest{
		ProjectId:       h.config.ProjectID,
		Branch:          h.config.Branch,
		Report:          reportName,
		OwnerId:         ownerID,
		EmailRecipients: emailRecipients,
		AnonRecipients:  anonRecipients,
		ExecutionTime:   timestamppb.New(executionTime),
		Resources:       resources,
		WebOpenMode:     webOpenMode,
	})
	if err != nil {
		return nil, err
	}

	recipientURLs := make(map[string]drivers.ReportURLs, len(res.RecipientUrls))
	for k, v := range res.RecipientUrls {
		recipientURLs[k] = drivers.ReportURLs{
			OpenURL:        v.OpenUrl,
			ExportURL:      v.ExportUrl,
			EditURL:        v.EditUrl,
			UnsubscribeURL: v.UnsubscribeUrl,
		}
	}

	return &drivers.ReportMetadata{
		RecipientURLs: recipientURLs,
	}, nil
}

func (h *Handle) GetAlertMetadata(ctx context.Context, alertName string, annotations map[string]string, queryForUserID, queryForUserEmail string) (*drivers.AlertMetadata, error) {
	req := &adminv1.GetAlertMetaRequest{
		ProjectId:   h.config.ProjectID,
		Branch:      h.config.Branch,
		Alert:       alertName,
		Annotations: annotations,
	}

	if queryForUserID != "" {
		req.QueryFor = &adminv1.GetAlertMetaRequest_QueryForUserId{QueryForUserId: queryForUserID}
	} else if queryForUserEmail != "" {
		req.QueryFor = &adminv1.GetAlertMetaRequest_QueryForUserEmail{QueryForUserEmail: queryForUserEmail}
	}

	res, err := h.admin.GetAlertMeta(ctx, req)
	if err != nil {
		return nil, err
	}

	meta := &drivers.AlertMetadata{
		OpenURL: res.OpenUrl,
		EditURL: res.EditUrl,
	}

	if res.QueryForAttributes != nil {
		meta.QueryForAttributes = res.QueryForAttributes.AsMap()
	}

	return meta, nil
}

func (h *Handle) ProvisionConnector(ctx context.Context, name, driver string, args map[string]any) (map[string]any, error) {
	argsPB, err := structpb.NewStruct(args)
	if err != nil {
		return nil, err
	}

	res, err := h.admin.Provision(ctx, &adminv1.ProvisionRequest{
		DeploymentId: "", // Will default to the deployment ID of the current access token.
		Type:         driver,
		Name:         name,
		Args:         argsPB,
	})
	if err != nil {
		return nil, err
	}

	return res.Resource.Config.AsMap(), nil
}
