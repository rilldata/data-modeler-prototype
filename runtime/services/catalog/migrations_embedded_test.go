package catalog_test

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/services/catalog"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

var EmbeddedSourceName = "local_file_data_AdBids_csv"
var EmbeddedSourcePath = "data/AdBids.csv"
var EmbeddedGzSourceName = "local_file_data_AdBids_csv_gz"
var EmbeddedGzSourcePath = "data/AdBids.csv.gz"
var ImpEmbeddedSourceName = "local_file_data_AdImpressions_csv"
var ImpEmbeddedSourcePath = "data/AdImpressions.csv"
var AdBidsNewModeName = "AdBids_new_model"
var AdBidsNewModelPath = "/models/AdBids_new_model.sql"

func TestEmbeddedSourcesHappyPath(t *testing.T) {
	s, dir := initBasicService(t)

	testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")

	addEmbeddedModel(t, s)
	addEmbeddedNewModel(t, s)
	testutils.AssertTable(t, s, "AdBids_new_model", AdBidsNewModelPath)

	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	// no errors when reconcile is run later
	testutils.AssertMigration(t, result, 0, 0, 0, 0, []string{})
	require.NoError(t, err)

	// delete on of the models
	err = os.Remove(path.Join(dir, AdBidsNewModelPath))
	time.Sleep(10 * time.Millisecond)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 0, 1, 1, []string{AdBidsNewModelPath, EmbeddedSourcePath})
	testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)

	// delete the other model
	err = os.Remove(path.Join(dir, AdBidsModelRepoPath))
	time.Sleep(10 * time.Millisecond)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(
		t,
		result,
		1,
		0,
		0,
		2,
		[]string{AdBidsModelRepoPath, AdBidsDashboardRepoPath, EmbeddedSourcePath},
	)
	testutils.AssertTableAbsence(t, s, EmbeddedSourceName)
}

func TestEmbeddedSourcesQueryChanging(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			testutils.CopyFileToData(t, dir, AdBidsCsvGzPath, "AdBids.csv.gz")

			addEmbeddedModel(t, s)
			addEmbeddedNewModel(t, s)

			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`select id, timestamp, publisher from "data/AdBids.csv.gz"`,
				AdBidsNewModelPath,
			)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath, EmbeddedGzSourcePath})
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.Equal(t, []string{"adbids_model"}, adBidsEntry.Embeds)
			adBidsGzEntry := testutils.AssertTable(t, s, EmbeddedGzSourceName, EmbeddedGzSourcePath)
			require.Equal(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsGzEntry.Embeds)
			testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)

			sc, _ := copyService(t, s)
			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`select id, timestamp, publisher from "data/AdBids.csv"`,
				AdBidsNewModelPath,
			)
			result, err = sc.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 2, 1, []string{AdBidsNewModelPath, EmbeddedSourcePath, EmbeddedGzSourcePath})
			adBidsEntry = testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.Equal(t, []string{"adbids_model", strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Embeds)
			testutils.AssertTableAbsence(t, s, EmbeddedGzSourceName)
		})
	}
}

func TestEmbeddedMultipleSources(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			testutils.CopyFileToData(t, dir, AdImpressionsCsvPath, "AdImpressions.csv")

			// create a model with 2 embedded sources, one repeated twice
			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`with
    NewYorkImpressions as (
        select imp.id, imp.city, imp.country from "data/AdImpressions.csv" imp where imp.city = 'NewYork'
    )
    select count(*) as impressions, bid.publisher, bid.domain, imp.city, imp.country
    from "data/AdBids.csv" bid join "data/AdImpressions.csv" imp on bid.id = imp.id
    group by bid.publisher, bid.domain, imp.city, imp.country`,
				AdBidsNewModelPath,
			)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 0, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath, ImpEmbeddedSourcePath})
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.Equal(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Embeds)
			require.Equal(t, 1, adBidsEntry.Links)
			adImpEntry := testutils.AssertTable(t, s, ImpEmbeddedSourceName, ImpEmbeddedSourcePath)
			require.Equal(t, []string{strings.ToLower(AdBidsNewModeName)}, adImpEntry.Embeds)
			require.Equal(t, 1, adImpEntry.Links)
			modelEntry := testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)
			require.ElementsMatch(t, []string{strings.ToLower(EmbeddedSourceName), strings.ToLower(ImpEmbeddedSourceName)}, modelEntry.Embeds)

			// update the model to have embedded sources without repetitions
			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`select count(*) as impressions, bid.publisher, bid.domain, imp.city, imp.country
    from "data/AdBids.csv" bid join "data/AdImpressions.csv" imp on bid.id = imp.id
    group by bid.publisher, bid.domain, imp.city, imp.country`,
				AdBidsNewModelPath,
			)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsNewModelPath})
			adBidsEntry = testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.Equal(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Embeds)
			require.Equal(t, 1, adBidsEntry.Links)
			adImpEntry = testutils.AssertTable(t, s, ImpEmbeddedSourceName, ImpEmbeddedSourcePath)
			require.Equal(t, []string{strings.ToLower(AdBidsNewModeName)}, adImpEntry.Embeds)
			require.Equal(t, 1, adImpEntry.Links)
			modelEntry = testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)
			require.ElementsMatch(t, []string{strings.ToLower(EmbeddedSourceName), strings.ToLower(ImpEmbeddedSourceName)}, modelEntry.Embeds)
		})
	}
}

func TestEmbeddedSourceOnNewService(t *testing.T) {
	s, dir := initBasicService(t)

	testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
	addEmbeddedModel(t, s)

	sc, result := copyService(t, s)
	// no updates other than when a new service is started
	// dashboards don't have equals check implemented right now. hence it is updated here
	testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsDashboardRepoPath})

	addEmbeddedNewModel(t, s)

	// change one model back to use AdBids from embedding the source
	testutils.CreateModel(
		t,
		sc,
		"AdBids_model",
		`select id, timestamp, publisher, domain, bid_price from AdBids`,
		AdBidsModelRepoPath,
	)
	// delete the other model embedding the source
	err := os.Remove(path.Join(dir, AdBidsNewModelPath))
	require.NoError(t, err)
	// create another copy
	sc, result = copyService(t, s)
	testutils.AssertMigration(
		t,
		result,
		0,
		0,
		3,
		1,
		[]string{EmbeddedSourcePath, AdBidsModelRepoPath, AdBidsDashboardRepoPath, AdBidsNewModelPath},
	)
}

func TestEmbeddingModelRename(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsModelRepoPath, AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			addEmbeddedModel(t, s)

			testutils.RenameFile(t, dir, AdBidsModelRepoPath, AdBidsNewModelPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(
				t,
				result,
				1,
				0,
				2,
				0,
				[]string{AdBidsDashboardRepoPath, EmbeddedSourcePath, AdBidsNewModelPath},
			)
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.Equal(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Embeds)
			require.Equal(t, 1, adBidsEntry.Links)
		})
	}
}

func TestEmbeddedSourceRefresh(t *testing.T) {
	s, dir := initBasicService(t)

	testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
	addEmbeddedModel(t, s)

	testutils.CopyFileToData(t, dir, AdImpressionsCsvPath, "AdBids.csv")
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{
		ChangedPaths: []string{EmbeddedSourcePath},
		ForcedPaths:  []string{EmbeddedSourcePath},
	})
	require.NoError(t, err)
	// refreshing the embedded source and replacing with different file caused errors
	// the model depended on column not present in the new file
	testutils.AssertMigration(
		t,
		result,
		2,
		0,
		1,
		0,
		[]string{EmbeddedSourcePath, AdBidsModelRepoPath, AdBidsDashboardRepoPath},
	)
}

func addEmbeddedModel(t *testing.T, s *catalog.Service) {
	testutils.CreateModel(
		t,
		s,
		"AdBids_model",
		`select id, timestamp, publisher, domain, bid_price from "data/AdBids.csv"`,
		AdBidsModelRepoPath,
	)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 2, 0, []string{AdBidsModelRepoPath, AdBidsDashboardRepoPath, EmbeddedSourcePath})
	testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
	testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
}

func addEmbeddedNewModel(t *testing.T, s *catalog.Service) {
	testutils.CreateModel(
		t,
		s,
		AdBidsNewModeName,
		`select id, timestamp, publisher from "data/AdBids.csv"`,
		AdBidsNewModelPath,
	)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 1, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath})
	testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
	testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)
}

func copyService(t *testing.T, s *catalog.Service) (*catalog.Service, *catalog.ReconcileResult) {
	sc := catalog.NewService(s.Catalog, s.Repo, s.Olap, s.InstID, nil)
	result, err := sc.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	return sc, result
}
