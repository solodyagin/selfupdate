package main

import (
	"fmt"
	"log"

	"github.com/solodyagin/selfupdate/cmd/selfupdatectl/internal/cloud"
	"github.com/urfave/cli/v2"
)

type awsConfig struct {
	endpoint   string
	region     string
	bucket     string
	accessKey  string
	secret     string
	baseS3Path string
}

func awsUpload() *cli.Command {
	a := &application{}
	config := &awsConfig{}

	return &cli.Command{
		Name:        "aws-upload",
		Usage:       "Upload an executable file to AWS S3, it will be signed and the signature uploaded too",
		Description: "The executable specified will get its signature generated and checked before being uploaded to a AWS S3 bucket location specified as the last arguments.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "private-key",
				Aliases:     []string{"priv"},
				Usage:       "The private key file to store the new key in.",
				Destination: &a.privateKey,
				Value:       "ed25519.key",
			},
			&cli.StringFlag{
				Name:        "public-key",
				Aliases:     []string{"pub"},
				Usage:       "The public key file to store the new key in.",
				Destination: &a.publicKey,
				Value:       "ed25519.pem",
			},
			&cli.StringFlag{
				Name:        "endpoint",
				Aliases:     []string{"e"},
				Usage:       "AWS endpoint to connect to (can be used to connect to non AWS S3 services)",
				EnvVars:     []string{"AWS_S3_ENDPOINT"},
				Destination: &config.endpoint,
			},
			&cli.StringFlag{
				Name:        "region",
				Aliases:     []string{"r"},
				Usage:       "AWS region to connect to",
				EnvVars:     []string{"AWS_S3_REGION"},
				Destination: &config.region,
			},
			&cli.StringFlag{
				Name:        "bucket",
				Aliases:     []string{"b"},
				Usage:       "AWS bucket to store data into",
				EnvVars:     []string{"AWS_S3_BUCKET"},
				Destination: &config.bucket,
			},
			&cli.StringFlag{
				Name:        "secret",
				Aliases:     []string{"s"},
				Usage:       "AWS secret to use to establish S3 connection",
				Destination: &config.secret,
			},
			&cli.StringFlag{
				Name:        "accesskey",
				Aliases:     []string{"a"},
				Usage:       "AWS Access Key ID to use to establish S3 connection",
				Destination: &config.accessKey,
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() != 2 {
				return fmt.Errorf("one executable and a S3 target path need to be specified")
			}

			log.Println("Connecting to AWS")
			session, err := cloud.NewAWSSession(config.accessKey, config.secret, config.endpoint, config.region, config.bucket)
			if err != nil {
				return err
			}

			config.baseS3Path = ctx.Args().Slice()[ctx.Args().Len()-1]

			exe := ctx.Args().First()
			s3path := buildS3Path(config.baseS3Path, exe)

			err = a.awsUpload(session, exe, s3path)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func (a *application) awsUpload(session *cloud.AWSSession, executable string, destination string) error {
	if a.check(executable) != nil {
		if err := a.sign(executable); err != nil {
			return err
		}
		if err := a.check(executable); err != nil {
			return err
		}
	}

	err := session.UploadFile(executable, destination)
	if err != nil {
		return err
	}
	fmt.Println()

	defer fmt.Println()
	return session.UploadFile(executable+".ed25519", destination+".ed25519")
}
