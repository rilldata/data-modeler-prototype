package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	wire "github.com/jeroenrinzema/psql-wire"
	"github.com/lib/pq/oid"
	"github.com/rilldata/rill/admin/server/auth"
	runtimeauth "github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

// psqlProxyQueryHandler handles queries on PSQL wire endpoint.
// It proxies queries to the runtime based on targeted project and org.
func (s *Server) psqlProxyQueryHandler(ctx context.Context, query string) (stmt wire.PreparedStatements, err error) {
	s.logger.Info("psql query", zap.String("query", query))
	now := time.Now()
	defer func() {
		// if there is no error then the handler returns a row handler and the query is not completed yet.
		if err != nil {
			s.logger.Info("psql query failed", zap.Duration("duration", time.Since(now)), zap.Error(err))
		}
	}()

	if strings.TrimSpace(query) == "" {
		return wire.Prepared(wire.NewStatement(func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
			return writer.Empty()
		})), nil
	}

	if hasPrefixFold(query, "SET") {
		return wire.Prepared(wire.NewStatement(func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
			return writer.Complete("SET")
		}, wire.WithColumns(nil))), nil
	}

	if hasPrefixFold(query, "BEGIN") || hasPrefixFold(query, "COMMIT") || hasPrefixFold(query, "ROLLBACK") {
		return wire.Prepared(wire.NewStatement(func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
			return writer.Complete(strings.ToUpper(strings.Trim(query, ";")))
		}, wire.WithColumns(nil))), nil
	}

	clientParams := wire.ClientParameters(ctx)
	pool, err := s.psqlConnectionPool.acquire(ctx, clientParams[wire.ParamDatabase])
	if err != nil {
		return nil, err
	}

	// acquire a connection from the pool and make sure to release the connection back to pool in all paths
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// We need to set the field descriptors to support extended protocol.
	// We create a prepared statement for the query but do not execute it to prevent buffering full results into memory.
	// The other approach is executing the query and only reading field descriptors but it keeps the connection in-use till handler is called via ClientExecute which can leak connections.
	// NOTE : at present runtime executes full query and sets result in the cache but it should be okay since results will be read immediately in most cases.
	//
	// Ideally we should use anonymous prepared statements(name="") but it is not supported by psql-wire.
	sd, err := conn.Conn().Prepare(ctx, query, query)
	if err != nil {
		return nil, err
	}

	// handle schema
	cols := make([]wire.Column, 0, len(sd.Fields))
	for _, fd := range sd.Fields {
		cols = append(cols, wire.Column{
			Table: int32(fd.TableOID),
			Name:  fd.Name,
			Oid:   oid.Oid(fd.DataTypeOID),
			Width: fd.DataTypeSize,
			Attr:  int16(fd.TableAttributeNumber),
		})
	}

	// handle data
	handle := func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) (err error) {
		defer func() {
			s.logger.Info("psql proxy query executed", zap.Duration("duration", time.Since(now)), zap.Error(err))
		}()
		rows, err := pool.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			d, err := rows.Values()
			if err != nil {
				return err
			}
			if err := writer.Row(d); err != nil {
				return err
			}
		}
		if rows.Err() != nil {
			return rows.Err()
		}

		return writer.Complete("OK")
	}
	return wire.Prepared(wire.NewStatement(handle, wire.WithColumns(cols))), nil
}

func (s *Server) runtimeJWT(ctx context.Context, org, project string) (string, error) {
	// Find the production deployment for the project we're proxying to
	proj, err := s.admin.DB.FindProjectByName(ctx, org, project)
	if err != nil {
		return "", fmt.Errorf("invalid org or project")
	}

	if proj.ProdDeploymentID == nil {
		return "", fmt.Errorf("no prod deployment for project")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return "", fmt.Errorf("no prod deployment for project")
	}

	claims := auth.GetClaims(ctx)
	switch claims.OwnerType() {
	case auth.OwnerTypeUser, auth.OwnerTypeService:
	default:
		return "", fmt.Errorf("runtime proxy not available for owner type %q", claims.OwnerType())
	}

	// The JWT should have the same permissions/configuration as one they would get by calling AdminService.GetProject.
	permissions := claims.ProjectPermissions(ctx, *proj.ProdDeploymentID, depl.ProjectID)
	if !permissions.ReadProd {
		return "", fmt.Errorf("does not have permission to access the production deployment")
	}

	var attr map[string]any
	if claims.OwnerType() == auth.OwnerTypeUser {
		attr, err = s.jwtAttributesForUser(ctx, claims.OwnerID(), proj.OrganizationID, permissions)
		if err != nil {
			return "", err
		}
	}

	return s.issuer.NewToken(runtimeauth.TokenOptions{
		AudienceURL: depl.RuntimeAudience,
		Subject:     claims.OwnerID(),
		TTL:         runtimeAccessTokenDefaultTTL,
		InstancePermissions: map[string][]runtimeauth.Permission{
			depl.RuntimeInstanceID: {
				// TODO: Remove ReadProfiling and ReadRepo (may require frontend changes)
				runtimeauth.ReadObjects,
				runtimeauth.ReadMetrics,
				runtimeauth.ReadProfiling,
				runtimeauth.ReadRepo,
				runtimeauth.ReadAPI,
			},
		},
		Attributes: attr,
	})
}

type psqlConnectionPool struct {
	runtimePool map[string]*pgxpool.Pool
	mu          sync.Mutex
	closed      bool
	server      *Server
}

func newPSQLConnectionPool(server *Server) *psqlConnectionPool {
	return &psqlConnectionPool{
		runtimePool: make(map[string]*pgxpool.Pool),
		server:      server,
	}
}

func (p *psqlConnectionPool) acquire(ctx context.Context, db string) (*pgxpool.Pool, error) {
	// database is org.project
	tokens := strings.Split(db, ".")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid org or project")
	}
	org := tokens[0]
	project := tokens[1]

	// Find the production deployment for the project we're proxying to
	proj, err := p.server.admin.DB.FindProjectByName(ctx, org, project)
	if err != nil {
		return nil, fmt.Errorf("invalid org or project")
	}

	if proj.ProdDeploymentID == nil {
		return nil, fmt.Errorf("no prod deployment for project")
	}

	depl, err := p.server.admin.DB.FindDeployment(ctx, *proj.ProdDeploymentID)
	if err != nil {
		return nil, fmt.Errorf("no prod deployment for project")
	}
	// Track usage of the deployment
	p.server.admin.Used.Deployment(depl.ID)

	// generate a postgres dsn for the runtime psql endpoint
	port := 5432
	// hack : the runtime port on localhost is 15432 so that it does not conflict with the postgres db port used in admin service
	if strings.Contains(depl.RuntimeHost, "localhost") {
		port = 15432
	}

	hostURL, err := url.Parse(depl.RuntimeHost)
	if err != nil {
		return nil, err
	}
	hostURL.Scheme = "postgres"
	hostURL.Host = hostURL.Hostname() + ":" + strconv.FormatInt(int64(port), 10)
	hostURL.User = url.UserPassword("postgres", "")
	hostURL.Path = depl.RuntimeInstanceID

	dsn := hostURL.String()

	// acquire connection pool from the cache
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, errors.New("pool is closed")
	}

	pool, ok := p.runtimePool[dsn]
	if ok {
		return pool, nil
	}

	// parse dsn into pgxpool.Config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	// Runtime JWts are valid for 30 minutes only
	config.MaxConnLifetime = time.Minute * 29
	// since runtimes get restarted more often than actual DB servers. Consider if this should be reduced to even less time
	// also consider if we should add some health check on connection acquisition
	config.HealthCheckPeriod = time.Minute
	// issues a runtime JWT and set it as password
	config.BeforeConnect = func(ctx context.Context, cc *pgx.ConnConfig) error {
		p.server.logger.Info("opening new connection to runtime", zap.String("host", cc.Host), zap.String("instance_id", cc.Database))
		tokens := strings.Split(db, ".")
		if len(tokens) != 2 {
			// this error will be captured much earlier
			return fmt.Errorf("developer error: invalid org or project")
		}
		org := tokens[0]
		project := tokens[1]

		jwt, err := p.server.runtimeJWT(ctx, org, project)
		if err != nil {
			return err
		}
		cc.Password = jwt
		return nil
	}

	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to psql: %w", err)
	}
	p.runtimePool[dsn] = pool
	return pool, nil
}

func (p *psqlConnectionPool) close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	for _, pool := range p.runtimePool {
		pool.Close()
	}
}

func hasPrefixFold(str, prefix string) bool {
	return len(str) >= len(prefix) && strings.EqualFold(str[:len(prefix)], prefix)
}
