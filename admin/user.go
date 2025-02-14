package admin

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

func (s *Service) CreateOrUpdateUser(ctx context.Context, email, name, photoURL string) (*database.User, error) {
	// Validate email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid user email address %q", email)
	}

	// Update user if exists
	user, err := s.DB.FindUserByEmail(ctx, email)
	if err == nil {
		return s.DB.UpdateUser(ctx, user.ID, &database.UpdateUserOptions{
			DisplayName:         name,
			PhotoURL:            photoURL,
			GithubUsername:      user.GithubUsername,
			GithubRefreshToken:  user.GithubRefreshToken,
			QuotaSingleuserOrgs: user.QuotaSingleuserOrgs,
			QuotaTrialOrgs:      user.QuotaTrialOrgs,
			PreferenceTimeZone:  user.PreferenceTimeZone,
		})
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// User does not exist. Creating a new user.

	// Get user invites if exists
	orgInvites, err := s.DB.FindOrganizationInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	projectInvites, err := s.DB.FindProjectInvitesByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	isFirstUser, err := s.DB.CheckUsersEmpty(ctx)
	if err != nil {
		return nil, err
	}

	opts := &database.InsertUserOptions{
		Email:               email,
		DisplayName:         name,
		PhotoURL:            photoURL,
		QuotaSingleuserOrgs: database.DefaultQuotaSingleuserOrgs,
		QuotaTrialOrgs:      database.DefaultQuotaTrialOrgs,
		Superuser:           isFirstUser,
	}

	// Create user
	user, err = s.DB.InsertUser(ctx, opts)
	if err != nil {
		return nil, err
	}

	// handle org invites
	addedToOrgIDs := make(map[string]bool)
	addedToOrgNames := make([]string, 0)
	for _, invite := range orgInvites {
		org, err := s.DB.FindOrganization(ctx, invite.OrgID)
		if err != nil {
			return nil, err
		}
		err = s.DB.InsertOrganizationMemberUser(ctx, invite.OrgID, user.ID, invite.OrgRoleID)
		if err != nil {
			return nil, err
		}

		for _, usergroupID := range invite.UsergroupIDs {
			// check if the user group exists, need to check explicitly as tx is not completed yet
			exists, err := s.DB.CheckUsergroupExists(ctx, usergroupID)
			if err != nil {
				return nil, err
			}

			if !exists {
				// ignore if usergroup does not exist, might have been deleted before invite was accepted
				continue
			}

			err = s.DB.InsertUsergroupMemberUser(ctx, usergroupID, user.ID)
			if err != nil {
				return nil, err
			}
		}

		err = s.DB.InsertUsergroupMemberUser(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			return nil, err
		}
		err = s.DB.DeleteOrganizationInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}
		addedToOrgIDs[org.ID] = true
		addedToOrgNames = append(addedToOrgNames, org.Name)
	}

	// check if users email domain is whitelisted for some organizations
	domain := email[strings.LastIndex(email, "@")+1:]
	organizationWhitelistedDomains, err := s.DB.FindOrganizationWhitelistedDomainsForDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, whitelist := range organizationWhitelistedDomains {
		// if user is already a member of the org then skip, prefer explicit invite to whitelist
		if _, ok := addedToOrgIDs[whitelist.OrgID]; ok {
			continue
		}
		org, err := s.DB.FindOrganization(ctx, whitelist.OrgID)
		if err != nil {
			return nil, err
		}
		err = s.DB.InsertOrganizationMemberUser(ctx, whitelist.OrgID, user.ID, whitelist.OrgRoleID)
		if err != nil {
			return nil, err
		}
		err = s.DB.InsertUsergroupMemberUser(ctx, *org.AllUsergroupID, user.ID)
		if err != nil {
			return nil, err
		}
		addedToOrgIDs[org.ID] = true
		addedToOrgNames = append(addedToOrgNames, org.Name)
	}

	// handle project invites
	addedToProjectIDs := make(map[string]bool)
	addedToProjectNames := make([]string, 0)
	for _, invite := range projectInvites {
		project, err := s.DB.FindProject(ctx, invite.ProjectID)
		if err != nil {
			return nil, err
		}
		err = s.DB.InsertProjectMemberUser(ctx, invite.ProjectID, user.ID, invite.ProjectRoleID)
		if err != nil {
			return nil, err
		}
		err = s.DB.DeleteProjectInvite(ctx, invite.ID)
		if err != nil {
			return nil, err
		}
		addedToProjectIDs[project.ID] = true
		addedToProjectNames = append(addedToProjectNames, project.Name)
	}

	// check if users email domain is whitelisted for some projects
	projectWhitelistedDomains, err := s.DB.FindProjectWhitelistedDomainsForDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	for _, whitelist := range projectWhitelistedDomains {
		// if user is already a member of the project then skip, prefer explicit invite to whitelist
		if _, ok := addedToProjectIDs[whitelist.ProjectID]; ok {
			continue
		}
		project, err := s.DB.FindProject(ctx, whitelist.ProjectID)
		if err != nil {
			return nil, err
		}
		err = s.DB.InsertProjectMemberUser(ctx, whitelist.ProjectID, user.ID, whitelist.ProjectRoleID)
		if err != nil {
			return nil, err
		}
		addedToProjectIDs[project.ID] = true
		addedToProjectNames = append(addedToProjectNames, project.Name)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	s.Logger.Info("created user",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("name", user.DisplayName),
		zap.String("org", strings.Join(addedToOrgNames, ",")),
		zap.String("project", strings.Join(addedToProjectNames, ",")),
	)

	return user, nil
}

func (s *Service) CreateOrganizationForUser(ctx context.Context, userID, email, orgName, description string) (*database.Organization, error) {
	txCtx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	quotaProjects := database.DefaultQuotaProjects
	quotaDeployments := database.DefaultQuotaDeployments
	quotaSlotsTotal := database.DefaultQuotaSlotsTotal
	quotaSlotsPerDeployment := database.DefaultQuotaSlotsPerDeployment
	quotaOutstandingInvites := database.DefaultQuotaOutstandingInvites
	quotaStorageLimitBytesPerDeployment := database.DefaultQuotaStorageLimitBytesPerDeployment

	org, err := s.DB.InsertOrganization(txCtx, &database.InsertOrganizationOptions{
		Name:                                orgName,
		DisplayName:                         "", // Populated later
		Description:                         description,
		LogoAssetID:                         nil,
		FaviconAssetID:                      nil,
		CustomDomain:                        "",
		QuotaProjects:                       quotaProjects,
		QuotaDeployments:                    quotaDeployments,
		QuotaSlotsTotal:                     quotaSlotsTotal,
		QuotaSlotsPerDeployment:             quotaSlotsPerDeployment,
		QuotaOutstandingInvites:             quotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: quotaStorageLimitBytesPerDeployment,
		BillingEmail:                        email,
		BillingCustomerID:                   "", // Populated later
		PaymentCustomerID:                   "", // Populated later
		CreatedByUserID:                     &userID,
	})
	if err != nil {
		return nil, err
	}

	org, err = s.prepareOrganization(txCtx, org.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	s.Logger.Info("created org", zap.String("name", orgName), zap.String("user_id", userID))

	// raise never subscribed billing issue in sync to prevent race condition where first project is deployed before issue is raised and thus start trial job not submitted
	if s.Biller.Name() != "noop" {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypeNeverSubscribed,
			Metadata:  database.BillingIssueMetadataNeverSubscribed{},
			EventTime: org.CreatedOn,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	// Submit job to init org billing // TODO modify river client to allow job submission as part of transaction
	_, err = s.Jobs.InitOrgBilling(ctx, org.ID)
	if err != nil {
		s.Logger.Named("billing").Error("failed to submit job to init org billing", zap.String("org_id", org.ID), zap.String("org_name", orgName), zap.Error(err))
		return org, nil
	}

	return org, nil
}

func (s *Service) prepareOrganization(ctx context.Context, orgID, userID string) (*database.Organization, error) {
	// create all user group for this org
	userGroup, err := s.DB.InsertUsergroup(ctx, &database.InsertUsergroupOptions{
		OrgID: orgID,
		Name:  "all-users",
	})
	if err != nil {
		return nil, err
	}
	// update org with all user group
	org, err := s.DB.UpdateOrganizationAllUsergroup(ctx, orgID, userGroup.ID)
	if err != nil {
		return nil, err
	}

	role, err := s.DB.FindOrganizationRole(ctx, database.OrganizationRoleNameAdmin)
	if err != nil {
		panic(err)
	}

	// Add user to created org with org admin role
	err = s.DB.InsertOrganizationMemberUser(ctx, orgID, userID, role.ID)
	if err != nil {
		return nil, err
	}
	// Add user to all user group
	err = s.DB.InsertUsergroupMemberUser(ctx, userGroup.ID, userID)
	if err != nil {
		return nil, err
	}
	return org, nil
}
