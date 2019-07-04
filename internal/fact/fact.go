package fact

import (
	"fmt"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/logging"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/report"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/util"
	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	jenkinsv1types "github.com/jenkins-x/jx/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appName = "jacoco"
)

var (
	logger = logging.AppLogger().WithFields(log.Fields{"component": "fact"})
)

// CreateFact creates the Fact CRD from the specified XML report.
func CreateFact(report report.Report, pipelineActivity *jenkinsv1.PipelineActivity, url string) *jenkinsv1.Fact {
	measurements := make([]jenkinsv1.Measurement, 0)
	for _, c := range report.Counters {
		t := ""
		switch c.Type {
		case "INSTRUCTION":
			t = jenkinsv1.CodeCoverageCountTypeInstructions
		case "LINE":
			t = jenkinsv1.CodeCoverageCountTypeLines
		case "METHOD":
			t = jenkinsv1.CodeCoverageCountTypeMethods
		case "COMPLEXITY":
			t = jenkinsv1.CodeCoverageCountTypeComplexity
		case "BRANCH":
			t = jenkinsv1.CodeCoverageCountTypeBranches
		case "CLASS":
			t = jenkinsv1.CodeCoverageCountTypeClasses
		}
		measurementCovered := createMeasurement(t, jenkinsv1.CodeCoverageMeasurementCoverage, c.Covered)
		measurementMissed := createMeasurement(t, jenkinsv1.CodeCoverageMeasurementMissed, c.Missed)
		measurementTotal := createMeasurement(t, jenkinsv1.CodeCoverageMeasurementTotal, c.Covered+c.Missed)
		measurements = append(measurements, measurementCovered, measurementMissed, measurementTotal)
	}

	name := fmt.Sprintf("%s-%s-%s", appName, jenkinsv1.FactTypeCoverage, pipelineActivity.Name)
	fact := jenkinsv1.Fact{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"subjectkind":  "PipelineActivity",
				"pipelineName": pipelineActivity.Name,
			},
		},
		Spec: jenkinsv1.FactSpec{
			Name:     name,
			FactType: jenkinsv1.FactTypeCoverage,
			Original: jenkinsv1.Original{
				URL:      url,
				MimeType: "application/xml",
				Tags: []string{
					"jacoco.xml",
				},
			},
			Tags: []string{
				appName,
			},
			Measurements: measurements,
			Statements:   []jenkinsv1.Statement{},
			SubjectReference: jenkinsv1.ResourceReference{
				APIVersion: pipelineActivity.APIVersion,
				Kind:       pipelineActivity.Kind,
				Name:       pipelineActivity.Name,
				UID:        pipelineActivity.UID,
			},
		},
	}
	logger.Tracef("created fact: %v", fact)
	return &fact
}

// StoreFact applies the specified Fact into the cluster.
func StoreFact(fact *jenkinsv1.Fact, factsInterface jenkinsv1types.FactInterface) error {
	f := func() error {
		_, err := factsInterface.Create(fact)
		if err != nil {
			switch err.(type) {
			case *errors.StatusError:
				status := err.(*errors.StatusError)
				if status.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
					logger.Debugf("fact with name '%s' already existed", fact.Name)
					return nil
				}
				return err
			default:
				return err
			}
		}
		return nil
	}
	return util.ApplyWithBackoff(f)
}

func createMeasurement(t string, measurement string, value int) jenkinsv1.Measurement {
	return jenkinsv1.Measurement{
		Name:             fmt.Sprintf("%s-%s", t, measurement),
		MeasurementType:  jenkinsv1.MeasurementCount,
		MeasurementValue: value,
	}
}
