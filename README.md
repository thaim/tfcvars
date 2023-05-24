# tfcvars
![test](https://github.com/thaim/tfcvars/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/thaim/tfcvars/branch/main/graph/badge.svg?token=8QSTFQX364)](https://codecov.io/gh/thaim/tfcvars)
[![Go Report Card](https://goreportcard.com/badge/github.com/thaim/tfcvars)](https://goreportcard.com/report/github.com/thaim/tfcvars)

tfcvars is a CLI tool that synchronizes variables managed in Terraform Cloud with local variable files.

## Usage
```
$ bin/tfcvars help
NAME:
   tfcvars - synchronize terraform cloud variables

USAGE:
   tfcvars [global options] command [command options] [arguments...]

COMMANDS:
   help  Show this help
   show  Show variables on Terraform Cloud
   diff  Show difference of variables between local tfvars and Terraform Cloud variables
   pull  update local tfvars with Terraform Cloud variables
   push  update Terraform Cloud variables with local tfvars

GLOBAL OPTIONS:
   --tfetoken value                The token used to authenticate with Terraform Cloud [$TFE_TOKEN]
   --organization value, -o value  Terraform Cloud organization name to deal with [$TFCVARS_ORGANIZATION]
   --workspace value, -w value     Terraform Cloud workspace name to deal with [$TFCVARS_WORKSPACE]
   --version, -v                   print the version
```

### Show command
show command print list of variables in Terraform Cloud.

```
$ tfcvars show --format tfvars
availability_zones = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
target_port        = "3000"
test               = "data"
environment        = "test2"
cidr_blocks        = ["10.1.0.0/24", "10.1.1.0/24", "10.1.2.0/24"]
tags = {
  tag_key1 = "tag_value3"
  tag_key2 = "tag_value3"
}
```

### Diff command
diff command print difference between local tfvars file and Terraform Cloud variables.

```
$ cat terraform.tfvars
availability_zones = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
target_port        = "3000"
environment        = "test"
cidr_blocks        = ["10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"]
tags = {
  tag_key1 = "tag_value1"
  tag_key2 = "tag_value2"
}

$ tfcvars diff
  availability_zones = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
  target_port        = "3000"
- test               = "data"
+ environment        = "test"
- environment        = "test2"
- cidr_blocks        = ["10.1.0.0/24", "10.1.1.0/24", "10.1.2.0/24"]
+ cidr_blocks        = ["10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"]
  tags = {
-   tag_key1 = "tag_value3"
+   tag_key1 = "tag_value1"
-   tag_key2 = "tag_value3"
+   tag_key2 = "tag_value2"
  }
```

### Pull command
pull command download Terraform Cloud variables and save as local terraform.tfvars file.

### Push command
push command update Terraform Cloud variables with local terraform.tfvars file.


## Limitation
### Sensitive Data
Terraform Cloud variables marked as "sensitive" cannot be shown or downloaded.

### Environment Variable
Terraform Cloud variables marked as "environment" can be shown or downloaded by setting the `--include-env` option. However, local environment variables are not taken into account in diff command or push command.

### Variable Set
Terraform Cloud variables stored as variable set can be shown or downloaded by setting the `--include-variable-set` option. However, those variables are not updated in push command.


## Install
### Install from binary
Binaries are available from [Github Releases](https://github.com/thaim/tfcvars/releases).

### Install from homebrew
```
$ brew install thaim/tap/tfcvars
```

### Install from go install
```
$ go install github.com/thaim/tfcvars@main
```


## LICENSE
MIT
