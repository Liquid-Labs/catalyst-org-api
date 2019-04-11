package orgs

import (
  "context"
  "database/sql"
  "fmt"
  "log"
  "strconv"

  "github.com/Liquid-Labs/go-api/sqldb"
  "github.com/Liquid-Labs/go-nullable-mysql/nulls"
  "github.com/Liquid-Labs/go-rest/rest"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/users"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/locations"
)

var OrgsSorts = map[string]string{
  "": `o.display_name ASC `,
  `name-asc`: `o.display_name ASC `,
  `name-desc`: `o.display_name DESC `,
}

func ScanOrgSummary(row *sql.Rows) (*OrgSummary, error) {
	var o OrgSummary

	if err := row.Scan(&o.PubId, &o.LastUpdated, &o.DisplayName, &o.Summary, &o.Phone, &o.Email, &o.Homepage, &o.LogoURL); err != nil {
		return nil, err
	}

	return &o, nil
}

func ScanOrgDetail(row *sql.Rows) (*Org, *locations.Address, error) {
  var o Org
  var a locations.Address

	if err := row.Scan(&o.PubId, &o.LastUpdated, &o.DisplayName, &o.Summary,
      &o.Phone, &o.Email, &o.Homepage, &o.LogoURL, &o.Active, &o.AuthId,
      &o.LegalID, &o.LegalIDType, &o.Id,
      &a.LocationId, &a.Idx, &a.Label, &a.Address1, &a.Address2, &a.City,
      &a.State, &a.Zip, &a.Lat, &a.Lng); err != nil {
		return nil, nil, err
	}

  // Negative locationIds are used by the UID for temporary identification.
  if a.LocationId.Int64 < 0 {
    a.LocationId = nulls.NewNullInt64()
  }

	return &o, &a, nil
}

// implement rest.ResultBuilder
func BuildOrgResults(rows *sql.Rows) (interface{}, error) {
  results := make([]*OrgSummary, 0)
  for rows.Next() {
    org, err := ScanOrgSummary(rows)
    if err != nil {
      return nil, err
    }

    results = append(results, org)
  }

  return results, nil
}

// Implements rest.GeneralSearchWhereBit
func OrgsGeneralWhereGenerator(term string, params []interface{}) (string, []interface{}, error) {
  likeTerm := `%`+term+`%`
  var whereBit string
  if _, err := strconv.ParseInt(term,10,64); err == nil {
    whereBit += "AND (o.phone LIKE ? OR o.phone_backup LIKE ?) "
    params = append(params, likeTerm, likeTerm)
  } else {
    whereBit += "AND (o.display_name LIKE ? OR o.email LIKE ?) "
    params = append(params, likeTerm, likeTerm)
  }

  return whereBit, params, nil
}

const CommonOrgFields = `e.pub_id, e.last_updated, o.display_name, o.summary, o.phone, o.email, o.homepage, o.logo_url, u.active, u.auth_id, u.legal_id, u.legal_id_type `
const CommonOrgsFrom = `FROM orgs o JOIN users u ON o.id=u.id JOIN entities e ON o.id=e.id `

const createOrgStatement = `INSERT INTO orgs (id, display_name, summary, phone, email, homepage, logo_url) VALUES(?,?,?,?,?,?,?)`
func CreateOrg(o *Org, ctx context.Context) (*Org, rest.RestError) {
  txn, err := sqldb.DB.Begin()
  if err != nil {
    defer txn.Rollback()
    return nil, rest.ServerError("Could not create org record. (txn error)", err)
  }
  newO, restErr := CreateOrgInTxn(o, ctx, txn)
  // txn already rolled back if in error, so we only need to commit if no error
  if err == nil {
    defer txn.Commit()
  }
  return newO, restErr
}

func CreateOrgInTxn(o *Org, ctx context.Context, txn *sql.Tx) (*Org, rest.RestError) {
  o.Addresses.CompleteAddresses(ctx)

  var err error
  newId, restErr := users.CreateUserInTxn(&o.User, txn)
  if restErr != nil {
    defer txn.Rollback()
		return nil, restErr
  }

  o.Id = nulls.NewInt64(newId)

	_, err = txn.Stmt(createOrgQuery).Exec(newId, o.DisplayName, o.Summary, o.Phone, o.Email, o.Homepage, o.LogoURL)
	if err != nil {
    // TODO: can we do more to tell the cause of the failure? We assume it's due to malformed data with the HTTP code
    defer txn.Rollback()
    log.Print(err)
		return nil, rest.UnprocessableEntityError("Failure creating org.", err)
	}

  if restErr := o.Addresses.CreateAddresses(nulls.NewInt64(newId), ctx, txn); restErr != nil {
    defer txn.Rollback()
    return nil, restErr
  }

  newOrg, err := GetOrgByIDInTxn(o.Id.Int64, ctx, txn)
  if err != nil {
    return nil, rest.ServerError("Problem retrieving newly updated org.", err)
  }
  // Carry any 'ChangeDesc' made by the geocoding out.
  o.PromoteChanges()
  newOrg.ChangeDesc = o.ChangeDesc

  return newOrg, nil
}

const CommonOrgGet string = `SELECT ` + CommonOrgFields + `, o.id, loc.id, ea.idx, ea.label, loc.address1, loc.address2, loc.city, loc.state, loc.zip, loc.lat, loc.lng ` + CommonOrgsFrom + ` LEFT JOIN entity_addresses ea ON o.id=ea.entity_id AND ea.idx >= 0 LEFT JOIN locations loc ON ea.location_id=loc.id `
const getOrgStatement string = CommonOrgGet + `WHERE e.pub_id=? `

// GetOrg retrieves a Org from a public ID string (UUID). Attempting to
// retrieve a non-existent Org results in a rest.NotFoundError. This is used
// primarily to retrieve a Org in response to an API request.
//
// Consider using GetOrgByID to retrieve a Org from another backend/DB
// function. TODO: reference discussion of internal vs public IDs.
func GetOrg(pubId string, ctx context.Context) (*Org, rest.RestError) {
  return getOrgHelper(getOrgQuery, pubId, ctx, nil)
}

// GetOrgInTxn retrieves a Org by public ID string (UUID) in the context
// of an existing transaction. See GetOrg.
func GetOrgInTxn(pubId string, ctx context.Context, txn *sql.Tx) (*Org, rest.RestError) {
  return getOrgHelper(getOrgQuery, pubId, ctx, txn)
}

const getOrgByAuthIdStatement string = CommonOrgGet + ` WHERE u.auth_id=? `
// GetOrgByAuthId retrieves a Org from a public authentication ID string
// provided by the authentication provider (firebase). Attempting to retrieve a
// non-existent Org results in a rest.NotFoundError. This is used primarily
// to retrieve a Org in response to an API request, especially
// '/orgs/self'.
func GetOrgByAuthId(authId string, ctx context.Context) (*Org, rest.RestError) {
  return getOrgHelper(getOrgByAuthIdQuery, authId, ctx, nil)
}

// GetOrgByAuthIdInTxn retrieves a Org by public authentication ID string
// in the context of an existing transaction. See GetOrgByAuthId.
func GetOrgByAuthIdInTxn(authId string, ctx context.Context, txn *sql.Tx) (*Org, rest.RestError) {
  return getOrgHelper(getOrgByAuthIdQuery, authId, ctx, txn)
}

const getOrgByIdStatement string = CommonOrgGet + ` WHERE o.id=? `
// GetOrgByID retrieves a Org by internal ID. As the internal ID must
// never be exposed to users, this method is exclusively for internal/backend
// use. Specifically, since Orgs are associated with other Entities through
// the internal ID (i.e., foreign keys use the internal ID), this function is
// most often used to retrieve a Org which is to be bundled in a response.
//
// Use GetOrg to retrieve a Org in response to an API request. TODO:
// reference discussion of internal vs public IDs.
func GetOrgByID(id int64, ctx context.Context) (*Org, rest.RestError) {
  return getOrgHelper(getOrgByIdQuery, id, ctx, nil)
}

// GetOrgByIDInTxn retrieves a Org by internal ID in the context of an
// existing transaction. See GetOrgByID.
func GetOrgByIDInTxn(id int64, ctx context.Context, txn *sql.Tx) (*Org, rest.RestError) {
  return getOrgHelper(getOrgByIdQuery, id, ctx, txn)
}

func getOrgHelper(stmt *sql.Stmt, id interface{}, ctx context.Context, txn *sql.Tx) (*Org, rest.RestError) {
  if txn != nil {
    stmt = txn.Stmt(stmt)
  }
	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return nil, rest.ServerError("Error retrieving org.", err)
	}
	defer rows.Close()

	var org *Org
  var address *locations.Address
  var addresses locations.Addresses = make(locations.Addresses, 0)
	for rows.Next() {
    var err error
    // The way the scanner works, it processes all the data each time. :(
    // 'org' gets updated with an equivalent structure while we gather up
    // the addresses.
    if org, address, err = ScanOrgDetail(rows); err != nil {
      return nil, rest.ServerError(fmt.Sprintf("Problem getting data for org: '%v'", id), err)
    }

    if address.LocationId.Valid {
	    addresses = append(addresses, address)
    }
	}
  if org != nil {
    org.Addresses = addresses
    org.FormatOut()
  } else {
    return nil, rest.NotFoundError(fmt.Sprintf(`Org '%s' not found.`, id), nil)
  }

	return org, nil
}

// BUG(zane@liquid-labs.com): UpdateOrg should use internal IDs if available
// on the Org struct. (I'm assuming this is slightly more efficient, though
// we should test.)

// UpdatesOrg updates the canonical Org record. Attempting to update a
// non-existent Org results in a rest.NotFoundError.
func UpdateOrg(o *Org, ctx context.Context) (*Org, rest.RestError) {
  txn, err := sqldb.DB.Begin()
  if err != nil {
    defer txn.Rollback()
    return nil, rest.ServerError("Could not update org record.", err)
  }

  newO, restErr := UpdateOrgInTxn(o, ctx, txn)
  // txn already rolled back if in error, so we only need to commit if no error
  if restErr == nil {
    defer txn.Commit()
  }

  return newO, restErr
}

// UpdatesOrgInTxn updates the canonical Org record within an existing
// transaction. See UpdateOrg.
func UpdateOrgInTxn(o *Org, ctx context.Context, txn *sql.Tx) (*Org, rest.RestError) {
  if o.Addresses != nil {
    o.Addresses.CompleteAddresses(ctx)
  }
  var err error
  var updateStmt *sql.Stmt = updateOrgQuery
  if o.Addresses != nil {
    if restErr := o.Addresses.Update(o.PubId.String, ctx, txn); restErr != nil {
      defer txn.Rollback()
      // TODO: this message could be misleading; like the org was updated, and just the addresses not
      return nil, restErr
    }
    updateStmt = txn.Stmt(updateOrgQuery)
  }

  _, err = updateStmt.Exec(o.Active, o.LegalID, o.LegalIDType, o.DisplayName, o.Summary, o.Phone, o.Email, o.Homepage, o.LogoURL, o.PubId)
  if err != nil {
    if txn != nil {
      defer txn.Rollback()
    }
    return nil, rest.ServerError("Could not update org record.", err)
  }

  newOrg, err := GetOrgInTxn(o.PubId.String, ctx, txn)
  if err != nil {
    return nil, rest.ServerError("Problem retrieving newly updated org.", err)
  }
  // Carry any 'ChangeDesc' made by the geocoding out.
  o.PromoteChanges()
  newOrg.ChangeDesc = o.ChangeDesc

  return newOrg, nil
}

// TODO: enable update of AuthID
const updateOrgStatement = `UPDATE orgs o JOIN users u ON u.id=o.id JOIN entities e ON o.id=e.id SET u.active=?, u.legal_id=?, u.legal_id_type=?, o.display_name=?, o.summary=?, o.phone=?, o.email=?, o.homepage=?, o.logo_url=?, e.last_updated=0 WHERE e.pub_id=?`
var createOrgQuery, updateOrgQuery, getOrgQuery, getOrgByAuthIdQuery, getOrgByIdQuery *sql.Stmt
func SetupDB(db *sql.DB) {
  var err error
  if createOrgQuery, err = db.Prepare(createOrgStatement); err != nil {
    log.Fatalf("mysql: prepare create org stmt:\n%v\n%s", err, createOrgStatement)
  }
  if getOrgQuery, err = db.Prepare(getOrgStatement); err != nil {
    log.Fatalf("mysql: prepare get org stmt: %v", err)
  }
  if getOrgByAuthIdQuery, err = db.Prepare(getOrgByAuthIdStatement); err != nil {
    log.Fatalf("mysql: prepare get org by auth ID stmt: %v", err)
  }
  if getOrgByIdQuery, err = db.Prepare(getOrgByIdStatement); err != nil {
    log.Fatalf("mysql: prepare get org by ID stmt: %v", err)
  }
  if updateOrgQuery, err = db.Prepare(updateOrgStatement); err != nil {
    log.Fatalf("mysql: prepare update org stmt: %v", err)
  }
}
