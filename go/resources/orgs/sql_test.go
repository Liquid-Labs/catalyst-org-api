package orgs_test

import (
  "context"
  "os"
  "testing"

  // the package we're testing
  . "github.com/Liquid-Labs/catalyst-orgs-api/go/resources/orgs"
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

const someOrgID=`E9EB036A-0194-4AD4-B598-2412FB9C8F5B`

func setupDB() {
  sqldb.RegisterSetup(entities.SetupDB, locations.SetupDB, users.SetupDB, /*orgs.*/SetupDB)
  sqldb.InitDB() // panics if unable to initialize
}

func testOrgDBSetup(t *testing.T) {
  setupDB()
}

func testOrgGet(t *testing.T) {
  org, err := GetOrg(someOrgID, context.Background())
  require.NoError(t, err, `Unexpected error getting Org.`)
  require.NotNil(t, org, `Unexpected nil Org on create (with no error); check ID.`)
  assert.Equal(t, `Some Org`, org.DisplayName.String, `Unexpected display name.`)
  assert.Equal(t, `janedoe@test.com`, org.Email.String, `Unexpected email.`)
  assert.Equal(t, `555-555-1111`, org.Phone.String, `Unexpected phone.`)
  assert.Equal(t, false, org.Active.Bool, `Unexpected active value.`)
  assert.NotEmpty(t, org.Id, `Unexpected empty ID.`)
  assert.Equal(t, someOrgID, org.PubId.String, `Unexpected public id.`)
}

func testOrgCreate(t *testing.T) {
  org, err := CreateOrg(someOrg, context.Background())
  require.NoError(t, err, `Unexpected error creating Org.`)
  require.NotNil(t, org, `Unexpected nil Org on create (with no error).`)
  assert.Equal(t, someOrg.DisplayName, org.DisplayName, `Unexpected display name.`)
  assert.Equal(t, someOrg.Email, org.Email, `Unexpected email.`)
  assert.Equal(t, someOrg.Phone, org.Phone, `Unexpected phone.`)
  assert.Equal(t, someOrg.Active, org.Active, `Unexpected active value.`)
  assert.NotEmpty(t, org.Id, `Unexpected empty ID.`)
  assert.NotEmpty(t, org.PubId, `Unexpected empty public id.`)
}

func testOrgUpdate(t *testing.T) {
  someOtherOrg, err := GetOrg(someOrgID, context.Background())
  require.NoError(t, err, `Unexpected error getting Org.`)
  require.NotNil(t, someOtherOrg, `Unexpected nil value retrieving org (with no error); check ID.`)
  someOtherOrg.SetActive(true)
  someOtherOrg.SetDisplayName(`Jane P. Doe`)
  someOtherOrg.SetEmail(`janepdoe@test.com`)
  someOtherOrg.SetPhone(`555-555-0001`)
  org, err := UpdateOrg(someOtherOrg, context.Background())
  require.NoError(t, err, `Unexpected error updating Org.`)
  require.NotNil(t, org, `Unexpected nil Org on create (with no error).`)
  assert.Equal(t, someOtherOrg.DisplayName, org.DisplayName, `Unexpected display name.`)
  assert.Equal(t, someOtherOrg.Email, org.Email, `Unexpected email.`)
  assert.Equal(t, someOtherOrg.Phone, org.Phone, `Unexpected phone.`)
  assert.Equal(t, someOtherOrg.Active, org.Active, `Unexpected active value.`)
  assert.NotEmpty(t, org.Id, `Unexpected empty ID.`)
  assert.NotEmpty(t, org.PubId, `Unexpected empty public id.`)
}

func testOrgGetInTxn(t *testing.T) {
  someOtherOrg, restErr := GetOrg(someOrgID, context.Background())
  assert.NoError(t, restErr, `Unexpected error getting org.`)
  txn, _ := sqldb.DB.Begin()
  orig := someOtherOrg.Clone()
  // if we get in a txn, we should see the changes
  someOtherOrg.SetPhone(`555-555-0003`)
  org, restErr := UpdateOrgInTxn(someOtherOrg, context.Background(), txn)
  someOtherTxn, restErr := GetOrgInTxn(someOrgID, context.Background(), txn)
  assert.Equal(t, *org, *someOtherTxn, `Update-Org and Get-Org do not match.`)
  assert.Equal(t, someOtherOrg.Phone, someOtherTxn.Phone, `Did not see change while getting in txn.`)
  assert.NotEqual(t, someOtherOrg.Phone, orig.Phone, `Phone number not changed.`)
  someOtherNoTxn, restErr := GetOrg(someOrgID, context.Background())
  assert.Equal(t, orig.Phone, someOtherNoTxn.Phone, `Non-txn org reflects changes.`)
  assert.NoError(t, txn.Commit(), `Error attempting commit.`)
  someOtherFinish, _ := GetOrg(someOrgID, context.Background())
  assert.Equal(t, *someOtherTxn, *someOtherFinish, `Post-commit Orgs didn't match.`)
}

func testOrgCreateInTxn(t *testing.T) {
  yetAnotherOrg := someOrg.Clone()
  yetAnotherOrg.SetDisplayName(`Jim Doe`)
  txn, _ := sqldb.DB.Begin()
  txnOrg, restErr := CreateOrgInTxn(yetAnotherOrg, context.Background(), txn)
  assert.NoError(t, restErr, `Unexpected error creating org in txn.`)
  noOrg, restErr := GetOrg(txnOrg.PubId.String, context.Background())
  assert.Nil(t, noOrg, `Unexpected retrieval of org outside of txn.`)
  assert.Error(t, restErr, `Unexpected non-error while retrieving org outside of txn.`)
  assert.NoError(t, txn.Commit(), `Error attempting commit.`)
  yetAnotherFinish, _ := GetOrg(txnOrg.PubId.String, context.Background())
  // We expect the created org to have a empty ('[]') ChangeDesc, but the
  // get-ed org's to be nil. So let's fix that before comparing.
  txnOrg.ChangeDesc = nil
  assert.Equal(t, *txnOrg, *yetAnotherFinish, `Post-commit Orgs didn't match.`)
}

func testOrgUpdateInTxn(t *testing.T) {
  /*txn, err := sqldb.DB.Begin()
  assert.NoError(t, err, `Unexpected error opening transaction.`)*/
}
