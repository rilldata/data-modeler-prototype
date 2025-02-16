package rillv1

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

// Built-in parser limits
const (
	maxFiles    = 10000
	maxFileSize = 1 << 17 // 128kb
)

// ignorePathPrefixes are prefixes of paths that should be ignored by the parser.
// Note: Generally these ignores should be applied at the repo level (through the defaults for ignore_paths in rill.yaml).
// This is only for files we DO want to list and show in the UI, but don't want to parse.
var ignorePathPrefixes = []string{
	"/.rillcloud/",
	"/.github/",
}

// Resource parsed from code files.
// One file may output multiple resources and multiple files may contribute config to one resource.
type Resource struct {
	// Metadata
	Name    ResourceName
	Paths   []string
	Refs    []ResourceName // Derived from rawRefs after parsing (can't contain ResourceKindUnspecified). Always sorted.
	rawRefs []ResourceName // Populated during parsing (may contain ResourceKindUnspecified)

	// Only one of these will be non-nil
	SourceSpec      *runtimev1.SourceSpec
	ModelSpec       *runtimev1.ModelSpec
	MetricsViewSpec *runtimev1.MetricsViewSpec
	ExploreSpec     *runtimev1.ExploreSpec
	MigrationSpec   *runtimev1.MigrationSpec
	ReportSpec      *runtimev1.ReportSpec
	AlertSpec       *runtimev1.AlertSpec
	ThemeSpec       *runtimev1.ThemeSpec
	ComponentSpec   *runtimev1.ComponentSpec
	CanvasSpec      *runtimev1.CanvasSpec
	APISpec         *runtimev1.APISpec
	ConnectorSpec   *runtimev1.ConnectorSpec
}

// ResourceName is a unique identifier for a resource
type ResourceName struct {
	Kind ResourceKind
	Name string
}

func (n ResourceName) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Name)
}

func (n ResourceName) Normalized() ResourceName {
	return ResourceName{
		Kind: n.Kind,
		Name: strings.ToLower(n.Name),
	}
}

// ResourceKind identifies a resource type supported by the parser
type ResourceKind int

const (
	ResourceKindUnspecified ResourceKind = iota
	ResourceKindSource
	ResourceKindModel
	ResourceKindMetricsView
	ResourceKindExplore
	ResourceKindMigration
	ResourceKindReport
	ResourceKindAlert
	ResourceKindTheme
	ResourceKindComponent
	ResourceKindCanvas
	ResourceKindAPI
	ResourceKindConnector
)

// ParseResourceKind maps a string to a ResourceKind.
// Note: The empty string is considered a valid kind (unspecified).
func ParseResourceKind(kind string) (ResourceKind, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "":
		return ResourceKindUnspecified, nil
	case "source":
		return ResourceKindSource, nil
	case "model":
		return ResourceKindModel, nil
	case "metricsview", "metrics_view":
		return ResourceKindMetricsView, nil
	case "explore":
		return ResourceKindExplore, nil
	case "migration":
		return ResourceKindMigration, nil
	case "report":
		return ResourceKindReport, nil
	case "alert":
		return ResourceKindAlert, nil
	case "theme":
		return ResourceKindTheme, nil
	case "component":
		return ResourceKindComponent, nil
	case "canvas":
		return ResourceKindCanvas, nil
	case "api":
		return ResourceKindAPI, nil
	case "connector":
		return ResourceKindConnector, nil
	default:
		return ResourceKindUnspecified, fmt.Errorf("invalid resource type %q", kind)
	}
}

func (k ResourceKind) String() string {
	switch k {
	case ResourceKindUnspecified:
		return ""
	case ResourceKindSource:
		return "Source"
	case ResourceKindModel:
		return "Model"
	case ResourceKindMetricsView:
		return "MetricsView"
	case ResourceKindExplore:
		return "Explore"
	case ResourceKindMigration:
		return "Migration"
	case ResourceKindReport:
		return "Report"
	case ResourceKindAlert:
		return "Alert"
	case ResourceKindTheme:
		return "Theme"
	case ResourceKindComponent:
		return "Component"
	case ResourceKindCanvas:
		return "Canvas"
	case ResourceKindAPI:
		return "API"
	case ResourceKindConnector:
		return "Connector"
	default:
		panic(fmt.Sprintf("unexpected resource type: %d", k))
	}
}

// Diff shows changes to Parser.Resources following an incremental reparse.
type Diff struct {
	Reloaded       bool
	Skipped        bool
	Added          []ResourceName
	Modified       []ResourceName
	ModifiedDotEnv bool
	Deleted        []ResourceName
}

// Parser parses a Rill project directory into a set of resources.
// After the initial parse, the parser can be used to incrementally reparse a subset of files.
// Parser is not concurrency safe.
type Parser struct {
	// Options
	Repo                 drivers.RepoStore
	InstanceID           string
	Environment          string
	DefaultOLAPConnector string

	// Output
	RillYAML  *RillYAML
	DotEnv    map[string]string
	Resources map[ResourceName]*Resource
	Errors    []*runtimev1.ParseError

	// Internal state
	resourcesForPath           map[string][]*Resource    // Reverse index of Resource.Paths
	resourcesForUnspecifiedRef map[string][]*Resource    // Reverse index of Resource.rawRefs where kind=ResourceKindUnspecified
	resourceNamesForDataPaths  map[string][]ResourceName // Index of local data files to resources that depend on them
	insertedResources          []*Resource
	updatedResources           []*Resource
	deletedResources           []*Resource
}

// ParseRillYAML parses only the project's rill.yaml (or rill.yml) file.
func ParseRillYAML(ctx context.Context, repo drivers.RepoStore, instanceID string) (*RillYAML, error) {
	files, err := repo.ListRecursive(ctx, "rill.{yaml,yml}", true)
	if err != nil {
		return nil, fmt.Errorf("could not list project files: %w", err)
	}

	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}

	p := Parser{Repo: repo, InstanceID: instanceID}
	err = p.parsePaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	if p.RillYAML == nil {
		return nil, ErrRillYAMLNotFound
	}

	return p.RillYAML, nil
}

// Parse creates a new parser and parses the entire project.
func Parse(ctx context.Context, repo drivers.RepoStore, instanceID, environment, defaultOLAPConnector string) (*Parser, error) {
	p := &Parser{
		Repo:                 repo,
		InstanceID:           instanceID,
		Environment:          environment,
		DefaultOLAPConnector: defaultOLAPConnector,
	}

	err := p.reload(ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Reparse re-parses the indicated file paths, updating the Parser's state.
// If rill.yaml has previously errored, or if rill.yaml is included in paths, it will reload the entire project.
// If a previous call to Reparse has returned an error, the Parser may not be accessed or called again.
func (p *Parser) Reparse(ctx context.Context, paths []string) (*Diff, error) {
	var changedRillYAML bool
	for _, path := range paths {
		if !pathIsRillYAML(path) {
			continue
		}
		oldRillYAML := p.RillYAML
		err := p.parseRillYAML(ctx, path)
		if err == nil {
			// Watcher sends multiple events for a single edit. We want to only restart the controller when rill.yaml actually changes.
			// So we check the new rill.yaml contents against the contents stored in parser state.
			changedRillYAML = !reflect.DeepEqual(oldRillYAML, p.RillYAML)
		} else {
			// any error including parse error means rill.yaml changed
			changedRillYAML = true
		}
		break
	}

	if changedRillYAML {
		err := p.reload(ctx)
		if err != nil {
			return nil, err
		}
		return &Diff{Reloaded: true}, nil
	}

	// If rill.yaml previously errored, we're not going to reparse anything until it's changed, at which point we'll reload the entire project.
	if p.RillYAML == nil {
		return &Diff{Skipped: true}, nil
	}

	return p.reparseExceptRillYAML(ctx, paths)
}

// IsSkippable returns true if the path will be skipped by Reparse.
// It's useful for callers to avoid triggering a reparse when they know the path is not relevant.
func (p *Parser) IsSkippable(path string) bool {
	if pathIsIgnored(path) {
		return true
	}
	_, ok := p.resourceNamesForDataPaths[path]
	if ok {
		return false
	}
	return !pathIsYAML(path) && !pathIsSQL(path) && !pathIsDotEnv(path)
}

// TrackedPathsInDir returns the paths under the given directory that the parser currently has cached results for.
func (p *Parser) TrackedPathsInDir(dir string) []string {
	// Ensure dir has a trailing path separator
	if dir != "" && dir[len(dir)-1] != '/' {
		dir += "/"
	}
	// Find paths
	var paths []string
	for path := range p.resourcesForPath {
		if strings.HasPrefix(path, dir) {
			paths = append(paths, path)
		}
	}
	return paths
}

// reload resets the parser's state and then parses the entire project.
func (p *Parser) reload(ctx context.Context) error {
	// Reset state
	p.RillYAML = nil
	p.DotEnv = nil
	p.Resources = make(map[ResourceName]*Resource)
	p.Errors = nil
	p.resourceNamesForDataPaths = make(map[string][]ResourceName)
	p.resourcesForPath = make(map[string][]*Resource)
	p.resourcesForUnspecifiedRef = make(map[string][]*Resource)
	p.insertedResources = nil
	p.updatedResources = nil
	p.deletedResources = nil

	// Load entire repo
	files, err := p.Repo.ListRecursive(ctx, "**/*.{env,sql,yaml,yml}", true)
	if err != nil {
		return fmt.Errorf("could not list project files: %w", err)
	}

	// Build paths slice
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}

	// Parse all files
	err = p.parsePaths(ctx, paths)
	if err != nil {
		return err
	}

	// Infer unspecified refs for all inserted resources
	for _, r := range p.insertedResources {
		p.inferUnspecifiedRefs(r)
	}

	return nil
}

// reparseExceptRillYAML re-parses the indicated file paths, updating the Parser's state.
// It assumes that p.RillYAML is valid and does not need to be reloaded.
func (p *Parser) reparseExceptRillYAML(ctx context.Context, paths []string) (*Diff, error) {
	// The logic here is slightly tricky because the relationship between files and resources can vary:
	//
	// - Case 1: one file created one resource
	// - Case 2: one file created multiple resources
	// - Case 3: multiple files contributed to one resource (for example, "model.sql" and "model.yaml")
	//
	// The high-level approach is: We'll delete all existing resources *related* to the changed paths and (re)parse them.
	// Then at the end, we build a diff that treats any resource that was both "deleted" and "added" as "modified".
	// (Renames are not supported in the parser. It needs to be handled by the caller, since parser state alone is insufficient to detect it.)
	//
	// Another wrinkle is that we need to re-infer unspecified refs for:
	// - any resources pointing to a changed resource, and
	// - any resources with previously unmatched unspecified refs that may match a new resource.

	// Reset insertedResources and updatedResources on reparse (used to construct Diff)
	p.insertedResources = nil
	p.updatedResources = nil
	p.deletedResources = nil

	// Phase 1: Clear existing state related to the paths.
	// Identify all paths directly passed and paths indirectly related through resourcesForPath and Resource.Paths.
	// And delete all resources and parse errors related to those paths.
	var parsePaths []string           // Paths we should pass to parsePaths
	checkPaths := slices.Clone(paths) // Paths we should visit in the loop
	for _, perr := range p.Errors {   // Also re-check paths with external parse errors
		if perr.External {
			checkPaths = append(checkPaths, perr.FilePath)
		}
	}
	seenPaths := make(map[string]bool) // Paths already visited by the loop
	modifiedDotEnv := false            // whether .env file was modified
	for i := 0; i < len(checkPaths); i++ {
		// Don't check the same path twice
		path := normalizePath(checkPaths[i])
		if seenPaths[path] {
			continue
		}
		seenPaths[path] = true

		// Skip ignored paths
		if pathIsIgnored(path) {
			continue
		}

		// add resources corresponding to local data files
		resources, ok := p.resourceNamesForDataPaths[path]
		if ok {
			for _, resource := range resources {
				if res, ok := p.Resources[resource]; ok {
					checkPaths = append(checkPaths, res.Paths...)
				}
			}
			continue
		}

		isSQL := pathIsSQL(path)
		isYAML := pathIsYAML(path)
		isDotEnv := pathIsDotEnv(path)
		if !isSQL && !isYAML && !isDotEnv {
			continue
		}

		// Check if path is .env and clear it (so we can re-parse it)
		// Watcher sends multiple events for a single edit. We want to only restart the controller when .env actually changes.
		// So we check the new .env contents against the contents stored in parser state.
		if isDotEnv {
			oldDotEnv := p.DotEnv
			err := p.parseDotEnv(ctx, checkPaths[i])
			if err == nil {
				modifiedDotEnv = !maps.Equal(p.DotEnv, oldDotEnv)
			} else {
				// any error means .env is under change
				modifiedDotEnv = true
			}
			if !modifiedDotEnv {
				continue
			}
			p.DotEnv = nil
		}

		// If a file exists at path, add it to the parse list
		info, err := p.Repo.Stat(ctx, path)
		if err == nil {
			if info.IsDir {
				continue
			}
			parsePaths = append(parsePaths, path)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("unexpected file stat error: %w", err)
		}
		// NOTE: Continue even if the file has been deleted because it may have associated state we need to clear.

		// Since .sql and .yaml files provide context for each other, if one was modified, we need to reparse both.
		// For cases where a file was modified or deleted, the transitive check through resourcesForPath will already take of that.
		// But this ensures the check also happens for cases where a companion file was added.
		stem := pathStem(path)
		if isSQL {
			checkPaths = append(checkPaths, stem+".yaml", stem+".yml")
		} else if isYAML {
			checkPaths = append(checkPaths, stem+".sql")
		}

		// Remove all resources derived from this path, and add any related paths to the check list
		rs := slices.Clone(p.resourcesForPath[path]) // Use Clone because deleteResource mutates resourcesForPath
		for _, resource := range rs {
			p.deleteResource(resource)

			// Make sure we-reparse all paths that contributed to the deleted resource.
			checkPaths = append(checkPaths, resource.Paths...)
		}

		// Remove all parse errors related to this path
		// (We can't mutate p.Errors while iterating over it, hence the nested loop here.)
		for {
			found := false
			for i, err := range p.Errors {
				if err.FilePath == path {
					found = true
					p.Errors = slices.Delete(p.Errors, i, i+1)
					break
				}
			}
			if !found {
				break
			}
		}
	}

	// Phase 2: Parse (or reparse) the related paths, adding back resources
	err := p.parsePaths(ctx, parsePaths)
	if err != nil {
		return nil, err
	}

	// Infer or re-infer refs for...
	inferRefsSeen := make(map[ResourceName]bool)
	// ... all inserted resources
	for _, r := range p.insertedResources {
		inferRefsSeen[r.Name.Normalized()] = true
		p.inferUnspecifiedRefs(r)
	}
	// ... all updated resources
	for _, r := range p.updatedResources {
		inferRefsSeen[r.Name.Normalized()] = true
		p.inferUnspecifiedRefs(r)
	}
	// ... any unchanged resource that may have an unspecified ref to a deleted resource
	for _, r1 := range p.deletedResources {
		for _, r2 := range p.resourcesForUnspecifiedRef[strings.ToLower(r1.Name.Name)] {
			n := r2.Name.Normalized()
			if !inferRefsSeen[n] {
				inferRefsSeen[n] = true
				p.inferUnspecifiedRefs(r2)
				p.updatedResources = append(p.updatedResources, r2) // inferRefsSeen ensures it's not already in insertedResources or updatedResources
			}
		}
	}
	// ... any unchanged resource that might have an unspecified ref (previously unmatched) that now matches a newly inserted resource
	for _, r1 := range p.insertedResources {
		for _, r2 := range p.resourcesForUnspecifiedRef[strings.ToLower(r1.Name.Name)] {
			n := r2.Name.Normalized()
			if !inferRefsSeen[n] {
				inferRefsSeen[n] = true
				p.inferUnspecifiedRefs(r2)
				p.updatedResources = append(p.updatedResources, r2) // inferRefsSeen ensures it's not already in insertedResources or updatedResources
			}
		}
	}

	// Phase 3: Build the diff using p.insertedResources, p.updatedResources and p.deletedResources
	diff := &Diff{
		ModifiedDotEnv: modifiedDotEnv,
	}
	for _, resource := range p.insertedResources {
		addedBack := false
		for _, deleted := range p.deletedResources {
			if resource.Name.Normalized() == deleted.Name.Normalized() {
				addedBack = true
				break
			}
		}
		if addedBack {
			diff.Modified = append(diff.Modified, resource.Name)
		} else {
			diff.Added = append(diff.Added, resource.Name)
		}
	}
	for _, resource := range p.updatedResources {
		diff.Modified = append(diff.Modified, resource.Name)
	}
	for _, deleted := range p.deletedResources {
		if p.Resources[deleted.Name.Normalized()] == nil {
			diff.Deleted = append(diff.Deleted, deleted.Name)
		}
	}

	return diff, nil
}

// parsePaths is the internal entrypoint for parsing a list of paths.
// It assumes that the caller has already checked that the paths exist.
// It also assumes that the caller has already removed any previous resources related to the paths,
// enabling parsePaths to insert changed resources without conflicts.
func (p *Parser) parsePaths(ctx context.Context, paths []string) error {
	// Check limits
	if len(paths) > maxFiles {
		return fmt.Errorf("project exceeds file limit of %d", maxFiles)
	}

	// Sort paths such that a) we always parse rill.yaml first (to pick up defaults),
	// and b) we align files with the same name but different extensions next to each other.
	slices.SortFunc(paths, func(a, b string) int {
		if pathIsRillYAML(a) {
			return -1
		}
		if pathIsRillYAML(b) {
			return 1
		}
		return strings.Compare(a, b)
	})

	// Iterate over the sorted paths, processing all paths with the same stem at once (stem = path without extension).
	sawRillYAML := false
	for i := 0; i < len(paths); {
		// Handle rill.yaml and .env separately
		path := paths[i]
		if pathIsRillYAML(path) {
			err := p.parseRillYAML(ctx, path)
			if err != nil {
				p.addParseError(path, err, false)
			}
			sawRillYAML = true
			i++
			continue
		} else if pathIsDotEnv(path) {
			err := p.parseDotEnv(ctx, path)
			if err != nil {
				p.addParseError(path, err, false)
			}
			i++
			continue
		}

		// Identify the range of paths with the same stem as paths[i]
		j := i + 1
		pathStemI := pathStem(paths[i])
		for j < len(paths) && pathStemI == pathStem(paths[j]) {
			j++
		}

		// Parse the paths with the same stem
		err := p.parseStemPaths(ctx, paths[i:j])
		if err != nil {
			return err
		}

		// Advance i to the next stem
		i = j
	}

	// If we didn't encounter rill.yaml (in this run or a previous run), that's a breaking error
	if !sawRillYAML && p.RillYAML == nil {
		p.addParseError("/rill.yaml", errors.New("rill.yaml not found"), false)
	}

	// As a special case, we need to check that there aren't any sources and models with the same name.
	// NOTE 1: We always attach the error to the model when there's a collision.
	// NOTE 2: Using a map since the two-way check (necessary for reparses) may match the same resource twice.
	modelsWithNameErrs := make(map[ResourceName]string)
	for _, r := range p.insertedResources {
		if r.Name.Kind == ResourceKindSource {
			n := ResourceName{Kind: ResourceKindModel, Name: r.Name.Name}.Normalized()
			if _, ok := p.Resources[n]; ok {
				modelsWithNameErrs[n] = r.Name.Name
			}
		} else if r.Name.Kind == ResourceKindModel {
			n := ResourceName{Kind: ResourceKindSource, Name: r.Name.Name}.Normalized()
			if r2, ok := p.Resources[n]; ok {
				modelsWithNameErrs[r.Name.Normalized()] = r2.Name.Name
			}
		}
	}
	for n, s := range modelsWithNameErrs {
		// NOTE: Setting external=true because removing the source should restore the model.
		p.replaceResourceWithError(n, fmt.Errorf("model name collides with source %q", s), true)
	}

	return nil
}

// parseStem parses a set of paths with the same stem (path without extension).
func (p *Parser) parseStemPaths(ctx context.Context, paths []string) error {
	// Load YAML and SQL contents
	var yaml, yamlPath, sql, sqlPath string
	var found bool
	for _, path := range paths {
		// Load contents
		data, err := p.Repo.Get(ctx, path)
		if err != nil {
			if os.IsNotExist(err) {
				// This is a dirty parse where a file disappeared during parsing.
				// But due to the clear-and-rebuild behavior, we can safely continue parsing.
				continue
			}
			return err
		}

		// Check size
		if len(data) > maxFileSize {
			p.Errors = append(p.Errors, &runtimev1.ParseError{
				Message:  fmt.Sprintf("size %d bytes exceeds max size of %d bytes", len(data), maxFileSize),
				FilePath: path,
			})
			continue
		}

		// Assign to correct variable
		if strings.HasSuffix(path, ".sql") {
			sql = data
			sqlPath = path
			found = true
			continue
		}
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			if yaml != "" {
				// Means there was both a .yaml and .yml file. We don't allow that!
				p.Errors = append(p.Errors, &runtimev1.ParseError{
					Message:  "skipping file because another YAML file has already been parsed for this path stem",
					FilePath: path,
				})
				continue
			}
			yaml = data
			yamlPath = path
			found = true
			continue
		}
		// The unhandled case should never happen, just being defensive
	}

	// There's a few cases above where we don't find YAML or SQL contents. It's fine to just skip the file in those cases.
	if !found {
		return nil
	}

	// Parse the SQL/YAML file pair to a Node, then parse the Node to p.Resources.
	node, err := p.parseStem(paths, yamlPath, yaml, sqlPath, sql)
	if err == nil {
		err = p.parseNode(ctx, node)
	}

	// Spread error across the node's paths (YAML and/or SQL files)
	if err != nil {
		var pathErr pathError
		if errors.As(err, &pathErr) {
			// If there's an error in either of the YAML or SQL files, we attach a "skipped" error to the other file as well.
			for _, path := range paths {
				if path == pathErr.path {
					p.addParseError(path, err, false)
				} else {
					p.addParseError(path, fmt.Errorf("skipping file due to error in companion SQL/YAML file"), false)
				}
			}
		} else {
			// Not a path error – we add the error to all paths
			for _, path := range paths {
				p.addParseError(path, err, false)
			}
		}
	}

	return nil
}

// inferUnspecifiedRefs populates r.Refs with a) all explicit refs from r.rawRefs, and b) any implicit refs that we can infer from context.
// An implicit ref is one where the kind is unspecified. They are common when extracted from SQL.
// For example, if a model contains "SELECT * FROM foo", we add "foo" to r.rawRefs, and need to infer whether "foo" is a source or a model.
//
// If an unspecified ref can't be matched to another resource, it is not added to r.Refs.
// That allows, for example, a model like "SELECT * FROM foo", to parse successfully even when no other model or source is named "foo" exists.
// This is necessary to support referencing existing tables in a database. Errors for such cases will be thrown from the downstream reconciliation logic instead.
// We may want to revisit this handling in the future.
func (p *Parser) inferUnspecifiedRefs(r *Resource) {
	var refs []ResourceName
	for _, ref := range r.rawRefs {
		if ref.Kind != ResourceKindUnspecified {
			refs = append(refs, ref)
			continue
		}

		// Rule 1: If it's a model and there's a source with that name, use it
		if r.Name.Kind == ResourceKindModel {
			n := ResourceName{Kind: ResourceKindSource, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				refs = append(refs, n)
				continue
			}
		}

		// Rule 2: If it's a metrics view and there's a model or source with that name, use it
		if r.Name.Kind == ResourceKindMetricsView {
			n := ResourceName{Kind: ResourceKindModel, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				refs = append(refs, n)
				continue
			}
			n = ResourceName{Kind: ResourceKindSource, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				refs = append(refs, n)
				continue
			}
		}

		// Rule 3: If it's a model and there's another model with that name, use it
		if r.Name.Kind == ResourceKindModel {
			n := ResourceName{Kind: r.Name.Kind, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				// NOTE: Not skipping self-references because we'd rather add them and error during cyclic dependency check
				refs = append(refs, n)
				continue
			}
		}

		// Rule 4: Skip it
	}

	slices.SortFunc(refs, func(a, b ResourceName) int {
		if a.Kind < b.Kind {
			return -1
		}
		if a.Kind > b.Kind {
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})

	r.Refs = refs
}

// insertDryRun returns an error if calling insertResource with the given resource name would fail.
func (p *Parser) insertDryRun(kind ResourceKind, name string) error {
	rn := ResourceName{Kind: kind, Name: name}
	_, ok := p.Resources[rn.Normalized()]
	if ok {
		return externalError{err: fmt.Errorf("name collision: another resource of type %q is also named %q", rn.Kind, rn.Name)}
	}
	return nil
}

// insertResource inserts a resource in the parser's internal state.
// After calling insertResource, the caller can directly modify the returned resource's spec.
func (p *Parser) insertResource(kind ResourceKind, name string, paths []string, refs ...ResourceName) (*Resource, error) {
	// Create the resource if not already present (ensures the spec for its kind is never nil)
	rn := ResourceName{Kind: kind, Name: name}
	_, ok := p.Resources[rn.Normalized()]
	if ok {
		return nil, externalError{err: fmt.Errorf("name collision: another resource of type %q is also named %q", rn.Kind, rn.Name)}
	}

	// Dedupe refs (it's not guaranteed by the upstream logic).
	// Doing a simple dedupe because there usually aren't many refs.
	var dedupedRefs []ResourceName
	for _, ref := range refs {
		found := false
		for _, existing := range dedupedRefs {
			if ref.Normalized() == existing.Normalized() {
				found = true
				break
			}
		}
		if !found {
			dedupedRefs = append(dedupedRefs, ref)
		}
	}
	refs = dedupedRefs

	// Create new resource
	r := &Resource{
		Name:    rn,
		Paths:   paths,
		rawRefs: refs,
	}
	switch kind {
	case ResourceKindModel:
		r.ModelSpec = &runtimev1.ModelSpec{}
	case ResourceKindMetricsView:
		r.MetricsViewSpec = &runtimev1.MetricsViewSpec{}
	case ResourceKindExplore:
		r.ExploreSpec = &runtimev1.ExploreSpec{}
	case ResourceKindMigration:
		r.MigrationSpec = &runtimev1.MigrationSpec{}
	case ResourceKindReport:
		r.ReportSpec = &runtimev1.ReportSpec{}
	case ResourceKindAlert:
		r.AlertSpec = &runtimev1.AlertSpec{}
	case ResourceKindTheme:
		r.ThemeSpec = &runtimev1.ThemeSpec{}
	case ResourceKindComponent:
		r.ComponentSpec = &runtimev1.ComponentSpec{}
	case ResourceKindCanvas:
		r.CanvasSpec = &runtimev1.CanvasSpec{}
	case ResourceKindAPI:
		r.APISpec = &runtimev1.APISpec{}
	case ResourceKindConnector:
		r.ConnectorSpec = &runtimev1.ConnectorSpec{}
	default:
		panic(fmt.Errorf("unexpected resource type: %s", kind.String()))
	}

	// Track it
	p.Resources[rn.Normalized()] = r
	p.insertedResources = append(p.insertedResources, r)

	// Index paths
	for _, path := range paths {
		p.resourcesForPath[path] = append(p.resourcesForPath[path], r)
	}

	// Index unspecified refs in p.resourcesForUnspecifiedRef
	for _, ref := range refs {
		if ref.Kind == ResourceKindUnspecified {
			n := strings.ToLower(ref.Name)
			p.resourcesForUnspecifiedRef[n] = append(p.resourcesForUnspecifiedRef[n], r)
		}
	}

	return r, nil
}

// deleteResource removes a resource from p.Resources as well as all internal indexes.
func (p *Parser) deleteResource(r *Resource) {
	// Remove from p.Resources
	delete(p.Resources, r.Name.Normalized())

	// Remove from p.insertedResources
	foundInInserted := false
	idx := slices.Index(p.insertedResources, r)
	if idx >= 0 {
		p.insertedResources = slices.Delete(p.insertedResources, idx, idx+1)
		foundInInserted = true
	}

	// Remove from p.updatedResources
	if !foundInInserted {
		idx = slices.Index(p.updatedResources, r)
		if idx >= 0 {
			p.updatedResources = slices.Delete(p.updatedResources, idx, idx+1)
		}
	}

	// Remove from p.resourcesForPath
	for _, path := range r.Paths {
		rs := p.resourcesForPath[path]
		idx := slices.Index(rs, r)
		if idx < 0 {
			panic(fmt.Errorf("resource %q not found in resourcesForPath", r.Name))
		}
		if len(rs) == 1 {
			delete(p.resourcesForPath, path)
		} else {
			p.resourcesForPath[path] = slices.Delete(rs, idx, idx+1)
		}
	}

	// Remove pointers indexed in resourcesForUnspecifiedRef
	for _, ref := range r.rawRefs {
		if ref.Kind != ResourceKindUnspecified {
			continue
		}
		n := strings.ToLower(ref.Name)
		rs := p.resourcesForUnspecifiedRef[n]
		idx := slices.Index(rs, r)
		if idx < 0 {
			panic(fmt.Errorf("resource %q not found in resourcesForUnspecifiedRef for ref %q", r.Name, ref.Name))
		}
		if len(rs) == 1 {
			delete(p.resourcesForUnspecifiedRef, n)
		} else {
			p.resourcesForUnspecifiedRef[n] = slices.Delete(rs, idx, idx+1)
		}
	}

	// Remove from p.resourceNamesForDataPaths
	for path, resources := range p.resourceNamesForDataPaths {
		idx := slices.Index(resources, r.Name.Normalized())
		if idx < 0 {
			continue
		}
		if len(resources) == 1 {
			delete(p.resourceNamesForDataPaths, path)
		} else {
			p.resourceNamesForDataPaths[path] = slices.Delete(resources, idx, idx+1)
		}
	}

	// Track in deleted resources (unless it was in insertedResources, in which case it's not a real deletion)
	if !foundInInserted {
		p.deletedResources = append(p.deletedResources, r)
	}
}

// replaceResourceWithError removes a resource from the parser's internal state and adds a parse error for its paths instead.
func (p *Parser) replaceResourceWithError(n ResourceName, err error, external bool) {
	r := p.Resources[n.Normalized()]
	p.deleteResource(r)
	for _, path := range r.Paths {
		p.addParseError(path, err, external)
	}
}

// addParseError adds a parse error to the p.Errors
func (p *Parser) addParseError(path string, err error, external bool) {
	var loc *runtimev1.CharLocation
	var locErr locationError
	if errors.As(err, &locErr) {
		loc = locErr.location
	}

	var extErr externalError
	if errors.As(err, &extErr) {
		external = true
	}

	p.Errors = append(p.Errors, &runtimev1.ParseError{
		Message:       err.Error(),
		FilePath:      path,
		StartLocation: loc,
		External:      external,
	})
}

// driverForConnector resolves a connector name to a connector driver.
// It should not be invoked until after rill.yaml has been parsed.
func (p *Parser) driverForConnector(name string) (string, drivers.Driver, error) {
	// Search rill.yaml and Connector resources for the connector's driver
	var driver string
	if p.RillYAML != nil {
		for _, c := range p.RillYAML.Connectors {
			if c.Name == name {
				driver = c.Type
				break
			}
		}
	}
	for _, c := range p.Resources {
		if c.ConnectorSpec != nil && c.Name.Name == name {
			driver = c.ConnectorSpec.Driver
			break
		}
	}

	// If we found a driver, return it
	if driver != "" {
		connector, ok := drivers.Connectors[driver]
		if !ok {
			return "", nil, fmt.Errorf("unknown connector type %q", driver)
		}
		return driver, connector, nil
	}

	// If the name matches a known driver, we consider it an implicitly defined connector
	connector, ok := drivers.Connectors[name]
	if !ok {
		return "", nil, fmt.Errorf("unknown connector %q", name)
	}
	return name, connector, nil
}

// defaultOLAPConnector resolves the project's default OLAP connector.
// It should not be invoked until after rill.yaml has been parsed.
func (p *Parser) defaultOLAPConnector() string {
	if p.RillYAML != nil && p.RillYAML.OLAPConnector != "" {
		return p.RillYAML.OLAPConnector
	}
	return p.DefaultOLAPConnector
}

// isDev returns true if the parser's instance's environment is "dev".
// Usually this means it's running on localhost with "rill start".
func (p *Parser) isDev() bool {
	return strings.EqualFold(p.Environment, "dev")
}

// pathIsSQL returns true if the path is a SQL file
func pathIsSQL(path string) bool {
	return strings.HasSuffix(path, ".sql")
}

// pathIsYAML returns true if the path is a YAML file
func pathIsYAML(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

// pathIsRillYAML returns true if the path is rill.yaml
func pathIsRillYAML(path string) bool {
	return path == "/rill.yaml" || path == "/rill.yml"
}

// pathIsDotEnv returns true if the path is .env
func pathIsDotEnv(path string) bool {
	return path == "/.env"
}

// pathIsIgnored returns true if the path should be ignored by the parser.
func pathIsIgnored(p string) bool {
	for _, prefix := range ignorePathPrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// normalizePath normalizes a user-provided path to the format returned from ListRecursive.
// TODO: Change this once ListRecursive returns paths without leading slash.
func normalizePath(path string) string {
	if path != "" && path[0] != '/' {
		return "/" + path
	}
	return path
}

// pathStem returns a slice of the path without the final file extension.
// If the path does not contain a file extension, the entire path is returned
func pathStem(path string) string {
	i := strings.LastIndexByte(path, '.')
	if i == -1 {
		return path
	}
	return path[:i]
}

// locationError wraps an error with source file character location information
type locationError struct {
	err      error
	location *runtimev1.CharLocation
}

func (e locationError) Error() string {
	return e.err.Error()
}

func (e locationError) Unwrap() error {
	return e.err
}

// pathError wraps an error with source file path information
type pathError struct {
	err  error
	path string
}

func (e pathError) Error() string {
	return e.err.Error()
}

func (e pathError) Unwrap() error {
	return e.err
}

// externalError wraps an error that should be emitted as a parse error with external=true.
// This means the error is not with the file itself, but with another file that interferes with it (e.g. name collision).
type externalError struct {
	err error
}

func (e externalError) Error() string {
	return e.err.Error()
}

func (e externalError) Unwrap() error {
	return e.err
}

// yamlErrLineRegexp matches the line number in a YAML error
var yamlErrLineRegexp = regexp.MustCompile(`^yaml: line (\d+):`)

// newYAMLError wraps a YAML error, extracting line number information if available
func newYAMLError(err error) error {
	res := yamlErrLineRegexp.FindStringSubmatch(err.Error())
	if len(res) != 2 {
		return err
	}

	line, err2 := strconv.ParseUint(res[1], 10, 32)
	if err2 != nil {
		return err
	}

	return locationError{
		err: err,
		location: &runtimev1.CharLocation{
			Line: uint32(line),
		},
	}
}

// duckDBErrLineRegexp matches the line number in a DuckDB parser error
var duckDBErrLineRegexp = regexp.MustCompile(`\nLINE (\d+):`)

// newDuckDBError wraps a DuckDB parser error, extracting line number information if available
func newDuckDBError(err error) error {
	res := duckDBErrLineRegexp.FindStringSubmatch(err.Error())
	if len(res) != 2 {
		return err
	}

	line, err2 := strconv.ParseUint(res[1], 10, 32)
	if err2 != nil {
		return err
	}

	return locationError{
		err: err,
		location: &runtimev1.CharLocation{
			Line: uint32(line),
		},
	}
}
