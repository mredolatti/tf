package main

func main() {
	qp, _ := NewProcessor()
	server, err := NewServer(&qp.schema, 8080, "/graphql")
	if err != nil {
		panic("error constructing graphql server: " + err.Error())
	}
	server.Start()
}
