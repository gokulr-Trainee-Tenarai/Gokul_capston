package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ecommerce-microservices/gateway/client"

	gql "github.com/graphql-go/graphql"
)

type Server struct {
	consulClient *client.ConsulClient
	httpClient   *http.Client
	schema       gql.Schema
}

func NewServer(consulClient *client.ConsulClient) (*Server, error) {
	s := &Server{
		consulClient: consulClient,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
	}
	schema, err := s.buildSchema()
	if err != nil {
		return nil, err
	}
	s.schema = schema
	return s, nil
}

func (s *Server) buildSchema() (gql.Schema, error) {
	productType := gql.NewObject(gql.ObjectConfig{
		Name: "Product",
		Fields: gql.Fields{
			"id":    &gql.Field{Type: gql.Int},
			"name":  &gql.Field{Type: gql.String},
			"price": &gql.Field{Type: gql.Float},
		},
	})

	orderType := gql.NewObject(gql.ObjectConfig{
		Name: "Order",
		Fields: gql.Fields{
			"id":         &gql.Field{Type: gql.Int},
			"customerId": &gql.Field{Type: gql.Int},
			"productId":  &gql.Field{Type: gql.Int},
			"quantity":   &gql.Field{Type: gql.Int},
			"total":      &gql.Field{Type: gql.Float},
			"status":     &gql.Field{Type: gql.String},
		},
	})

	query := gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"product": &gql.Field{
				Type: productType,
				Args: gql.FieldConfigArgument{"id": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)}},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(int)
					baseURL, err := s.consulClient.GetServiceURL("product-service")
					if err != nil {
						return nil, err
					}
					resp, err := s.httpClient.Get(fmt.Sprintf("%s/products/%d", baseURL, id))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					var res map[string]interface{}
					_ = json.NewDecoder(resp.Body).Decode(&res)
					return res, nil
				},
			},
			"products": &gql.Field{
				Type: gql.NewList(productType),
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					baseURL, err := s.consulClient.GetServiceURL("product-service")
					if err != nil {
						return nil, err
					}
					resp, err := s.httpClient.Get(fmt.Sprintf("%s/products", baseURL))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					var res []map[string]interface{}
					_ = json.NewDecoder(resp.Body).Decode(&res)
					return res, nil
				},
			},
			"order": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{"id": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)}},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id := p.Args["id"].(int)
					baseURL, err := s.consulClient.GetServiceURL("order-service")
					if err != nil {
						return nil, err
					}
					resp, err := s.httpClient.Get(fmt.Sprintf("%s/orders/%d", baseURL, id))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					var res map[string]interface{}
					_ = json.NewDecoder(resp.Body).Decode(&res)
					return res, nil
				},
			},
			"orders": &gql.Field{
				Type: gql.NewList(orderType),
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					baseURL, err := s.consulClient.GetServiceURL("order-service")
					if err != nil {
						return nil, err
					}
					resp, err := s.httpClient.Get(fmt.Sprintf("%s/orders", baseURL))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					var res []map[string]interface{}
					_ = json.NewDecoder(resp.Body).Decode(&res)
					return res, nil
				},
			},
		},
	})

	mutation := gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			"createProduct": &gql.Field{
				Type: productType,
				Args: gql.FieldConfigArgument{
					"name":  &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
					"price": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Float)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					baseURL, err := s.consulClient.GetServiceURL("product-service")
					if err != nil {
						return nil, err
					}
					bodyData, _ := json.Marshal(p.Args)
					resp, err := s.httpClient.Post(fmt.Sprintf("%s/products", baseURL), "application/json", bytes.NewBuffer(bodyData))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					var res map[string]interface{}
					_ = json.NewDecoder(resp.Body).Decode(&res)
					return res, nil
				},
			},
			"createOrder": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"customerId": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
					"productId":  &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
					"quantity":   &gql.ArgumentConfig{Type: gql.NewNonNull(gql.Int)},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					baseURL, err := s.consulClient.GetServiceURL("order-service")
					if err != nil {
						return nil, err
					}
					bodyData, _ := json.Marshal(p.Args)
					resp, err := s.httpClient.Post(fmt.Sprintf("%s/orders", baseURL), "application/json", bytes.NewBuffer(bodyData))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					var res map[string]interface{}
					_ = json.NewDecoder(resp.Body).Decode(&res)
					return res, nil
				},
			},
		},
	})

	return gql.NewSchema(gql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
}
