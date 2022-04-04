# test_terraform

Terraform files to provision an Azure Service Bus, for testing sb-shovel.

**Prerequisites:**

- [Terraform v1.0.0+](https://www.terraform.io/downloads)
- [az CLI](https://docs.microsoft.com/en-us/cli/azure/)
- Create a file with a name like `*.tfvars` (e.g. `test.tfvars`)

## Creating your tfvars file

To run this Terraform configuration, you need a `.tfvars` file. It does not matter what the name is, as long as you remember to refer to it correctly during the `plan`, `apply` and `destroy` steps when running Terraform.

Example containing required variables: `test.tfvars`
```
rg_name     = "sb-shovel-test"
rg_location = "UK West"
sb_name     = "testservicebus"
q_name      = "testqueue"
```

## Guide to run

Log in to `az CLI`, a tool by Microsoft allowing you to interact with Azure Resource Manager (create / update / change Azure resources)

```
az login -t <tenantid>

# if you have multiple Azure subscriptions, run the following to change your default
az account set --subscription <subscriptionid>
```

Initialise Terraform

```
terraform init
```

Perform Terraform Plan

```
terraform plan -var-file="test.tfvars"
# plans to create 3 resources
```

If the above was successful and looks as expected, apply the plan to create the Azure resources

```
terraform apply -var-file="test.tfvars" --auto-approve
# auto approve is not recommended, without first having proven that a plan shows expected changes
```

Once complete, verify the resources are in your chosen subscription.

## After testing

To keep any personal costs down, destroy resources after use

```
terraform destroy -var-file="test.tfvars" --auto-approve
```