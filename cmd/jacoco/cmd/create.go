package cmd

import (
	"encoding/xml"
	"fmt"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/fact"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/logging"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/report"
	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"strings"
)

const (
	workspace = "workspace/source"
)

var (
	createCmdLogger = logging.AppLogger().WithFields(log.Fields{"command": "create"})

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "creates the JaCoCo Fact CRD",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("snafu")
		},
	}
)

func run(cmd *cobra.Command, args []string) {
	factory := clients.NewFactory()
	jxClient, ns, err := factory.CreateJXClient()
	if err != nil {
		createCmdLogger.Fatal(err)
	}

	if err != nil {
		createCmdLogger.Fatal(err)
	}

	createFact(jxClient, ns)
}

func createFact(jxClient versioned.Interface, ns string) {
	r := loadReport()

	act, err := findActivity(jxClient, ns)
	if err != nil {
		logger.Fatal(err)
	}

	f := fact.CreateFact(r, act, "")
	factsInterface := jxClient.JenkinsV1().Facts(ns)
	err = fact.StoreFact(f, factsInterface)
	if err != nil {
		logger.Errorf("error storing Fact %s: %s", f.Spec.Name, err)
	} else {
		logger.Infof("successfully stored JaCoCo fact '%s' for report from %s", f.Spec.Name, f.Spec.Original.URL)
	}
}

func loadReport() report.Report {
	data, _ := ioutil.ReadFile(filepath.Join(workspace, "target", "site", "jacoco", "jacoco.xml"))
	r := &report.Report{}
	_ = xml.Unmarshal([]byte(data), &r)
	return *r
}

func findActivity(jxClient versioned.Interface, ns string) (*jenkinsv1.PipelineActivity, error) {
	getOptions := metav1.GetOptions{}
	name := pipelineActivityName()
	// TODO probably should use labels here (HF)
	activity, err := jxClient.JenkinsV1().PipelineActivities(ns).Get(name, getOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving pipeline activity")
	}

	return activity, nil
}

func pipelineActivityName() string {
	sourceURL := viper.GetString(sourceURLOptionName)
	gitInfo, err := gits.ParseGitURL(sourceURL)
	if err != nil {
		logger.Fatal(err)
	}

	branch := viper.GetString(branchOptionName)
	buildNumber := viper.GetString(buildNumberOptionName)

	return strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", gitInfo.Organisation, gitInfo.Name, branch, buildNumber))
}
