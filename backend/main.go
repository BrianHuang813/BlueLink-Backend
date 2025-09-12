package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Sui RPC client configuration
const DEFAULT_SUI_RPC = "https://fullnode.testnet.sui.io:443"

// Data structures matching the Move contract
type Project struct {
	ID          string `json:"id"`
	Creator     string `json:"creator"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FundingGoal string `json:"funding_goal"`
	TotalRaised string `json:"total_raised"`
	DonorCount  string `json:"donor_count"`
}

type DonationReceipt struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Donor     string `json:"donor"`
	Amount    string `json:"amount"`
}

type SuiObjectResponse struct {
	Data struct {
		ObjectID string `json:"objectId"`
		Content  struct {
			DataType string `json:"dataType"`
			Type     string `json:"type"`
			Fields   struct {
				ID struct {
					ID string `json:"id"`
				} `json:"id"`
				Creator     string `json:"creator"`
				Name        string `json:"name"`
				Description string `json:"description"`
				FundingGoal string `json:"funding_goal"`
				TotalRaised struct {
					Value string `json:"value"`
				} `json:"total_raised"`
				DonorCount string `json:"donor_count"`
				ProjectID  struct {
					ID string `json:"id"`
				} `json:"project_id,omitempty"`
				Donor  string `json:"donor,omitempty"`
				Amount string `json:"amount,omitempty"`
			} `json:"fields"`
		} `json:"content"`
	} `json:"data"`
}

type SuiRPCClient struct {
	URL string
}

func NewSuiRPCClient() *SuiRPCClient {
	url := os.Getenv("SUI_RPC_URL")
	if url == "" {
		url = DEFAULT_SUI_RPC
	}
	return &SuiRPCClient{URL: url}
}

func (client *SuiRPCClient) makeRPCCall(method string, params interface{}) (json.RawMessage, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(client.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResponse struct {
		Result json.RawMessage `json:"result"`
		Error  interface{}     `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResponse); err != nil {
		return nil, err
	}

	if rpcResponse.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", rpcResponse.Error)
	}

	return rpcResponse.Result, nil
}

func (client *SuiRPCClient) GetObjectsByType(objectType string) ([]SuiObjectResponse, error) {
	params := []interface{}{
		map[string]interface{}{
			"MatchAll": []interface{}{
				map[string]interface{}{
					"StructType": objectType,
				},
			},
		},
		map[string]interface{}{
			"showType":    true,
			"showContent": true,
			"showOwner":   true,
		},
		nil,
		50,
	}

	result, err := client.makeRPCCall("suix_queryObjects", params)
	if err != nil {
		return nil, err
	}

	var queryResult struct {
		Data []SuiObjectResponse `json:"data"`
	}

	if err := json.Unmarshal(result, &queryResult); err != nil {
		return nil, err
	}

	return queryResult.Data, nil
}

func (client *SuiRPCClient) GetObject(objectId string) (*SuiObjectResponse, error) {
	params := []interface{}{
		objectId,
		map[string]interface{}{
			"showType":    true,
			"showContent": true,
			"showOwner":   true,
		},
	}

	result, err := client.makeRPCCall("sui_getObject", params)
	if err != nil {
		return nil, err
	}

	var object SuiObjectResponse
	if err := json.Unmarshal(result, &object); err != nil {
		return nil, err
	}

	return &object, nil
}

func (client *SuiRPCClient) GetOwnedObjects(owner string, objectType string) ([]SuiObjectResponse, error) {
	params := []interface{}{
		owner,
		map[string]interface{}{
			"filter": map[string]interface{}{
				"StructType": objectType,
			},
		},
		nil,
		50,
	}

	result, err := client.makeRPCCall("suix_getOwnedObjects", params)
	if err != nil {
		return nil, err
	}

	var queryResult struct {
		Data []struct {
			Data SuiObjectResponse `json:"data"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &queryResult); err != nil {
		return nil, err
	}

	var objects []SuiObjectResponse
	for _, item := range queryResult.Data {
		objects = append(objects, item.Data)
	}

	return objects, nil
}

// API Handlers
func setupRouter() *gin.Engine {
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	client := NewSuiRPCClient()

	// Get all projects
	r.GET("/api/projects", func(c *gin.Context) {
		// Replace with actual package address after deployment
		objectType := "0x0::bluelink::Project"

		objects, err := client.GetObjectsByType(objectType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var projects []Project
		for _, obj := range objects {
			if obj.Data.Content.DataType == "moveObject" {
				project := Project{
					ID:          obj.Data.ObjectID,
					Creator:     obj.Data.Content.Fields.Creator,
					Name:        obj.Data.Content.Fields.Name,
					Description: obj.Data.Content.Fields.Description,
					FundingGoal: obj.Data.Content.Fields.FundingGoal,
					TotalRaised: obj.Data.Content.Fields.TotalRaised.Value,
					DonorCount:  obj.Data.Content.Fields.DonorCount,
				}
				projects = append(projects, project)
			}
		}

		c.JSON(http.StatusOK, projects)
	})

	// Get specific project by ID
	r.GET("/api/projects/:id", func(c *gin.Context) {
		projectId := c.Param("id")

		object, err := client.GetObject(projectId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if object.Data.Content.DataType != "moveObject" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		project := Project{
			ID:          object.Data.ObjectID,
			Creator:     object.Data.Content.Fields.Creator,
			Name:        object.Data.Content.Fields.Name,
			Description: object.Data.Content.Fields.Description,
			FundingGoal: object.Data.Content.Fields.FundingGoal,
			TotalRaised: object.Data.Content.Fields.TotalRaised.Value,
			DonorCount:  object.Data.Content.Fields.DonorCount,
		}

		c.JSON(http.StatusOK, project)
	})

	// Get donation receipts for an address
	r.GET("/api/donors/:address", func(c *gin.Context) {
		address := c.Param("address")

		// Replace with actual package address after deployment
		objectType := "0x0::bluelink::DonationReceipt"

		objects, err := client.GetOwnedObjects(address, objectType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var receipts []DonationReceipt
		for _, obj := range objects {
			if obj.Data.Content.DataType == "moveObject" {
				receipt := DonationReceipt{
					ID:        obj.Data.ObjectID,
					ProjectID: obj.Data.Content.Fields.ProjectID.ID,
					Donor:     obj.Data.Content.Fields.Donor,
					Amount:    obj.Data.Content.Fields.Amount,
				}
				receipts = append(receipts, receipt)
			}
		}

		c.JSON(http.StatusOK, receipts)
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default values")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := setupRouter()

	fmt.Printf("BlueLink Backend API starting on port %s\n", port)
	fmt.Printf("Health check: http://localhost:%s/health\n", port)
	fmt.Printf("API Base URL: http://localhost:%s/api\n", port)

	log.Fatal(r.Run(":" + port))
}
