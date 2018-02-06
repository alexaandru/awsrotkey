# AWS Rotate Key

Refactored from https://github.com/Fullscreen/aws-rotate-key

The credit for the original idea and implementation belongs to them.

This tool **WILL REMOVE** all keys associated with your account (as determined from the profile)
and replace them with a NEW key. No deactivation/no cluttering with old/unused keys. You've
been warned.

**DISCLAIMER: Use this tool at your own risk.** If you use the AWS key for things other than
the aws cli (you use it for services/web apps/etc.) then this is probably not the tool you
are looking for. Removing an existing key WILL result in downtime until the apps/services/etc.
are updated & deployed to use the new key.

## Assumptions

The tool is supposed to _"work out of the box"_ and for that a few assumptions are made.
If any of them is invalidated you should NOT be using this tool.

1. You are using shared profiles;
2. You are NOT using multiple keys.

As long as you use a single key and you want it rotated this tool will work just fine.

If you want to test it without actually affecting your existing key(s) then run it in "dry mode".
You must always give it a profile name, it will NOT assume the default profile is "default".
You may either pass it as a flag, or if `AWS_PROFILE` is defined it will be used. Fur further
details see `./awsrotkey -h`.

## Install

```
go get -u github.com/alexaandru/awsrotkey
```

No binaries provided, but you do get a free advice: don't install binaries from untrusted sources
even if they are provided :-) Cheers!
