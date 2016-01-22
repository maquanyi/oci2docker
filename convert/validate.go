package convert

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

const (
	// ConfigFile is the path to config file inside the bundle
	ConfigFile = "config.json"
	// RuntimeFile is the path to runtime.json
	RuntimeFile = "runtime.json"
	// RootfsDir is the path to rootfs directory inside the bundle
	RootfsDir = "rootfs"
)

var (
	// ErrNoRootFS ...
	ErrNoRootFS = errors.New("no rootfs found in bundle")
	// ErrNoConfig ...
	ErrNoConfig = errors.New("no config json file found in bundle")
	// ErrNoRun ...
	ErrNoRun = errors.New("no runtime json file found in bundle")
)

type validateRes struct {
	cfgOK   bool
	runOK   bool
	rfsOK   bool
	config  io.Reader
	runtime io.Reader
}

func validateOCIProc(path string) bool {
	var bRes bool
	if err := validateBundle(path); err != nil {
		logrus.Debugf("%s: invalid oci bundle: %v.", path, err)
		bRes = false
	} else {
		logrus.Debugf("%s: valid oci bundle.", path)
		bRes = true
	}
	return bRes
}

func validateBundle(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error accessing bundle: %v", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("given path %q is not a directory", path)
	}
	var flist []string
	var res validateRes
	walkBundle := func(fpath string, fi os.FileInfo, err error) error {
		rpath, err := filepath.Rel(path, fpath)
		if err != nil {
			return err
		}
		switch rpath {
		case ".":
		case ConfigFile:
			res.config, err = os.Open(fpath)
			if err != nil {
				return err
			}
			res.cfgOK = true
		case RuntimeFile:
			res.runtime, err = os.Open(fpath)
			if err != nil {
				return err
			}
			res.runOK = true
		case RootfsDir:
			if !fi.IsDir() {
				return errors.New("rootfs is not a directory")
			}
			res.rfsOK = true
		default:
			flist = append(flist, rpath)
		}
		return nil
	}
	if err := filepath.Walk(path, walkBundle); err != nil {
		return err
	}
	return checkBundle(res, flist)
}

func checkBundle(res validateRes, files []string) error {
	defer func() {
		if rc, ok := res.config.(io.Closer); ok {
			rc.Close()
		}
		if rc, ok := res.runtime.(io.Closer); ok {
			rc.Close()
		}
	}()
	if !res.cfgOK {
		return ErrNoConfig
	}
	if !res.runOK {
		return ErrNoRun
	}
	if !res.rfsOK {
		return ErrNoRootFS
	}
	_, err := ioutil.ReadAll(res.config)
	if err != nil {
		return fmt.Errorf("error reading the bundle: %v", err)
	}
	_, err = ioutil.ReadAll(res.runtime)
	if err != nil {
		return fmt.Errorf("error reading the bundle: %v", err)
	}

	for _, f := range files {
		if !strings.HasPrefix(f, "rootfs") {
			return fmt.Errorf("unrecognized file path in bundle: %q", f)
		}
	}
	return nil
}
