package orgs_test

import (
  "encoding/json"
  "reflect"
  "strconv"
  "strings"
  "testing"

  . "github.com/Liquid-Labs/catalyst-orgs-api/go/resources/orgs"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/entities"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/locations"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/users"
  "github.com/Liquid-Labs/go-nullable-mysql/nulls"
  "github.com/stretchr/testify/assert"
)

var trivialOrgSummary = &OrgSummary{
  users.User{
    entities.Entity{
      nulls.NewInt64(1),
      nulls.NewString(`a`),
      nulls.NewInt64(2),
    },
    nulls.NewString(`xzc098`),
    nulls.NewString(`555-55-5555`),
    nulls.NewString(`SSN`),
    nulls.NewBool(false),
  },
  nulls.NewString(`displayName`),
  nulls.NewString(`A great company.`),
  nulls.NewString(`foo@test.com`),
  nulls.NewString(`555-555-9999`),
  nulls.NewString(`https://google.com`),
  nulls.NewString(`http://foo.com/logo`),
}

func TestOrgSummaryClone(t *testing.T) {
  clone := trivialOrgSummary.Clone()
  assert.Equal(t, trivialOrgSummary, clone, `Original does not match clone.`)
  clone.Id = nulls.NewInt64(3)
  clone.PubId = nulls.NewString(`b`)
  clone.LastUpdated = nulls.NewInt64(4)
  clone.SetActive(true)
  clone.SetDisplayName(`different name`)
  clone.SetSummary(`A new summary.`)
  clone.SetEmail(`blah@test.com`)
  clone.SetPhone(`555-555-9997`)
  clone.SetHomepage(`https://bar.com`)
  clone.SetLogoURL(`http://bar.com/image`)

  oReflection := reflect.ValueOf(trivialOrgSummary).Elem()
  cReflection := reflect.ValueOf(clone).Elem()
  for i := 0; i < oReflection.NumField(); i++ {
    assert.NotEqualf(
      t,
      oReflection.Field(i).Interface(),
      cReflection.Field(i).Interface(),
      `Fields '%s' unexpectedly match with: %s`,
      oReflection.Type().Field(i).Name, oReflection.Field(i).Interface(),
    )
  }
}

var trivialOrg = &Org{
  *trivialOrgSummary,
  locations.Addresses{
    &locations.Address{
      locations.Location{
        nulls.NewInt64(1),
        nulls.NewString(`a`),
        nulls.NewString(`b`),
        nulls.NewString(`c`),
        nulls.NewString(`d`),
        nulls.NewString(`e`),
        nulls.NewFloat64(2.0),
        nulls.NewFloat64(3.0),
        []string{`f`, `g`},
      },
      nulls.NewInt64(1),
      nulls.NewString(`label a`),
    },
  },
  []string{`h`, `i`},
}

func TestOrgClone(t *testing.T) {
  clone := trivialOrg.Clone()
  assert.Equal(t, trivialOrg, clone, `Original does not match clone.`)
  clone.Id = nulls.NewInt64(3)
  clone.PubId = nulls.NewString(`b`)
  clone.LastUpdated = nulls.NewInt64(4)
  clone.Active = nulls.NewBool(true)
  clone.DisplayName = nulls.NewString(`different name`)
  clone.Summary = nulls.NewString(`A new company.`)
  clone.Email = nulls.NewString(`blah@test.com`)
  clone.Phone = nulls.NewString(`555-555-9997`)
  clone.Homepage = nulls.NewString(`https://bar.com`)
  clone.LogoURL = nulls.NewString(`http://bar.com/image`)
  clone.Addresses = locations.Addresses{
    &locations.Address{
      locations.Location{
        nulls.NewInt64(2),
        nulls.NewString(`z`),
        nulls.NewString(`y`),
        nulls.NewString(`x`),
        nulls.NewString(`w`),
        nulls.NewString(`u`),
        nulls.NewFloat64(4.0),
        nulls.NewFloat64(5.0),
        []string{`i`},
      },
      nulls.NewInt64(2),
      nulls.NewString(`label b`),
    },
  }
  clone.ChangeDesc = []string{`j`}

  assert.NotEqual(t, trivialOrg.Addresses, clone.Addresses, `Addresses unexpectedly equal.`)
  aoReflection := reflect.ValueOf(trivialOrg.Addresses[0]).Elem()
  acReflection := reflect.ValueOf(clone.Addresses[0]).Elem()
  for i := 0; i < aoReflection.NumField(); i++ {
    assert.NotEqualf(
      t,
      aoReflection.Field(i).Interface(),
      acReflection.Field(i).Interface(),
      `Fields '%s' unexpectedly match.`,
      aoReflection.Type().Field(i),
    )
  }

  oReflection := reflect.ValueOf(trivialOrg).Elem()
  cReflection := reflect.ValueOf(clone).Elem()
  for i := 0; i < oReflection.NumField(); i++ {
    assert.NotEqualf(
      t,
      oReflection.Field(i).Interface(),
      cReflection.Field(i).Interface(),
      `Fields '%s' unexpectedly match.`,
      oReflection.Type().Field(i),
    )
  }
}

const jdDisplayName = "John Doe"
const jdEmail = "johndoe@test.com"
const jdPhone = "555-555-0000"
const jdActive = false

var johnDoeJson string = `
  {
    "displayName": "` + jdDisplayName + `",
    "email": "` + jdEmail + `",
    "phone": "` + jdPhone + `",
    "active": ` + strconv.FormatBool(jdActive) + `
  }`

var decoder *json.Decoder = json.NewDecoder(strings.NewReader(johnDoeJson))
var someOrg = &Org{}
var decodeErr = decoder.Decode(someOrg)

func TestOrgsDecode(t *testing.T) {
  assert.NoError(t, decodeErr, "Unexpected error decoding org JSON.")
  assert.Equal(t, jdDisplayName, someOrg.DisplayName.String, "Unexpected display name.")
  assert.Equal(t, jdEmail, someOrg.Email.String, "Unexpected email.")
  assert.Equal(t, jdPhone, someOrg.Phone.String, "Unexpected phone.")
  assert.Equal(t, jdActive, someOrg.Active.Bool, "Unexpected active value.")
}

func TestOrgFormatter(t *testing.T) {
  testP := &Org{OrgSummary: OrgSummary{
    Phone: nulls.NewString(`5555555555`),
  }}
  testP.FormatOut()
  assert.Equal(t, `555-555-5555`, testP.Phone.String)
}
