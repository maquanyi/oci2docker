package convert

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/specs"
)

type DockerInfo struct {
	Appdir      string
	Entrypoint  string
	Expose      string
	Environment string
	Port        bool
	Env         bool
}

const (
	buildTemplate = `
FROM scratch
MAINTAINER ChengTiesheng <chengtiesheng@huawei.com>
ADD {{.Appdir}} .
{{if .Env}} 
ENV {{.Environment}}
{{end}}
ENTRYPOINT ["{{.Entrypoint}}"]
{{if .Port}}
EXPOSE {{.Expose}}
{{end}}
`
)

func RunOCI2Docker(path string, imgName string, port string) error {
	appdir := "./rootfs"
	entrypoint := getEntrypointFromSpecs(path)
	env := getEnvFromSpecs(path)

	bPort := false
	if port != "" {
		bPort = true
	}

	bEnv := false
	if env != "" {
		bEnv = true
	}

	dockerInfo := DockerInfo{
		Appdir:      appdir,
		Entrypoint:  entrypoint,
		Expose:      port,
		Environment: env,
		Port:        bPort,
		Env:         bEnv,
	}

	generateDockerfile(dockerInfo)

	dirWork := createWorkDir()

	run(exec.Command("mv", "./Dockerfile", dirWork+"/Dockerfile"))
	run(exec.Command("cp", "-rf", path+"/rootfs", dirWork))

	run(exec.Command("docker", "build", "-t", imgName, dirWork))

	return nil
}

func generateDockerfile(dockerInfo DockerInfo) {
	t := template.Must(template.New("buildTemplate").Parse(buildTemplate))

	f, err := os.Create("Dockerfile")
	if err != nil {
		log.Fatal("Error wrinting Dockerfile %v", err.Error())
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
	spec := *pSpec

	for index := range spec.Process.Env {
		env = env + spec.Process.Env[index] + " "

	}

	return env
}
