package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
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
	Can(ctx context.Context, orgID string, op OrganizationPermission, projID string, pp ProjectPermission) bool
	OrganizationPermissions(ctx context.Context, orgID string) (*adminv1.OrganizationPermissions, error)
	ProjectPermissions(ctx context.Context, projectID string) (*adminv1.ProjectPermissions, error)
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

func (c anonClaims) Can(ctx context.Context, orgID string, op OrganizationPermission, projectID string, pp ProjectPermission) bool {
	return false
}

func (c anonClaims) OrganizationPermissions(ctx context.Context, orgID string) (*adminv1.OrganizationPermissions, error) {
	return &adminv1.OrganizationPermissions{}, nil
}

func (c anonClaims) ProjectPermissions(ctx context.Context, projectID string) (*adminv1.ProjectPermissions, error) {
	return &adminv1.ProjectPermissions{}, nil
}

// authTokenClaims represents claims for an admin.AuthToken.
type authTokenClaims struct {
	token                   admin.AuthToken
	admin                   *admin.Service
	orgPermissionsCache     map[string]*adminv1.OrganizationPermissions
	projectPermissionsCache map[string]*adminv1.ProjectPermissions
	orgCacheLock            sync.Mutex
	projectCacheLock        sync.Mutex
}

func newAuthTokenClaims(token admin.AuthToken, adminService *admin.Service) Claims {
	return &authTokenClaims{
		token:                   token,
		admin:                   adminService,
		orgPermissionsCache:     make(map[string]*adminv1.OrganizationPermissions),
		projectPermissionsCache: make(map[string]*adminv1.ProjectPermissions),
		orgCacheLock:            sync.Mutex{},
		projectCacheLock:        sync.Mutex{},
	}
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
		permissions, err := c.OrganizationPermissions(ctx, orgID)
		if err != nil {
			panic(fmt.Errorf("failed to get organization permissions: %w", err))
		}
		switch p {
		case ReadOrg:
			return permissions.ReadOrg
		case ManageOrg:
			return permissions.ManageOrg
		case ReadProjects:
			return permissions.ReadProjects
		case CreateProjects:
			return permissions.CreateProjects
		case ManageProjects:
			return permissions.ManageProjects
		case ReadOrgMembers:
			return permissions.ReadOrgMembers
		case ManageOrgMembers:
			return permissions.ManageOrgMembers
		default:
			panic(fmt.Errorf("unexpected organization permission %q", p))
		}
	case authtoken.TypeService:
		panic(errors.New("service tokens not supported"))
	default:
		panic(fmt.Errorf("unexpected token type %q", t))
	}
}

func (c *authTokenClaims) CanProject(ctx context.Context, projectID string, p ProjectPermission) bool {
	t := c.token.Token().Type
	switch t {
	case authtoken.TypeUser:
		permissions, err := c.ProjectPermissions(ctx, projectID)
		if err != nil {
			panic(fmt.Errorf("failed to get project permissions: %w", err))
		}
		switch p {
		case ReadProject:
			return permissions.ReadProject
		case ManageProject:
			return permissions.ManageProject
		case ReadProdBranch:
			return permissions.ReadProdBranch
		case ManageProdBranch:
			return permissions.ManageProdBranch
		case ReadDevBranches:
			return permissions.ReadDevBranches
		case ManageDevBranches:
			return permissions.ManageDevBranches
		case ReadProjectMembers:
			return permissions.ReadProjectMembers
		case ManageProjectMembers:
			return permissions.ManageProjectMembers
		default:
			panic(fmt.Errorf("unexpected organization permission %q", p))
		}
	case authtoken.TypeService:
		panic(errors.New("service tokens not supported"))
	default:
		panic(fmt.Errorf("unexpected token type %q", t))
	}
}

func (c *authTokenClaims) Can(ctx context.Context, orgID string, op OrganizationPermission, projectID string, pp ProjectPermission) bool {
	return c.CanOrganization(ctx, orgID, op) || c.CanProject(ctx, projectID, pp)
}

func (c *authTokenClaims) OrganizationPermissions(ctx context.Context, orgID string) (*adminv1.OrganizationPermissions, error) {
	c.orgCacheLock.Lock()
	if perm, ok := c.orgPermissionsCache[orgID]; ok {
		return perm, nil
	}
	c.orgCacheLock.Unlock()

	composite := &adminv1.OrganizationPermissions{}
	roles, err := c.admin.DB.ResolveOrganizationMemberUserRoles(ctx, c.token.OwnerID(), orgID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		composite = unionOrgRoles(composite, role)
	}

	c.orgCacheLock.Lock()
	c.orgPermissionsCache[orgID] = composite
	c.orgCacheLock.Unlock()
	return composite, nil
}

func (c *authTokenClaims) ProjectPermissions(ctx context.Context, projectID string) (*adminv1.ProjectPermissions, error) {
	c.projectCacheLock.Lock()
	if perm, ok := c.projectPermissionsCache[projectID]; ok {
		return perm, nil
	}
	c.projectCacheLock.Unlock()

	composite := &adminv1.ProjectPermissions{}
	roles, err := c.admin.DB.ResolveProjectMemberUserRoles(ctx, c.token.OwnerID(), projectID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		composite = unionProjectRoles(composite, role)
	}

	c.projectCacheLock.Lock()
	c.projectPermissionsCache[projectID] = composite
	c.projectCacheLock.Unlock()
	return composite, nil
}

func unionOrgRoles(a *adminv1.OrganizationPermissions, b *database.OrganizationRole) *adminv1.OrganizationPermissions {
	return &adminv1.OrganizationPermissions{
		ReadOrg:          a.ReadOrg || b.ReadOrg,
		ManageOrg:        a.ManageOrg || b.ManageOrg,
		ReadProjects:     a.ReadProjects || b.ReadProjects,
		CreateProjects:   a.CreateProjects || b.CreateProjects,
		ManageProjects:   a.ManageProjects || b.ManageProjects,
		ReadOrgMembers:   a.ReadOrgMembers || b.ReadOrgMembers,
		ManageOrgMembers: a.ManageOrgMembers || b.ManageOrgMembers,
	}
}

func unionProjectRoles(a *adminv1.ProjectPermissions, b *database.ProjectRole) *adminv1.ProjectPermissions {
	return &adminv1.ProjectPermissions{
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
