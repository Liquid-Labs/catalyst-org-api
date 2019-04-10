package orgs_test

import (
  "context"
  "os"
  "testing"

  // the package we're testing
  . "github.com/Liquid-Labs/catalyst-persons-api/go/persons"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/entities"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/locations"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/users"
  "github.com/Liquid-Labs/go-api/sqldb"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
)

func TestOrgsDBIntegration(t *testing.T) {
  if os.Getenv(`SKIP_INTEGRATION`) == `true` {
    t.Skip()
  }

  if someOrg == nil {
    t.Error(`Org struct not define; can't continue. This probbaly indicates a setup failure in 'model_test.go'.`)
  } else {
    if t.Run(`OrgsDBSetup`, testOrgDBSetup) {
      if sqldb.DB == nil { // test was skipped, but we still need to setup
        setupDB()
      }
      t.Run(`OrgGet`, testOrgGet)
      t.Run(`OrgCreate`, testOrgCreate)
      t.Run(`OrgUpdate`, testOrgUpdate)
      t.Run(`OrgGetInTxn`, testOrgGetInTxn)
      t.Run(`OrgCreateInTxn`, testOrgCreateInTxn)
      t.Run(`OrgUpdateInTxn`, testOrgUpdateInTxn)
    }
  }
}

const someOtherId=`4BE66BE5-2A62-11E9-B987-42010A8003FF`

func setupDB() {
  sqldb.RegisterSetup(entities.SetupDB, locations.SetupDB, users.SetupDB, /*persons.*/SetupDB)
  sqldb.InitDB() // panics if unable to initialize
}

func testOrgDBSetup(t *testing.T) {
  setupDB()
}

func testOrgGet(t *testing.T) {
  person, err := GetOrg(someOtherId, context.Background())
  require.NoError(t, err, `Unexpected error getting Org.`)
  require.NotNil(t, person, `Unexpected nil Org on create (with no error).`)
  assert.Equal(t, `Jane Doe`, person.DisplayName.String, `Unexpected display name.`)
  assert.Equal(t, `janedoe@test.com`, person.Email.String, `Unexpected email.`)
  assert.Equal(t, `555-555-1111`, person.Phone.String, `Unexpected phone.`)
  assert.Equal(t, false, person.Active.Bool, `Unexpected active value.`)
  assert.NotEmpty(t, person.Id, `Unexpected empty ID.`)
  assert.Equal(t, someOtherId, person.PubId.String, `Unexpected public id.`)
}

func testOrgCreate(t *testing.T) {
  person, err := CreateOrg(someOrg, context.Background())
  require.NoError(t, err, `Unexpected error creating Org.`)
  require.NotNil(t, person, `Unexpected nil Org on create (with no error).`)
  assert.Equal(t, someOrg.DisplayName, person.DisplayName, `Unexpected display name.`)
  assert.Equal(t, someOrg.Email, person.Email, `Unexpected email.`)
  assert.Equal(t, someOrg.Phone, person.Phone, `Unexpected phone.`)
  assert.Equal(t, someOrg.Active, person.Active, `Unexpected active value.`)
  assert.NotEmpty(t, person.Id, `Unexpected empty ID.`)
  assert.NotEmpty(t, person.PubId, `Unexpected empty public id.`)
}

func testOrgUpdate(t *testing.T) {
  someOtherOrg, err := GetOrg(someOtherId, context.Background())
  require.NoError(t, err, `Unexpected error getting Org.`)
  someOtherOrg.SetActive(true)
  someOtherOrg.SetDisplayName(`Jane P. Doe`)
  someOtherOrg.SetEmail(`janepdoe@test.com`)
  someOtherOrg.SetPhone(`555-555-0001`)
  someOtherOrg.SetPhoneBackup(`555-555-0002`)
  person, err := UpdateOrg(someOtherOrg, context.Background())
  require.NoError(t, err, `Unexpected error updating Org.`)
  require.NotNil(t, person, `Unexpected nil Org on create (with no error).`)
  assert.Equal(t, someOtherOrg.DisplayName, person.DisplayName, `Unexpected display name.`)
  assert.Equal(t, someOtherOrg.Email, person.Email, `Unexpected email.`)
  assert.Equal(t, someOtherOrg.Phone, person.Phone, `Unexpected phone.`)
  assert.Equal(t, someOtherOrg.Active, person.Active, `Unexpected active value.`)
  assert.NotEmpty(t, person.Id, `Unexpected empty ID.`)
  assert.NotEmpty(t, person.PubId, `Unexpected empty public id.`)
}

func testOrgGetInTxn(t *testing.T) {
  someOtherOrg, restErr := GetOrg(someOtherId, context.Background())
  assert.NoError(t, restErr, `Unexpected error getting person.`)
  txn, _ := sqldb.DB.Begin()
  orig := someOtherOrg.Clone()
  // if we get in a txn, we should see the changes
  someOtherOrg.SetPhone(`555-555-0003`)
  person, restErr := UpdateOrgInTxn(someOtherOrg, context.Background(), txn)
  someOtherTxn, restErr := GetOrgInTxn(someOtherId, context.Background(), txn)
  assert.Equal(t, *person, *someOtherTxn, `Update-Org and Get-Org do not match.`)
  assert.Equal(t, someOtherOrg.Phone, someOtherTxn.Phone, `Did not see change while getting in txn.`)
  assert.NotEqual(t, someOtherOrg.Phone, orig.Phone, `Phone number not changed.`)
  someOtherNoTxn, restErr := GetOrg(someOtherId, context.Background())
  assert.Equal(t, orig.Phone, someOtherNoTxn.Phone, `Non-txn person reflects changes.`)
  assert.NoError(t, txn.Commit(), `Error attempting commit.`)
  someOtherFinish, _ := GetOrg(someOtherId, context.Background())
  assert.Equal(t, *someOtherTxn, *someOtherFinish, `Post-commit Orgs didn't match.`)
}

func testOrgCreateInTxn(t *testing.T) {
  yetAnotherOrg := someOrg.Clone()
  yetAnotherOrg.SetDisplayName(`Jim Doe`)
  txn, _ := sqldb.DB.Begin()
  txnOrg, restErr := CreateOrgInTxn(yetAnotherOrg, context.Background(), txn)
  assert.NoError(t, restErr, `Unexpected error creating person in txn.`)
  noOrg, restErr := GetOrg(txnOrg.PubId.String, context.Background())
  assert.Nil(t, noOrg, `Unexpected retrieval of person outside of txn.`)
  assert.Error(t, restErr, `Unexpected non-error while retrieving person outside of txn.`)
  assert.NoError(t, txn.Commit(), `Error attempting commit.`)
  yetAnotherFinish, _ := GetOrg(txnOrg.PubId.String, context.Background())
  // We expect the created person to have a empty ('[]') ChangeDesc, but the
  // get-ed person's to be nil. So let's fix that before comparing.
  txnOrg.ChangeDesc = nil
  assert.Equal(t, *txnOrg, *yetAnotherFinish, `Post-commit Orgs didn't match.`)
}

func testOrgUpdateInTxn(t *testing.T) {
  /*txn, err := sqldb.DB.Begin()
  assert.NoError(t, err, `Unexpected error opening transaction.`)*/
}
