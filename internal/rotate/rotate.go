package rotate

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

// Key rotates the key associated with profile, as follows:
//
// 1. it deletes all but the current key (regardless of their age)
// 2. it creates a new key and pushes it to shared credentials file
// 3. it deletes the current key.
func Key(profile string, dryMode bool) error {
	credsPath, err := setup(profile)
	if err != nil {
		return err
	}

	credentialsProvider := credentials.NewSharedCredentials(credsPath, profile)
	creds, err := credentialsProvider.Get()
	if err != nil {
		return err
	}

	fmt.Printf("Using access key %s from profile '%s'.\n", creds.AccessKeyID, profile)

	sess, err := session.NewSession(&aws.Config{Credentials: credentialsProvider})
	if err != nil {
		return err
	}

	stsc := sts.New(sess)
	respGetCallerIdentity, err := stsc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return errors.Wrap(err, "unable to get caller id (is the key disabled?)")
	}

	fmt.Println("Your user arn is:", *respGetCallerIdentity.Arn)

	iamc := iam.New(sess)

	keys, err := iamc.ListAccessKeys(&iam.ListAccessKeysInput{})
	if err != nil {
		return err
	}

	for _, k := range keys.AccessKeyMetadata {
		if *k.AccessKeyId == creds.AccessKeyID {
			continue
		}

		if dryMode {
			fmt.Println("Pretending to delete", *k.AccessKeyId)
			continue
		}

		_, err = iamc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			AccessKeyId: k.AccessKeyId,
		})

		if err != nil {
			return err
		}
	}

	if dryMode {
		fmt.Println("Pretending to create a new key...")
		fmt.Println("Pretending to update credentials file...")
		fmt.Println("Pretending to delete current key...")
		fmt.Println("All good, pretending worked.")
		return nil
	}

	nkey, err := iamc.CreateAccessKey(&iam.CreateAccessKeyInput{})
	if err != nil {
		return err
	}

	if err = updateCredentialsFile(credsPath, creds, nkey); err != nil {
		if _, err2 := iamc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			AccessKeyId: nkey.AccessKey.AccessKeyId,
		}); err2 != nil {
			return fmt.Errorf("%v, additionally failed to remove new key: %v", err, err2)
		}

		return err
	}

	_, err = iamc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		AccessKeyId: &creds.AccessKeyID,
	})

	return err
}
