package cmd

import (
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/logging"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/pipeline"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/pom"
	jacocoutil "github.com/jenkins-x-apps/jx-app-jacoco/internal/util"
	"github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
)

const (
	sourceDirOptionName   = "source-dir"
	sourceURLOptionName   = "source-url"
	branchOptionName      = "branch-name"
	buildNumberOptionName = "build-number"
	contextOptionName     = "pipeline-context"
	pomXML                = "pom.xml"
)

var (
	configureCmdLogger = logging.AppLogger().WithFields(log.Fields{"command": "configure"})

	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "configures pom.xml and effective pipeline config",
		Run:   configure,
	}
	sourceDir   string
	sourceURL   string
	branchName  string
	buildNumber string
	context     string
)

func init() {
	configureCmd.Flags().StringVar(&sourceDir, sourceDirOptionName, "", "The directory containing the source for the Maven project for which to generate the JaCoCo report")
	_ = viper.BindPFlag(sourceDirOptionName, configureCmd.Flags().Lookup(sourceDirOptionName))
	viper.SetDefault(sourceDirOptionName, ".")

	configureCmd.Flags().StringVar(&sourceURL, sourceURLOptionName, "", "The git repo URL.")
	_ = viper.BindPFlag(sourceURLOptionName, configureCmd.Flags().Lookup(sourceURLOptionName))

	configureCmd.Flags().StringVar(&branchName, branchOptionName, "", "The git branch name.")
	_ = viper.BindPFlag(branchOptionName, configureCmd.Flags().Lookup(branchOptionName))

	configureCmd.Flags().StringVar(&buildNumber, buildNumberOptionName, "", "The build number.")
	_ = viper.BindPFlag(buildNumberOptionName, configureCmd.Flags().Lookup(buildNumberOptionName))

	configureCmd.Flags().StringVar(&context, contextOptionName, "", "The build context")
	_ = viper.BindPFlag(contextOptionName, configureCmd.Flags().Lookup(contextOptionName))
	viper.SetDefault(contextOptionName, "")
}

func configure(cmd *cobra.Command, args []string) {
	multiError := verify()
	if !multiError.Empty() {
		for _, err := range multiError.Errors {
			configureCmdLogger.Error(err.Error())
		}

		configureCmdLogger.Fatal("not all required parameters for this command execution specified")
	}

	sourceDir := viper.GetString(sourceDirOptionName)
	pomPath := filepath.Join(sourceDir, pomXML)
	if !jacocoutil.Exists(pomPath) {
		configureCmdLogger.Infof("nothing to do, no pom.xml in '%s'", sourceDir)
		return
	}

	pomConfigurator := pom.NewPomConfigurator(pomPath)
	err := pomConfigurator.ConfigurePom()
	if err != nil {
		configureCmdLogger.Fatal(errors.Wrap(err, "unable to enhance pom.xml with JaCoCo configuration"))
	}

	pipelineExtender := pipeline.NewMetaPipelineConfigurator(sourceDir, viper.GetString(contextOptionName))
	err = pipelineExtender.ConfigurePipeline()
	if err != nil {
		configureCmdLogger.Fatal(err)
	}
	allInOne()
}

func verify() jacocoutil.MultiError {
	validationErrors := jacocoutil.MultiError{}

	validationErrors.Collect(jacocoutil.IsNotEmpty(viper.GetString(sourceURLOptionName), sourceURLOptionName))
	validationErrors.Collect(jacocoutil.IsNotEmpty(viper.GetString(branchOptionName), branchOptionName))
	validationErrors.Collect(jacocoutil.IsNotEmpty(viper.GetString(buildNumberOptionName), buildNumberOptionName))

	return validationErrors
}

// this code is temporary until https://github.com/jenkins-x/jx/issues/4660 is resolved (HF)
func allInOne() {
	factory := clients.NewFactory()
	jxClient, ns, err := factory.CreateJXClient()
	if err != nil {
		configureCmdLogger.Fatal(err)
	}

	cmd := util.Command{
		Dir:  "/workspace/source",
		Name: "mvn",
		Args: []string{"verify"},
	}
	output, err := cmd.RunWithoutRetry()
	configureCmdLogger.Info(output)

	if err != nil {
		configureCmdLogger.Fatal(err)
	}

	createFact(jxClient, ns)
}
