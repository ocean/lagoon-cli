package api

import (
	"encoding/json"
	"github.com/machinebox/graphql"
)

// UpdateTask .
func (api *Interface) UpdateTask(task UpdateTask) ([]byte, error) {
	req := graphql.NewRequest(`
	mutation ($id: Int!, $patch: UpdateTaskPatchInput!) {
		updateTask(input: {
			id: $id
			patch: $patch
		}) {
		  	...Task
		}
	}` + taskFragment)
	generateVars(req, task)
	returnType, err := api.RunQuery(req, Data{})
	if err != nil {
		return []byte(""), err
	}
	reMappedResult := returnType.(map[string]interface{})
	jsonBytes, err := json.Marshal(reMappedResult["updateTask"])
	if err != nil {
		return []byte(""), err
	}
	return jsonBytes, nil
}
