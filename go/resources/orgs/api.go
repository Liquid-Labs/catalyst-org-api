package orgs

import (
  "fmt"
  "net/http"

  "github.com/gorilla/mux"

  "github.com/Liquid-Labs/catalyst-core-api/go/handlers"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, "/orgs is alive\n")
}

func createHandler(w http.ResponseWriter, r *http.Request) {
  var org *Org = &Org{}
  if _, restErr := handlers.CheckAndExtract(w, r, org, `Org`); restErr != nil {
    return // response handled by CheckAndExtract
  } else {
    handlers.DoCreate(w, r, CreateOrg, org, `Org`)
  }
}

func detailHandler(w http.ResponseWriter, r *http.Request) {
  if _, restErr := handlers.BasicAuthCheck(w, r); restErr != nil {
    return // response handled by BasicAuthCheck
  } else {
    vars := mux.Vars(r)
    pubID := vars["pubId"]

    handlers.DoGetDetail(w, r, GetOrg, pubID, `Org`)
  }
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
  var newData *Org = &Org{}
  if _, restErr := handlers.CheckAndExtract(w, r, newData, `Org`); restErr != nil {
    return // response handled by CheckAndExtract
  } else {
    vars := mux.Vars(r)
    pubID := vars["pubId"]

    handlers.DoUpdate(w, r, UpdateOrg, newData, pubID, `Org`)
  }
}

const uuidRE = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`

func InitAPI(r *mux.Router) {
  r.HandleFunc("/orgs/", pingHandler).Methods("PING")
  r.HandleFunc("/orgs/", createHandler).Methods("POST")
  r.HandleFunc("/orgs/{pubId:" + uuidRE + "}/", detailHandler).Methods("GET")
  r.HandleFunc("/orgs/{pubId:" + uuidRE + "}/", updateHandler).Methods("PUT")
}
