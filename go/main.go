package main

import (
  "github.com/Liquid-Labs/catalyst-core-api/go/restserv"
  // core resources
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/entities"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/locations"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/users"

  "github.com/Liquid-Labs/catalyst-orgs-api/go/resources/orgs"
  "github.com/Liquid-Labs/go-api/sqldb"
)

func main() {
  sqldb.RegisterSetup(entities.SetupDB)
  sqldb.RegisterSetup(locations.SetupDB)
  sqldb.RegisterSetup(users.SetupDB)
  sqldb.RegisterSetup(orgs.SetupDB)
  sqldb.InitDB()
  restserv.RegisterResource(orgs.InitAPI)
  restserv.Init()
}
