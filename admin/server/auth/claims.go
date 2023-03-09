package auth

import (
	"context"
	"fmt"
	"github.com/rilldata/rill/admin/database"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/pkg/authtoken"
)

// OwnerType is an enum of types of claim owners
type OwnerType string

const (
	OwnerTypeAnon OwnerType = "anon"
	OwnerTypeUser OwnerType = "user"
)

// Claims resolves permissions for a requester.
type Claims interface {
	OwnerType() OwnerType
	OwnerID() string
	AuthTokenID() string
	CanOrganization(ctx context.Context, orgID string, p OrganizationPermission) bool
	CanProject(ctx context.Context, projectID string, p ProjectPermission) bool
}

// claimsContextKey is used to set and get Claims on a request context.
type claimsContextKey struct{}

// GetClaims retrieves Claims from a request context.
// It should only be used in handlers intercepted by UnaryServerInterceptor or StreamServerInterceptor.
func GetClaims(ctx context.Context) Claims {
	claims, ok := ctx.Value(claimsContextKey{}).(Claims)
	if !ok {
		return nil
	}

	return claims
}

// anonClaims represents claims for an unauthenticated user.
type anonClaims struct{}

func (c anonClaims) OwnerType() OwnerType {
	return OwnerTypeAnon
}

func (c anonClaims) OwnerID() string {
	return ""
}

func (c anonClaims) AuthTokenID() string {
	return ""
}

func (c anonClaims) CanOrganization(ctx context.Context, orgID string, p OrganizationPermission) bool {
	return false
}

func (c anonClaims) CanProject(ctx context.Context, projectID string, p ProjectPermission) bool {
	return false
}

// authTokenClaims represents claims for an admin.AuthToken.
type authTokenClaims struct {
	token admin.AuthToken
	admin *admin.Service
}

func (c *authTokenClaims) OwnerType() OwnerType {
	t := c.token.Token().Type
	switch t {
	case authtoken.TypeUser:
		return OwnerTypeUser
	default:
		panic(fmt.Errorf("unexpected token type %q", t))
	}
}

func (c *authTokenClaims) OwnerID() string {
	return c.token.OwnerID()
}

func (c *authTokenClaims) AuthTokenID() string {
	return c.token.Token().ID.String()
}

func (c *authTokenClaims) CanOrganization(ctx context.Context, orgID string, p OrganizationPermission) bool {
	t := c.token.Token().Type
	switch t {
	case authtoken.TypeUser:
		role, err := c.composeOrgPermissions(ctx, orgID)
		if err != nil {
			// TODO: log error
			return false
		}
		switch p {
		case ReadOrg:
			return role.ReadOrganization
		case ManageOrg:
			return role.ManageOrganization
		case ReadProjects:
			return role.ReadProjects
		case CreateProjects:
			return role.CreateProjects
		case ManageProjects:
			return role.ManageProjects
		case ReadOrgMembers:
			return role.ReadOrgMembers
		case ManageOrgMembers:
			return role.ManageOrgMembers
		default:
			panic(fmt.Errorf("unexpected organization permission %q", p))
		}
	case authtoken.TypeService:
		return false
	default:
		panic(fmt.Errorf("unexpected token type %q", t))
	}
}

func (c *authTokenClaims) CanProject(ctx context.Context, projectID string, p ProjectPermission) bool {
	t := c.token.Token().Type
	switch t {
	case authtoken.TypeUser:
		role, err := c.composeProjectPermissions(ctx, projectID)
		if err != nil {
			// TODO: log error
			return false
		}
		switch p {
		case ReadProject:
			return role.ReadProject
		case ManageProject:
			return role.ManageProject
		case ReadProdBranch:
			return role.ReadProdBranch
		case ManageProdBranch:
			return role.ManageProdBranch
		case ReadDevBranches:
			return role.ReadDevBranches
		case ManageDevBranches:
			return role.ManageDevBranches
		case ReadProjectMembers:
			return role.ReadProjectMembers
		case ManageProjectMembers:
			return role.ManageProjectMembers
		default:
			// TODO: log error
			return false
		}
	default:
		// TODO: log error
		return false
	}
}

func (c *authTokenClaims) composeOrgPermissions(ctx context.Context, orgID string) (*database.OrganizationRole, error) {
	var composite *database.OrganizationRole
	directRole, err := c.admin.DB.ResolveUserOrganizationRole(ctx, c.token.OwnerID(), orgID)
	if err != nil {
		// no direct role, check group roles
		if err != database.ErrNotFound {
			return nil, err
		}
	}
	if directRole != nil {
		composite = unionOrgRoles(composite, directRole)
	}
	groupRoles, err := c.admin.DB.ResolveUserGroupOrgRoles(ctx, c.token.OwnerID(), orgID)
	if err != nil {
		return nil, err
	}

	for _, role := range groupRoles {
		composite = unionOrgRoles(composite, role)
	}
	return composite, nil
}

func unionOrgRoles(a, b *database.OrganizationRole) *database.OrganizationRole {
	return &database.OrganizationRole{
		ReadOrganization:   a.ReadOrganization || b.ReadOrganization,
		ManageOrganization: a.ManageOrganization || b.ManageOrganization,
		ReadProjects:       a.ReadProjects || b.ReadProjects,
		CreateProjects:     a.CreateProjects || b.CreateProjects,
		ManageProjects:     a.ManageProjects || b.ManageProjects,
		ReadOrgMembers:     a.ReadOrgMembers || b.ReadOrgMembers,
		ManageOrgMembers:   a.ManageOrgMembers || b.ManageOrgMembers,
	}
}

func (c *authTokenClaims) composeProjectPermissions(ctx context.Context, projectID string) (*database.ProjectRole, error) {
	var composite *database.ProjectRole
	directRole, err := c.admin.DB.ResolveUserProjectRole(ctx, c.token.OwnerID(), projectID)
	if err != nil {
		// no direct role, check group roles
		if err != database.ErrNotFound {
			return nil, err
		}
	}
	if directRole != nil {
		composite = unionProjectRoles(composite, directRole)
	}
	groupRoles, err := c.admin.DB.ResolveUserGroupProjectRoles(ctx, c.token.OwnerID(), projectID)
	if err != nil {
		return nil, err
	}

	for _, role := range groupRoles {
		composite = unionProjectRoles(composite, role)
	}
	return composite, nil
}

func unionProjectRoles(a, b *database.ProjectRole) *database.ProjectRole {
	return &database.ProjectRole{
		ReadProject:          a.ReadProject || b.ReadProject,
		ManageProject:        a.ManageProject || b.ManageProject,
		ReadProdBranch:       a.ReadProdBranch || b.ReadProdBranch,
		ManageProdBranch:     a.ManageProdBranch || b.ManageProdBranch,
		ReadDevBranches:      a.ReadDevBranches || b.ReadDevBranches,
		ManageDevBranches:    a.ManageDevBranches || b.ManageDevBranches,
		ReadProjectMembers:   a.ReadProjectMembers || b.ReadProjectMembers,
		ManageProjectMembers: a.ManageProjectMembers || b.ManageProjectMembers,
	}
}
