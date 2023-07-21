package certificates

// import (
// 	"github.com/cjlapao/locally-cli/common"
// 	"errors"
// 	"os"
// 	"path/filepath"

// 	"github.com/cjlapao/common-go/execution_context"
// 	"github.com/cjlapao/common-go/helper"
// 	"gopkg.in/yaml.v3"
// )

// var ctx = execution_context.Get()
// var config = ctx.Configuration

// func Init() *CertificateGeneratorConfig {
// 	config := CertificateGeneratorConfig{
// 		Root: make([]*RootCertificate, 0),
// 	}
// 	return &config
// }

// func (c *CertificateGeneratorConfig) ReadFromFile() error {
// 	fileName := c.GetFile()
// 	if fileName == "" {
// 		fileName = ".\\certificates.yml"
// 	}

// 	if !helper.FileExists(fileName) {
// 		return errors.New("file not found")
// 	}

// 	yamlFile, err := helper.ReadFromFile(fileName)
// 	if err != nil {
// 		return err
// 	}

// 	if err := yaml.Unmarshal(yamlFile, c); err != nil {
// 		return err
// 	}

// 	if c.OutputToFile {
// 		config.UpsertKey(common.OUTPUT_TO_FILE, c.OutputToFile)
// 	}

// 	return nil
// }

// func (c *CertificateGeneratorConfig) SaveToFile() error {
// 	fileName := c.GetFile()
// 	if fileName == "" {
// 		fileName = ".\\certificates.yml"
// 	}

// 	if helper.FileExists(fileName) {
// 		helper.DeleteFile(fileName)
// 	}

// 	yamlFile, err := yaml.Marshal(c)
// 	if err != nil {
// 		return err
// 	}

// 	helper.WriteToFile(string(yamlFile), fileName)

// 	return nil
// }

// func (svc *CertificateGeneratorConfig) GetFile() string {
// 	if file := helper.GetFlagValue("cert-config-file", ""); file != "" {
// 		logger.Info("Loading file from %v", file)
// 		return file
// 	}
// 	if file := helper.GetFlagValue("c", ""); file != "" {
// 		logger.Info("Loading file from %v", file)
// 		return file
// 	}

// 	ex, err := os.Executable()
// 	if err != nil {
// 		panic(err)
// 	}
// 	exPath := filepath.Dir(ex)
// 	// Testing local personal file
// 	if helper.FileExists(helper.JoinPath(exPath, "certificates.personal.yml")) {
// 		logger.Info("Loading certificate configuration file from local certificates.personal.yml")
// 		configFilename := helper.JoinPath(exPath, "certificates.personal.yml")
// 		return configFilename
// 	}

// 	// Testing local personal file
// 	if helper.FileExists(helper.JoinPath(exPath, "certificates.personal.yaml")) {
// 		logger.Info("Loading certificate configuration file from local certificates.personal.yaml")
// 		configFilename := helper.JoinPath(exPath, "certificates.personal.yml")
// 		return configFilename
// 	}

// 	// Testing local file
// 	if helper.FileExists(helper.JoinPath(exPath, "certificates.yml")) {
// 		logger.Info("Loading certificate configuration file from local certificates.yml")
// 		configFilename := helper.JoinPath(exPath, "certificates.yml")
// 		return configFilename
// 	}

// 	// Testing local file
// 	if helper.FileExists(helper.JoinPath(exPath, "certificates.yaml")) {
// 		logger.Info("Loading certificate configuration file from local certificates.yaml")
// 		configFilename := helper.JoinPath(exPath, "certificates.yaml")
// 		return configFilename
// 	}

// 	return ""
// }
