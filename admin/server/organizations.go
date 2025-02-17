package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/publicemail"
	"github.com/rilldata/rill/admin/server/auth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListOrganizations(ctx context.Context, req *adminv1.ListOrganizationsRequest) (*adminv1.ListOrganizationsResponse, error) {
	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	pageSize := validPageSize(req.PageSize)

	orgs, err := s.admin.DB.FindOrganizationsForUser(ctx, claims.OwnerID(), token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(orgs) >= pageSize {
		nextToken = marshalPageToken(orgs[len(orgs)-1].Name)
	}

	pbs := make([]*adminv1.Organization, len(orgs))
	for i, org := range orgs {
		pbs[i] = s.organizationToDTO(org, false)
	}

	return &adminv1.ListOrganizationsResponse{Organizations: pbs, NextPageToken: nextToken}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *adminv1.GetOrganizationRequest) (*adminv1.GetOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Name))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	claims := auth.GetClaims(ctx)
	perms := claims.OrganizationPermissions(ctx, org.ID)
	if !perms.ReadOrg && !claims.Superuser(ctx) {
		ok, err := s.admin.DB.CheckOrganizationHasPublicProjects(ctx, org.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "not allowed to read org")
		}

		perms.ReadOrg = true
		perms.ReadProjects = true
	}

	// TODO: This is used to update plan name cache and can be removed a few months after Feb 2025 when plans have been cached for most orgs.
	// after that we can return empty plan name for uncached orgs, discussion - https://github.com/rilldata/rill/pull/6338#discussion_r1952713404
	if org.BillingPlanName == nil {
		_, org, err = s.getSubscriptionAndUpdateOrg(ctx, org)
		if err != nil {
			return nil, err
		}
	}

	return &adminv1.GetOrganizationResponse{
		Organization: s.organizationToDTO(org, perms.ManageOrg),
		Permissions:  perms,
	}, nil
}

func (s *Server) GetOrganizationNameForDomain(ctx context.Context, req *adminv1.GetOrganizationNameForDomainRequest) (*adminv1.GetOrganizationNameForDomainResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.domain", req.Domain))

	org, err := s.admin.DB.FindOrganizationByCustomDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}

	// NOTE: Not checking auth on purpose. This needs to be a public endpoint.

	return &adminv1.GetOrganizationNameForDomainResponse{
		Name: org.Name,
	}, nil
}

func (s *Server) CreateOrganization(ctx context.Context, req *adminv1.CreateOrganizationRequest) (*adminv1.CreateOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Name),
		attribute.String("args.description", req.Description),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	if !claims.Superuser(ctx) {
		// check single user org limit for this user
		count, err := s.admin.DB.CountSingleuserOrganizationsForMemberUser(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		if user.QuotaSingleuserOrgs >= 0 && count >= user.QuotaSingleuserOrgs {
			return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: you can only create %d single-user orgs", user.QuotaSingleuserOrgs)
		}
	}

	org, err := s.admin.CreateOrganizationForUser(ctx, user.ID, user.Email, req.Name, req.Description)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &adminv1.CreateOrganizationResponse{
		Organization: s.organizationToDTO(org, true),
	}, nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *adminv1.DeleteOrganizationRequest) (*adminv1.DeleteOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Name))

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to delete org")
	}

	_, err = s.admin.Jobs.DeleteOrg(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.DeleteOrganizationResponse{}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *adminv1.UpdateOrganizationRequest) (*adminv1.UpdateOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Name))
	if req.Description != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.description", *req.Description))
	}
	if req.NewName != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.new_name", *req.NewName))
	}
	if req.BillingEmail != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.billing_email", *req.BillingEmail))
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
		return nil, status.Error(codes.PermissionDenied, "not allowed to update org")
	}

	logoAssetID := org.LogoAssetID
	if req.LogoAssetId != nil { // Means it should be updated
		if *req.LogoAssetId == "" { // Means it should be cleared
			logoAssetID = nil
		} else {
			logoAssetID = req.LogoAssetId
		}
	}

	faviconAssetID := org.FaviconAssetID
	if req.FaviconAssetId != nil { // Means it should be updated
		if *req.FaviconAssetId == "" { // Means it should be cleared
			faviconAssetID = nil
		} else {
			faviconAssetID = req.FaviconAssetId
		}
	}

	nameChanged := req.NewName != nil && *req.NewName != org.Name
	emailChanged := req.BillingEmail != nil && *req.BillingEmail != org.BillingEmail
	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                valOrDefault(req.NewName, org.Name),
		DisplayName:                         valOrDefault(req.DisplayName, org.DisplayName),
		Description:                         valOrDefault(req.Description, org.Description),
		LogoAssetID:                         logoAssetID,
		FaviconAssetID:                      faviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        valOrDefault(req.BillingEmail, org.BillingEmail),
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}

	if nameChanged {
		err := s.admin.UpdateOrgDeploymentAnnotations(ctx, org)
		if err != nil {
			return nil, err
		}
	}

	if emailChanged {
		if org.BillingCustomerID != "" {
			err = s.admin.Biller.UpdateCustomerEmail(ctx, org.BillingCustomerID, org.BillingEmail)
			if err != nil {
				return nil, fmt.Errorf("failed to update billing email in biller: %w", err)
			}
		}
		if org.PaymentCustomerID != "" {
			err = s.admin.PaymentProvider.UpdateCustomerEmail(ctx, org.PaymentCustomerID, org.BillingEmail)
			if err != nil {
				return nil, fmt.Errorf("failed to update billing email in payment provider: %w", err)
			}
		}
	}

	return &adminv1.UpdateOrganizationResponse{
		Organization: s.organizationToDTO(org, true),
	}, nil
}

func (s *Server) ListOrganizationMemberUsers(ctx context.Context, req *adminv1.ListOrganizationMemberUsersRequest) (*adminv1.ListOrganizationMemberUsersResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) && !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	members, err := s.admin.DB.FindOrganizationMemberUsers(ctx, org.ID, token.Val, pageSize)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	nextToken := ""
	if len(members) >= pageSize {
		nextToken = marshalPageToken(members[len(members)-1].Email)
	}

	dtos := make([]*adminv1.MemberUser, len(members))
	for i, user := range members {
		dtos[i] = memberUserToPB(user)
	}

	return &adminv1.ListOrganizationMemberUsersResponse{
		Members:       dtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) ListOrganizationInvites(ctx context.Context, req *adminv1.ListOrganizationInvitesRequest) (*adminv1.ListOrganizationInvitesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ReadOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not authorized to read org members")
	}

	token, err := unmarshalPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	pageSize := validPageSize(req.PageSize)

	// get pending user invites for this org
	userInvites, err := s.admin.DB.FindOrganizationInvites(ctx, org.ID, token.Val, pageSize)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	if len(userInvites) >= pageSize {
		nextToken = marshalPageToken(userInvites[len(userInvites)-1].Email)
	}

	invitesDtos := make([]*adminv1.UserInvite, len(userInvites))
	for i, invite := range userInvites {
		invitesDtos[i] = inviteToPB(invite)
	}

	return &adminv1.ListOrganizationInvitesResponse{
		Invites:       invitesDtos,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) AddOrganizationMemberUser(ctx context.Context, req *adminv1.AddOrganizationMemberUserRequest) (*adminv1.AddOrganizationMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	forceAccess := claims.Superuser(ctx) && req.SuperuserForceAccess
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers && !forceAccess {
		return nil, status.Error(codes.PermissionDenied, "not allowed to add org members")
	}

	count, err := s.admin.DB.CountInvitesForOrganization(ctx, org.ID)
	if err != nil {
		return nil, err
	}
	if org.QuotaOutstandingInvites >= 0 && count >= org.QuotaOutstandingInvites {
		return nil, status.Errorf(codes.FailedPrecondition, "quota exceeded: org can at most have %d outstanding invitations", org.QuotaOutstandingInvites)
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var invitedByUserID, invitedByName string
	if claims.OwnerType() == auth.OwnerTypeUser {
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		invitedByUserID = user.ID
		invitedByName = user.DisplayName
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}

		// Invite user to join org
		err := s.admin.DB.InsertOrganizationInvite(ctx, &database.InsertOrganizationInviteOptions{
			Email:     req.Email,
			InviterID: invitedByUserID,
			OrgID:     org.ID,
			RoleID:    role.ID,
		})
		if err != nil {
			if !errors.Is(err, database.ErrNotUnique) {
				return nil, err
			}
			// Already invited. Update the invitation role.
			invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
			if err != nil {
				return nil, err
			}
			// Update the role of the invite
			err = s.admin.DB.UpdateOrganizationInviteRole(ctx, invite.ID, role.ID)
			if err != nil {
				return nil, err
			}
			// Fallthrough so we send the email again.
		}

		// Send invitation email
		err = s.admin.Email.SendOrganizationInvite(&email.OrganizationInvite{
			ToEmail:       req.Email,
			ToName:        "",
			AcceptURL:     s.admin.URLs.WithCustomDomain(org.CustomDomain).OrganizationInviteAccept(org.Name),
			OrgName:       org.Name,
			RoleName:      role.Name,
			InvitedByName: invitedByName,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &adminv1.AddOrganizationMemberUserResponse{
			PendingSignup: true,
		}, nil
	}

	// Insert the user in the org and AllUsergroup transactionally.
	err = func() error {
		ctx, tx, err := s.admin.DB.NewTx(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		err = s.admin.DB.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID)
		if err != nil {
			return err
		}

		err = s.admin.DB.InsertUsergroupMemberUser(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			return fmt.Errorf("failed to add user to all user group: %w", err)
		}

		return tx.Commit()
	}()
	if err != nil {
		if !errors.Is(err, database.ErrNotUnique) {
			return nil, err
		}

		// The user is already in the org. Instead of erroring, we update their role and fallthrough to send the email again.
		err = s.admin.DB.UpdateOrganizationMemberUserRole(ctx, org.ID, user.ID, role.ID)
		if err != nil {
			return nil, err
		}
	}

	err = s.admin.Email.SendOrganizationAddition(&email.OrganizationAddition{
		ToEmail:       req.Email,
		ToName:        "",
		OpenURL:       s.admin.URLs.WithCustomDomain(org.CustomDomain).Organization(org.Name),
		OrgName:       org.Name,
		RoleName:      role.Name,
		InvitedByName: invitedByName,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.AddOrganizationMemberUserResponse{
		PendingSignup: false,
	}, nil
}

func (s *Server) RemoveOrganizationMemberUser(ctx context.Context, req *adminv1.RemoveOrganizationMemberUserRequest) (*adminv1.RemoveOrganizationMemberUserResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.Bool("args.keep_project_roles", req.KeepProjectRoles),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, err
		}

		// Only admins can remove pending invites.
		// NOTE: If we change invites to accept/decline (instead of auto-accept on signup), we need to revisit this.
		claims := auth.GetClaims(ctx)
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
			return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
		}

		// Check if there is a pending invite
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
		if err != nil {
			return nil, err
		}

		err = s.admin.DB.DeleteOrganizationInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}

		return &adminv1.RemoveOrganizationMemberUserResponse{}, nil
	}

	// The caller must either have ManageOrgMembers permission or be the user being removed.
	claims := auth.GetClaims(ctx)
	isManager := claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers
	isSelf := claims.OwnerType() == auth.OwnerTypeUser && claims.OwnerID() == user.ID
	if !isManager && !isSelf {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	if org.BillingEmail == user.Email {
		return nil, status.Error(codes.InvalidArgument, "this user is the billing email for the organization, please update the billing email before removing")
	}

	// Check that the user is not the last admin
	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		return nil, err
	}
	users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, role.ID)
	if err != nil {
		return nil, err
	}
	if len(users) == 1 && users[0].ID == user.ID {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last admin member")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	err = s.admin.DB.DeleteOrganizationMemberUser(ctx, org.ID, user.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// delete from all user groups of the org
	err = s.admin.DB.DeleteUsergroupsMemberUser(ctx, org.ID, user.ID)
	if err != nil {
		return nil, err
	}

	// delete from projects if KeepProjectRoles flag is set
	if !req.KeepProjectRoles {
		err = s.admin.DB.DeleteAllProjectMemberUserForOrganization(ctx, org.ID, user.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveOrganizationMemberUserResponse{}, nil
}

func (s *Server) SetOrganizationMemberUserRole(ctx context.Context, req *adminv1.SetOrganizationMemberUserRoleRequest) (*adminv1.SetOrganizationMemberUserRoleResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.role", req.Role),
	)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	claims := auth.GetClaims(ctx)
	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to set org members role")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		// Check if there is a pending invite for this user
		invite, err := s.admin.DB.FindOrganizationInvite(ctx, org.ID, req.Email)
		if err != nil {
			return nil, err
		}
		err = s.admin.DB.UpdateOrganizationInviteRole(ctx, invite.ID, role.ID)
		if err != nil {
			return nil, err
		}
		return &adminv1.SetOrganizationMemberUserRoleResponse{}, nil
	}

	// Check if the user is the last owner
	if role.Name != database.OrganizationRoleNameAdmin {
		adminRole, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
		if err != nil {
			panic(err)
		}
		// TODO optimize this, may be extract roles during auth token validation
		//  and store as part of the claims and fetch admins only if the user is an admin
		users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, adminRole.ID)
		if err != nil {
			return nil, err
		}
		if len(users) == 1 && users[0].ID == user.ID {
			return nil, status.Error(codes.InvalidArgument, "cannot change role of the last owner")
		}
	}

	err = s.admin.DB.UpdateOrganizationMemberUserRole(ctx, org.ID, user.ID, role.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.SetOrganizationMemberUserRoleResponse{}, nil
}

func (s *Server) LeaveOrganization(ctx context.Context, req *adminv1.LeaveOrganizationRequest) (*adminv1.LeaveOrganizationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	// Check the request is made by an authenticated user
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers {
		return nil, status.Error(codes.PermissionDenied, "not allowed to remove org members")
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(err)
	}

	user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	if org.BillingEmail == user.Email {
		return nil, status.Error(codes.InvalidArgument, "this user is the billing email for the organization, please update the billing email before leaving")
	}

	// check if the user is the last owner
	// TODO optimize this, may be extract roles during auth token validation
	//  and store as part of the claims and fetch admins only if the user is an admin
	users, err := s.admin.DB.FindOrganizationMemberUsersByRole(ctx, org.ID, role.ID)
	if err != nil {
		return nil, err
	}

	if len(users) == 1 && users[0].ID == claims.OwnerID() {
		return nil, status.Error(codes.InvalidArgument, "cannot remove the last owner")
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	err = s.admin.DB.DeleteOrganizationMemberUser(ctx, org.ID, claims.OwnerID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// delete from all user groups of the org
	err = s.admin.DB.DeleteUsergroupsMemberUser(ctx, org.ID, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &adminv1.LeaveOrganizationResponse{}, nil
}

func (s *Server) CreateWhitelistedDomain(ctx context.Context, req *adminv1.CreateWhitelistedDomainRequest) (*adminv1.CreateWhitelistedDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.domain", req.Domain),
		attribute.String("args.role", req.Role),
	)

	claims := auth.GetClaims(ctx)
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, err
	}

	if !claims.Superuser(ctx) {
		if !claims.OrganizationPermissions(ctx, org.ID).ManageOrg {
			return nil, status.Error(codes.PermissionDenied, "only org admins can add whitelisted domain")
		}
		// check if the user's domain matches the whitelist domain
		user, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(user.Email, "@"+req.Domain) {
			return nil, status.Error(codes.PermissionDenied, "Domain name doesn’t match verified email domain. Please contact Rill support.")
		}

		if publicemail.IsPublic(req.Domain) {
			return nil, status.Errorf(codes.InvalidArgument, "Public Domain %s cannot be whitelisted", req.Domain)
		}
	}

	role, err := s.admin.DB.FindOrganizationRole(ctx, req.Role)
	if err != nil {
		return nil, err
	}

	// find existing users belonging to the whitelisted domain to the org
	users, err := s.admin.DB.FindUsersByEmailPattern(ctx, "%@"+req.Domain, "", math.MaxInt)
	if err != nil {
		return nil, err
	}

	// filter out users who are already members of the org
	newUsers := make([]*database.User, 0)
	for _, user := range users {
		// check if user is already a member of the org
		exists, err := s.admin.DB.CheckUserIsAnOrganizationMember(ctx, user.ID, org.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			newUsers = append(newUsers, user)
		}
	}

	ctx, tx, err := s.admin.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = s.admin.DB.InsertOrganizationWhitelistedDomain(ctx, &database.InsertOrganizationWhitelistedDomainOptions{
		OrgID:     org.ID,
		OrgRoleID: role.ID,
		Domain:    req.Domain,
	})
	if err != nil {
		return nil, err
	}

	for _, user := range newUsers {
		err = s.admin.DB.InsertOrganizationMemberUser(ctx, org.ID, user.ID, role.ID)
		if err != nil {
			return nil, err
		}

		// add to all user group
		err = s.admin.DB.InsertUsergroupMemberUser(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &adminv1.CreateWhitelistedDomainResponse{}, nil
}

func (s *Server) RemoveWhitelistedDomain(ctx context.Context, req *adminv1.RemoveWhitelistedDomainRequest) (*adminv1.RemoveWhitelistedDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
		attribute.String("args.domain", req.Domain),
	)

	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, err
	}

	if !(claims.OrganizationPermissions(ctx, org.ID).ManageOrg || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only org admins can remove whitelisted domain")
	}

	invite, err := s.admin.DB.FindOrganizationWhitelistedDomain(ctx, org.ID, req.Domain)
	if err != nil {
		return nil, err
	}

	err = s.admin.DB.DeleteOrganizationWhitelistedDomain(ctx, invite.ID)
	if err != nil {
		return nil, err
	}

	return &adminv1.RemoveWhitelistedDomainResponse{}, nil
}

func (s *Server) ListWhitelistedDomains(ctx context.Context, req *adminv1.ListWhitelistedDomainsRequest) (*adminv1.ListWhitelistedDomainsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Organization),
	)

	claims := auth.GetClaims(ctx)

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, err
	}

	if !(claims.OrganizationPermissions(ctx, org.ID).ManageOrg || claims.Superuser(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "only org admins can list whitelisted domain")
	}

	whitelistedDomains, err := s.admin.DB.FindOrganizationWhitelistedDomainForOrganizationWithJoinedRoleNames(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	whitelistedDomainDtos := make([]*adminv1.WhitelistedDomain, len(whitelistedDomains))
	for i, whitelistedDomain := range whitelistedDomains {
		whitelistedDomainDtos[i] = whitelistedDomainToPB(whitelistedDomain)
	}

	return &adminv1.ListWhitelistedDomainsResponse{
		Domains: whitelistedDomainDtos,
	}, nil
}

func (s *Server) SudoUpdateOrganizationQuotas(ctx context.Context, req *adminv1.SudoUpdateOrganizationQuotasRequest) (*adminv1.SudoUpdateOrganizationQuotasResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.org", req.Organization))
	if req.Projects != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.projects", int(*req.Projects)))
	}
	if req.Deployments != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.deployments", int(*req.Deployments)))
	}
	if req.SlotsTotal != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.slots_total", int(*req.SlotsTotal)))
	}
	if req.SlotsPerDeployment != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.slots_per_deployment", int(*req.SlotsPerDeployment)))
	}
	if req.OutstandingInvites != nil {
		observability.AddRequestAttributes(ctx, attribute.Int("args.outstanding_invites", int(*req.OutstandingInvites)))
	}

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage quotas")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Organization)
	if err != nil {
		return nil, err
	}

	opts := &database.UpdateOrganizationOptions{
		Name:                                req.Organization,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       int(valOrDefault(req.Projects, int32(org.QuotaProjects))),
		QuotaDeployments:                    int(valOrDefault(req.Deployments, int32(org.QuotaDeployments))),
		QuotaSlotsTotal:                     int(valOrDefault(req.SlotsTotal, int32(org.QuotaSlotsTotal))),
		QuotaSlotsPerDeployment:             int(valOrDefault(req.SlotsPerDeployment, int32(org.QuotaSlotsPerDeployment))),
		QuotaOutstandingInvites:             int(valOrDefault(req.OutstandingInvites, int32(org.QuotaOutstandingInvites))),
		QuotaStorageLimitBytesPerDeployment: valOrDefault(req.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	}

	updatedOrg, err := s.admin.DB.UpdateOrganization(ctx, org.ID, opts)
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateOrganizationQuotasResponse{
		Organization: s.organizationToDTO(updatedOrg, true),
	}, nil
}

func (s *Server) SudoUpdateOrganizationCustomDomain(ctx context.Context, req *adminv1.SudoUpdateOrganizationCustomDomainRequest) (*adminv1.SudoUpdateOrganizationCustomDomainResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.org", req.Name),
		attribute.String("args.custom_domain", req.CustomDomain),
	)

	claims := auth.GetClaims(ctx)
	if !claims.Superuser(ctx) {
		return nil, status.Error(codes.PermissionDenied, "only superusers can manage custom domains")
	}

	org, err := s.admin.DB.FindOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	org, err = s.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		CustomDomain:                        req.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}

	return &adminv1.SudoUpdateOrganizationCustomDomainResponse{
		Organization: s.organizationToDTO(org, true),
	}, nil
}

func (s *Server) organizationToDTO(o *database.Organization, privileged bool) *adminv1.Organization {
	var logoURL string
	if o.LogoAssetID != nil {
		logoURL = s.admin.URLs.WithCustomDomain(o.CustomDomain).Asset(*o.LogoAssetID)
	}

	var faviconURL string
	if o.FaviconAssetID != nil {
		faviconURL = s.admin.URLs.WithCustomDomain(o.CustomDomain).Asset(*o.FaviconAssetID)
	}

	res := &adminv1.Organization{
		Id:           o.ID,
		Name:         o.Name,
		DisplayName:  o.DisplayName,
		Description:  o.Description,
		LogoUrl:      logoURL,
		FaviconUrl:   faviconURL,
		CustomDomain: o.CustomDomain,
		Quotas: &adminv1.OrganizationQuotas{
			Projects:                       int32(o.QuotaProjects),
			Deployments:                    int32(o.QuotaDeployments),
			SlotsTotal:                     int32(o.QuotaSlotsTotal),
			SlotsPerDeployment:             int32(o.QuotaSlotsPerDeployment),
			OutstandingInvites:             int32(o.QuotaOutstandingInvites),
			StorageLimitBytesPerDeployment: o.QuotaStorageLimitBytesPerDeployment,
		},
		CreatedOn: timestamppb.New(o.CreatedOn),
		UpdatedOn: timestamppb.New(o.UpdatedOn),
	}

	if privileged {
		res.BillingCustomerId = o.BillingCustomerID
		res.PaymentCustomerId = o.PaymentCustomerID
		res.BillingEmail = o.BillingEmail
		res.BillingPlanName = o.BillingPlanName
		res.BillingPlanDisplayName = o.BillingPlanDisplayName
	}

	return res
}

func valOrEmptyString(v *int) string {
	if v != nil {
		return strconv.Itoa(*v)
	}
	return ""
}

func val64OrEmptyString(v *int64) string {
	if v != nil {
		return strconv.FormatInt(*v, 10)
	}
	return ""
}

func whitelistedDomainToPB(a *database.OrganizationWhitelistedDomainWithJoinedRoleNames) *adminv1.WhitelistedDomain {
	return &adminv1.WhitelistedDomain{
		Domain: a.Domain,
		Role:   a.RoleName,
	}
}
