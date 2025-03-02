package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

const MAJOR_SEPARATOR = "====================================="
const MINOR_SEPARATOR = "-------------------------------------"

func main() {

	// Take command line arguments for maximum number of policies to list
	// If no arguments are provided, list up to 10 policies

	// maxPols := flag.Int("max", 10, "Maximum number of policies to list")
	// flag.Parse()

	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}
	iamClient := iam.NewFromConfig(sdkConfig)

	fmt.Println("Getting details for the current user...")

	// Call the get-user API to get the details of the current user and print them
	// i.e. aws iam get-user
	fmt.Println(MAJOR_SEPARATOR)
	currentUserDetails, err := GetUserDetails(ctx, iamClient)
	if err != nil {
		fmt.Println("Couldn't get details for the current user. Exiting...")
		return
	}

	fmt.Println("User details:")
	fmt.Printf("\tUsername: %v\n", *currentUserDetails.User.UserName)
	fmt.Printf("\tUser ARN: %v\n", *currentUserDetails.User.Arn)
	fmt.Printf("\tUser ID: %v\n", *currentUserDetails.User.UserId)
	fmt.Printf("\tCreated on: %v\n", *currentUserDetails.User.CreateDate)
	fmt.Println(MAJOR_SEPARATOR)

	// Call the list-groups-for-user API to get the policies attached to the current user and print them
	// i.e. aws iam list-groups-for-user --user-name <username>
	fmt.Println(MAJOR_SEPARATOR)
	fmt.Println("Getting groups for the current user...")
	fmt.Println(MAJOR_SEPARATOR)
	userGroups, err := ListUserGroups(ctx, iamClient, *currentUserDetails.User.UserName)
	if err != nil {
		fmt.Println("Couldn't get groups for the current user. Exiting...")
		return
	}

	for _, group := range userGroups.Groups {
		fmt.Printf("\tGroup name: %v\n", *group.GroupName)
		fmt.Printf("\tGroup ARN: %v\n", *group.Arn)
		fmt.Printf("\tGroup ID: %v\n", *group.GroupId)
		fmt.Printf("\tCreated on: %v\n", *group.CreateDate)
		fmt.Println(MINOR_SEPARATOR)
	}

	// Call the list-attached-user-policies API to get the policies attached to the current user and print them
	// i.e. aws iam list-attached-user-policies --user-name <username>
	fmt.Println(MAJOR_SEPARATOR)
	fmt.Println("Getting attached policies for the current user...")
	fmt.Println(MAJOR_SEPARATOR)
	userPolicies, err := ListAttachedUserPolicies(ctx, iamClient, *currentUserDetails.User.UserName)
	if err != nil {
		fmt.Println("Couldn't get attached policies for the current user. Exiting...")
		return
	}

	for _, policy := range userPolicies.AttachedPolicies {
		fmt.Printf("\tPolicy name: %v\n", *policy.PolicyName)
		fmt.Printf("\tPolicy ARN: %v\n", *policy.PolicyArn)
		fmt.Println(MINOR_SEPARATOR)
	}

	// Prompt the user if they want to get the details of any policy's latest version
	PromptUserForPolicyVersionDetails(ctx, iamClient)

	// Call the list-user-policies API to get the inline policies attached to the current user and print them
	// i.e. aws iam list-user-policies --user-name <username>
	fmt.Println(MAJOR_SEPARATOR)
	fmt.Println("Getting inline policies for the current user...")
	fmt.Println(MAJOR_SEPARATOR)
	userInlinePolicies, err := ListInlineUserPolicies(ctx, iamClient, *currentUserDetails.User.UserName)
	if err != nil {
		fmt.Println("Couldn't get inline policies for the current user. Exiting...")
		return
	}

	for _, policy := range userInlinePolicies.PolicyNames {
		fmt.Printf("\tPolicy name: %v\n", policy)
		fmt.Println(MINOR_SEPARATOR)
	}

	fmt.Println("All done!")

}

func PromptUserForPolicyVersionDetails(ctx context.Context, iamClient *iam.Client) {
	// Prompt if the user wants to get policy version details
	// If yes, call the get-policy-version API to get the details of the policy version
	// i.e. aws iam get-policy-version --policy-arn <policy-arn> --version-id <version-id>
	// If no, exit
	fmt.Print("Do you want the details of any policy's version? (y/n): ")
	var input string
	fmt.Scanln(&input)
	if input == "y" {
		fmt.Print("Enter the ARN of the policy: ")
		var policyArn string
		fmt.Scanln(&policyArn)

		fmt.Println("Available versions: ")
		policyVersions, err := ListLatestPolicyVersions(ctx, iamClient, policyArn)
		if err != nil {
			fmt.Println("Couldn't get details for the policy version")
			return
		}

		for _, version := range policyVersions.Versions {
			fmt.Printf("\tVersion ID: %v\n", *version.VersionId)
			fmt.Printf("\tCreated on: %v\n", *version.CreateDate)
			fmt.Println()
		}

		fmt.Print("Enter the version ID to retrieve: ")
		var versionId string
		fmt.Scanln(&versionId)
		fmt.Println("Getting details for version ", versionId)
		policyVersionDetails, err := GetPolicyVersionDetails(ctx, iamClient, policyArn, versionId)
		if err != nil {
			fmt.Println("Couldn't get details for the policy version")
			return
		}

		// Print out the VersionID, CreateDate, and Document of the policy version
		fmt.Println(MAJOR_SEPARATOR)
		fmt.Println("Policy version details:")
		fmt.Printf("\tVersion ID: %v\n", *policyVersionDetails.PolicyVersion.VersionId)
		fmt.Printf("\tCreated on: %v\n", *policyVersionDetails.PolicyVersion.CreateDate)
		decodedDocument, err := url.QueryUnescape(*policyVersionDetails.PolicyVersion.Document)
		if err != nil {
			fmt.Println("Couldn't encode the document. Exiting...")
			return
		}

		fmt.Printf("\tDocument: \n%v\n", decodedDocument)
		fmt.Println(MAJOR_SEPARATOR)

	} else {
		return
	}

}

func ListLatestPolicyVersions(ctx context.Context, iamClient *iam.Client, policyArn string) (*iam.ListPolicyVersionsOutput, error) {
	// Get the details of the policy version
	policyVersions, err := iamClient.ListPolicyVersions(ctx, &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		fmt.Printf("Couldn't get details for the policy version. Here's why: %v\n", err)
		return nil, err
	}

	return policyVersions, nil
}

func GetPolicyVersionDetails(ctx context.Context, iamClient *iam.Client, policyArn string, versionId string) (*iam.GetPolicyVersionOutput, error) {
	// Get the details of the policy version
	policyVersionDetails, err := iamClient.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(policyArn),
		VersionId: aws.String(versionId),
	})
	if err != nil {
		fmt.Printf("Couldn't get details for the policy version. Here's why: %v\n", err)
		return nil, err
	}

	return policyVersionDetails, nil
}

func GetUserDetails(ctx context.Context, iamClient *iam.Client) (*iam.GetUserOutput, error) {
	// Get the details of the user
	userDetails, err := iamClient.GetUser(ctx, &iam.GetUserInput{})
	if err != nil {
		fmt.Printf("Couldn't get details for the user. Here's why: %v\n", err)
		return nil, err
	}

	return userDetails, nil
}

func ListUserGroups(ctx context.Context, iamClient *iam.Client, username string) (*iam.ListGroupsForUserOutput, error) {
	// Get the groups that the user belongs to
	userGroups, err := iamClient.ListGroupsForUser(ctx, &iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})
	if err != nil {
		fmt.Printf("Couldn't get the groups for the user. Here's why: %v\n", err)
		return nil, err
	}

	return userGroups, nil
}

func ListAttachedUserPolicies(ctx context.Context, iamClient *iam.Client, username string) (*iam.ListAttachedUserPoliciesOutput, error) {
	// Get the policies attached to the user
	userPolicies, err := iamClient.ListAttachedUserPolicies(ctx, &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(username),
	})
	if err != nil {
		fmt.Printf("Couldn't get the policies attached to the user. Here's why: %v\n", err)
		return nil, err
	}

	return userPolicies, nil
}

func ListInlineUserPolicies(ctx context.Context, iamClient *iam.Client, username string) (*iam.ListUserPoliciesOutput, error) {
	// Get the inline policies attached to the user
	userPolicies, err := iamClient.ListUserPolicies(ctx, &iam.ListUserPoliciesInput{
		UserName: aws.String(username),
	})
	if err != nil {
		fmt.Printf("Couldn't get the inline policies attached to the user. Here's why: %v\n", err)
		return nil, err
	}

	return userPolicies, nil
}
