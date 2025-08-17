package rotate

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Key rotates the key associated with profile, as follows:
//
// 1. it deletes all but the current key (regardless of their age)
// 2. it creates a new key and pushes it to shared credentials file
// 3. it deletes the current key.
func Key(profile string, dryMode bool) (err error) {
	ctx := context.Background()

	credsPath, err := setup(profile)
	if err != nil {
		return
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return
	}

	fmt.Printf("Using access key %s from profile '%s'.\n", creds.AccessKeyID, profile)

	stsc := sts.NewFromConfig(cfg)
	respGetCallerIdentity, err := stsc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return
	}

	fmt.Println("Your user arn is:", *respGetCallerIdentity.Arn)

	iamc := iam.NewFromConfig(cfg)

	keys, err := iamc.ListAccessKeys(ctx, &iam.ListAccessKeysInput{})
	if err != nil {
		return
	}

	for _, k := range keys.AccessKeyMetadata {
		if *k.AccessKeyId == creds.AccessKeyID {
			continue
		}

		if dryMode {
			fmt.Println("Pretending to delete", *k.AccessKeyId)
			continue
		}

		_, err = iamc.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
			AccessKeyId: k.AccessKeyId,
		})
		if err != nil {
			return
		}

		fmt.Println("Deleted access key", *k.AccessKeyId)
	}

	if dryMode {
		fmt.Println("Pretending to create a new key...")
		fmt.Println("Pretending to update credentials file...")
		fmt.Println("Pretending to delete current key...")
		fmt.Println("All good, pretending worked.")

		return
	}

	nkey, err := iamc.CreateAccessKey(ctx, &iam.CreateAccessKeyInput{})
	if err != nil {
		return
	}

	fmt.Println("Created new access key", *nkey.AccessKey.AccessKeyId)

	if err = updateCredentialsFile(credsPath, creds, nkey); err != nil {
		// If updating credentials file fails, clean up the new key
		if _, err2 := iamc.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
			AccessKeyId: nkey.AccessKey.AccessKeyId,
		}); err2 != nil {
			return fmt.Errorf("%v, additionally failed to remove new key: %v", err, err2)
		}

		return
	}

	_, err = iamc.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
		AccessKeyId: &creds.AccessKeyID,
	})
	if err != nil {
		return
	}

	fmt.Println("Deleted old access key", creds.AccessKeyID)
	fmt.Println("Key rotation completed successfully!")

	return
}
