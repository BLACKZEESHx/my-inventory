package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(t *testing.M) {
	err := a.Initialize(Dbuser, Dbpassword, Dbname)
	if err != nil {
		log.Fatal(err)
	}
	createTable()
	t.Run()
}

func createTable() {
	createTablequery := `CREATE TABLE IF NOT EXISTS products (
		id int(11) NOT NULL AUTO_INCREMENT,
        name varchar(255) NOT NULL,
        quantity int(11) NOT NULL,
        price int(11) NOT NULL,
        PRIMARY KEY (id)
	);`
	_, err := a.DB.Exec(createTablequery)
	if err != nil {
		log.Fatal(err)
	}
}
func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER TABLE products AUTO_INCREMENT=1")
}

func addProduct(name string, quantity int, price float64) {
	query := fmt.Sprintf("insert into products(name, quantity, price) values('%v', %v, %v)", name, quantity, price)
	a.DB.Exec(query)
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProduct("Keyboard", 122, 299)
	request, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)

	// _, err := a.getProdcut(nil, nil)
	// if err == nil {
	// t.Error("expected error")
	// }
}

func checkStatusCode(t *testing.T, expectedStatusCode int, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("expected status code %d, got %d", expectedStatusCode, actualStatusCode)
	}
}

func sendRequest(req *http.Request) httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, req)
	return *recorder

}

func TestCreateProduct(t *testing.T) {
	clearTable()
	var product = []byte(`{"name":"chair", "quantity":12, "price":120}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(product))
	req.Header.Set("Content-type", "application/json")

	response := sendRequest(req)
	checkStatusCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "chair" {
		t.Errorf("expected chair, got %v", m["name"])
	}

	if m["quantity"] != 12.0 {
		t.Errorf("expected 12.0, got %v", m["quantity"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("connector", 10, 10)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("connector", 10, 10)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	var oldvalue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &oldvalue)

	var product = []byte(`{"name":"connector", "quantity":12, "price":150}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(product))
	req.Header.Set("Content-type", "application/json")

	response = sendRequest(req)

	var newvalue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &newvalue)

	if oldvalue["id"] != newvalue["id"] {
		t.Errorf("Expected id %v, got %v", newvalue["id"], oldvalue["id"])
	}

	if oldvalue["name"] != newvalue["name"] {
		t.Errorf("Expected name %v, got %v", newvalue["name"], oldvalue["name"])
	}

	if oldvalue["price"] == newvalue["price"] {
		t.Errorf("Expected name %v, got %v", newvalue["price"], oldvalue["price"])
	}

	if oldvalue["quantity"] == newvalue["quantity"] {
		t.Errorf("Expected name %v, got %v", newvalue["quantity"], oldvalue["quantity"])
	}
}
