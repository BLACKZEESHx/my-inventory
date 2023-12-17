package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// App represents the application
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// New creates a new instance of App
func (app *App) Initialize(Dbuser string, Dbpassword string, Dbname string) error {
	connectionString := fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v", Dbuser, Dbpassword, Dbname)
	var err error
	app.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	app.Router = mux.NewRouter().StrictSlash(true)
	app.handleRoutes()
	return nil
}

// SetDB sets the database connection
func (app *App) Run(address string) {
	log.Fatal(http.ListenAndServe(address, app.Router))
}

func sendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)

}

func sendError(w http.ResponseWriter, statusCode int, err string) {
	errormessage := map[string]string{"error": err}
	sendResponse(w, statusCode, errormessage)
}

func (app *App) getProdcuts(w http.ResponseWriter, r *http.Request) {
	Products, err := getProducts(app.DB)

	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, Products)
}

func (app *App) getProdcut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])

	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	p := product{ID: key}
	err = p.getProduct(app.DB)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			sendError(w, http.StatusNotFound, "product not found")
		default:
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	sendResponse(w, http.StatusOK, p)
}

func (app *App) createProdcut(w http.ResponseWriter, r *http.Request) {
	var p product
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid requested payload")
		return
	}
	err = p.createProduct(app.DB)

	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusCreated, p)
}

func (app *App) updateProdcut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])

	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var p product
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid requested payload")
		return
	}
	p.ID = key
	err = p.updateProdcut(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, p)
}

func (app *App) deleteProdcut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi(vars["id"])

	p := product{ID: key}
	err = p.deleteProdcut(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"result": "successfully deleted"})

}

func (app *App) handleRoutes() {
	app.Router.HandleFunc("/products", app.getProdcuts).Methods("GET")
	app.Router.HandleFunc("/product/{id}", app.getProdcut).Methods("GET")
	app.Router.HandleFunc("/product", app.createProdcut).Methods("POST")
	app.Router.HandleFunc("/product/{id}", app.updateProdcut).Methods("PUT")
	app.Router.HandleFunc("/product/{id}", app.deleteProdcut).Methods("DELETE")
}
