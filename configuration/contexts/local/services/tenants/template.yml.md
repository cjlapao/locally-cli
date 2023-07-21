# Tenants Schema template

The tenants schema tells locally what tenants exist and what is the expected tenant url, **this will not create the tenant**, you will still need to use the EMS api endpoints to do so. It is only used to create dns records in the hosts files and proxy configuration so that the locally proxy can work with the UI

```yaml
# All tenants definitions needs to start with this, the reason is locally is folder agnostic
# so while we places all the tenants definitions in the same folder it does necessarily
# needs to be so, adding that allows locally to set the configuration in the right place
tenants:
    # name of the tenant, this is just an identifier and can be anything
  - name: Local Tenant 1
    # uri will be used by locally to build configuration for the proxy so that the tenant can work as
    # expected this uri will not be the full url but just a sub domain, for example if the uri would
    # be example then the spa service would be listening on example.locally.team
    uri: local-t1
```
