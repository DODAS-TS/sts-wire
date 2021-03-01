package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/host"

	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

func instanceLog() string {
	instanceLogFile, errOpenLog := os.OpenFile(instanceLogFilename, os.O_RDONLY, fileMode)
	if errOpenLog != nil {
		if !strings.Contains(errOpenLog.Error(), "no such file or directory") {
			panic(errOpenLog)
		} else {
			return ""
		}
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

	hostStat, _ := host.Info()

	report.WriteString(divider)
	report.WriteString("\n| Version\n")
	report.WriteString(buildCmdVersion())
	report.WriteRune('\n')

	report.WriteString("| System\n")
	report.WriteString(divider)
	report.WriteRune('\n')
	report.WriteString(fmt.Sprintf("Hostname:\t\t\t\t%s\n", hostStat.Hostname))
	report.WriteString(fmt.Sprintf("OS:\t\t\t\t\t\t%s\n", hostStat.OS))
	report.WriteString(fmt.Sprintf("Platform:\t\t\t\t%s\n", hostStat.Platform))
	report.WriteString(fmt.Sprintf("PlatformFamily:\t\t\t%s\n", hostStat.PlatformFamily))
	report.WriteString(fmt.Sprintf("PlatformVersion:\t\t%s\n", hostStat.PlatformVersion))
	report.WriteString(fmt.Sprintf("KernelVersion:\t\t\t%s\n", hostStat.KernelVersion))
	report.WriteString(fmt.Sprintf("KernelArch:\t\t\t\t%s\n", hostStat.KernelArch))
	report.WriteString(fmt.Sprintf("VirtualizationSystem:\t%s\n", hostStat.VirtualizationSystem))
	report.WriteString(fmt.Sprintf("VirtualizationRole:\t%s\n", hostStat.VirtualizationRole))
	report.WriteString(divider)
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

	baseDir := filepath.Clean(filepath.Dir(instanceLogFilename))

	if baseDir == "." {
		baseDir = filepath.Join(".", ".generic")

		errMkdirs := os.MkdirAll(baseDir, os.ModePerm)
		if errMkdirs != nil {
			panic(errMkdirs)
		}
	}

	reportFilename := filepath.Join(baseDir, fmt.Sprintf("report_%d.out", time.Now().Unix()))

	reportFile, errCreateReport := os.OpenFile(reportFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_SYNC, fileMode)
	if errCreateReport != nil {
		panic(errCreateReport)
	}

	defer reportFile.Close()

	reportFile.WriteString(report.String())

	log.Err(fmt.Errorf("%s", mainErr)).Msg("ERROR")
	log.Info().Msg(divider)
	log.Info().Str("report", reportFilename).Msg("Report created")
	log.Warn().Msg("Please, use the report to have more details...")
	log.Warn().Msg("You can search for reports using the command \"sts-wire report\"")
	log.Info().Msg(divider)
}
