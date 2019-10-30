package cmd

import (
	"fmt"
	"os"

	"github.com/amazeeio/lagoon-cli/api"
	"github.com/amazeeio/lagoon-cli/graphql"

	"encoding/json"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project name]",
	Short: "Update a project",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Not enough arguments. Requires: project name")
			cmd.Help()
			os.Exit(1)
		}
		projectName := args[0]

		lagoonAPI, err := graphql.LagoonAPI()
		if err != nil {
			fmt.Println(err)
			return
		}

		var jsonBytes []byte
		// get the project id from name
		projectBName := api.Project{
			Name: projectName,
		}
		projectByName, err := lagoonAPI.GetProjectByName(projectBName, graphql.ProjectByNameFragment)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonBytes, _ = json.Marshal(projectByName)

		reMappedResult := projectByName.(map[string]interface{})
		var projects api.Project
		jsonBytes, _ = json.Marshal(reMappedResult["project"])
		err = json.Unmarshal([]byte(jsonBytes), &projects)
		if err != nil {
			fmt.Println(err)
			return
		}
		projectID := projects.ID

		// patch the project by id
		projectUpdate := api.UpdateProject{}
		project := api.ProjectPatch{}
		err = json.Unmarshal([]byte(jsonPatch), &project)
		if err != nil {
			fmt.Println(err)
			return
		}
		projectUpdate = api.UpdateProject{
			ID:    projectID,
			Patch: project,
		}

		projectUpdateID, err := lagoonAPI.UpdateProject(projectUpdate, graphql.ProjectByNameFragment)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonBytes, _ = json.Marshal(projectUpdateID)

		reMappedResult = projectUpdateID.(map[string]interface{})
		var updatedProject api.Project
		jsonBytes, _ = json.Marshal(reMappedResult["updateProject"])
		err = json.Unmarshal([]byte(jsonBytes), &updatedProject)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(fmt.Sprintf("Result: %s\n", aurora.Green("success")))
		fmt.Println(fmt.Sprintf("%s: %s", aurora.Yellow("Project"), updatedProject.Name))

	},
}

func init() {
	projectCmd.AddCommand(projectUpdateCmd)
	// use json to patch, maybe re-introduce these better later on @TODO
	// projectUpdateCmd.Flags().StringVarP(&projectGitURL, "giturl", "g", "", "GitURL of the project")
	// projectUpdateCmd.Flags().StringVarP(&projectBranches, "branches", "b", "", "Branches of the project")
	// projectUpdateCmd.Flags().StringVarP(&projectProductionEnvironment, "prod-env", "P", "", "Production environment of the project")
	// projectUpdateCmd.Flags().StringVarP(&projectPullRequests, "pull-requests", "r", "", "Pull requests of the project")
	// projectUpdateCmd.Flags().IntVarP(projectAutoIdle, "auto-idle", "a", 1, "Auto idle setting of the project")
	// projectUpdateCmd.Flags().IntVarP(&projectDevelopmentEnvironmentsLimit, "dev-env-limit", "d", 5, "Auto idle setting of the project")
	projectUpdateCmd.Flags().StringVarP(&jsonPatch, "json", "j", "", "JSON string to patch")
}
