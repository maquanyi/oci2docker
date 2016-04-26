package convert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/Sirupsen/logrus"
	specs "github.com/opencontainers/specs/specs-go"
)

// DockerInfo stores data for generating Dockerfile.
type DockerInfo struct {
	Appdir      string
	Entrypoint  string
	Expose      string
	Environment string
	Workdir     string
	Command     string
	User        string
	Port        bool
	Env         bool
	Add         bool
	Cmd         bool
	Usr         bool
	Cwd         bool
}

const (
	buildTemplate = `
FROM scratch
MAINTAINER ChengTiesheng <chengtiesheng@huawei.com>
{{if .Add}}
ADD {{.Appdir}} .
{{end}}
{{if .Env}} 
ENV {{.Environment}}
{{end}}
{{if .Usr}}
USER {{.User}}
{{end}}
{{if .Cwd}}
WORKDIR {{.Workdir}}
{{end}}
{{if .Cmd}}
CMD {{.Command}}
{{end}}
ENTRYPOINT ["{{.Entrypoint}}"]
{{if .Port}}
EXPOSE {{.Expose}}
{{end}}
`
)

// RunOCI2Docker is the entrypoint for oci2docker CLI tool.
func RunOCI2Docker(path string, flagDebug bool, imgName string, port string) {
	if flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if bValidate := validateOCIProc(path); bValidate != true {
		logrus.Infof("Invalid oci bundle.")
		return
	}

	appdir := getRootPathFromSpecs(path)
	entrypoint, workdir := getEntrypointFromSpecs(path)
	env := getEnvFromSpecs(path)
	command := getPoststartFromSpecs(path)
	user := getUserFromSpecs(path)

	bPort := false
	if port != "" {
		bPort = true
	}

	bEnv := false
	if env != "" {
		bEnv = true
	}

	bAdd := false
	if appdir != "" {
		bAdd = true
		appdir = "./" + appdir
	}

	bCmd := false
	if command != "" {
		bCmd = true
	}

	bUsr := false
	if user != "" {
		bUsr = true
	}

	bCwd := false
	if workdir != "" {
		bCwd = true
	}

	dockerInfo := DockerInfo{
		Appdir:      appdir,
		Entrypoint:  entrypoint,
		Expose:      port,
		Environment: env,
		Workdir:     workdir,
		Command:     command,
		User:        user,
		Port:        bPort,
		Env:         bEnv,
		Add:         bAdd,
		Cmd:         bCmd,
		Usr:         bUsr,
		Cwd:         bCwd,
	}

	generateDockerfile(dockerInfo)

	dirWork := createWorkDir()

	err := run(exec.Command("mv", "./Dockerfile", dirWork+"/Dockerfile"))
	if err != nil {
		logrus.Debugf("Store Dockerfile failed: %v", err)
	}
	err = run(exec.Command("cp", "-rf", path+"/rootfs", dirWork))
	if err != nil {
		logrus.Debugf("Store rootfs failed: %v", err)
	}

	buildOut := ""
	if flagDebug == false {
		buildOut = " > /dev/null"
	}
	cmdStr := fmt.Sprintf("docker build -t %s %s %s", imgName, dirWork, buildOut)
	logrus.Debugf("Docker build log is:")

	err = run(exec.Command("/bin/sh", "-c", cmdStr))
	if err != nil {
		logrus.Debugf("Docker build failed: %v", err)
	} else {
		logrus.Infof("Docker image %v generated successfully.", imgName)
	}

	return
}

func generateDockerfile(dockerInfo DockerInfo) {
	t := template.Must(template.New("buildTemplate").Parse(buildTemplate))

	f, err := os.Create("Dockerfile")
	if err != nil {
		logrus.Debugf("Error wrinting Dockerfile %v", err.Error())
		return
	}
	defer f.Close()

	t.Execute(f, dockerInfo)

	return
}

// Create work directory for the conversion output
func createWorkDir() string {
	idir, err := ioutil.TempDir("", "oci2docker")
	if err != nil {
		return ""
	}
	rootfs := filepath.Join(idir, "rootfs")
	os.MkdirAll(rootfs, 0755)

	data := []byte{}
	if err := ioutil.WriteFile(filepath.Join(idir, "Dockerfile"), data, 0644); err != nil {
		return ""
	}

	logrus.Debugf("Docker build context is in %s\n", idir)
	return idir
}

func getConfigSpec(path string) *specs.Spec {

	configPath := path + "/config.json"
	config, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Debugf("Open file config.json failed: %v", err)
		return nil
	}

	spec := new(specs.Spec)
	err = json.Unmarshal(config, spec)
	if err != nil {
		logrus.Debugf("Unmarshal config.json failed: %v", err)
		return nil
	}

	return spec
}

func getEntrypointFromSpecs(path string) (string, string) {

	pSpec := getConfigSpec(path)
	if pSpec == nil {
		return "", ""
	}
	spec := *pSpec

	workDir := spec.Process.Cwd
	entryPoint := spec.Process.Args
	if entryPoint == nil {
		return "/bin/sh", ""
	}

	return entryPoint[0], workDir
}

func getEnvFromSpecs(path string) string {
	env := ""
	pSpec := getConfigSpec(path)
	if pSpec == nil {
		return ""
	}
	spec := *pSpec

	for index := range spec.Process.Env {
		env = env + spec.Process.Env[index] + " "

	}

	return env
}

func getRootPathFromSpecs(path string) string {
	pSpec := getConfigSpec(path)
	if pSpec == nil {
		return ""
	}
	spec := *pSpec

	rootpath := spec.Root.Path

	return rootpath
}

func getUserFromSpecs(path string) string {
	pSpec := getConfigSpec(path)
	if pSpec == nil {
		return ""
	}
	spec := *pSpec

	uid := fmt.Sprintf("%d", spec.Process.User.UID)
	gid := fmt.Sprintf("%d", spec.Process.User.GID)

	user := uid + "/" + gid

	return user
}

func getPoststartFromSpecs(path string) string {
	pSpec := getConfigSpec(path)
	if pSpec == nil {
		return ""
	}
	spec := *pSpec

	if len(spec.Hooks.Poststart) == 0 {
		return ""
	}

	poststart := spec.Hooks.Poststart[0].Path
	if poststart != "" {
		for i := range spec.Hooks.Poststart[0].Args {
			poststart = poststart + " " + spec.Hooks.Poststart[0].Args[i]
		}
		for i := range spec.Hooks.Poststart[0].Env {
			poststart = poststart + " " + spec.Hooks.Poststart[0].Env[i]
		}
	}

	return poststart
}
