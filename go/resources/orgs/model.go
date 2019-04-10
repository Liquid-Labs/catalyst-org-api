package orgs

import (
  "regexp"

  "github.com/Liquid-Labs/catalyst-core-api/go/resources/users"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources/locations"
  "github.com/Liquid-Labs/catalyst-core-api/go/resources"
  "github.com/Liquid-Labs/go-nullable-mysql/nulls"
)

var phoneOutFormatter *regexp.Regexp = regexp.MustCompile(`^(\d{3})(\d{3})(\d{4})$`)

// On summary, we don't include address. Note leaving it empty and using
// 'omitempty' on the Org struct won't work because then Orgs without an address
// will appear 'incomplete' in the front-end model and never resolve.
type OrgSummary struct {
  users.User
  DisplayName   nulls.String `json:"displayName"`
  Summary       nulls.String `json:"summary"`
  Email         nulls.String `json:"email"`
  Phone         nulls.String `json:"phone,string"`
  Homepage       nulls.String `json:"webpage"`
  LogoURL       nulls.String `json:"logoURL"`
}

func (o *OrgSummary) FormatOut() {
  o.Phone.String = phoneOutFormatter.ReplaceAllString(o.Phone.String, `$1-$2-$3`)
}

func (o *OrgSummary) SetDisplayName(val string) {
  o.DisplayName = nulls.NewString(val)
}

func (o *OrgSummary) SetSummary(val string) {
  o.Summary = nulls.NewString(val)
}

func (o *OrgSummary) SetEmail(val string) {
  o.Email = nulls.NewString(val)
}

func (o *OrgSummary) SetPhone(val string) {
  o.Phone = nulls.NewString(val)
}

func (o *OrgSummary) SetHomepage(val string) {
  o.Homepage = nulls.NewString(val)
}

func (o *OrgSummary) SetLogoURL(val string) {
  o.LogoURL = nulls.NewString(val)
}

func (o *OrgSummary) Clone() *OrgSummary {
  return &OrgSummary{
    *o.User.Clone(),
    o.DisplayName,
    o.Summary,
    o.Email,
    o.Phone,
    o.Homepage,
    o.LogoURL,
  }
}

// We expect an empty address array if no addresses on detail
type Org struct {
  OrgSummary
  Addresses     locations.Addresses  `json:"addresses"`
  ChangeDesc    []string             `json:"changeDesc,omitempty"`
}

func (o *Org) Clone() *Org {
  newChangeDesc := make([]string, len(o.ChangeDesc))
  copy(newChangeDesc, o.ChangeDesc)

  return &Org{
    *o.OrgSummary.Clone(),
    *o.Addresses.Clone(),
    newChangeDesc,
  }
}

func (o *Org) PromoteChanges() {
  o.ChangeDesc = resources.PromoteChanges(o.Addresses, o.ChangeDesc)
}
