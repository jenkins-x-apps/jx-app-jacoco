package pom

import (
	"github.com/beevik/etree"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/logging"
	"github.com/jenkins-x-apps/jx-app-jacoco/internal/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

const (
	jaCoCoArtifactID = "jacoco-maven-plugin"
	jaCoCoConfig     = `
     <plugin>
        <groupId>org.jacoco</groupId>
        <artifactId>jacoco-maven-plugin</artifactId>
        <version>0.8.4</version>
        <executions>
           <execution>
              <id>default-prepare-agent</id>
              <goals>
                 <goal>prepare-agent</goal>
              </goals>
           </execution>
           <execution>
              <id>prepare-xml-report</id>
              <goals>
                 <goal>report</goal>
              </goals>
              <phase>verify</phase>
           </execution>
        </executions>
     </plugin>
`
)

var (
	logger = logging.AppLogger().WithFields(log.Fields{"component": "meta-pipeline-extender"})
)

// Configurator configures a given pom.xml to create the JaCoCo coverage report.
type Configurator struct {
	pomPath string
}

// NewPomConfigurator creates a new instance of Configurator.
func NewPomConfigurator(pomPath string) Configurator {
	return Configurator{
		pomPath: pomPath,
	}
}

// ConfigurePom triggers the actual POM configuration.
func (c *Configurator) ConfigurePom() error {
	f, err := os.Open(c.pomPath)
	if err != nil {
		return errors.Wrapf(err, "unable to open pom.xml for reading")
	}
	defer func() {
		_ = f.Close()
	}()

	pom, err := c.loadPom(f)
	if err != nil {
		return errors.Wrapf(err, "unable to load pom.xml '%s'", c.pomPath)
	}

	err = c.configureJaCoCo(pom)
	if err != nil {
		return errors.Wrap(err, "unable to enhance pom.xml ")
	}

	err = f.Close()
	if err != nil {
		return errors.Wrapf(err, "unable to close pom.xml '%s'", c.pomPath)
	}

	err = c.backupAndWrite(pom)
	if err != nil {
		return err
	}

	return nil
}

func (c *Configurator) loadPom(in io.Reader) (*etree.Document, error) {
	doc := etree.NewDocument()
	_, err := doc.ReadFrom(in)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (c *Configurator) backupAndWrite(pom *etree.Document) error {
	// TODO - make backup optional/configurable (HF)
	err := util.CopyFile(c.pomPath, c.pomPath+".jacoco.orig")
	if err != nil {
		return errors.Wrapf(err, "unable to backup pom.xml")
	}

	f, err := os.Create(c.pomPath)
	if err != nil {
		return errors.Wrapf(err, "unable to open pom.xml for writing")
	}
	defer func() {
		_ = f.Close()
	}()
	err = c.writePom(pom, f)
	if err != nil {
		return errors.Wrap(err, "unable to write pom.xml")
	}
	return nil
}

func (c *Configurator) writePom(pom *etree.Document, out io.Writer) error {
	count, err := pom.WriteTo(out)
	logger.Debugf("written %d of data", count)
	if err != nil {
		return err
	}
	return nil
}

func (c *Configurator) configureJaCoCo(pom *etree.Document) error {
	// TODO - see whether it makes sense trying to modify existing config (HF)
	err := c.unlinkExistingJaCoCoConfig(pom)
	if err != nil {
		return err
	}

	plugins := pom.FindElement("//plugins")
	plugins.AddChild(c.jaCoCoPluginConfig())
	return nil
}

func (c *Configurator) unlinkExistingJaCoCoConfig(pom *etree.Document) error {
	// for now we don't try to update any existing config, but instead replace it
	for _, plugin := range pom.FindElements("//plugin") {
		artifactID := plugin.SelectElement("artifactId").Text()
		if artifactID == jaCoCoArtifactID {
			logger.Info("replacing existing JaCoCo config in pom.xml")
			parent := plugin.Parent()
			parent.RemoveChild(plugin)
			break
		}
	}

	return nil
}

func (c *Configurator) jaCoCoPluginConfig() *etree.Element {
	doc := etree.NewDocument()
	err := doc.ReadFromBytes([]byte(jaCoCoConfig))
	if err != nil {
		logger.Errorf("unable to load JaCoCo config: %s", err.Error())
	}
	return doc.Root()
}
