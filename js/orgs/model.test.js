/* globals beforeAll describe expect test */
import { resourcesSettings, verifyCatalystSetup } from '@liquid-labs/catalyst-core-api'
import { Org, orgResourceConf } from './model'

const orgFooModel = {
  pubId       : '630AC9ED-3531-41E3-BD87-E26ADA74ECBC',
  lastUpdated : null,
  active      : true,
  authId      : null,
  legalID     : null,
  legalIDType : null,
  displayName : 'foo',
  summary     : null,
  phone       : null,
  email       : null,
  homepage    : null,
  logoURL     : null,
  addresses   : undefined
}

const orgBarModel = {
  pubId       : '23DB5195-67FF-4709-9033-7F9F5C5A6C6F',
  lastUpdated : null,
  active      : true,
  authId      : null,
  legalID     : null,
  legalIDType : null,
  displayName : 'bar',
  summary     : null,
  phone       : null,
  email       : null,
  homepage    : null,
  logoURL     : null,
  addresses   : []
}

describe('Org', () => {
  beforeAll(() => {
    const resourceList = [ orgResourceConf ]
    resourcesSettings.setResources(resourceList)
    verifyCatalystSetup()
  })

  test("should identify self as a 'orgs' resource", () => {
    const org = new Org(orgFooModel)
    expect(org.resourceName).toBe('orgs')
  })

  test("should be incomplete if address is 'null'", () => {
    const org = new Org(orgFooModel)
    expect(org.isComplete()).toBe(false)
    expect(org.getMissing()).toHaveLength(1)
    expect(org.getMissing()[0]).toBe('addresses')
  })

  test("should provide ascending and descending display name sort options", () => {
    const orgFoo = new Org(orgFooModel)
    const orgBar = new Org(orgBarModel)

    const orgs = [ orgFoo, orgBar ]
    expect(typeof resourcesSettings.getResourcesMap()['orgs'].sortMap['displayName-asc'])
      .toBe('function')
    orgs.sort(resourcesSettings.getResourcesMap()['orgs'].sortMap['displayName-asc'])
    expect(orgs[0]).toBe(orgBar)
    expect(orgs[1]).toBe(orgFoo)

    expect(typeof resourcesSettings.getResourcesMap()['orgs'].sortMap['displayName-desc'])
      .toBe('function')
    orgs.sort(resourcesSettings.getResourcesMap()['orgs'].sortMap['displayName-desc'])
    expect(orgs[0]).toBe(orgFoo)
    expect(orgs[1]).toBe(orgBar)
    // and verify that we test all the options
    expect(resourcesSettings.getResourcesMap()['orgs'].sortOptions).toHaveLength(2)
  })

  test("should define default sort options", () => {
    expect(resourcesSettings.getResourcesMap()['orgs'].sortDefault).toBe('displayName-asc')
  })
})
