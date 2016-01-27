package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/huawei-openlab/oci2docker/convert"
)

func main() {
	app := cli.NewApp()
	app.Name = "oci2docker"
	app.Usage = "A tool for coverting oci bundle to docker image"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:  "convert",
			Usage: "format converting operation",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "oci-bundle",
					Value: "",
					Usage: "path of oci-bundle to convert",
				},
				cli.BoolFlag{
					Name:  "debug",
					Usage: "debug messages switch, default false",
				},
				cli.StringFlag{
					Name:  "image-name",
					Value: "",
					Usage: "docker image name",
				},
				cli.StringFlag{
					Name:  "port",
					Value: "",
					Usage: "exposed port of docker images",
				},
			},
			Action: oci2docker,
		},
	}

	app.Run(os.Args)
}

func oci2docker(c *cli.Context) {
	ociPath := c.String("oci-bundle")
	imgName := c.String("image-name")
	port := c.String("port")
	flagDebug := c.Bool("debug")

	if c.NumFlags() == 0 {
		cli.ShowCommandHelp(c, "convert")
		return
	}

	if ociPath == "" {
		logrus.Infof("Please specify OCI bundle path.")
		return
	}

	_, err := os.Stat(ociPath)
	if os.IsNotExist(err) {
		logrus.Infof("OCI bundle path does not exsit.")
		return
	}

	if imgName == "" {
		logrus.Infof("Please specify docker image name for output.")
		return
	}

	convert.RunOCI2Docker(ociPath, flagDebug, imgName, port)

	return
}
