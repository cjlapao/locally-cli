package npm

var globalNpmService *NpmService

type NpmService struct {
	wrapper *NpmCommandWrapper
}

func New() *NpmService {
	svc := NpmService{
		wrapper: GetWrapper(),
	}
	return &svc
}

func Get() *NpmService {
	if globalNpmService != nil {
		return globalNpmService
	}

	return New()
}

func (svc *NpmService) CheckForNpm(softFail bool) {
	notify.Rocket("Running npm tool checker for locally service")

	svc.wrapper.CheckForNpm(softFail)
}

func (svc *NpmService) CI(workingDir string, minVersion string) error {
	notify.Rocket("Running npm ci for locally service")

	if err := svc.wrapper.CI(workingDir, minVersion); err != nil {
		return err
	}

	return nil
}

func (svc *NpmService) Install(workingDir string, minVersion string) error {
	notify.Rocket("Running npm install for locally service")

	if err := svc.wrapper.Install(workingDir, minVersion); err != nil {
		return err
	}

	return nil
}

func (svc *NpmService) Publish(workingDir string, minVersion string) error {
	notify.Rocket("Running npm publish for locally service")

	if err := svc.wrapper.Publish(workingDir, minVersion); err != nil {
		return err
	}

	return nil
}

func (svc *NpmService) Custom(customCommand string, workingDir string, minVersion string) error {
	notify.Rocket("Running npm custom command [%s] for locally service", customCommand)

	if err := svc.wrapper.Custom(customCommand, workingDir, minVersion); err != nil {
		notify.Error(err.Error())
		return err
	}

	return nil
}
