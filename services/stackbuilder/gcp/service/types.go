package service

type Endpoints map[string]Endpoint

type Endpoint struct {
	Uri string
}
