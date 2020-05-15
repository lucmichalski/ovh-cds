package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ovh/cds/engine/api/application"
	"github.com/ovh/cds/engine/api/bootstrap"
	"github.com/ovh/cds/engine/api/event"
	"github.com/ovh/cds/engine/api/pipeline"
	"github.com/ovh/cds/engine/api/project"
	"github.com/ovh/cds/engine/api/test"
	"github.com/ovh/cds/engine/api/test/assets"
	"github.com/ovh/cds/engine/api/workflow"
	"github.com/ovh/cds/sdk"
)

func TestLoadByNameAsAdmin(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	_ = event.Initialize(context.Background(), db, cache)
	key := sdk.RandomString(10)
	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name: "my-app",
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	actual, err := application.LoadByProjectIDAndName(context.TODO(), db, proj.ID, "my-app")
	require.NoError(t, err)

	assert.Equal(t, app.Name, actual.Name)
	assert.Equal(t, proj.ID, actual.ProjectID)
}

func TestLoadByNameAsUser(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)
	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name: "my-app",
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	_, _ = assets.InsertLambdaUser(t, db, &proj.ProjectGroups[0].Group)

	actual, err := application.LoadByProjectIDAndName(context.TODO(), db, proj.ID, "my-app")
	require.NoError(t, err)
	assert.Equal(t, app.Name, actual.Name)
	assert.Equal(t, proj.ID, actual.ProjectID)
}

func TestLoadByIDAsAdmin(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)
	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name: "my-app",
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	actual, err := application.LoadByID(context.TODO(), db, app.ID)
	require.NoError(t, err)
	assert.Equal(t, app.Name, actual.Name)
	assert.Equal(t, proj.ID, actual.ProjectID)
}

func TestLoadByIDAsUser(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)

	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name: "my-app",
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	_, _ = assets.InsertLambdaUser(t, db, &proj.ProjectGroups[0].Group)

	actual, err := application.LoadByID(context.TODO(), db, app.ID)
	assert.NoError(t, err)
	assert.Equal(t, app.Name, actual.Name)
	assert.Equal(t, proj.ID, actual.ProjectID)
}

func TestLoadAllAsAdmin(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)
	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name: "my-app",
		Metadata: sdk.Metadata{
			"bla": "bla",
		},
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	app2 := sdk.Application{
		Name: "my-app2",
		Metadata: sdk.Metadata{
			"bla": "bla",
		},
	}
	require.NoError(t, application.Insert(db, proj.ID, &app2))

	actual, err := application.LoadAll(context.TODO(), db, proj.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(actual))

	for _, a := range actual {
		assert.EqualValues(t, app.Metadata, a.Metadata)
		assert.EqualValues(t, app2.Metadata, a.Metadata)
	}
}

func TestLoadAllAsUser(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)
	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name: "my-app",
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	app2 := sdk.Application{
		Name: "my-app2",
	}
	require.NoError(t, application.Insert(db, proj.ID, &app2))

	_, _ = assets.InsertLambdaUser(t, db, &proj.ProjectGroups[0].Group)

	actual, err := application.LoadAll(context.TODO(), db, proj.ID)
	test.NoError(t, err)
	assert.Equal(t, 2, len(actual))
}

func TestLoadByWorkflowID(t *testing.T) {
	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)

	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := sdk.Application{
		Name:      "my-app",
		ProjectID: proj.ID,
	}
	require.NoError(t, application.Insert(db, proj.ID, &app))

	pip := sdk.Pipeline{
		ProjectID:  proj.ID,
		ProjectKey: proj.Key,
		Name:       "pip1",
	}
	require.NoError(t, pipeline.InsertPipeline(db, &pip))

	w := sdk.Workflow{
		Name:       "test_1",
		ProjectID:  proj.ID,
		ProjectKey: proj.Key,
		WorkflowData: sdk.WorkflowData{
			Node: sdk.Node{
				Type: sdk.NodeTypePipeline,
				Context: &sdk.NodeContext{
					PipelineID:    pip.ID,
					ApplicationID: app.ID,
				},
			},
		},
	}
	require.NoError(t, workflow.RenameNode(context.TODO(), db, &w))

	proj, _ = project.LoadByID(db, proj.ID, project.LoadOptions.WithApplications, project.LoadOptions.WithPipelines, project.LoadOptions.WithEnvironments, project.LoadOptions.WithGroups)

	require.NoError(t, workflow.Insert(context.TODO(), db, cache, *proj, &w))

	actuals, err := application.LoadByWorkflowID(context.TODO(), db, w.ID)
	require.NoError(t, err)
	require.Len(t, actuals, 1)
	assert.Equal(t, app.Name, actuals[0].Name)
	assert.Equal(t, proj.ID, actuals[0].ProjectID)
}

func TestWithRepositoryStrategy(t *testing.T) {

	db, cache, end := test.SetupPG(t, bootstrap.InitiliazeDB)
	defer end()
	key := sdk.RandomString(10)

	proj := assets.InsertTestProject(t, db, cache, key, key)
	app := &sdk.Application{
		Name:       "my-app",
		ProjectKey: proj.Key,
		ProjectID:  proj.ID,
	}
	require.NoError(t, application.Insert(db, proj.ID, app))

	app.RepositoryStrategy = sdk.RepositoryStrategy{
		Branch:         "{{.git.branch}}",
		ConnectionType: "https",
		DefaultBranch:  "master",
		User:           "user",
		Password:       "password",
		SSHKeyContent:  "content",
	}

	require.NoError(t, application.Update(context.TODO(), db, app))
	require.Equal(t, "user", app.RepositoryStrategy.User)
	require.Equal(t, sdk.PasswordPlaceholder, app.RepositoryStrategy.Password)
	require.Equal(t, "", app.RepositoryStrategy.SSHKeyContent) // it depends on the connection type

	var err error

	app, err = application.LoadByID(context.TODO(), db, app.ID)
	require.NoError(t, err)
	app.RepositoryStrategy.Password = "password2"
	require.NoError(t, application.Update(context.TODO(), db, app))

	app, err = application.LoadByIDWithClearVCSStrategyPassword(context.TODO(), db, app.ID)
	require.NoError(t, err)
	require.Equal(t, "user", app.RepositoryStrategy.User)
	require.Equal(t, "password2", app.RepositoryStrategy.Password)
	require.Equal(t, "", app.RepositoryStrategy.SSHKeyContent) // it depends on the connection type

	app, err = application.LoadByID(context.TODO(), db, app.ID)
	require.NoError(t, err)
	require.Equal(t, "user", app.RepositoryStrategy.User)
	require.Equal(t, sdk.PasswordPlaceholder, app.RepositoryStrategy.Password)
	require.Equal(t, "", app.RepositoryStrategy.SSHKeyContent) // it depends on the connection type

	app.RepositoryStrategy.ConnectionType = "ssh"
	app.RepositoryStrategy.SSHKeyContent = "ssh_key"
	app.RepositoryStrategy.SSHKey = "ssh_key"

	require.NoError(t, application.Update(context.TODO(), db, app))
	require.Equal(t, "user", app.RepositoryStrategy.User)
	require.Equal(t, sdk.PasswordPlaceholder, app.RepositoryStrategy.Password)
	require.Equal(t, "", app.RepositoryStrategy.SSHKeyContent) // it depends on the connection type

}
