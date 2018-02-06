package rotate

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
)

func setup(profile string) (credsPath string, err error) {
	if profile == "" {
		return "", fmt.Errorf("no profile given")
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	credsPath = filepath.Join(usr.HomeDir, ".aws", "credentials")

	return
}

func updateCredentialsFile(credsPath string, oldKey credentials.Value, newKey *iam.CreateAccessKeyOutput) error {
	content, err := ioutil.ReadFile(credsPath)
	if err != nil {
		return err
	}

	re1, err := regexp.Compile(`(?m)^aws_access_key_id\s*=\s*` + regexp.QuoteMeta(oldKey.AccessKeyID))
	if err != nil {
		return err
	}
	if !re1.Match(content) {
		return fmt.Errorf("unable to locate key id in credentials file")
	}

	content = re1.ReplaceAll(content, []byte(`aws_access_key_id = `+*newKey.AccessKey.AccessKeyId))

	re2, err := regexp.Compile(`(?m)^aws_secret_access_key\s*=\s*` + regexp.QuoteMeta(oldKey.SecretAccessKey))
	if err != nil {
		return err
	}
	if !re2.Match(content) {
		return fmt.Errorf("unable to locate key secret in credentials file")
	}

	content = re2.ReplaceAll(content, []byte(`aws_secret_access_key = `+*newKey.AccessKey.SecretAccessKey))

	if err = ioutil.WriteFile(credsPath, content, 0600); err != nil {
		return err
	}

	fmt.Println("Wrote new key pair to", credsPath)

	return nil
}
