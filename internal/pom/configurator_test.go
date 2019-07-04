package pom

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	updateGolden = false
)

var update = flag.Bool("update", false, "update .golden files")

func TestConfigureJaCoCo(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	tests := []struct {
		testName string
	}{
		{
			testName: "pom-no-jacoco-plugin-configured.xml",
		},
		{
			testName: "pom-with-jacoco-plugin-configured.xml",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.testName, func(t *testing.T) {
			testFilePath := filepath.Join("testdata", filepath.FromSlash(t.Name()))

			f, err := os.Open(testFilePath)
			defer func() {
				_ = f.Close()
			}()
			assert.NoError(t, err)

			configurator := NewPomConfigurator(testFilePath)
			pom, err := configurator.loadPom(f)
			assert.NoError(t, err)

			err = configurator.configureJaCoCo(pom)
			assert.NoError(t, err)

			goldenPath := testFilePath + ".golden"
			if updateGolden {
				golden, err := os.Create(goldenPath)
				assert.NoError(t, err)
				err = configurator.writePom(pom, golden)
				assert.NoError(t, err)
				err = golden.Close()
				assert.NoError(t, err)
			}

			goldenBytes, err := ioutil.ReadFile(goldenPath)
			assert.NoError(t, err, fmt.Sprintf("failed reading .golden: %s", goldenPath))

			var enhancedPom bytes.Buffer
			w := bufio.NewWriter(&enhancedPom)
			err = configurator.writePom(pom, w)
			assert.NoError(t, err)
			err = w.Flush()
			assert.NoError(t, err)

			assert.Equal(t, string(goldenBytes), string(enhancedPom.Bytes()))
		})
	}
}
