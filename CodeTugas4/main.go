package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var templates map[string]*template.Template

var ctx = context.Background()

type employee struct {
	Id      bson.ObjectId `bson:"_id"`
	Name    string        `bson:"name"`
	Email   string        `bson:"email"`
	Phone   string        `bson:"phone"`
	Address string        `bson:"address"`
}

func connect() (*mongo.Database, error) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://datastore:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client.Database("employee_db"), nil
}

func init() {
	loadTemplates()
}

func main() {

	router := mux.NewRouter()

	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public/"))))
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/tambah", tambah).Methods("POST")
	router.HandleFunc("/update", update).Methods("POST")
	router.HandleFunc("/hapus", hapus).Methods("POST")

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func index(res http.ResponseWriter, req *http.Request) {
	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	csr, err := db.Collection("employee").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err.Error())
	}
	defer csr.Close(ctx)

	result := make([]employee, 0)
	for csr.Next(ctx) {
		var row employee
		err := csr.Decode(&row)
		if err != nil {
			log.Fatal(err.Error())
		}

		result = append(result, row)
	}

	var data = bson.M{"employee": result}

	if err := templates["index"].Execute(res, data); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func tambah(res http.ResponseWriter, req *http.Request) {
	var name = req.FormValue("name")
	var email = req.FormValue("email")
	var phone = req.FormValue("phone")
	var address = req.FormValue("address")

	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = db.Collection("employee").InsertOne(ctx, employee{bson.NewObjectId(), name, email, phone, address})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Insert success!")

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func update(res http.ResponseWriter, req *http.Request) {
	var name_before = req.FormValue("name-before")
	var email_before = req.FormValue("email-before")
	var phone_before = req.FormValue("phone-before")
	var address_before = req.FormValue("address-before")

	var id = req.FormValue("id")
	var name = req.FormValue("name")
	var email = req.FormValue("email")
	var phone = req.FormValue("phone")
	var address = req.FormValue("address")

	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	var selector = bson.M{"name": name_before, "email": email_before, "phone": phone_before, "address": address_before}
	var changes = employee{bson.ObjectIdHex(id), name, email, phone, address}

	_, err = db.Collection("employee").UpdateOne(ctx, selector, bson.M{"$set": changes})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Update success!")

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func hapus(res http.ResponseWriter, req *http.Request) {
	var name_before = req.FormValue("name-before")
	var email_before = req.FormValue("email-before")
	var phone_before = req.FormValue("phone-before")
	var address_before = req.FormValue("address-before")

	db, err := connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	var selector = bson.M{"name": name_before, "email": email_before, "phone": phone_before, "address": address_before}
	_, err = db.Collection("employee").DeleteOne(ctx, selector)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Remove success!")

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func loadTemplates() {
	templates = make(map[string]*template.Template)

	templates["index"] = template.Must(template.ParseFiles("index.html"))
}
