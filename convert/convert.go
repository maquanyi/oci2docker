package convert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/specs"
)

// DockerInfo stores data for generating Dockerfile.
type DockerInfo struct {
	Appdir      string
	Entrypoint  string
	Expose      string
	Environment string
	Command     string
	Port        bool
	Env         bool
	Add         bool
	Cmd         bool
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
func RunOCI2Docker(path string, flagDebug bool, imgName string, port string) error {
	if flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	appdir := getRootPathFromSpecs(path)
	entrypoint := getEntrypointFromSpecs(path)
	env := getEnvFromSpecs(path)
	command := getPoststartFromSpecs(path)

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

	dockerInfo := DockerInfo{
		Appdir:      appdir,
		Entrypoint:  entrypoint,
		Expose:      port,
		Environment: env,
		Command:     command,
		Port:        bPort,
		Env:         bEnv,
		Add:         bAdd,
		Cmd:         bCmd,
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

	err = run(exec.Command("/bin/sh", "-c", cmdStr))
	if err != nil {
		logrus.Debugf("Docker build failed: %v", err)
	} else {
		logrus.Infof("Docker image %v generated successfully.", imgName)
	}

	return nil
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

func getConfigSpec(path string) *specs.LinuxSpec {

	configPath := path + "/config.json"
	config, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Debugf("Open file config.json failed: %v", err)
		return nil
	}

	spec := new(specs.LinuxSpec)
	err = json.Unmarshal(config, spec)
	if err != nil {
		logrus.Debugf("Unmarshal config.json failed: %v", err)
		return nil
	}

	return spec
}

func getRuntimeSpec(path string) *specs.LinuxRuntimeSpec {

	runtimePath := path + "/runtime.json"
	runtime, err := ioutil.ReadFile(runtimePath)
	if err != nil {
		logrus.Debugf("Open file runtime.json failed: %v", err)
		return nil
	}

	spec := new(specs.LinuxRuntimeSpec)
	err = json.Unmarshal(runtime, spec)
	if err != nil {
		logrus.Debugf("Unmarshal runtime.json failed: %v", err)
		return nil
	}

	return spec
}

func getEntrypointFromSpecs(path string) string {

	pSpec := getConfigSpec(path)
	if pSpec == nil {
		return ""
	}
	spec := *pSpec

	prefixDir := ""
	entryPoint := spec.Process.Args
	if entryPoint == nil {
		return "/bin/sh"
	}
	if !filepath.IsAbs(entryPoint[0]) {
		if spec.Process.Cwd == "" {
			prefixDir = "/"
		} else {
			prefixDir = spec.Process.Cwd
		}
	}
	entryPoint[0] = prefixDir + entryPoint[0]

	var res []string
	res = strings.SplitAfter(entryPoint[0], "/")
	if len(res) <= 2 {
		entryPoint[0] = "/bin" + entryPoint[0]
	}

	return entryPoint[0]
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

func getPoststartFromSpecs(path string) string {
	pSpec := getRuntimeSpec(path)
	if pSpec == nil {
		return ""
	}
	spec := *pSpec

	if len(spec.Hooks.Poststart) == 0 {
		return ""
	}

	poststart := spec.Hooks.Prestart[0].Path
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
