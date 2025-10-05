package sui

type Client struct {
}

var clients = []Client{}

func NewClient() *Client {
	newClient := Client{}

	clients = append(clients, newClient)

	return &newClient
}
