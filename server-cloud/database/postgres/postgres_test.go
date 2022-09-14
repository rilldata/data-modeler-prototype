package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rilldata/rill/server-cloud/database"
)

// TestPostgres starts Postgres using testcontainers and runs all other tests in
// this file as sub-tests (to prevent spawning many clusters).
func TestPostgres(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s:%d/postgres", host, port.Int())

	db, err := database.Open("postgres", databaseURL)
	require.NoError(t, err)
	require.NotNil(t, db)

	require.NoError(t, db.Migrate(ctx))

	t.Run("TestOrganizations", func(t *testing.T) { testOrganizations(t, db) })
	t.Run("TestProjects", func(t *testing.T) { testProjects(t, db) })
	// Add new tests here

	require.NoError(t, db.Close())
}

func testOrganizations(t *testing.T, db database.DB) {
	ctx := context.Background()

	org, err := db.FindOrganizationByName(ctx, "foo")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, org)

	org, err = db.CreateOrganization(ctx, "foo", "hello world")
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)
	require.Equal(t, "hello world", org.Description)
	require.Less(t, time.Since(org.CreatedOn), 10*time.Second)
	require.Less(t, time.Since(org.UpdatedOn), 10*time.Second)

	org, err = db.FindOrganizationByName(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)
	require.Equal(t, "hello world", org.Description)

	org, err = db.UpdateOrganization(ctx, org.Name, "")
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)
	require.Equal(t, "", org.Description)

	err = db.DeleteOrganization(ctx, org.Name)
	require.NoError(t, err)

	org, err = db.FindOrganizationByName(ctx, "foo")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, org)
}

func testProjects(t *testing.T, db database.DB) {
	ctx := context.Background()

	org, err := db.CreateOrganization(ctx, "foo", "")
	require.NoError(t, err)
	require.Equal(t, "foo", org.Name)

	proj, err := db.FindProjectByName(ctx, org.Name, "bar")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, proj)

	proj, err = db.CreateProject(ctx, org.ID, "bar", "hello world")
	require.NoError(t, err)
	require.Equal(t, org.ID, proj.OrganizationID)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, "hello world", proj.Description)
	require.Less(t, time.Since(proj.CreatedOn), 10*time.Second)
	require.Less(t, time.Since(proj.UpdatedOn), 10*time.Second)

	proj, err = db.FindProjectByName(ctx, org.Name, proj.Name)
	require.NoError(t, err)
	require.Equal(t, org.ID, proj.OrganizationID)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, "hello world", proj.Description)

	proj, err = db.UpdateProject(ctx, proj.ID, "")
	require.NoError(t, err)
	require.Equal(t, org.ID, proj.OrganizationID)
	require.Equal(t, "bar", proj.Name)
	require.Equal(t, "", proj.Description)

	err = db.DeleteOrganization(ctx, org.Name)
	require.ErrorContains(t, err, "constraint")

	err = db.DeleteProject(ctx, proj.ID)
	require.NoError(t, err)

	proj, err = db.FindProjectByName(ctx, org.Name, "bar")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, proj)

	err = db.DeleteOrganization(ctx, org.Name)
	require.NoError(t, err)

	org, err = db.FindOrganizationByName(ctx, "foo")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, org)
}
