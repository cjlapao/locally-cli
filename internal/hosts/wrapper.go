package hosts

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

type HostsCommandWrapper struct {
	HostsFilePath string
	Hosts         []*HostEntry
}

func GetWrapper() *HostsCommandWrapper {
	result := &HostsCommandWrapper{
		Hosts: make([]*HostEntry, 0),
	}

	opSystem := helper.GetOperatingSystem()
	if opSystem == helper.LinuxOs || opSystem == helper.UnknownOs {
		result.HostsFilePath = "/etc/hosts"
	}
	if opSystem == helper.WindowsOs {
		result.HostsFilePath = "c:\\windows\\system32\\drivers\\etc\\hosts"
	}
	return result
}

func (svc *HostsCommandWrapper) Read() error {
	current, errorsInHost := svc.getHostsFileContent()
	if errorsInHost != nil {
		return fmt.Errorf("there was %v errors in the host file", len(errorsInHost))
	}

	svc.Hosts = current
	return nil
}

func (svc *HostsCommandWrapper) getHostsFileContent() ([]*HostEntry, []error) {
	result := make([]*HostEntry, 0)
	errorsArr := make([]error, 0)
	notify.Info("Getting the host file content from %v", svc.HostsFilePath)
	if !helper.FileExists(svc.HostsFilePath) {
		err := fmt.Errorf("there was an error getting the content of the host file on %v", svc.HostsFilePath)
		notify.Error(err.Error())
		errorsArr = append(errorsArr, err)
		return nil, errorsArr
	}

	file, err := os.Open(svc.HostsFilePath)
	if err != nil {
		errorsArr = append(errorsArr, err)
		return nil, errorsArr
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount += 1
		entry, err := svc.parse(scanner.Text())
		if err != nil {
			errorsArr = append(errorsArr, err)
		}
		if entry != nil {
			result = append(result, entry)
		}
	}

	if len(errorsArr) == 0 {
		return result, nil
	} else {
		return result, errorsArr
	}
}

func (svc *HostsCommandWrapper) parse(value string) (*HostEntry, error) {
	result := HostEntry{
		IsNew:     false,
		InSection: false,
		State:     StateNone,
	}
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.ReplaceAll(value, "\r\n", "")
	value = strings.ReplaceAll(value, "\r", "")

	if strings.HasPrefix(value, "#") || value == "#" || strings.TrimSpace(value) == "#" || strings.TrimSpace(value) == "" {
		return nil, nil
	}

	if _, err := strconv.Atoi(string(value[0])); err != nil {
		return nil, nil
	}
	// let's first check for comments
	commentParts := strings.Split(value, "#")
	if len(commentParts) > 1 {
		comment := strings.TrimSpace(commentParts[1])
		if strings.HasPrefix(comment, locally_COMMENT) {
			comment = strings.ReplaceAll(comment, fmt.Sprintf("%v: ", locally_COMMENT), "")
			comment = strings.ReplaceAll(comment, fmt.Sprintf("%v", locally_COMMENT), "")
		}
		result.Comment = comment
	}

	parts := strings.Split(value, " ")
	if len(parts) <= 1 {
		return nil, fmt.Errorf("%v does not seem to be valid format", value)
	}
	ip := parts[0]

	if !svc.ValidateIp(ip) {
		return nil, fmt.Errorf("ip %v does not seem to be in a valid format in %v", ip, value)
	}

	result.IP = ip
	for idx, part := range parts {
		if idx > 0 {
			if part != "" && part != " " && svc.ValidateHostname(part) {
				result.Hosts = append(result.Hosts, part)
			}
		}
	}

	return &result, nil
}

func (svc *HostsCommandWrapper) Exists(hostnameFrom string) (bool, *HostEntry) {
	for _, entry := range svc.Hosts {
		for _, hostname := range entry.Hosts {
			if strings.EqualFold(hostname, hostnameFrom) {
				return true, entry
			}
		}
	}

	return false, nil
}

func (svc *HostsCommandWrapper) Add(ip string, hostname string, comment string) (*HostEntry, error) {
	var entry *HostEntry
	if !svc.ValidateIp(ip) {
		return nil, fmt.Errorf("%v is not valid ip", ip)
	}
	if !svc.ValidateHostname(hostname) {
		return nil, fmt.Errorf("%v is not valid hostname", hostname)
	}

	if exists, host := svc.Exists(hostname); exists {
		host.State = StateClean
		host.Comment = comment
		return nil, fmt.Errorf("%v already exists with ip %v", hostname, host.IP)
	}

	entry = &HostEntry{
		IP:      ip,
		Hosts:   make([]string, 0),
		IsNew:   true,
		State:   StateNew,
		Comment: comment,
	}
	entry.Hosts = append(entry.Hosts, hostname)

	svc.Hosts = append(svc.Hosts, entry)
	return entry, nil
}

func (svc *HostsCommandWrapper) ValidateIp(ip string) bool {
	exp := "^((25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)\\.){3}(25[0-5]|(2[0-4]|1\\d|[1-9]|)\\d)$"
	result, err := regexp.MatchString(exp, ip)
	if err != nil {
		return false
	}

	return result
}

func (svc *HostsCommandWrapper) ValidateHostname(hostname string) bool {
	exp := "[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}\\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)"
	result, err := regexp.MatchString(exp, hostname)
	if err != nil {
		return false
	}

	return result
}

func (svc *HostsCommandWrapper) Clean() error {
	file, err := os.Open(svc.HostsFilePath)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0
	cleanState := NONE
	cleanedContent := make([]string, 0)
	for scanner.Scan() {
		var existingHost *HostEntry
		var exists bool
		line := scanner.Text()
		host, err := svc.parse(line)
		if err != nil {
			return err
		}

		if host != nil {
			for _, hostname := range host.Hosts {
				exists, existingHost = svc.Exists(hostname)
			}
		}

		if exists && existingHost.State == StateClean && cleanState == NONE {
			notify.Debug("This ip %v for hosts %v SHOULD BE CLEANED and added to the locally section", host.IP, strings.Join(host.Hosts, " "))
			existingHost.State = StateAdd
			continue
		}

		if line == START_SECTION {
			notify.Debug("Found start section in host file")
			cleanState = STARTED
		}
		if line == END_SECTION {
			notify.Debug("Found end section in host file")
			cleanState = ENDED
		}
		if cleanState == NONE {
			if host != nil {
				notify.Debug("This ip %v for hosts %v WILL NOT BE CLEANED as it is outside the section", host.IP, strings.Join(host.Hosts, " "))
			}
			cleanedContent = append(cleanedContent, line)
		} else {
			if host != nil {
				notify.Debug("This ip %v for hosts %v WILL BE CLEANED as it is inside the section", host.IP, strings.Join(host.Hosts, " "))
			}
			if exists {
				existingHost.State = StateAdd
			}
			if cleanState == ENDED {
				cleanState = NONE
			}
		}

		lineCount += 1
	}

	file.Close()
	if !strings.HasSuffix(cleanedContent[len(cleanedContent)-1], "\n") {
		cleanedContent[len(cleanedContent)-1] = fmt.Sprintf("%v\n", cleanedContent[len(cleanedContent)-1])
	}

	err = helper.WriteToFile(strings.Join(cleanedContent, "\n"), svc.HostsFilePath)
	if err != nil {
		return err
	}
	return nil
}

func (svc *HostsCommandWrapper) Save() error {
	svc.backupHostFile()

	if err := svc.Clean(); err != nil {
		return err
	}

	contentToWrite := make([]string, 0)

	for _, host := range svc.Hosts {
		if host.State == StateNew || host.State == StateAdd {
			if host.Comment != "" {
				contentToWrite = append(contentToWrite, fmt.Sprintf("%v    %v # %v: %v\n", host.IP, strings.Join(host.Hosts, " "), locally_COMMENT, host.Comment))
			} else {
				contentToWrite = append(contentToWrite, fmt.Sprintf("%v    %v # %v\n", host.IP, strings.Join(host.Hosts, " "), locally_COMMENT))
			}
		}
	}

	if len(contentToWrite) > 0 {
		f, err := os.OpenFile(svc.HostsFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			notify.FromError(err, "Could not open host file")
			return err
		}
		defer f.Close()

		if _, err := f.Write([]byte(fmt.Sprintf("%v\n", START_SECTION))); err != nil {
			notify.FromError(err, "Could not write to host file")
			return err
		}

		for _, host := range contentToWrite {
			if _, err := f.WriteString(host); err != nil {
				notify.FromError(err, "Could not write to host file")
				return err
			}
		}

		if _, err := f.Write([]byte(fmt.Sprintf("%v\n", END_SECTION))); err != nil {
			notify.FromError(err, "Could not write to host file")
			return err
		}
	}

	return nil
}

func (svc *HostsCommandWrapper) backupHostFile() error {
	backupFileName := svc.HostsFilePath + ".bck"
	notify.Flag("Backing up host file to %v", backupFileName)
	if helper.FileExists(backupFileName) {
		helper.DeleteFile(backupFileName)
	}

	if helper.FileExists(svc.HostsFilePath) {
		source, err := os.Open(svc.HostsFilePath)
		if err != nil {
			return err
		}

		defer source.Close()

		destination, err := os.Create(backupFileName)
		if err != nil {
			return err
		}

		defer destination.Close()

		_, err = io.Copy(destination, source)

		return err
	}

	return errors.New("host file was not found to backup")
}
