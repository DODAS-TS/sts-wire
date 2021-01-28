package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

func instanceLog() string {
	instanceLogFile, errOpenLog := os.OpenFile(instanceLogFilename, os.O_RDONLY, fileMode)
	if errOpenLog != nil {
		panic(errOpenLog)
	}

	defer instanceLogFile.Close()

	var buffer bytes.Buffer

	_, errRead := buffer.ReadFrom(instanceLogFile)
	if errRead != nil {
		return fmt.Errorf("cannot read instance log: %w", errRead).Error()
	}

	return buffer.String()
}

func WriteReport(mainErr interface{}) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	report := strings.Builder{}
	report.WriteString(divider)
	report.WriteString("\n| Version\n")
	report.WriteString(buildCmdVersion())
	report.WriteRune('\n')

	c := viper.AllSettings()

	bs, errMarshall := yaml.Marshal(c)
	if errMarshall != nil {
		log.Err(errMarshall).Msg("unable to marshal config to YAML")
	}

	report.WriteString("| Parameters\n")
	report.WriteString(divider)
	report.Write(bs)
	report.WriteString(divider)
	report.WriteRune('\n')

	report.WriteString("| Instance log\n")
	report.WriteString(divider)
	report.WriteRune('\n')
	report.WriteString(instanceLog())
	report.WriteRune('\n')

	report.WriteString(divider)
	report.WriteString("\n| Error\n")
	report.WriteString(divider)
	report.WriteRune('\n')
	report.WriteString(fmt.Sprintf("%s\n", mainErr))

	reportFilename := path.Join(path.Dir(instanceLogFilename), fmt.Sprintf("report_%d.out", time.Now().Unix()))

	reportFile, errCreateReport := os.OpenFile(reportFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_SYNC, fileMode)
	if errCreateReport != nil {
		panic(errCreateReport)
	}

	defer reportFile.Close()

	reportFile.WriteString(report.String())

	log.Err(fmt.Errorf("%s", mainErr)).Msg("ERROR")
	log.Info().Str("report", reportFilename).Msg("Report created")
	log.Warn().Msg("Please, use the report to have more details...")
}
