import {
  Address,
  arrayType,
  CommonResourceConf,
  Model,
  userPropsModel
} from '@liquid-labs/catalyst-core-api'

const orgPropsModel = [
  'displayName',
  'summary',
  'phone',
  'email',
  'homepage',
  'logoURL']
  .map((propName) => ({ propName : propName, writable : true }))
orgPropsModel.push(...userPropsModel)
orgPropsModel.push({
  propName  : 'addresses',
  model     : Address,
  valueType : arrayType,
  writable  : true})
orgPropsModel.push({
  propName            : 'changeDesc',
  unsetForNew         : true,
  writable            : true,
  optionalForComplete : true
})

const Org = class extends Model {
  get resourceName() { return 'orgs' }
}
Model.finalizeConstructor(Org, orgPropsModel)

const orgResourceConf = new CommonResourceConf('org', {
  model       : Org,
  sortOptions : [
    { label : 'Dispaly name (asc)',
      value : 'displayName-asc',
      func  : (a, b) => a.displayName.localeCompare(b.displayName) },
    { label : 'Display name (desc)',
      value : 'displayName-desc',
      func  : (a, b) => -a.displayName.localeCompare(b.displayName) }
  ],
  sortDefault : 'displayName-asc'
})

export { Org, orgPropsModel, orgResourceConf }
