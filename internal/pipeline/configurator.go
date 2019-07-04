package pipeline

import (
	"fmt"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/logging"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/util"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/version"
	"github.com/jenkins-x/jx/pkg/config"
	"github.com/jenkins-x/jx/pkg/jenkinsfile"
	"github.com/jenkins-x/jx/pkg/tekton/syntax"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

var (
	logger = logging.AppLogger().WithFields(log.Fields{"component": "meta-pipeline-extender"})
)

// MetaPipelineConfigurator is responsible to prepare the POM and pipeline for generating the JaCoCo report.
type MetaPipelineConfigurator struct {
	sourceDir string
	context   string
}

// NewMetaPipelineConfigurator creates a new instance of MetaPipelineConfigurator.
func NewMetaPipelineConfigurator(sourceDir string, context string) MetaPipelineConfigurator {
	return MetaPipelineConfigurator{
		sourceDir: sourceDir,
		context:   context,
	}
}

// ConfigurePipeline configures the Jenkins-X pipeline.
func (e *MetaPipelineConfigurator) ConfigurePipeline() error {
	log.Infof("processing directory '%s'", e.sourceDir)

	if !util.IsDirectory(e.sourceDir) {
		return errors.Errorf("specified directory '%s' does not exist", e.sourceDir)
	}

	effectiveConfig := "jenkins-x-effective.yml"
	if e.context != "" {
		effectiveConfig = fmt.Sprintf("jenkins-x-%s-effective.yml", e.context)
	}

	pipelineConfigPath := filepath.Join(e.sourceDir, effectiveConfig)
	if !util.Exists(pipelineConfigPath) {
		return errors.Errorf("unable to find effective pipeline config in '%s'", e.sourceDir)
	}

	projectConfig, err := config.LoadProjectConfigFile(pipelineConfigPath)
	err = e.addFactCreationStep(projectConfig)
	if err != nil {
		return errors.Wrap(err, "unable to enhance pipeline with JaCoCo configuration")
	}

	err = e.writeProjectConfig(projectConfig, pipelineConfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to write modified project config")
	}
	return nil
}

func (e *MetaPipelineConfigurator) addFactCreationStep(projectConfig *config.ProjectConfig) error {
	// insert us into all pipeline kinds for now
	for _, pipelineKind := range jenkinsfile.PipelineKinds {
		pipeline, err := projectConfig.PipelineConfig.Pipelines.GetPipeline(pipelineKind, false)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve pipeline for type %s", pipelineKind)
		}

		if pipeline == nil {
			continue
		}

		stages := pipeline.Pipeline.Stages

		createFactStage := e.createFactStep()

		lastStage := stages[len(stages)-1]
		steps := lastStage.Steps
		steps = append(steps, createFactStage)

		lastStage.Steps = steps
		stages[len(stages)-1] = lastStage
		pipeline.Pipeline.Stages = stages
	}
	return nil
}

func (e *MetaPipelineConfigurator) writeProjectConfig(projectConfig *config.ProjectConfig, pipelineConfigPath string) error {
	err := util.CopyFile(pipelineConfigPath, pipelineConfigPath+".jacoco.orig")
	if err != nil {
		return errors.Wrapf(err, "unable to backup '%s'", pipelineConfigPath)
	}

	logger.Infof("writing '%s'", pipelineConfigPath)
	err = projectConfig.SaveConfig(pipelineConfigPath)
	if err != nil {
		return errors.Wrapf(err, "unable to write '%s'", pipelineConfigPath)
	}
	return nil
}

func (e *MetaPipelineConfigurator) createFactStep() syntax.Step {
	step := syntax.Step{
		Name:      "jacoco-create-fact",
		Image:     version.GetFQImage(),
		Command:   "/jx-app-jacoco",
		Arguments: []string{"create"},
	}
	return step
}
