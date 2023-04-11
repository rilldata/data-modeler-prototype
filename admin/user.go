package admin

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/pkg/errors"
	"github.com/rilldata/rill/admin/database"
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
		return s.DB.UpdateUser(ctx, user.ID, name, photoURL)
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// User does not exist. Creating a new user.
	user, err = s.DB.InsertUser(ctx, email, name, photoURL)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) CreateOrganizationForUser(ctx context.Context, userID, orgName, description string) (*database.Organization, error) {
	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	org, err := s.DB.InsertOrganization(ctx, orgName, description)
	if err != nil {
		return nil, err
	}

	org, err = s.prepareOrganization(ctx, org.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (s *Service) prepareOrganization(ctx context.Context, orgID, userID string) (*database.Organization, error) {
	// create all user group for this org
	userGroup, err := s.DB.InsertOrganizationMemberUsergroup(ctx, orgID, "all-users")
	if err != nil {
		return nil, err
	}
	// update org with all user group
	org, err := s.DB.UpdateOrganizationMemberAllUsergroup(ctx, orgID, userGroup.ID)
	if err != nil {
		return nil, err
	}

	role, err := s.DB.FindOrganizationRole(ctx, database.OrganizationAdminRoleName)
	if err != nil {
		panic(errors.Wrap(err, "failed to find organization admin role"))
	}

	// Add user to created org with org admin role
	err = s.DB.InsertOrganizationMemberUser(ctx, orgID, userID, role.ID)
	if err != nil {
		return nil, err
	}
	// Add user to all user group
	err = s.DB.InsertUserInUsergroup(ctx, userID, userGroup.ID)
	if err != nil {
		return nil, err
	}
	return org, nil
}
